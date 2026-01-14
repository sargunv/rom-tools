// Package core provides shared types for ROM identification.
package core

import "io"

// Format indicates the detected file format.
type Format string

const (
	FormatUnknown Format = "unknown"
	FormatCHD     Format = "chd"
	FormatXISO    Format = "xiso"
	FormatXBE     Format = "xbe"
	FormatISO9660 Format = "iso9660"
	FormatZIP     Format = "zip"
	FormatGBA     Format = "gba"
	FormatZ64     Format = "z64" // N64 big-endian (native)
	FormatV64     Format = "v64" // N64 byte-swapped
	FormatN64     Format = "n64" // N64 word-swapped (little-endian)
	FormatGB      Format = "gb"
	FormatMD      Format = "md"
	FormatSMD     Format = "smd"
	FormatNDS     Format = "nds"
	FormatNES     Format = "nes"
	FormatSNES    Format = "snes"
	FormatGCM     Format = "gcm" // GameCube/Wii uncompressed disc
	FormatRVZ     Format = "rvz" // GameCube/Wii RVZ/WIA compressed disc
)

// IdentifyFunc is the signature for format-specific identification functions.
// Each function verifies the format and extracts game metadata.
type IdentifyFunc func(r io.ReaderAt, size int64) (*GameIdent, error)

// GameIdent represents platform-specific identification data.
type GameIdent struct {
	Platform   Platform `json:"platform"`
	TitleID    string   `json:"title_id,omitempty"`
	Title      string   `json:"title,omitempty"`
	Regions    []Region `json:"regions,omitempty"`
	MakerCode  string   `json:"maker_code,omitempty"`
	Version    *int     `json:"version,omitempty"`
	DiscNumber *int     `json:"disc_number,omitempty"`
	Extra      any      `json:"extra,omitempty"`
}

// Region represents a game region using ISO country codes, continent codes, and some other non-country codes.
type Region string

const (
	RegionJP Region = "JP" // Japan
	RegionUS Region = "US" // USA
	RegionFR Region = "FR" // France
	RegionES Region = "ES" // Spain
	RegionDE Region = "DE" // Germany
	RegionIT Region = "IT" // Italy
	RegionAU Region = "AU" // Australia
	RegionBR Region = "BR" // Brazil
	RegionCN Region = "CN" // China
	RegionNL Region = "NL" // Netherlands
	RegionKR Region = "KR" // Korea
	RegionCA Region = "CA" // Canada
	RegionSE Region = "SE" // Sweden
	RegionFI Region = "FI" // Finland
	RegionDK Region = "DK" // Denmark

	RegionNA     Region = "NA"     // North America
	RegionEU     Region = "EU"     // Europe
	RegionNordic Region = "Nordic" // Scandinavia

	RegionNTSC Region = "NTSC"
	RegionPAL  Region = "PAL"

	RegionWorld   Region = "World"
	RegionUnknown Region = "Unknown"
)

// Platform represents a gaming platform.
type Platform string

const (
	PlatformNES     Platform = "famicom"
	PlatformSNES    Platform = "superfamicom"
	PlatformN64     Platform = "nintendo64"
	PlatformGC      Platform = "gamecube"
	PlatformWii     Platform = "wii"
	PlatformWiiU    Platform = "wiiu"
	PlatformSwitch  Platform = "switch"
	PlatformSwitch2 Platform = "switch2"

	PlatformGB     Platform = "gameboy"
	PlatformGBC    Platform = "gameboycolor"
	PlatformGBA    Platform = "gameboyadvance"
	PlatformNDS    Platform = "ds"
	PlatformDSi    Platform = "dsi"
	Platform3DS    Platform = "3ds"
	PlatformNew3DS Platform = "new3ds"

	PlatformPS1 Platform = "playstation"
	PlatformPS2 Platform = "playstation2"
	PlatformPS3 Platform = "playstation3"
	PlatformPS4 Platform = "playstation4"
	PlatformPS5 Platform = "playstation5"

	PlatformPSP    Platform = "psp"
	PlatformPSVita Platform = "psvita"

	PlatformMS        Platform = "mastersystem"
	PlatformMD        Platform = "megadrive"
	PlatformSaturn    Platform = "saturn"
	PlatformDreamcast Platform = "dreamcast"

	PlatformGameGear Platform = "gamegear"

	PlatformXbox       Platform = "xbox"
	PlatformXbox360    Platform = "xbox360"
	PlatformXboxOne    Platform = "xboxone"
	PlatformXboxSeries Platform = "xboxseries"
)

// ReaderAtSeekCloser combines io.ReaderAt, io.Seeker, and io.Closer with Size().
// This is needed for format detection and identification which require random access.
type ReaderAtSeekCloser interface {
	io.ReaderAt
	io.Seeker
	io.Closer
	Size() int64
}

// FileEntry represents a file within a container.
type FileEntry struct {
	Name  string // Relative path within container
	Size  int64  // Uncompressed size
	CRC32 uint32 // Pre-computed CRC32 (0 if not available, e.g., for folders)
}

// Container represents a container format (ZIP, folder, etc.) that can enumerate
// and provide access to its contents.
type Container interface {
	// Entries returns all files in the container.
	Entries() []FileEntry

	// OpenFile opens a file for sequential reading.
	OpenFile(name string) (io.ReadCloser, error)

	// OpenFileAt opens a file with random access support (for format detection/identification).
	// Only available in slow mode for containers that require decompression (e.g., ZIP).
	OpenFileAt(name string) (ReaderAtSeekCloser, error)

	// Close releases resources associated with the container.
	Close() error
}
