package screenscraper

import (
	"encoding/json"
	"fmt"
)

// ROMTypesListResponse is the complete response for the ROM types list endpoint
type ROMTypesListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers  ServerInfo `json:"serveurs"`
		SSUser   *UserInfo  `json:"ssuser,omitempty"`
		ROMTypes []string   `json:"romtypes"`
	} `json:"response"`
}

// GetROMTypesList retrieves the list of ROM types
func (c *Client) GetROMTypesList() (*ROMTypesListResponse, error) {
	body, err := c.get("romTypesListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp ROMTypesListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse ROM types list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
