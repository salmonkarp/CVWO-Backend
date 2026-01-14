package handlers

import (
	"backend/internal/models"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
)

func GetTopics(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
            SELECT name, description, image IS NOT NULL
            FROM topics
        `)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var topics []models.Topic

		for rows.Next() {
			var (
				t        models.Topic
				hasImage bool
			)

			if err := rows.Scan(&t.Name, &t.Description, &hasImage); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if hasImage {
				url := "/topics/" + t.Name + "/image"
				t.ImageURL = &url
			}

			topics = append(topics, t)
		}

		json.NewEncoder(w).Encode(topics)
	}
}

func GetTopic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		var (
			t        models.Topic
			hasImage bool
		)

		err := db.QueryRow(
			`SELECT name, description, image IS NOT NULL
             FROM topics WHERE name = $1`,
			name,
		).Scan(&t.Name, &t.Description, &hasImage)

		if err != nil {
			http.Error(w, "topic not found", http.StatusNotFound)
			return
		}

		if hasImage {
			url := "/topics/" + t.Name + "/image"
			t.ImageURL = &url
		}

		json.NewEncoder(w).Encode(t)
	}
}

func GetTopicImage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topicName := r.PathValue("name")

		row := db.QueryRow(
			`SELECT image FROM topics WHERE name = $1`,
			topicName,
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

func AddTopic(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var t struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			ImageBase64 string `json:"image,omitempty"`
		}

		// log.Print(r.Body)

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		var imgBytes []byte
		if t.ImageBase64 != "" {
			var err error
			imgBytes, err = base64.StdEncoding.DecodeString(t.ImageBase64)
			if err != nil {
				http.Error(w, "invalid base64 image", http.StatusBadRequest)
				return
			}
			if len(imgBytes) > 2<<20 {
				http.Error(w, "image too large", http.StatusBadRequest)
				return
			}
		}

		_, err := db.Exec(
			`INSERT INTO topics (name, description, image) VALUES ($1, $2, $3)`,
			t.Name,
			t.Description,
			imgBytes,
		)
		if err != nil {
			if err.Error() == "pq: duplicate key value violates unique constraint \"topics_pkey\"" {
				http.Error(w, "Topic already exists.", http.StatusConflict)
				return
			} else {
				log.Println("Database error:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	})
}
