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

func GetCommentsByPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")

		rows, err := db.Query(
			`SELECT id, body, post, creator, created_at, is_edited
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
			rows.Scan(&c.ID, &c.Body, &c.Post, &c.Creator, &c.CreatedAt, &c.IsEdited)
			comments = append(comments, c)
		}

		json.NewEncoder(w).Encode(comments)
	}
}

func AddComment(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := auth.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token.", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid Token Claims.", http.StatusUnauthorized)
			return
		}

		userID := int(claims["sub"].(float64))

		var t models.Comment

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		if len(t.Body) > 500 {
			http.Error(w, "Comment body too long.", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			`INSERT INTO comments (post, creator, body) VALUES ($1, $2, $3)`,
			t.Post,
			userID,
			t.Body,
		)
		if err != nil {
			log.Println("Database error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func EditComment(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := auth.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token.", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid Token Claims.", http.StatusUnauthorized)
			return
		}

		userID := int(claims["sub"].(float64))

		var t models.Comment

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		if len(t.Body) > 500 {
			http.Error(w, "Comment body too long.", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			`UPDATE comments SET body = $1, is_edited = TRUE WHERE id = $2 AND creator = $3`,
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

func DeleteComment(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := auth.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid Token.", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid Token Claims.", http.StatusUnauthorized)
			return
		}

		userID := int(claims["sub"].(float64))

		var t models.Comment

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			log.Println("Error decoding JSON:", err)
			return
		}

		_, err = db.Exec(
			`DELETE FROM comments WHERE id = $1 AND creator = $2`,
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
