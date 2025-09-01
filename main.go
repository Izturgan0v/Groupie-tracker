package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func main() {
	server := &Server{}
	if err := server.loadData(); err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	// Define custom template functions
	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	// Load all templates, including the new 405.html
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		"templates/index.html",
		"templates/400.html",
		"templates/404.html",
		"templates/405.html", // Added 405 template
		"templates/500.html",
	)
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
	server.template = tmpl

	// Setup routes
	http.HandleFunc("/", server.homeHandler)
	http.HandleFunc("/artist/", server.artistHandler)
	http.HandleFunc("/filter", server.filterHandler)
	http.HandleFunc("/concerts/data", server.concertsHandler)

	// Route for serving static files (CSS, JS, images)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start the server
	fmt.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
