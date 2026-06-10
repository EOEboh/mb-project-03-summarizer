package main

import (
	"log"
	"net/http"

	"github.com/EOEboh/mb-project-03-summarizer/handlers"
)

func main() {
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	// GET  /           : serve the summariser page
	// POST /summarize  : receive text + format, return HTML fragment (consumed by HTMX)
	mux.HandleFunc("/", handlers.Index)
	mux.HandleFunc("POST /summarize", handlers.Summarize)

	log.Println("🚀 Smart Summariser running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
