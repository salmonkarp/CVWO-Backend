package handlers

import (
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
)

func GetUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var (
			t        models.User
			hasImage bool
		)

		err := db.QueryRow(
			`SELECT id, username, image IS NOT NULL
             FROM users WHERE id = $1`,
			id,
		).Scan(&t.ID, &t.Username, &hasImage)

		if err != nil {
			err2 := db.QueryRow(
				`SELECT id, username, image IS NOT NULL
				FROM users WHERE username = $1`,
				id,
			).Scan(&t.ID, &t.Username, &hasImage)
			if err2 != nil {
				http.Error(w, err2.Error(), http.StatusInternalServerError)
				return
			}
		}

		if hasImage {
			url := "/users/" + t.Username + "/image"
			t.ImageURL = &url
		}

		json.NewEncoder(w).Encode(t)
	}
}

func GetUserImage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		row := db.QueryRow(
			`SELECT image FROM users WHERE id = $1`,
			id,
		)

		var image []byte
		if err := row.Scan(&image); err != nil {
			http.Error(w, "image not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Write(image)
	}
}
