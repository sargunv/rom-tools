package screenscraper

import (
	"encoding/json"
	"fmt"
)

// Genre represents a game genre
type Genre struct {
	ID        int     `json:"id"`
	ShortName string  `json:"nomcourt"`
	NameDE    string  `json:"nom_de"`
	NameEN    string  `json:"nom_en"`
	NameES    string  `json:"nom_es"`
	NameFR    string  `json:"nom_fr"`
	NameIT    string  `json:"nom_it"`
	NamePT    string  `json:"nom_pt"`
	ParentID  string  `json:"parent"`
	Medias    []Media `json:"medias,omitempty"`
}

// GenresListResponse is the complete response for the genres list endpoint
type GenresListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers ServerInfo       `json:"serveurs"`
		SSUser  *UserInfo        `json:"ssuser,omitempty"`
		Genres  map[string]Genre `json:"genres"`
	} `json:"response"`
}

// GetGenresList retrieves the list of genres
func (c *Client) GetGenresList() (*GenresListResponse, error) {
	body, err := c.get("genresListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp GenresListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse genres list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
