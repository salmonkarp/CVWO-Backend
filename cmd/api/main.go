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

	mux.HandleFunc("/user/{id}", handlers.GetUser(db.Conn))
	mux.HandleFunc("/user/{id}/image", handlers.GetUserImage(db.Conn))

	mux.Handle("/topics", middleware.Auth(handlers.GetTopics(db.Conn)))
	mux.HandleFunc("/topics/{name}", handlers.GetTopic(db.Conn))
	mux.HandleFunc("/topics/{name}/posts", handlers.GetPostsByTopic(db.Conn))
	mux.HandleFunc("/topics/{name}/image", handlers.GetTopicImage(db.Conn))

	mux.Handle("/addtopic", middleware.Auth(handlers.AddTopic(db.Conn)))

	mux.HandleFunc("/posts/{id}", handlers.GetPost(db.Conn))
	mux.HandleFunc("/posts/{id}/comments", handlers.GetCommentsByPost(db.Conn))

	mux.Handle("/addpost", middleware.Auth(handlers.AddPost(db.Conn)))

	handler := middleware.CORS(mux)

	log.Println("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
