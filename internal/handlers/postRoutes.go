package handlers

import (
	"backend/internal/auth"
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func GetPostsByTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topicName := r.PathValue("name")

		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		var userID int
		if uid, err := auth.VerifyToken(tokenStr); err == nil {
			userID = uid
		}

		rows, err := db.Query(`
			SELECT
				p.id,
				p.title,
				p.body,
				p.topic,
				p.creator,
				p.created_at,
				p.is_edited,
				COALESCE(SUM(
					CASE
						WHEN pv.is_positive THEN 1
						ELSE -1
					END
				), 0) AS score,
				MAX(
					CASE
						WHEN pv.user_id = $2 AND pv.is_positive THEN 1
						WHEN pv.user_id = $2 AND NOT pv.is_positive THEN -1
					END
				) AS user_vote
			FROM posts p
			LEFT JOIN post_votes pv ON pv.post_id = p.id
			WHERE p.topic = $1
			GROUP BY
				p.id, p.title, p.body, p.topic,
				p.creator, p.created_at, p.is_edited
			ORDER BY score DESC
		`,
			topicName,
			userID,
		)
		if err != nil {
			http.Error(w, "Query failed.", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		posts := []models.Post{}
		for rows.Next() {
			var p models.Post
			var userVote sql.NullInt64
			rows.Scan(&p.ID, &p.Title, &p.Body, &p.Topic, &p.Creator, &p.CreatedAt, &p.IsEdited, &p.Score, &userVote)
			if userVote.Valid {
				p.UserVote = int(userVote.Int64)
			}
			posts = append(posts, p)
		}

		json.NewEncoder(w).Encode(posts)
	}
}

func GetPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")

		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		var userID int
		if uid, err := auth.VerifyToken(tokenStr); err == nil {
			userID = uid
		}

		row := db.QueryRow(
			`SELECT 
				id, 
				title, 
				body, 
				topic, 
				creator, 
				created_at, 
				is_edited,
				COALESCE(
					(SELECT SUM(CASE WHEN is_positive THEN 1 ELSE -1 END)
					FROM post_votes
					WHERE post_id = p.id),
				0) AS score,
				(SELECT 
					CASE 
						WHEN is_positive IS TRUE THEN 1 
						WHEN is_positive IS FALSE THEN -1 
						ELSE NULL 
					END 
				FROM post_votes WHERE post_id = p.id AND user_id = $2 LIMIT 1) AS user_vote
			 FROM posts WHERE id = $1`,
			postID,
			userID,
		)

		var p models.Post
		var userVote sql.NullInt64
		if err := row.Scan(&p.ID, &p.Title, &p.Body, &p.Topic, &p.Creator, &p.CreatedAt, &p.IsEdited, &p.Score, &userVote); err != nil {
			http.Error(w, "Post not found.", http.StatusNotFound)
			return
		}
		if userVote.Valid {
			p.UserVote = int(userVote.Int64)
		}

		json.NewEncoder(w).Encode(p)
	}
}

func VotePost(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.VerifyToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token.", http.StatusUnauthorized)
			return
		}

		var payload struct {
			PostID     int   `json:"post_id"`
			IsPositive *bool `json:"is_positive"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		if payload.IsPositive == nil {
			_, err = db.Exec(`DELETE FROM post_votes WHERE post_id = $1 AND user_id = $2`, payload.PostID, userID)
			if err != nil {
				log.Println("Database error:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		_, err = db.Exec(
			`INSERT INTO post_votes (post_id, user_id, is_positive) VALUES ($1, $2, $3)
			 ON CONFLICT (post_id, user_id) DO UPDATE SET is_positive = EXCLUDED.is_positive`,
			payload.PostID,
			userID,
			*payload.IsPositive,
		)
		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func AddPost(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.VerifyToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token.", http.StatusUnauthorized)
			return
		}

		var t models.Post

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		if len(t.Title) > 100 {
			http.Error(w, "Post title too long.", http.StatusBadRequest)
			return
		} else if len(t.Body) > 3000 {
			http.Error(w, "Post description too long.", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			`INSERT INTO posts (title, body, topic, creator) VALUES ($1, $2, $3, $4)`,
			t.Title,
			t.Body,
			t.Topic,
			userID,
		)
		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func EditPost(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var t models.Post

		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.VerifyToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token.", http.StatusUnauthorized)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		if len(t.Title) > 100 {
			http.Error(w, "Post title too long.", http.StatusBadRequest)
			return
		} else if len(t.Body) > 3000 {
			http.Error(w, "Post description too long.", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			`UPDATE posts 
			SET title = $1, body = $2, is_edited = TRUE
			WHERE id = $3 AND creator = $4`,
			t.Title,
			t.Body,
			t.ID,
			userID,
		)

		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func DeletePost(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var t models.Post

		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.VerifyToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token.", http.StatusUnauthorized)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		_, err = db.Exec(
			`DELETE FROM posts 
			WHERE id = $1 AND creator = $2`,
			t.ID,
			userID,
		)

		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
