package screenscraper

import (
	"encoding/json"
	"fmt"
)

// RegionMedias contains media URLs for regions
type RegionMedias struct {
	MediaPictoMonochrome    string `json:"media_pictomonochrome,omitempty"`
	MediaPictoMonochromeSVG string `json:"media_pictomonochromesvg,omitempty"`
	MediaPictoCouleur       string `json:"media_pictocouleur,omitempty"`
	MediaPictoCouleurSVG    string `json:"media_pictocouleursvg,omitempty"`
	MediaBackground         string `json:"media_background,omitempty"`
}

// Region represents a region
type Region struct {
	ID        int           `json:"id"`
	ShortName string        `json:"nomcourt"`
	NameDE    string        `json:"nom_de"`
	NameEN    string        `json:"nom_en"`
	NameES    string        `json:"nom_es"`
	NameFR    string        `json:"nom_fr"`
	NameIT    string        `json:"nom_it"`
	NamePT    string        `json:"nom_pt"`
	ParentID  string        `json:"parent"`
	Medias    *RegionMedias `json:"medias,omitempty"`
}

// RegionsListResponse is the complete response for the regions list endpoint
type RegionsListResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers ServerInfo        `json:"serveurs"`
		SSUser  *UserInfo         `json:"ssuser,omitempty"`
		Regions map[string]Region `json:"regions"`
	} `json:"response"`
}

// GetRegionsList retrieves the list of regions
func (c *Client) GetRegionsList() (*RegionsListResponse, error) {
	body, err := c.get("regionsListe.php", nil)
	if err != nil {
		return nil, err
	}

	var resp RegionsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse regions list response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
