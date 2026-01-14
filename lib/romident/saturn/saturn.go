package saturn

import (
	"strconv"
	"strings"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/romident/core"
)

// Sega Saturn disc identification from ISO 9660 system area.
//
// Saturn discs store metadata in the ISO 9660 system area (sectors 0-15).
// The System ID structure is at the start of sector 0.
//
// System ID structure (first 256 bytes):
//   - 0x00: Hardware ID (16 bytes) - "SEGA SEGASATURN "
//   - 0x10: Maker ID (16 bytes) - e.g., "SEGA ENTERPRISES" or "SEGA TP T-xxx"
//   - 0x20: Product Number (10 bytes) - e.g., "MK-81022" or "T-17602G"
//   - 0x2A: Version (6 bytes) - e.g., "V1.000"
//   - 0x30: Release Date (8 bytes) - YYYYMMDD format
//   - 0x38: Device Info (8 bytes) - e.g., "CD-1/1"
//   - 0x40: Area Symbols (16 bytes) - Region codes (J, T, U, B, K, A, E, L)
//   - 0x50: Peripherals (16 bytes) - Controller compatibility codes
//   - 0x60: Title (112 bytes) - Game title (space-padded)

const (
	Magic      = "SEGA SEGASATURN "
	HeaderSize = 256

	makerOffset      = 0x10
	makerSize        = 16
	productOffset    = 0x20
	productSize      = 10
	versionOffset    = 0x2A
	versionSize      = 6
	dateOffset       = 0x30
	dateSize         = 8
	deviceOffset     = 0x38
	deviceSize       = 8
	areaOffset       = 0x40
	areaSize         = 16
	peripheralOffset = 0x50
	peripheralSize   = 16
	titleOffset      = 0x60
	titleSize        = 112
)

// Info contains Saturn-specific metadata extracted from the disc header.
type Info struct {
	MakerID       string `json:"maker_id,omitempty"`
	ProductNumber string `json:"product_number,omitempty"`
	Version       string `json:"version,omitempty"`
	ReleaseDate   string `json:"release_date,omitempty"`
	DeviceInfo    string `json:"device_info,omitempty"`
	AreaSymbols   string `json:"area_symbols,omitempty"`
	Peripherals   string `json:"peripherals,omitempty"`
}

// IdentifyFromSystemArea parses Saturn metadata from ISO 9660 system area data.
// Returns nil if the data doesn't contain a valid Saturn header.
func IdentifyFromSystemArea(data []byte) *core.GameIdent {
	if len(data) < HeaderSize {
		return nil
	}

	// Validate magic
	if string(data[:len(Magic)]) != Magic {
		return nil
	}

	// Extract fields
	info := &Info{
		MakerID:       util.ExtractASCII(data[makerOffset : makerOffset+makerSize]),
		ProductNumber: util.ExtractASCII(data[productOffset : productOffset+productSize]),
		Version:       util.ExtractASCII(data[versionOffset : versionOffset+versionSize]),
		ReleaseDate:   util.ExtractASCII(data[dateOffset : dateOffset+dateSize]),
		DeviceInfo:    util.ExtractASCII(data[deviceOffset : deviceOffset+deviceSize]),
		AreaSymbols:   util.ExtractASCII(data[areaOffset : areaOffset+areaSize]),
		Peripherals:   util.ExtractASCII(data[peripheralOffset : peripheralOffset+peripheralSize]),
	}

	title := util.ExtractASCII(data[titleOffset : titleOffset+titleSize])

	// Parse disc number from DeviceInfo (e.g., "CD-1/2" -> disc 1 of 2)
	var discNumber *int
	if info.DeviceInfo != "" {
		if dn := parseDiscNumber(info.DeviceInfo); dn > 0 {
			discNumber = &dn
		}
	}

	return &core.GameIdent{
		Platform:   core.PlatformSaturn,
		TitleID:    info.ProductNumber,
		Title:      title,
		MakerCode:  extractMakerCode(info.MakerID),
		DiscNumber: discNumber,
		Extra:      info,
	}
}

// extractMakerCode extracts the maker code from the Maker ID field.
// First-party: "SEGA ENTERPRISES" -> "SEGA"
// Third-party: "SEGA TP T-176   " -> "T-176"
func extractMakerCode(makerID string) string {
	// Third-party games have format "SEGA TP X-xxx"
	if strings.HasPrefix(makerID, "SEGA TP ") {
		code := strings.TrimPrefix(makerID, "SEGA TP ")
		return strings.TrimSpace(code)
	}

	// First-party is just "SEGA ENTERPRISES" or similar
	if strings.HasPrefix(makerID, "SEGA") {
		return "SEGA"
	}

	return strings.TrimSpace(makerID)
}

// parseDiscNumber extracts disc number from device info like "CD-1/2".
func parseDiscNumber(deviceInfo string) int {
	// Format: "CD-X/Y" where X is disc number, Y is total discs
	if !strings.HasPrefix(deviceInfo, "CD-") {
		return 0
	}

	rest := strings.TrimPrefix(deviceInfo, "CD-")
	parts := strings.Split(rest, "/")
	if len(parts) < 1 {
		return 0
	}

	n, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0
	}
	return n
}
