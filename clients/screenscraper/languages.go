package screenscraper

import (
	"encoding/json"
	"fmt"
)

// Language represents a language
type Language struct {
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

// LanguagesListResponse is the complete response for the languages list endpoint
type LanguagesListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers   ServerInfo          `json:"serveurs"`
		SSUser    *UserInfo           `json:"ssuser,omitempty"`
		Languages map[string]Language `json:"langues"`
	} `json:"response"`
}

// GetLanguagesList retrieves the list of languages
func (c *Client) GetLanguagesList() (*LanguagesListResponse, error) {
	body, err := c.get("languesListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp LanguagesListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse languages list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
