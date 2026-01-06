package screenscraper

import (
	"encoding/json"
	"fmt"
)

// UserInfoResponse is the complete response for the user info endpoint
type UserInfoResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers ServerInfo `json:"serveurs"`
		SSUser  UserInfo   `json:"ssuser"`
	} `json:"response"`
}

// GetUserInfo retrieves user information and quotas
func (c *Client) GetUserInfo() (*UserInfoResponse, error) {
	body, err := c.get("ssuserInfos.php", nil)
	if err != nil {
		return nil, err
	}

	var resp UserInfoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
