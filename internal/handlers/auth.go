package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/auth"
)

type LoginRequest struct {
	UserID int `json:"user_id"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := auth.GenerateToken(req.UserID)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func Protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("you are authenticated"))
}
