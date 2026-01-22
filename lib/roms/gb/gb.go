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
	gbTitleMaxLen          = 15 // Max title length (0x134-0x142, CGB flag at 0x143)
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

// GBROMSize represents the ROM size code from the header.
type GBROMSize byte

// GBROMSize values
const (
	GBROMSize32KB  GBROMSize = 0x00 // 32 KB (2 banks)
	GBROMSize64KB  GBROMSize = 0x01 // 64 KB (4 banks)
	GBROMSize128KB GBROMSize = 0x02 // 128 KB (8 banks)
	GBROMSize256KB GBROMSize = 0x03 // 256 KB (16 banks)
	GBROMSize512KB GBROMSize = 0x04 // 512 KB (32 banks)
	GBROMSize1MB   GBROMSize = 0x05 // 1 MB (64 banks)
	GBROMSize2MB   GBROMSize = 0x06 // 2 MB (128 banks)
	GBROMSize4MB   GBROMSize = 0x07 // 4 MB (256 banks)
	GBROMSize8MB   GBROMSize = 0x08 // 8 MB (512 banks)
	GBROMSize1_1MB GBROMSize = 0x52 // 1.1 MB (72 banks)
	GBROMSize1_2MB GBROMSize = 0x53 // 1.2 MB (80 banks)
	GBROMSize1_5MB GBROMSize = 0x54 // 1.5 MB (96 banks)
)

// GBRAMSize represents the external RAM size code from the header.
type GBRAMSize byte

// GBRAMSize values
const (
	GBRAMSizeNone  GBRAMSize = 0x00 // No RAM
	GBRAMSize2KB   GBRAMSize = 0x01 // 2 KB (unofficial)
	GBRAMSize8KB   GBRAMSize = 0x02 // 8 KB (1 bank)
	GBRAMSize32KB  GBRAMSize = 0x03 // 32 KB (4 banks)
	GBRAMSize128KB GBRAMSize = 0x04 // 128 KB (16 banks)
	GBRAMSize64KB  GBRAMSize = 0x05 // 64 KB (8 banks)
)

// GBInfo contains metadata extracted from a GB/GBC ROM file.
type GBInfo struct {
	// Title is the game title (11-15 chars, space-padded).
	Title string
	// ManufacturerCode is the 4-char code in newer cartridges (empty for older games).
	ManufacturerCode string
	// CGBFlag is the Color Game Boy compatibility flag.
	CGBFlag GBCGBFlag
	// SGBFlag is the Super Game Boy compatibility flag.
	SGBFlag GBSGBFlag
	// CartridgeType is the MBC type and features code.
	CartridgeType byte
	// ROMSize is the ROM size code.
	ROMSize GBROMSize
	// RAMSize is the external RAM size code.
	RAMSize GBRAMSize
	// DestinationCode indicates the region (0x00=Japan, 0x01=Overseas).
	DestinationCode byte
	// LicenseeCode is the publisher identifier (old or new format).
	LicenseeCode string
	// Version is the ROM version number.
	Version int
	// HeaderChecksum is the checksum of the header bytes at 0x14D.
	HeaderChecksum byte
	// GlobalChecksum is the checksum of the entire ROM at 0x14E-0x14F.
	GlobalChecksum uint16
	// Platform is GB or GBC based on the CGB flag.
	Platform core.Platform
}

// hasManufacturerCode checks if the manufacturer bytes contain valid uppercase ASCII.
// Early CGB games (pre-1998) used 15-16 char titles without manufacturer codes.
func hasManufacturerCode(header []byte) bool {
	mfgStart := gbManufacturerOffset - gbHeaderStart
	for _, b := range header[mfgStart : mfgStart+gbManufacturerLen] {
		if b < 'A' || b > 'Z' {
			return false
		}
	}
	return true
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

	// Check if this cartridge has a valid manufacturer code by inspecting the bytes.
	// Only newer cartridges have 11-char title + 4-char uppercase manufacturer code.
	// Early CGB games and all original GB games use the full title area.
	if hasManufacturerCode(header) {
		// Newer format: 11-char title + 4-char manufacturer
		title = util.ExtractASCII(header[titleStart : titleStart+gbTitleNewLen])
		mfgStart := gbManufacturerOffset - gbHeaderStart
		manufacturerCode = util.ExtractASCII(header[mfgStart : mfgStart+gbManufacturerLen])
	} else {
		// Original format: title up to 15 chars (0x134-0x142)
		title = util.ExtractASCII(header[titleStart : titleStart+gbTitleMaxLen])
	}

	// Extract SGB flag
	sgbFlag := GBSGBFlag(header[gbSGBFlagOffset-gbHeaderStart])

	// Extract cartridge type
	cartType := header[gbCartTypeOffset-gbHeaderStart]

	// Extract ROM/RAM size
	romSize := GBROMSize(header[gbROMSizeOffset-gbHeaderStart])
	ramSize := GBRAMSize(header[gbRAMSizeOffset-gbHeaderStart])

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

	// Extract checksums
	headerChecksum := header[gbHeaderChecksumOffset-gbHeaderStart]
	globalChecksumHi := header[gbGlobalChecksumOffset-gbHeaderStart]
	globalChecksumLo := header[gbGlobalChecksumOffset-gbHeaderStart+1]
	globalChecksum := uint16(globalChecksumHi)<<8 | uint16(globalChecksumLo)

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
		HeaderChecksum:   headerChecksum,
		GlobalChecksum:   globalChecksum,
		Platform:         platform,
	}, nil
}
