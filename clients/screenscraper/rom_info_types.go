package screenscraper

import (
	"encoding/json"
	"fmt"
)

// ROMInfoType represents a ROM info type definition
type ROMInfoType struct {
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

// ROMInfoTypesListResponse is the complete response for the ROM info types list endpoint
type ROMInfoTypesListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers   ServerInfo             `json:"serveurs"`
		SSUser    *UserInfo              `json:"ssuser,omitempty"`
		InfoTypes map[string]ROMInfoType `json:"infos"`
	} `json:"response"`
}

// GetROMInfoTypesList retrieves the list of available info types for ROMs
func (c *Client) GetROMInfoTypesList() (*ROMInfoTypesListResponse, error) {
	body, err := c.get("infosRomListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp ROMInfoTypesListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse ROM info types list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
