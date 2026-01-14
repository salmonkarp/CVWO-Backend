package handlers

import (
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
)

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
