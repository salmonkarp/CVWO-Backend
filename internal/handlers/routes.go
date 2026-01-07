package handlers

import (
	"backend/internal/models"
	"database/sql"
	"encoding/json"
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
