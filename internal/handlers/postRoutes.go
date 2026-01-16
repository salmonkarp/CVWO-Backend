package handlers

import (
	"backend/internal/auth"
	"backend/internal/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

func AddPost(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := auth.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid Token Claims", http.StatusUnauthorized)
			return
		}

		userID := int(claims["sub"].(float64))

		var t models.Post

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
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
		token, err := auth.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid Token Claims", http.StatusUnauthorized)
			return
		}

		userID := int(claims["sub"].(float64))

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		_, err = db.Exec(
			`UPDATE posts 
			SET title = $1, body = $2
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
		token, err := auth.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid Token Claims", http.StatusUnauthorized)
			return
		}

		userID := int(claims["sub"].(float64))

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
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
