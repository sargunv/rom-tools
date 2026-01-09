package screenscraper

import (
	"encoding/json"
	"fmt"
)

// UserLevel represents a user level
type UserLevel struct {
	ID     int    `json:"id"`
	NameFR string `json:"nom_fr"`
}

// UserLevelsListResponse is the complete response for the user levels list endpoint
type UserLevelsListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers    ServerInfo           `json:"serveurs"`
		SSUser     *UserInfo            `json:"ssuser,omitempty"`
		UserLevels map[string]UserLevel `json:"userlevels"`
	} `json:"response"`
}

// GetUserLevelsList retrieves the list of user levels
func (c *Client) GetUserLevelsList() (*UserLevelsListResponse, error) {
	body, err := c.get("userlevelsListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp UserLevelsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse user levels list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
