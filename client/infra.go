package screenscraper

import (
	"encoding/json"
	"fmt"
)

// InfraInfoResponse is the complete response for the infrastructure info endpoint
type InfraInfoResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers ServerInfo `json:"serveurs"`
	} `json:"response"`
}

// GetInfraInfo retrieves infrastructure/server information
func (c *Client) GetInfraInfo() (*InfraInfoResponse, error) {
	body, err := c.get("ssinfraInfos.php", nil)
	if err != nil {
		return nil, err
	}

	var resp InfraInfoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse infra info response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
