package main

import (
	"html/template"
	"sync"
)

// Artist represents band information
type Artist struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Members      []string `json:"members"`
}

// Location represents concert locations
type Location struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
}

// Date represents concert dates
type Date struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

// Relation links artists with dates and locations
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// APIResponse holds all API data
type APIResponse struct {
	Artists   []Artist
	Locations []Location
	Dates     []Date
	Relations []Relation
}

// Server holds the API data and template
type Server struct {
	data     APIResponse
	mutex    sync.RWMutex
	template *template.Template
}
