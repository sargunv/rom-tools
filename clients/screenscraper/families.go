package screenscraper

import (
	"encoding/json"
	"fmt"
)

// Family represents a game family
type Family struct {
	ID     int     `json:"id"`
	Name   string  `json:"nom"`
	Medias []Media `json:"medias,omitempty"`
}

// FamiliesListResponse is the complete response for the families list endpoint
type FamiliesListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers  ServerInfo        `json:"serveurs"`
		SSUser   *UserInfo         `json:"ssuser,omitempty"`
		Families map[string]Family `json:"familles"`
	} `json:"response"`
}

// GetFamiliesList retrieves the list of families
func (c *Client) GetFamiliesList() (*FamiliesListResponse, error) {
	body, err := c.get("famillesListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp FamiliesListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse families list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
