package screenscraper

import (
	"encoding/json"
	"fmt"
)

// SupportTypesListResponse is the complete response for the support types list endpoint
type SupportTypesListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers      ServerInfo `json:"serveurs"`
		SSUser       *UserInfo  `json:"ssuser,omitempty"`
		SupportTypes []string   `json:"supporttypes"`
	} `json:"response"`
}

// GetSupportTypesList retrieves the list of support types
func (c *Client) GetSupportTypesList() (*SupportTypesListResponse, error) {
	body, err := c.get("supportTypesListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp SupportTypesListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse support types list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
