package saturn

import (
	"fmt"
	"io"
	"time"

	"github.com/sargunv/rom-tools/internal/util"
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
	// AreaSymbols contains region codes (J, T, U, B, K, A, E, L).
	AreaSymbols string `json:"area_symbols,omitempty"`
	// Peripherals contains controller compatibility codes.
	Peripherals string `json:"peripherals,omitempty"`
}

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

	info := &SaturnInfo{
		Title:         util.ExtractASCII(data[titleOffset : titleOffset+titleSize]),
		MakerID:       util.ExtractASCII(data[makerOffset : makerOffset+makerSize]),
		ProductNumber: util.ExtractASCII(data[productOffset : productOffset+productSize]),
		Version:       util.ExtractASCII(data[versionOffset : versionOffset+versionSize]),
		ReleaseDate:   releaseDate,
		DeviceInfo:    util.ExtractASCII(data[deviceOffset : deviceOffset+deviceSize]),
		AreaSymbols:   string(data[areaOffset : areaOffset+areaSize]), // Don't trim - positions matter
		Peripherals:   util.ExtractASCII(data[peripheralOffset : peripheralOffset+peripheralSize]),
	}

	return info, nil
}
