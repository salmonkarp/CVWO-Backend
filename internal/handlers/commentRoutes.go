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

func GetCommentsByPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")

		rows, err := db.Query(
			`SELECT id, body, post, creator, created_at, is_edited, parent
			 FROM comments
			 WHERE post = $1
			 ORDER BY created_at ASC`,
			postID,
		)
		if err != nil {
			http.Error(w, "Query failed.", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		comments := []models.Comment{}
		for rows.Next() {
			var c models.Comment
			rows.Scan(&c.ID, &c.Body, &c.Post, &c.Creator, &c.CreatedAt, &c.IsEdited, &c.Parent)
			comments = append(comments, c)
		}

		json.NewEncoder(w).Encode(comments)
	}
}

func AddComment(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.VerifyToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token.", http.StatusUnauthorized)
			return
		}

		var c models.Comment

		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		if len(c.Body) > 500 {
			http.Error(w, "Comment body too long.", http.StatusBadRequest)
			return
		}

		if c.Parent != nil {
			var parentPostID int
			err = db.QueryRow(
				`SELECT post FROM comments WHERE id = $1`,
				*c.Parent,
			).Scan(&parentPostID)
			if err != nil {
				http.Error(w, "Parent comment not found.", http.StatusBadRequest)
				return
			}
			if parentPostID != c.Post {
				http.Error(w, "Parent comment does not belong to the same post.", http.StatusBadRequest)
				return
			}
		}
		_, err = db.Exec(
			`INSERT INTO comments (post, creator, body, parent) VALUES ($1, $2, $3, $4)`,
			c.Post,
			userID,
			c.Body,
			c.Parent,
		)
		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})
}

func EditComment(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.VerifyToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token.", http.StatusUnauthorized)
			return
		}

		var c models.Comment

		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		if len(c.Body) > 500 {
			http.Error(w, "Comment body too long.", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			`UPDATE comments SET body = $1, is_edited = TRUE WHERE id = $2 AND creator = $3`,
			c.Body,
			c.ID,
			userID,
		)
		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})
}

func DeleteComment(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.VerifyToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token.", http.StatusUnauthorized)
			return
		}

		var c models.Comment

		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		_, err = db.Exec(
			`DELETE FROM comments WHERE id = $1 AND creator = $2`,
			c.ID,
			userID,
		)
		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})
}
