package handlers

import (
	"backend/internal/models"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
)

func GetTopics(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
            SELECT name, description, image IS NOT NULL, EXTRACT(EPOCH FROM image_updated_at)
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
				t          models.Topic
				hasImage   bool
				imageEpoch float64
			)

			if err := rows.Scan(&t.Name, &t.Description, &hasImage, &imageEpoch); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err.Error())
				return
			}

			if hasImage {
				url := "/topics/" + t.Name + "/image"
				t.ImageURL = &url
				t.ImageUpdatedAt = int64(imageEpoch)
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
			t          models.Topic
			hasImage   bool
			imageEpoch float64
		)

		err := db.QueryRow(
			`SELECT name, description, image IS NOT NULL, EXTRACT(EPOCH FROM image_updated_at)
             FROM topics WHERE name = $1`,
			name,
		).Scan(&t.Name, &t.Description, &hasImage, &imageEpoch)

		if err != nil {
			http.Error(w, "Topic not found.", http.StatusNotFound)
			log.Println(err.Error())
			return
		}

		if hasImage {
			url := "/topics/" + t.Name + "/image"
			t.ImageURL = &url
			t.ImageUpdatedAt = int64(imageEpoch)
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
			http.Error(w, "Image not found.", http.StatusNotFound)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
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

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		if len(t.Name) > 50 {
			http.Error(w, "Topic name too long.", http.StatusBadRequest)
			return
		} else if !regexp.MustCompile("^[a-zA-Z0-9]*$").MatchString(t.Name) {
			http.Error(w, "Topic name must contain only alphanumeric characters.", http.StatusBadRequest)
			return
		} else if len(t.Description) > 1000 {
			http.Error(w, "Topic description too long.", http.StatusBadRequest)
			return
		}

		var imgBytes interface{} = nil

		if t.ImageBase64 != "" {
			decoded, err := base64.StdEncoding.DecodeString(t.ImageBase64)
			if err != nil {
				http.Error(w, "Invalid base64 image.", http.StatusBadRequest)
				return
			}
			if len(decoded) > 2<<20 {
				http.Error(w, "Image too large.", http.StatusBadRequest)
				return
			}
			imgBytes = decoded
		}

		_, err := db.Exec(
			`INSERT INTO topics (name, description, image) VALUES ($1, $2, $3)`,
			t.Name,
			t.Description,
			imgBytes,
		)
		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func EditTopic(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var t struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			ImageBase64 string `json:"image,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		if len(t.Name) > 50 {
			http.Error(w, "Topic name too long.", http.StatusBadRequest)
			return
		} else if !regexp.MustCompile("^[a-zA-Z0-9]*$").MatchString(t.Name) {
			http.Error(w, "Topic name must contain only alphanumeric characters.", http.StatusBadRequest)
			return
		} else if len(t.Description) > 1000 {
			http.Error(w, "Topic description too long.", http.StatusBadRequest)
			return
		}

		var imgBytes interface{} = nil

		if t.ImageBase64 != "" {
			decoded, err := base64.StdEncoding.DecodeString(t.ImageBase64)
			if err != nil {
				http.Error(w, "Invalid base64 image.", http.StatusBadRequest)
				return
			}
			if len(decoded) > 2<<20 {
				http.Error(w, "Image too large.", http.StatusBadRequest)
				return
			}
			imgBytes = decoded
		}

		_, err := db.Exec(
			`UPDATE topics 
			SET description = $2, image = COALESCE($3, image), 
				image_updated_at = CASE WHEN $3 IS NOT NULL THEN now() 
										ELSE image_updated_at END
			WHERE name = $1`,
			t.Name,
			t.Description,
			imgBytes,
		)

		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})
}

func DeleteTopic(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var t models.Topic

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		_, err := db.Exec(
			`DELETE FROM TOPICS
			WHERE name = $1`,
			t.Name,
		)

		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})
}
