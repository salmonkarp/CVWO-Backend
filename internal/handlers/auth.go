package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

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

	var userID int

	err := db.Conn.QueryRow(
		"SELECT id FROM users WHERE username = $1",
		req.Username,
	).Scan(&userID)

	if err == sql.ErrNoRows {
		http.Error(w, "invalid username", http.StatusUnauthorized)
		return
	}

	token, _ := auth.GenerateToken(userID)

	if err != nil {
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
