package main

import (
	"log"
	"net/http"

	"backend/internal/db"
	"backend/internal/handlers"
	"backend/internal/middleware"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	if err := db.Connect(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login", handlers.Login)
	mux.Handle("/protected", middleware.Auth(http.HandlerFunc(handlers.Protected)))
	mux.Handle("/topics", middleware.Auth(handlers.GetTopics(db.Conn)))

	log.Println("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
