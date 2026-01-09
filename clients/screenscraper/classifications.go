package screenscraper

import (
	"encoding/json"
	"fmt"
)

// Classification represents a game classification/rating
type Classification struct {
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

// ClassificationsListResponse is the complete response for the classifications list endpoint
type ClassificationsListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers         ServerInfo                `json:"serveurs"`
		SSUser          *UserInfo                 `json:"ssuser,omitempty"`
		Classifications map[string]Classification `json:"classifications"`
	} `json:"response"`
}

// GetClassificationsList retrieves the list of classifications
func (c *Client) GetClassificationsList() (*ClassificationsListResponse, error) {
	body, err := c.get("classificationsListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp ClassificationsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse classifications list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
