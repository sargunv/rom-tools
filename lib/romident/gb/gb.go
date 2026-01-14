package gb

import (
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/romident/core"
)

// Game Boy (GB) and Game Boy Color (GBC) ROM format parsing.
//
// GB/GBC cartridge header specification:
// https://gbdev.io/pandocs/The_Cartridge_Header.html
//
// Header layout (starting at 0x100):
//
//	Offset  Size  Description
//	0x100   4     Entry Point (usually NOP + JP)
//	0x104   48    Nintendo Logo (required for boot)
//	0x134   16    Title (uppercase ASCII, may be shorter in newer games)
//	0x13F   4     Manufacturer Code (in newer cartridges, overlaps title)
//	0x143   1     CGB Flag (0x80 = CGB support, 0xC0 = CGB only)
//	0x144   2     New Licensee Code
//	0x146   1     SGB Flag (0x03 = SGB support)
//	0x147   1     Cartridge Type (MBC type)
//	0x148   1     ROM Size
//	0x149   1     RAM Size
//	0x14A   1     Destination Code (0x00 = Japan, 0x01 = Overseas)
//	0x14B   1     Old Licensee Code (0x33 = use new licensee)
//	0x14C   1     ROM Version
//	0x14D   1     Header Checksum
//	0x14E   2     Global Checksum

const (
	gbHeaderStart        = 0x100
	gbHeaderSize         = 0x50 // 0x100 to 0x14F
	gbTitleOffset        = 0x134
	gbTitleMaxLen        = 16
	gbTitleNewLen        = 11 // Title length in newer cartridges with manufacturer code
	gbManufacturerOffset = 0x13F
	gbManufacturerLen    = 4
	gbCGBFlagOffset      = 0x143
	gbNewLicenseeOffset  = 0x144
	gbNewLicenseeLen     = 2
	gbSGBFlagOffset      = 0x146
	gbCartTypeOffset     = 0x147
	gbROMSizeOffset      = 0x148
	gbRAMSizeOffset      = 0x149
	gbDestCodeOffset     = 0x14A
	gbOldLicenseeOffset  = 0x14B
	gbVersionOffset      = 0x14C
)

// CGBFlag values
const (
	CGBFlagNone      = 0x00 // No CGB support (original GB)
	CGBFlagSupported = 0x80 // Supports CGB functions, works on GB too
	CGBFlagRequired  = 0xC0 // CGB only
)

// SGBFlag values
const (
	SGBFlagNone      = 0x00 // No SGB support
	SGBFlagSupported = 0x03 // Supports SGB functions
)

// GBPlatform indicates whether this is a GB or GBC game
type GBPlatform string

const (
	GBPlatformGB  GBPlatform = "gb"  // Original Game Boy
	GBPlatformGBC GBPlatform = "gbc" // Game Boy Color (or dual-compatible)
)

// GBInfo contains metadata extracted from a GB/GBC ROM file.
type GBInfo struct {
	Title            string
	ManufacturerCode string // 4-char code in newer cartridges
	CGBFlag          byte
	SGBFlag          byte
	CartridgeType    byte
	ROMSize          byte
	RAMSize          byte
	DestinationCode  byte   // 0x00 = Japan, 0x01 = Overseas
	LicenseeCode     string // Old or New licensee code
	Version          int
	Platform         GBPlatform // GB or GBC based on CGB flag
}

// parseGB extracts game information from a GB/GBC ROM file.
func parseGB(r io.ReaderAt, size int64) (*GBInfo, error) {
	if size < gbHeaderStart+gbHeaderSize {
		return nil, fmt.Errorf("file too small for GB header: %d bytes", size)
	}

	header := make([]byte, gbHeaderSize)
	if _, err := r.ReadAt(header, gbHeaderStart); err != nil {
		return nil, fmt.Errorf("failed to read GB header: %w", err)
	}

	// Extract CGB flag to determine title length
	cgbFlagIdx := gbCGBFlagOffset - gbHeaderStart
	cgbFlag := header[cgbFlagIdx]

	// Determine platform based on CGB flag
	var platform GBPlatform
	if cgbFlag == CGBFlagSupported || cgbFlag == CGBFlagRequired {
		platform = GBPlatformGBC
	} else {
		platform = GBPlatformGB
	}

	// Extract title - length depends on whether manufacturer code is present
	titleStart := gbTitleOffset - gbHeaderStart
	var title string
	var manufacturerCode string

	// Check if this might be a newer cartridge with manufacturer code
	// Newer cartridges have title at 0x134-0x13E (11 chars) and manufacturer at 0x13F-0x142 (4 chars)
	// We can detect this by checking if CGB flag is set (newer format)
	if cgbFlag == CGBFlagSupported || cgbFlag == CGBFlagRequired {
		// Newer format: 11-char title + 4-char manufacturer
		title = util.ExtractASCII(header[titleStart : titleStart+gbTitleNewLen])
		mfgStart := gbManufacturerOffset - gbHeaderStart
		manufacturerCode = util.ExtractASCII(header[mfgStart : mfgStart+gbManufacturerLen])
	} else {
		// Original format: 16-char title (but CGB flag byte overlaps, so effectively 15 max)
		title = util.ExtractASCII(header[titleStart : titleStart+gbTitleMaxLen])
	}

	// Extract SGB flag
	sgbFlag := header[gbSGBFlagOffset-gbHeaderStart]

	// Extract cartridge type
	cartType := header[gbCartTypeOffset-gbHeaderStart]

	// Extract ROM/RAM size
	romSize := header[gbROMSizeOffset-gbHeaderStart]
	ramSize := header[gbRAMSizeOffset-gbHeaderStart]

	// Extract destination code
	destCode := header[gbDestCodeOffset-gbHeaderStart]

	// Extract licensee code
	oldLicensee := header[gbOldLicenseeOffset-gbHeaderStart]
	var licenseeCode string
	if oldLicensee == 0x33 {
		// Use new licensee code
		newLicStart := gbNewLicenseeOffset - gbHeaderStart
		licenseeCode = string(header[newLicStart : newLicStart+gbNewLicenseeLen])
	} else {
		// Use old licensee code as hex
		licenseeCode = fmt.Sprintf("%02X", oldLicensee)
	}

	// Extract version
	version := int(header[gbVersionOffset-gbHeaderStart])

	return &GBInfo{
		Title:            title,
		ManufacturerCode: manufacturerCode,
		CGBFlag:          cgbFlag,
		SGBFlag:          sgbFlag,
		CartridgeType:    cartType,
		ROMSize:          romSize,
		RAMSize:          ramSize,
		DestinationCode:  destCode,
		LicenseeCode:     licenseeCode,
		Version:          version,
		Platform:         platform,
	}, nil
}

// Identify verifies the format and extracts game identification from a GB/GBC ROM.
func Identify(r io.ReaderAt, size int64) (*core.GameIdent, error) {
	info, err := parseGB(r, size)
	if err != nil {
		return nil, err
	}

	version := info.Version

	// Determine platform based on CGB flag
	var platform core.Platform
	if info.Platform == GBPlatformGBC {
		platform = core.PlatformGBC
	} else {
		platform = core.PlatformGB
	}

	return &core.GameIdent{
		Platform:  platform,
		Title:     info.Title,
		MakerCode: info.LicenseeCode,
		Version:   &version,
		Extra:     info,
	}, nil
}
