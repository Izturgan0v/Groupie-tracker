package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// fetchAPI fetches data from a given URL and decodes it into the target
func fetchAPI(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code for %s: %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response from %s: %v", url, err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		log.Printf("Raw response from %s: %s", url, string(body))
		return fmt.Errorf("failed to parse JSON from %s: %v", url, err)
	}
	return nil
}

// loadData fetches and parses API data
func (s *Server) loadData() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := fetchAPI("https://groupietrackers.herokuapp.com/api/artists", &s.data.Artists); err != nil {
		return err
	}

	var locIndex struct {
		Index []Location `json:"index"`
	}
	if err := fetchAPI("https://groupietrackers.herokuapp.com/api/locations", &locIndex); err != nil {
		return err
	}
	s.data.Locations = locIndex.Index

	var dateIndex struct {
		Index []Date `json:"index"`
	}
	if err := fetchAPI("https://groupietrackers.herokuapp.com/api/dates", &dateIndex); err != nil {
		return err
	}
	s.data.Dates = dateIndex.Index

	var relIndex struct {
		Index []Relation `json:"index"`
	}
	if err := fetchAPI("https://groupietrackers.herokuapp.com/api/relation", &relIndex); err != nil {
		return err
	}
	s.data.Relations = relIndex.Index

	return nil
}

// renderErrorPage renders a specific error page based on the status code
func (s *Server) renderErrorPage(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode) // Explicitly set the status code
	var templateName string
	switch statusCode {
	case http.StatusBadRequest:
		templateName = "400.html"
	case http.StatusNotFound:
		templateName = "404.html"
	case http.StatusMethodNotAllowed:
		templateName = "405.html" // Added 405 case
	case http.StatusInternalServerError:
		templateName = "500.html"
	default:
		// Fallback to a generic server error for unhandled codes
		templateName = "500.html"
	}

	if err := s.template.ExecuteTemplate(w, templateName, nil); err != nil {
		log.Printf("Error rendering %s: %v", templateName, err)
		// Fallback error message if template fails
		http.Error(w, fmt.Sprintf("Error rendering error page: %v", err), http.StatusInternalServerError)
	}
}

// homeHandler serves the main page
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure only GET requests are processed
	if r.Method != http.MethodGet {
		s.renderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}
	// If the path is not the root, it's a 404
	if r.URL.Path != "/" {
		s.renderErrorPage(w, http.StatusNotFound)
		return
	}
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data := struct {
		Artists        []Artist
		Artist         Artist
		IsSingleArtist bool
	}{
		Artists: s.data.Artists,
	}
	if err := s.template.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Template error: %v", err)
		s.renderErrorPage(w, http.StatusInternalServerError)
	}
}

// artistHandler serves a specific band's card or returns an error
func (s *Server) artistHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure only GET requests are processed
	if r.Method != http.MethodGet {
		s.renderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(idStr)
	// If the ID is not a valid integer or is invalid, treat as Not Found
	if err != nil || id < 1 {
		s.renderErrorPage(w, http.StatusNotFound)
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var artist Artist
	for _, a := range s.data.Artists {
		if a.ID == id {
			artist = a
			break
		}
	}
	// If no artist was found for the given ID, return Not Found
	if artist.ID == 0 {
		s.renderErrorPage(w, http.StatusNotFound)
		return
	}

	data := struct {
		Artists        []Artist
		Artist         Artist
		IsSingleArtist bool
	}{
		Artist:         artist,
		IsSingleArtist: true,
	}
	if err := s.template.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Template error: %v", err)
		s.renderErrorPage(w, http.StatusInternalServerError)
	}
}

// filterHandler handles band filtering via query parameters
func (s *Server) filterHandler(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	idStr := r.URL.Query().Get("id")

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var filtered []Artist
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			// Silently fail or return bad request for invalid filter ID
			s.renderErrorPage(w, http.StatusBadRequest)
			return
		}
		for _, artist := range s.data.Artists {
			if artist.ID == id {
				filtered = append(filtered, artist)
				break
			}
		}
	} else if yearStr == "" || yearStr == "0" {
		filtered = s.data.Artists
	} else {
		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 1900 || year > 2025 {
			s.renderErrorPage(w, http.StatusBadRequest)
			return
		}
		for _, artist := range s.data.Artists {
			if artist.CreationDate == year {
				filtered = append(filtered, artist)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filtered); err != nil {
		log.Printf("JSON encode error: %v", err)
		// Avoid rendering an HTML error page on an API endpoint
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// getConcerts returns sorted concert details for an artist
func (s *Server) getConcerts(artistID int) []map[string]string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var concerts []map[string]string
	for _, rel := range s.data.Relations {
		if rel.ID == artistID {
			for location, dates := range rel.DatesLocations {
				cleanLocation := strings.TrimSpace(location)
				for _, date := range dates {
					concerts = append(concerts, map[string]string{
						"date":     strings.TrimSpace(date),
						"location": cleanLocation,
					})
				}
			}
			break
		}
	}

	// Sort concerts by date
	sort.Slice(concerts, func(i, j int) bool {
		// A more robust date comparison might be needed if formats vary
		return concerts[i]["date"] < concerts[j]["date"]
	})

	return concerts
}

// concertsHandler serves concert details as JSON
func (s *Server) concertsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid artist ID", http.StatusBadRequest)
		return
	}

	concerts := s.getConcerts(id)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(concerts); err != nil {
		log.Printf("JSON encode error: %v", err)
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}
