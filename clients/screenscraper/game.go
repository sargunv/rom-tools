package screenscraper

import (
	"encoding/json"
	"fmt"
)

// IDText represents an entity with an ID and text name
type IDText struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// GameClassification represents a game rating/classification in game responses
type GameClassification struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// DateEntry represents a release date for a specific region
type DateEntry struct {
	Region string `json:"region"`
	Text   string `json:"text"`
}

// Players represents player count information
type Players struct {
	Text string `json:"text"`
}

// SystemInfo represents system information in game responses
type SystemInfo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// ROMRegions represents region information for a ROM
type ROMRegions struct {
	RegionsDE        []string `json:"regions_de,omitempty"`
	RegionsEN        []string `json:"regions_en,omitempty"`
	RegionsES        []string `json:"regions_es,omitempty"`
	RegionsFR        []string `json:"regions_fr,omitempty"`
	RegionsID        []string `json:"regions_id,omitempty"`
	RegionsPT        []string `json:"regions_pt,omitempty"`
	RegionsShortName []string `json:"regions_shortname,omitempty"`
}

// ROM represents a ROM file associated with a game
type ROM struct {
	ID              string      `json:"id"`
	ROMNumSupport   string      `json:"romnumsupport"`
	ROMTotalSupport string      `json:"romtotalsupport"`
	ROMFilename     string      `json:"romfilename"`
	ROMSize         string      `json:"romsize"`
	ROMCRC          string      `json:"romcrc"`
	ROMMD5          string      `json:"rommd5"`
	ROMSHA1         string      `json:"romsha1"`
	ROMCloneOf      string      `json:"romcloneof"`
	Beta            string      `json:"beta"`
	Demo            string      `json:"demo"`
	Trad            string      `json:"trad"`
	Hack            string      `json:"hack"`
	Unl             string      `json:"unl"`
	Alt             string      `json:"alt"`
	Best            string      `json:"best"`
	Netplay         string      `json:"netplay"`
	Proto           string      `json:"proto"`
	NbScrap         string      `json:"nbscrap"`
	Regions         *ROMRegions `json:"regions,omitempty"`
}

// GameGenre represents genre information in game responses
type GameGenre struct {
	ID         string          `json:"id"`
	ShortName  string          `json:"nomcourt"`
	Names      []LocalizedName `json:"noms"`
	ParentID   string          `json:"parentid"`
	Principale string          `json:"principale"`
}

// GameFamily represents family information in game responses
type GameFamily struct {
	ID         string          `json:"id"`
	ShortName  string          `json:"nomcourt"`
	Names      []LocalizedName `json:"noms"`
	ParentID   string          `json:"parentid"`
	Principale string          `json:"principale"`
}

// Game represents a game with all its information
type Game struct {
	ID               string      `json:"id"`
	ROMID            string      `json:"romid,omitempty"`
	NotGame          string      `json:"notgame,omitempty"`
	Name             string      `json:"nom,omitempty"`
	Names            []NameEntry `json:"noms,omitempty"`
	RegionShortNames []string    `json:"regionshortnames,omitempty"`
	CloneOf          string      `json:"cloneof,omitempty"`
	System           SystemInfo  `json:"systeme"`
	Publisher        IDText      `json:"editeur"`
	PublisherMedias  *struct {
		MediaPictoMonochrome string `json:"editeurmedia_pictomonochrome,omitempty"`
		MediaPictoCouleur    string `json:"editeurmedia_pictocouleur,omitempty"`
	} `json:"editeurmedias,omitempty"`
	Developer       IDText `json:"developpeur"`
	DeveloperMedias *struct {
		MediaPictoMonochrome string `json:"developpeurmedia_pictomonochrome,omitempty"`
		MediaPictoCouleur    string `json:"developpeurmedia_pictocouleur,omitempty"`
	} `json:"developpeurmedias,omitempty"`
	Players       Players `json:"joueurs"`
	PlayersMedias *struct {
		MediaPictoListe      string `json:"joueursmedia_pictoliste,omitempty"`
		MediaPictoMonochrome string `json:"joueursmedia_pictomonochrome,omitempty"`
		MediaPictoCouleur    string `json:"joueursmedia_pictocouleur,omitempty"`
	} `json:"joueursmedias,omitempty"`
	Note       *Players `json:"note,omitempty"`
	NoteMedias *struct {
		MediaPictoListe      string `json:"notemedia_pictoliste,omitempty"`
		MediaPictoMonochrome string `json:"notemedia_pictomonochrome,omitempty"`
		MediaPictoCouleur    string `json:"notemedia_pictocouleur,omitempty"`
	} `json:"notemedias,omitempty"`
	TopStaff        string               `json:"topstaff,omitempty"`
	Rotation        string               `json:"rotation,omitempty"`
	Resolution      string               `json:"resolution,omitempty"`
	Synopsis        []LocalizedName      `json:"synopsis,omitempty"`
	Classifications []GameClassification `json:"classifications,omitempty"`
	Dates           []DateEntry          `json:"dates,omitempty"`
	Genres          []GameGenre          `json:"genres,omitempty"`
	Modes           []GameGenre          `json:"modes,omitempty"`
	Families        []GameFamily         `json:"familles,omitempty"`
	Numbers         []GameGenre          `json:"numeros,omitempty"`
	Themes          []GameGenre          `json:"themes,omitempty"`
	Styles          []GameGenre          `json:"styles,omitempty"`
	Medias          []Media              `json:"medias,omitempty"`
	ROMs            []ROM                `json:"roms,omitempty"`
	ROM             *ROM                 `json:"rom,omitempty"`
}

// GameInfoParams parameters for game info lookup
type GameInfoParams struct {
	// ROM identification (at least one hash + size recommended)
	CRC     string
	MD5     string
	SHA1    string
	ROMSize string

	// Required context
	SystemID string
	ROMType  string // "rom", "iso", or "folder"
	ROMName  string

	// Alternative: direct game lookup
	GameID string

	// Optional
	SerialNum string
}

// GameInfoResponse is the complete response for the game info endpoint
type GameInfoResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers ServerInfo `json:"serveurs"`
		SSUser  *UserInfo  `json:"ssuser,omitempty"`
		Game    Game       `json:"jeu"`
	} `json:"response"`
}

// GetGameInfo retrieves detailed game information and media
func (c *Client) GetGameInfo(params GameInfoParams) (*GameInfoResponse, error) {
	p := map[string]string{
		"crc":       params.CRC,
		"md5":       params.MD5,
		"sha1":      params.SHA1,
		"romtaille": params.ROMSize,
		"systemeid": params.SystemID,
		"romtype":   params.ROMType,
		"romnom":    params.ROMName,
		"gameid":    params.GameID,
		"serialnum": params.SerialNum,
	}
	body, err := c.get("jeuInfos.php", p)
	if err != nil {
		return nil, err
	}

	var resp GameInfoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse game info response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
