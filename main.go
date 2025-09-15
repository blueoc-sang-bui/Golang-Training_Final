package main

import (
	"blogapi/internal/cache"
	"blogapi/internal/search"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"blogapi/internal/handlers"
)

func main() {
	cache.InitRedis()
	search.InitES()
	r := mux.NewRouter()
	// Register handlers
	importHandlers(r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))

}

func importHandlers(r *mux.Router) {
	// Import các handler từ package handlers
	// Đăng ký các endpoint chính
	// Bạn cần import "handlers" ở đầu file
	// Đăng ký các endpoint cụ thể trước endpoint động
	r.HandleFunc("/posts/search", handlers.SearchPosts).Methods("GET")
	r.HandleFunc("/posts/tag/{tag}", handlers.SearchByTag).Methods("GET")
	r.HandleFunc("/posts", handlers.CreatePost).Methods("POST")
	r.HandleFunc("/posts/{id}", handlers.GetPost).Methods("GET")
	r.HandleFunc("/posts/{id}", handlers.UpdatePost).Methods("PUT")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Blog API is running!"))
	})
}
