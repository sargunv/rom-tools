package screenscraper

import (
	"encoding/json"
	"fmt"
)

// MediaType represents a media type definition
type MediaType struct {
	ID            int    `json:"id"`
	ShortName     string `json:"nomcourt"`
	Name          string `json:"nom"`
	Category      string `json:"categorie"`
	PlatformTypes string `json:"plateformtypes"`
	Platforms     string `json:"plateforms"`
	Type          string `json:"type"`
	FileFormat    string `json:"fileformat"`
	FileFormat2   string `json:"fileformat2"`
	AutoGen       string `json:"autogen"`
	MultiRegions  string `json:"multiregions"`
	MultiSupports string `json:"multisupports"`
	MultiVersions string `json:"multiversions"`
	ExtraInfosTxt string `json:"extrainfostxt"`
}

// GameMediaListResponse is the complete response for the game media types list endpoint
type GameMediaListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers ServerInfo           `json:"serveurs"`
		SSUser  *UserInfo            `json:"ssuser,omitempty"`
		Medias  map[string]MediaType `json:"medias"`
	} `json:"response"`
}

// SystemMediaListResponse is the complete response for the system media types list endpoint
type SystemMediaListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers ServerInfo           `json:"serveurs"`
		SSUser  *UserInfo            `json:"ssuser,omitempty"`
		Medias  map[string]MediaType `json:"medias"`
	} `json:"response"`
}

// GetGameMediaList retrieves the list of available media types for games
func (c *Client) GetGameMediaList() (*GameMediaListResponse, error) {
	body, err := c.get("mediasJeuListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp GameMediaListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse game media list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetSystemMediaList retrieves the list of available media types for systems
func (c *Client) GetSystemMediaList() (*SystemMediaListResponse, error) {
	body, err := c.get("mediasSystemeListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp SystemMediaListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse system media list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
