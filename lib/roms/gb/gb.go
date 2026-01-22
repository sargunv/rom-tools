package gb

import (
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
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

// GBCGBFlag represents the Color Game Boy compatibility flag.
type GBCGBFlag byte

// GBCGBFlag values
const (
	GBCGBFlagNone      GBCGBFlag = 0x00 // No CGB support (original GB)
	GBCGBFlagSupported GBCGBFlag = 0x80 // Supports CGB functions, works on GB too
	GBCGBFlagRequired  GBCGBFlag = 0xC0 // CGB only
)

// GBSGBFlag represents the Super Game Boy compatibility flag.
type GBSGBFlag byte

// GBSGBFlag values
const (
	GBSGBFlagNone      GBSGBFlag = 0x00 // No SGB support
	GBSGBFlagSupported GBSGBFlag = 0x03 // Supports SGB functions
)

// GBInfo contains metadata extracted from a GB/GBC ROM file.
type GBInfo struct {
	// Title is the game title (11-16 chars, space-padded).
	Title string
	// ManufacturerCode is the 4-char code in newer cartridges (empty for older games).
	ManufacturerCode string
	// CGBFlag is the Color Game Boy compatibility flag.
	CGBFlag GBCGBFlag
	// SGBFlag is the Super Game Boy compatibility flag.
	SGBFlag GBSGBFlag
	// CartridgeType is the MBC type and features code.
	CartridgeType byte
	// ROMSize is the ROM size code (32KB << n).
	ROMSize byte
	// RAMSize is the external RAM size code.
	RAMSize byte
	// DestinationCode indicates the region (0x00=Japan, 0x01=Overseas).
	DestinationCode byte
	// LicenseeCode is the publisher identifier (old or new format).
	LicenseeCode string
	// Version is the ROM version number.
	Version int
	// Platform is GB or GBC based on the CGB flag.
	Platform core.Platform
}

// ParseGB extracts game information from a GB/GBC ROM file.
func ParseGB(r io.ReaderAt, size int64) (*GBInfo, error) {
	if size < gbHeaderStart+gbHeaderSize {
		return nil, fmt.Errorf("file too small for GB header: %d bytes", size)
	}

	header := make([]byte, gbHeaderSize)
	if _, err := r.ReadAt(header, gbHeaderStart); err != nil {
		return nil, fmt.Errorf("failed to read GB header: %w", err)
	}

	// Extract CGB flag to determine title length
	cgbFlagIdx := gbCGBFlagOffset - gbHeaderStart
	cgbFlag := GBCGBFlag(header[cgbFlagIdx])

	// Determine platform based on CGB flag
	var platform core.Platform
	if cgbFlag == GBCGBFlagSupported || cgbFlag == GBCGBFlagRequired {
		platform = core.PlatformGBC
	} else {
		platform = core.PlatformGB
	}

	// Extract title - length depends on whether manufacturer code is present
	titleStart := gbTitleOffset - gbHeaderStart
	var title string
	var manufacturerCode string

	// Check if this might be a newer cartridge with manufacturer code
	// Newer cartridges have title at 0x134-0x13E (11 chars) and manufacturer at 0x13F-0x142 (4 chars)
	// We can detect this by checking if CGB flag is set (newer format)
	if cgbFlag == GBCGBFlagSupported || cgbFlag == GBCGBFlagRequired {
		// Newer format: 11-char title + 4-char manufacturer
		title = util.ExtractASCII(header[titleStart : titleStart+gbTitleNewLen])
		mfgStart := gbManufacturerOffset - gbHeaderStart
		manufacturerCode = util.ExtractASCII(header[mfgStart : mfgStart+gbManufacturerLen])
	} else {
		// Original format: 16-char title (but CGB flag byte overlaps, so effectively 15 max)
		title = util.ExtractASCII(header[titleStart : titleStart+gbTitleMaxLen])
	}

	// Extract SGB flag
	sgbFlag := GBSGBFlag(header[gbSGBFlagOffset-gbHeaderStart])

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
