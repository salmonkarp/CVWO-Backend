package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"

	"backend/internal/auth"
	"backend/internal/db"
)

type LoginRequest struct {
	Username string `json:"username"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if len(req.Username) > 20 {
		http.Error(w, "Username is too long.", http.StatusBadRequest)
		return
	} else if !regexp.MustCompile(`^\d*[a-zA-Z][a-zA-Z0-9]*$`).MatchString(req.Username) {
		http.Error(w, "Username must be alphanumeric, and must have at least one alphabetical character.", http.StatusBadRequest)
		return
	}

	var userID int

	err := db.Conn.QueryRow(
		"SELECT id FROM users WHERE username = $1",
		req.Username,
	).Scan(&userID)

	if err == sql.ErrNoRows {
		_, insertErr := db.Conn.Exec(
			"INSERT INTO users (username) VALUES ($1)",
			req.Username,
		)

		if insertErr != nil {
			http.Error(w, "Failed to register. "+insertErr.Error(), http.StatusInternalServerError)
			return
		}

		db.Conn.QueryRow(
			"SELECT id FROM users WHERE username = $1",
			req.Username,
		).Scan(&userID)

	}

	token, tokenErr := auth.GenerateToken(userID)

	if tokenErr != nil {
		http.Error(w, "Token Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func Protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Authenticated"))
}
