package screenscraper

import (
	"encoding/json"
	"fmt"
)

// PlayerCount represents a player count
type PlayerCount struct {
	ID       int    `json:"id"`
	Name     string `json:"nom"`
	ParentID string `json:"parent"`
}

// PlayerCountsListResponse is the complete response for the player counts list endpoint
type PlayerCountsListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers      ServerInfo             `json:"serveurs"`
		SSUser       *UserInfo              `json:"ssuser,omitempty"`
		PlayerCounts map[string]PlayerCount `json:"nbjoueurs"`
	} `json:"response"`
}

// GetPlayerCountsList retrieves the list of player counts
func (c *Client) GetPlayerCountsList() (*PlayerCountsListResponse, error) {
	body, err := c.get("nbJoueursListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp PlayerCountsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse player counts list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
