package dreamcast

import (
	"fmt"
	"io"
	"time"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
)

// Sega Dreamcast disc identification from ISO 9660 system area.
//
// Dreamcast discs store metadata in the ISO 9660 system area (sectors 0-15).
// The IP.BIN header is at the start of sector 0.
//
// Specification: https://mc.pp.se/dc/ip0000.bin.html
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

// Area represents Dreamcast area codes as a bitfield.
// Multiple areas can be combined with bitwise OR.
type Area uint8

const (
	AreaJapan  Area = 1 << 0 // J - Japan/East Asia (Taiwan, Philippines)
	AreaUSA    Area = 1 << 1 // U - USA/Canada
	AreaEurope Area = 1 << 2 // E - Europe
)

const (
	magic      = "SEGA SEGAKATANA "
	headerSize = 256

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

// Info contains metadata extracted from a Dreamcast disc header.
type Info struct {
	// Title is the game title (space-padded, up to 128 characters).
	Title string `json:"title,omitempty"`
	// ProductNumber is the product code (e.g., "MK-51058" or "T-xxxxx").
	ProductNumber string `json:"product_number,omitempty"`
	// MakerID identifies the publisher (e.g., "SEGA ENTERPRISES").
	MakerID string `json:"maker_id,omitempty"`
	// DeviceInfo describes the disc format (e.g., "D018 GD-ROM1/1").
	DeviceInfo string `json:"device_info,omitempty"`
	// Area is a bitfield of supported areas.
	Area Area `json:"area,omitempty"`
	// Peripherals contains controller compatibility hex flags.
	Peripherals string `json:"peripherals,omitempty"`
	// Version is the disc version (e.g., "V1.005").
	Version string `json:"version,omitempty"`
	// ReleaseDate is the release date parsed from YYYYMMDD format.
	// Zero value indicates the date could not be parsed.
	ReleaseDate time.Time `json:"release_date,omitempty"`
	// BootFilename is the boot executable filename (e.g., "1ST_READ.BIN").
	BootFilename string `json:"boot_filename,omitempty"`
	// SWMakerName is the software maker/developer name.
	SWMakerName string `json:"sw_maker_name,omitempty"`
}

// GamePlatform implements core.GameInfo.
func (i *Info) GamePlatform() core.Platform { return core.PlatformDreamcast }

// GameTitle implements core.GameInfo.
func (i *Info) GameTitle() string { return i.Title }

// GameSerial implements core.GameInfo.
func (i *Info) GameSerial() string { return i.ProductNumber }

// GameRegions implements core.GameInfo.
func (i *Info) GameRegions() []core.Region {
	var regions []core.Region
	if i.Area&AreaJapan != 0 {
		regions = append(regions, core.RegionJapan)
	}
	if i.Area&AreaUSA != 0 {
		regions = append(regions, core.RegionUSA)
	}
	if i.Area&AreaEurope != 0 {
		regions = append(regions, core.RegionEurope)
	}
	return regions
}

// Parse parses Dreamcast metadata from a reader.
// The reader should contain the ISO 9660 system area data.
func Parse(r io.ReaderAt, size int64) (*Info, error) {
	if size < headerSize {
		return nil, fmt.Errorf("data too small for Dreamcast header: %d bytes", size)
	}

	data := make([]byte, headerSize)
	if _, err := r.ReadAt(data, 0); err != nil {
		return nil, fmt.Errorf("failed to read Dreamcast header: %w", err)
	}

	return parseDreamcastBytes(data)
}

func parseDreamcastBytes(data []byte) (*Info, error) {
	// Validate magic
	if string(data[:len(magic)]) != magic {
		return nil, fmt.Errorf("not a valid Dreamcast disc: invalid magic")
	}

	// Parse release date
	dateStr := util.ExtractASCII(data[dateOffset : dateOffset+dateSize])
	releaseDate := util.ParseYYYYMMDD(dateStr)

	// Parse area codes
	area := parseAreaSymbols(data[areaOffset : areaOffset+areaSize])

	info := &Info{
		Title:         util.ExtractShiftJIS(data[titleOffset : titleOffset+titleSize]),
		ProductNumber: util.ExtractASCII(data[productOffset : productOffset+productSize]),
		MakerID:       util.ExtractASCII(data[makerOffset : makerOffset+makerSize]),
		DeviceInfo:    util.ExtractASCII(data[deviceOffset : deviceOffset+deviceSize]),
		Area:          area,
		Peripherals:   util.ExtractASCII(data[peripheralOffset : peripheralOffset+peripheralSize]),
		Version:       util.ExtractASCII(data[versionOffset : versionOffset+versionSize]),
		ReleaseDate:   releaseDate,
		BootFilename:  util.ExtractASCII(data[bootFileOffset : bootFileOffset+bootFileSize]),
		SWMakerName:   util.ExtractASCII(data[swMakerOffset : swMakerOffset+swMakerSize]),
	}

	return info, nil
}

// parseAreaSymbols extracts area codes from the area symbols field.
// Dreamcast uses ASCII characters: J (Japan/East Asia), U (USA/Canada), E (Europe).
func parseAreaSymbols(data []byte) Area {
	var area Area
	for _, b := range data {
		switch b {
		case 'J':
			area |= AreaJapan
		case 'U':
			area |= AreaUSA
		case 'E':
			area |= AreaEurope
		}
	}
	return area
}
