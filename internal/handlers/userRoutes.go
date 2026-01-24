package handlers

import (
	"backend/internal/auth"
	"backend/internal/models"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"unicode"
)

func GetUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var (
			t          models.User
			hasImage   bool
			imageEpoch float64
			err        error
		)

		if unicode.IsDigit(rune(id[0])) {
			err = db.QueryRow(
				`SELECT id, username, image IS NOT NULL, EXTRACT(EPOCH FROM image_updated_at)
	             FROM users WHERE id = $1`,
				id,
			).Scan(&t.ID, &t.Username, &hasImage, &imageEpoch)
		} else {
			err = db.QueryRow(
				`SELECT id, username, image IS NOT NULL, EXTRACT(EPOCH FROM image_updated_at)
				FROM users WHERE username = $1`,
				id,
			).Scan(&t.ID, &t.Username, &hasImage, &imageEpoch)
		}

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if hasImage {
			url := "/user/" + t.Username + "/image"
			t.ImageURL = &url
			t.ImageUpdatedAt = int64(imageEpoch)
		}

		json.NewEncoder(w).Encode(t)
	}
}

func GetUserImage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var (
			image []byte
			err   error
		)

		if unicode.IsDigit(rune(id[0])) {
			err = db.QueryRow(
				`SELECT image FROM users WHERE id = $1`,
				id,
			).Scan(&image)
		} else {
			err = db.QueryRow(
				`SELECT image FROM users WHERE username = $1`,
				id,
			).Scan(&image)
		}

		if err != nil {
			http.Error(w, "Image not found.", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Write(image)
	}
}

func EditUser(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var t struct {
			ImageBase64 string `json:"image,omitempty"`
		}

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

		var imgBytes any = nil

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

		_, err = db.Exec(
			`UPDATE users 
			SET image = COALESCE($1, image), 
				image_updated_at = CASE WHEN $1 IS NOT NULL THEN now() 
										ELSE image_updated_at END
			WHERE id = $2`,
			imgBytes,
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
