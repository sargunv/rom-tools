package screenscraper

import (
	"encoding/json"
	"fmt"
)

// System represents a game system/console
type System struct {
	ID          int               `json:"id"`
	ParentID    int               `json:"parentid"`
	Company     string            `json:"compagnie"`
	Type        string            `json:"type"`
	StartDate   string            `json:"datedebut"`
	EndDate     string            `json:"datefin"`
	Extensions  string            `json:"extensions"`
	ROMType     string            `json:"romtype"`
	SupportType string            `json:"supporttype"`
	Names       map[string]string `json:"noms"`
	Medias      []Media           `json:"medias"`
}

// SystemsListResponse is the complete response for the systems list endpoint
type SystemsListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers ServerInfo `json:"serveurs"`
		SSUser  *UserInfo  `json:"ssuser,omitempty"`
		Systems []System   `json:"systemes"`
	} `json:"response"`
}

// GetSystemsList retrieves the list of systems/consoles
func (c *Client) GetSystemsList() (*SystemsListResponse, error) {
	body, err := c.get("systemesListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp SystemsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse systems list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
