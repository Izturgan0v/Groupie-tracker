package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

type Artist struct {
	ID           int
	Image        string
	Name         string
	Members      []string
	CreationDate int
	FirstAlbum   string
}

func fetchArtists() ([]Artist, error) {
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var artists []Artist
	err = json.Unmarshal(body, &artists)
	return artists, nil
}
