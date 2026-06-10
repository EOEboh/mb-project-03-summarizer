package handlers

import (
	"html/template"
	"log"
	"net/http"
)

// tmpl parses the template once at startup.
// All handlers in this package share this template instance.
var tmpl = template.Must(template.ParseFiles("templates/index.html"))

// Index serves the main page of the application.
// Every project will replace or extend this with their project-specific data.
func Index(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
