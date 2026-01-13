package handlers

import (
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

func GetTopics(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT name, description FROM topics")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var topics []models.Topic
		for rows.Next() {
			var t models.Topic
			if err := rows.Scan(&t.Name, &t.Description); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			topics = append(topics, t)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topics)
	}
}

func GetTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topicName := r.PathValue("name")

		row := db.QueryRow(
			`SELECT name, description FROM topics WHERE name = $1`,
			topicName,
		)

		var name, description string
		if err := row.Scan(&name, &description); err != nil {
			http.Error(w, "topic not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"name":        name,
			"description": description,
		})
	}
}

func AddTopic(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var t models.Topic
		log.Print(r.Body)

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		_, err := db.Exec(
			`INSERT INTO topics (name, description) VALUES ($1, $2)`,
			t.Name,
			t.Description,
		)
		if err != nil {
			http.Error(w, "failed to add topic", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func GetPostsByTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topicName := r.PathValue("name")

		rows, err := db.Query(
			`SELECT id, title, body, topic, creator, created_at
			 FROM posts
			 WHERE topic = $1
			 ORDER BY created_at DESC`,
			topicName,
		)
		if err != nil {
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		posts := []models.Post{}
		for rows.Next() {
			var p models.Post
			rows.Scan(&p.ID, &p.Title, &p.Body, &p.Topic, &p.Creator, &p.CreatedAt)
			posts = append(posts, p)
		}

		json.NewEncoder(w).Encode(posts)
	}
}

func GetPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")

		row := db.QueryRow(
			`SELECT id, title, body, topic, creator, created_at
			 FROM posts WHERE id = $1`,
			postID,
		)

		var p models.Post
		if err := row.Scan(&p.ID, &p.Title, &p.Body, &p.Topic, &p.Creator, &p.CreatedAt); err != nil {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(p)
	}
}

func GetCommentsByPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")

		rows, err := db.Query(
			`SELECT id, body, post, creator, created_at
			 FROM comments
			 WHERE post = $1
			 ORDER BY created_at ASC`,
			postID,
		)
		if err != nil {
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		comments := []models.Comment{}
		for rows.Next() {
			var c models.Comment
			rows.Scan(&c.ID, &c.Body, &c.Post, &c.Creator, &c.CreatedAt)
			comments = append(comments, c)
		}

		json.NewEncoder(w).Encode(comments)
	}
}
