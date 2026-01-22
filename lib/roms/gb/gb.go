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
	gbHeaderStart          = 0x100
	gbHeaderSize           = 0x50 // 0x100 to 0x14F
	gbTitleOffset          = 0x134
	gbTitleMaxLen          = 16
	gbTitleNewLen          = 11 // Title length in newer cartridges with manufacturer code
	gbManufacturerOffset   = 0x13F
	gbManufacturerLen      = 4
	gbCGBFlagOffset        = 0x143
	gbNewLicenseeOffset    = 0x144
	gbNewLicenseeLen       = 2
	gbSGBFlagOffset        = 0x146
	gbCartTypeOffset       = 0x147
	gbROMSizeOffset        = 0x148
	gbRAMSizeOffset        = 0x149
	gbDestCodeOffset       = 0x14A
	gbOldLicenseeOffset    = 0x14B
	gbVersionOffset        = 0x14C
	gbHeaderChecksumOffset = 0x14D
	gbGlobalChecksumOffset = 0x14E
)

// CGBFlag represents the Color Game Boy compatibility flag.
type CGBFlag byte

// CGBFlag values
const (
	CGBFlagNone      CGBFlag = 0x00 // No CGB support (original GB)
	CGBFlagSupported CGBFlag = 0x80 // Supports CGB functions, works on GB too
	CGBFlagRequired  CGBFlag = 0xC0 // CGB only
)

// SGBFlag represents the Super Game Boy compatibility flag.
type SGBFlag byte

// SGBFlag values
const (
	SGBFlagNone      SGBFlag = 0x00 // No SGB support
	SGBFlagSupported SGBFlag = 0x03 // Supports SGB functions
)

// ROMSize represents the ROM size code from the cartridge header.
type ROMSize byte

// ROMSize values (raw header codes).
const (
	ROMSize32KB  ROMSize = 0x00 // 32 KB (2 banks)
	ROMSize64KB  ROMSize = 0x01 // 64 KB (4 banks)
	ROMSize128KB ROMSize = 0x02 // 128 KB (8 banks)
	ROMSize256KB ROMSize = 0x03 // 256 KB (16 banks)
	ROMSize512KB ROMSize = 0x04 // 512 KB (32 banks)
	ROMSize1MB   ROMSize = 0x05 // 1 MB (64 banks)
	ROMSize2MB   ROMSize = 0x06 // 2 MB (128 banks)
	ROMSize4MB   ROMSize = 0x07 // 4 MB (256 banks)
	ROMSize8MB   ROMSize = 0x08 // 8 MB (512 banks)
)

// RAMSize represents the external RAM size code from the cartridge header.
type RAMSize byte

// RAMSize values (raw header codes).
const (
	RAMSizeNone  RAMSize = 0x00 // No RAM
	RAMSize2KB   RAMSize = 0x01 // 2 KB (unofficial, used by some games)
	RAMSize8KB   RAMSize = 0x02 // 8 KB (1 bank)
	RAMSize32KB  RAMSize = 0x03 // 32 KB (4 banks)
	RAMSize128KB RAMSize = 0x04 // 128 KB (16 banks)
	RAMSize64KB  RAMSize = 0x05 // 64 KB (8 banks)
)

// Destination represents the destination code indicating the target region.
type Destination byte

// Destination values.
const (
	DestinationJapan    Destination = 0x00 // Japanese market
	DestinationOverseas Destination = 0x01 // Non-Japanese markets
)

// GBInfo contains metadata extracted from a GB/GBC ROM file.
type GBInfo struct {
	// Title is the game title (11-16 chars, space-padded).
	Title string
	// ManufacturerCode is the 4-char code in newer cartridges (empty for older games).
	ManufacturerCode string
	// CGBFlag is the Color Game Boy compatibility flag.
	CGBFlag CGBFlag
	// SGBFlag is the Super Game Boy compatibility flag.
	SGBFlag SGBFlag
	// CartridgeType is the MBC type and features code.
	CartridgeType byte
	// ROMSize is the ROM size code.
	ROMSize ROMSize
	// RAMSize is the external RAM size code.
	RAMSize RAMSize
	// Destination indicates the target region.
	Destination Destination
	// LicenseeCode is the publisher identifier (old or new format).
	LicenseeCode string
	// Version is the ROM version number.
	Version int
	// HeaderChecksum is the header checksum byte (0x14D).
	HeaderChecksum byte
	// GlobalChecksum is the 16-bit global checksum (0x14E-0x14F, big-endian).
	GlobalChecksum uint16
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
	cgbFlag := CGBFlag(header[cgbFlagIdx])

	// Determine platform based on CGB flag
	var platform core.Platform
	if cgbFlag == CGBFlagSupported || cgbFlag == CGBFlagRequired {
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
	sgbFlag := SGBFlag(header[gbSGBFlagOffset-gbHeaderStart])

	// Extract cartridge type
	cartType := header[gbCartTypeOffset-gbHeaderStart]

	// Extract ROM/RAM size
	romSize := ROMSize(header[gbROMSizeOffset-gbHeaderStart])
	ramSize := RAMSize(header[gbRAMSizeOffset-gbHeaderStart])

	// Extract destination
	destination := Destination(header[gbDestCodeOffset-gbHeaderStart])

	// Extract checksums
	headerChecksum := header[gbHeaderChecksumOffset-gbHeaderStart]
	globalChecksum := uint16(header[gbGlobalChecksumOffset-gbHeaderStart])<<8 |
		uint16(header[gbGlobalChecksumOffset-gbHeaderStart+1])

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
		Destination:      destination,
		LicenseeCode:     licenseeCode,
		Version:          version,
		HeaderChecksum:   headerChecksum,
		GlobalChecksum:   globalChecksum,
		Platform:         platform,
	}, nil
}
