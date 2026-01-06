package screenscraper

import (
	"encoding/json"
	"fmt"
)

// SearchGameParams parameters for game search
type SearchGameParams struct {
	Query    string
	SystemID string
}

// GameSearchResponse is the complete response for the game search endpoint
type GameSearchResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers ServerInfo `json:"serveurs"`
		SSUser  *UserInfo  `json:"ssuser,omitempty"`
		Games   []Game     `json:"jeux"`
	} `json:"response"`
}

// SearchGame searches for games by name
func (c *Client) SearchGame(params SearchGameParams) (*GameSearchResponse, error) {
	p := map[string]string{
		"recherche": params.Query,
		"systemeid": params.SystemID,
	}
	body, err := c.get("jeuRecherche.php", p)
	if err != nil {
		return nil, err
	}

	var resp GameSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse game search response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
