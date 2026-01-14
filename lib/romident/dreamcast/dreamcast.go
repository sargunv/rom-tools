package dreamcast

import (
	"strconv"
	"strings"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/romident/core"
)

// Sega Dreamcast disc identification from ISO 9660 system area.
//
// Dreamcast discs store metadata in the ISO 9660 system area (sectors 0-15).
// The IP.BIN header is at the start of sector 0.
//
// IP.BIN header structure (first 256 bytes):
//   - 0x00: Hardware ID (16 bytes) - "SEGA SEGAKATANA " (Dreamcast codename)
//   - 0x10: Maker ID (16 bytes) - e.g., "SEGA ENTERPRISES"
//   - 0x20: Device Info (16 bytes) - CRC + "GD-ROM" + disc numbering (e.g., "D018 GD-ROM1/1")
//   - 0x30: Area Symbols (8 bytes) - Region codes (J, U, E, etc.)
//   - 0x38: Peripherals (8 bytes) - Controller compatibility hex flags
//   - 0x40: Product Number (10 bytes) - e.g., "MK-51058" or "T-xxxxx"
//   - 0x4A: Version (6 bytes) - e.g., "V1.005"
//   - 0x50: Release Date (8 bytes) - YYYYMMDD format
//   - 0x60: Boot Filename (16 bytes) - e.g., "1ST_READ.BIN"
//   - 0x70: SW Maker Name (16 bytes) - Publisher/Developer name
//   - 0x80: Title (128 bytes) - Game title (space-padded)

const (
	Magic      = "SEGA SEGAKATANA "
	HeaderSize = 256

	makerOffset      = 0x10
	makerSize        = 16
	deviceOffset     = 0x20
	deviceSize       = 16
	areaOffset       = 0x30
	areaSize         = 8
	peripheralOffset = 0x38
	peripheralSize   = 8
	productOffset    = 0x40
	productSize      = 10
	versionOffset    = 0x4A
	versionSize      = 6
	dateOffset       = 0x50
	dateSize         = 8
	bootFileOffset   = 0x60
	bootFileSize     = 16
	swMakerOffset    = 0x70
	swMakerSize      = 16
	titleOffset      = 0x80
	titleSize        = 128
)

// Info contains Dreamcast-specific metadata extracted from the disc header.
type Info struct {
	MakerID      string `json:"maker_id,omitempty"`
	DeviceInfo   string `json:"device_info,omitempty"`
	AreaSymbols  string `json:"area_symbols,omitempty"`
	Peripherals  string `json:"peripherals,omitempty"`
	Version      string `json:"version,omitempty"`
	ReleaseDate  string `json:"release_date,omitempty"`
	BootFilename string `json:"boot_filename,omitempty"`
	SWMakerName  string `json:"sw_maker_name,omitempty"`
}

// IdentifyFromSystemArea parses Dreamcast metadata from ISO 9660 system area data.
// Returns nil if the data doesn't contain a valid Dreamcast header.
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
		MakerID:      util.ExtractASCII(data[makerOffset : makerOffset+makerSize]),
		DeviceInfo:   util.ExtractASCII(data[deviceOffset : deviceOffset+deviceSize]),
		AreaSymbols:  string(data[areaOffset : areaOffset+areaSize]), // Don't trim - positions matter
		Peripherals:  util.ExtractASCII(data[peripheralOffset : peripheralOffset+peripheralSize]),
		Version:      util.ExtractASCII(data[versionOffset : versionOffset+versionSize]),
		ReleaseDate:  util.ExtractASCII(data[dateOffset : dateOffset+dateSize]),
		BootFilename: util.ExtractASCII(data[bootFileOffset : bootFileOffset+bootFileSize]),
		SWMakerName:  util.ExtractASCII(data[swMakerOffset : swMakerOffset+swMakerSize]),
	}

	productNumber := util.ExtractASCII(data[productOffset : productOffset+productSize])
	title := util.ExtractASCII(data[titleOffset : titleOffset+titleSize])

	// Parse disc number from DeviceInfo (e.g., "D018 GD-ROM1/2" -> disc 1 of 2)
	var discNumber *int
	if info.DeviceInfo != "" {
		if dn := parseDiscNumber(info.DeviceInfo); dn > 0 {
			discNumber = &dn
		}
	}

	// Decode regions from raw area symbols (positions matter, don't trim)
	areaSymbolsRaw := string(data[areaOffset : areaOffset+areaSize])

	return &core.GameIdent{
		Platform:   core.PlatformDreamcast,
		TitleID:    productNumber,
		Title:      title,
		Regions:    decodeRegions(areaSymbolsRaw),
		MakerCode:  extractMakerCode(info.MakerID),
		DiscNumber: discNumber,
		Extra:      info,
	}
}

// decodeRegions converts Dreamcast area symbols to regions.
// The 8-character area symbols field uses positions for regions:
// Position 0: Japan, Position 1: USA, Position 2: Europe, etc.
// A letter at a position means supported; a space means unsupported.
func decodeRegions(areaSymbols string) []core.Region {
	var regions []core.Region

	// Area symbols are positional - check each position
	for i, c := range areaSymbols {
		if c == ' ' {
			continue
		}
		switch i {
		case 0: // Japan
			regions = append(regions, core.RegionJP)
		case 1: // USA
			regions = append(regions, core.RegionUS)
		case 2: // Europe
			regions = append(regions, core.RegionEU)
		// Positions 3-7 are less common regions (Taiwan, etc.)
		// For now, map them to closest standard region
		case 3, 4, 5, 6, 7:
			// Asia/other regions - could be more specific but this covers basics
		}
	}

	// Deduplicate
	seen := make(map[core.Region]bool)
	unique := make([]core.Region, 0, len(regions))
	for _, r := range regions {
		if !seen[r] {
			seen[r] = true
			unique = append(unique, r)
		}
	}

	if len(unique) == 0 {
		return []core.Region{core.RegionUnknown}
	}
	return unique
}

// extractMakerCode extracts the maker code from the Maker ID field.
// First-party: "SEGA ENTERPRISES" -> "SEGA"
// Third-party: Keep full maker ID
func extractMakerCode(makerID string) string {
	if strings.HasPrefix(makerID, "SEGA") {
		return "SEGA"
	}
	return strings.TrimSpace(makerID)
}

// parseDiscNumber extracts disc number from device info like "D018 GD-ROM1/2".
func parseDiscNumber(deviceInfo string) int {
	// Format: "Dxxx GD-ROMX/Y" where X is disc number, Y is total discs
	idx := strings.Index(deviceInfo, "GD-ROM")
	if idx == -1 {
		return 0
	}

	rest := deviceInfo[idx+len("GD-ROM"):]
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
