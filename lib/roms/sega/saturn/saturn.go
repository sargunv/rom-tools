package saturn

import (
	"fmt"
	"io"
	"time"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
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

// Area represents Saturn area codes as a bitfield.
// Multiple areas can be combined with bitwise OR.
type Area uint8

const (
	AreaJapan        Area = 1 << 0 // J - Japan
	AreaAsiaNTSC     Area = 1 << 1 // T - Asia NTSC (Taiwan, Philippines)
	AreaNorthAmerica Area = 1 << 2 // U - North America NTSC (USA, Canada)
	AreaBrazil       Area = 1 << 3 // B - Brazil
	AreaKorea        Area = 1 << 4 // K - Korea
	AreaAsiaPAL      Area = 1 << 5 // A - Asia PAL
	AreaEurope       Area = 1 << 6 // E - Europe PAL
	AreaLatinAmerica Area = 1 << 7 // L - Latin America
)

const (
	magic      = "SEGA SEGASATURN "
	headerSize = 256

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

// SaturnInfo contains metadata extracted from a Saturn disc header.
type SaturnInfo struct {
	// Title is the game title (space-padded, up to 112 characters).
	Title string `json:"title,omitempty"`
	// MakerID identifies the publisher (e.g., "SEGA ENTERPRISES").
	MakerID string `json:"maker_id,omitempty"`
	// ProductNumber is the product code (e.g., "MK-81022" or "T-17602G").
	ProductNumber string `json:"product_number,omitempty"`
	// Version is the disc version (e.g., "V1.000").
	Version string `json:"version,omitempty"`
	// ReleaseDate is the release date parsed from YYYYMMDD format.
	// Zero value indicates the date could not be parsed.
	ReleaseDate time.Time `json:"release_date,omitempty"`
	// DeviceInfo describes the disc format (e.g., "CD-1/1").
	DeviceInfo string `json:"device_info,omitempty"`
	// Area is a bitfield of supported areas.
	Area Area `json:"area,omitempty"`
	// Peripherals contains controller compatibility codes.
	Peripherals string `json:"peripherals,omitempty"`
}

// GamePlatform implements identify.GameInfo.
func (i *SaturnInfo) GamePlatform() core.Platform { return core.PlatformSaturn }

// GameTitle implements identify.GameInfo.
func (i *SaturnInfo) GameTitle() string { return i.Title }

// GameSerial implements identify.GameInfo.
func (i *SaturnInfo) GameSerial() string { return i.ProductNumber }

// ParseSaturn parses Saturn metadata from a reader.
// The reader should contain the ISO 9660 system area data.
func ParseSaturn(r io.ReaderAt, size int64) (*SaturnInfo, error) {
	if size < headerSize {
		return nil, fmt.Errorf("file too small for Saturn header: need %d bytes, got %d", headerSize, size)
	}

	data := make([]byte, headerSize)
	if _, err := r.ReadAt(data, 0); err != nil {
		return nil, fmt.Errorf("failed to read Saturn header: %w", err)
	}

	return parseSaturnBytes(data)
}

func parseSaturnBytes(data []byte) (*SaturnInfo, error) {
	// Validate magic
	if string(data[:len(magic)]) != magic {
		return nil, fmt.Errorf("not a valid Saturn disc: invalid magic")
	}

	// Parse release date
	dateStr := util.ExtractASCII(data[dateOffset : dateOffset+dateSize])
	releaseDate := util.ParseYYYYMMDD(dateStr)

	// Parse area codes
	area := parseAreaSymbols(data[areaOffset : areaOffset+areaSize])

	info := &SaturnInfo{
		Title:         util.ExtractShiftJIS(data[titleOffset : titleOffset+titleSize]),
		MakerID:       util.ExtractASCII(data[makerOffset : makerOffset+makerSize]),
		ProductNumber: util.ExtractASCII(data[productOffset : productOffset+productSize]),
		Version:       util.ExtractASCII(data[versionOffset : versionOffset+versionSize]),
		ReleaseDate:   releaseDate,
		DeviceInfo:    util.ExtractASCII(data[deviceOffset : deviceOffset+deviceSize]),
		Area:          area,
		Peripherals:   util.ExtractASCII(data[peripheralOffset : peripheralOffset+peripheralSize]),
	}

	return info, nil
}

// parseAreaSymbols extracts area codes from the area symbols field.
// Saturn uses ASCII characters: J (Japan), T (Asia NTSC), U (North America),
// B (Brazil), K (Korea), A (Asia PAL), E (Europe), L (Latin America).
func parseAreaSymbols(data []byte) Area {
	var area Area
	for _, b := range data {
		switch b {
		case 'J':
			area |= AreaJapan
		case 'T':
			area |= AreaAsiaNTSC
		case 'U':
			area |= AreaNorthAmerica
		case 'B':
			area |= AreaBrazil
		case 'K':
			area |= AreaKorea
		case 'A':
			area |= AreaAsiaPAL
		case 'E':
			area |= AreaEurope
		case 'L':
			area |= AreaLatinAmerica
		}
	}
	return area
}
