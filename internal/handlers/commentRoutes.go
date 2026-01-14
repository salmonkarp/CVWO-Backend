package handlers

import (
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
)

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
