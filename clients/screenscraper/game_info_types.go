package screenscraper

import (
	"encoding/json"
	"fmt"
)

// GameInfoType represents a game info type definition
type GameInfoType struct {
	ID            int    `json:"id"`
	ShortName     string `json:"nomcourt"`
	Name          string `json:"nom"`
	Category      string `json:"categorie"`
	PlatformTypes string `json:"plateformtypes"`
	Platforms     string `json:"plateforms"`
	Type          string `json:"type"`
	AutoGen       string `json:"autogen"`
	MultiRegions  string `json:"multiregions"`
	MultiSupports string `json:"multisupports"`
	MultiVersions string `json:"multiversions"`
	MultiChoix    string `json:"multichoix"`
}

// GameInfoTypesListResponse is the complete response for the game info types list endpoint
type GameInfoTypesListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers   ServerInfo              `json:"serveurs"`
		SSUser    *UserInfo               `json:"ssuser,omitempty"`
		InfoTypes map[string]GameInfoType `json:"infos"`
	} `json:"response"`
}

// GetGameInfoTypesList retrieves the list of available info types for games
func (c *Client) GetGameInfoTypesList() (*GameInfoTypesListResponse, error) {
	body, err := c.get("infosJeuListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp GameInfoTypesListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse game info types list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
