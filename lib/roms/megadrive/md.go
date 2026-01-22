package megadrive

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/sargunv/rom-tools/internal/util"
)

// Mega Drive (Genesis) ROM format parsing.
//
// Mega Drive ROM header specification:
// https://plutiedev.com/rom-header
//
// Header layout (starting at $100):
//
//	Offset   Size  Description
//	$100     16    System Type (e.g., "SEGA MEGA DRIVE" or "SEGA GENESIS")
//	$110     16    Copyright/Release Date (e.g., "(C)SEGA YYYY.MM")
//	$120     48    Domestic Title (Japanese)
//	$150     48    Overseas Title (International)
//	$180     14    Serial Number/Product Code
//	$18E     2     Checksum (big-endian)
//	$190     16    Device Support (I/O info)
//	$1A0     8     ROM Address Range
//	$1A8     8     RAM Address Range
//	$1B0     12    Extra Memory (SRAM info)
//	$1BC     12    Modem Support
//	$1C8     40    Reserved
//	$1F0     16    Region Support (first 3 chars typically significant)

const (
	mdHeaderStart      = 0x100
	mdHeaderSize       = 0x100 // 256 bytes ($100-$1FF)
	mdSystemTypeOffset = 0x100
	mdSystemTypeLen      = 16
	mdCopyrightOffset    = 0x110
	mdCopyrightLen       = 16
	mdDomesticTitleOff   = 0x120
	mdDomesticTitleLen   = 48
	mdOverseasTitleOff   = 0x150
	mdOverseasTitleLen   = 48
	mdSerialNumberOffset = 0x180
	mdSerialNumberLen    = 14
	mdChecksumOffset     = 0x18E
	mdDeviceSupportOff   = 0x190
	mdDeviceSupportLen   = 16
	mdRegionOffset       = 0x1F0
	mdRegionLen          = 16
)

// MDInfo contains metadata extracted from a Mega Drive/Genesis ROM file.
type MDInfo struct {
	// SystemType identifies the console (e.g., "SEGA MEGA DRIVE", "SEGA GENESIS").
	SystemType string
	// Copyright contains copyright and release date info.
	Copyright string
	// DomesticTitle is the Japanese title.
	DomesticTitle string
	// OverseasTitle is the international title.
	OverseasTitle string
	// SerialNumber is the product code (e.g., "GM XXXXXXXX-XX").
	SerialNumber string
	// Checksum is the ROM checksum (big-endian).
	Checksum uint16
	// DeviceSupport indicates I/O device support.
	DeviceSupport string
	// Regions contains region codes (J, U, E, or new-style hex).
	Regions []byte
}

// ParseMD extracts game information from a Mega Drive ROM.
func ParseMD(r io.ReaderAt, size int64) (*MDInfo, error) {
	if size < mdHeaderStart+mdHeaderSize {
		return nil, fmt.Errorf("file too small for Mega Drive header: %d bytes", size)
	}

	data := make([]byte, mdHeaderStart+mdHeaderSize)
	if _, err := r.ReadAt(data, 0); err != nil {
		return nil, fmt.Errorf("failed to read Mega Drive ROM: %w", err)
	}

	return parseMDBytes(data)
}

func parseMDBytes(data []byte) (*MDInfo, error) {
	// Extract system type and verify
	systemType := util.ExtractASCII(data[mdSystemTypeOffset : mdSystemTypeOffset+mdSystemTypeLen])
	if !strings.Contains(systemType, "SEGA") {
		return nil, fmt.Errorf("not a valid Mega Drive ROM: system type is %q", systemType)
	}

	// Extract all fields
	copyright := util.ExtractASCII(data[mdCopyrightOffset : mdCopyrightOffset+mdCopyrightLen])
	domesticTitle := util.ExtractASCII(data[mdDomesticTitleOff : mdDomesticTitleOff+mdDomesticTitleLen])
	overseasTitle := util.ExtractASCII(data[mdOverseasTitleOff : mdOverseasTitleOff+mdOverseasTitleLen])
	serialNumber := util.ExtractASCII(data[mdSerialNumberOffset : mdSerialNumberOffset+mdSerialNumberLen])

	// Extract checksum (big-endian)
	checksum := binary.BigEndian.Uint16(data[mdChecksumOffset:])

	// Extract device support
	deviceSupport := util.ExtractASCII(data[mdDeviceSupportOff : mdDeviceSupportOff+mdDeviceSupportLen])

	// Extract region codes
	regionData := data[mdRegionOffset : mdRegionOffset+mdRegionLen]
	regions := parseRegionCodes(regionData)

	return &MDInfo{
		SystemType:    systemType,
		Copyright:     copyright,
		DomesticTitle: domesticTitle,
		OverseasTitle: overseasTitle,
		SerialNumber:  serialNumber,
		Checksum:      checksum,
		DeviceSupport: deviceSupport,
		Regions:       regions,
	}, nil
}

// parseRegionCodes extracts region codes from the region field.
// Mega Drive uses two styles:
// - Old style: ASCII chars like J (Japan), U (USA), E (Europe) - multiple chars
// - New style: Single hex digit as bitfield (bit 0=Japan, bit 2=USA, bit 3=Europe)
func parseRegionCodes(data []byte) []byte {
	var regions []byte

	// First, count non-space/null characters to determine style
	var chars []byte
	for _, b := range data {
		if b == 0x00 || b == ' ' {
			break
		}
		chars = append(chars, b)
	}

	if len(chars) == 0 {
		return regions
	}

	// New-style: single hex digit (0-9 or A-F, but NOT J/U/E which are old-style)
	// If there's exactly one character and it's a digit 0-9 or A-F (excluding ambiguous letters),
	// treat it as new-style bitfield
	if len(chars) == 1 && isNewStyleRegionHex(chars[0]) {
		b := chars[0]
		var val byte
		if b >= '0' && b <= '9' {
			val = b - '0'
		} else if b >= 'A' && b <= 'F' {
			val = b - 'A' + 10
		} else if b >= 'a' && b <= 'f' {
			val = b - 'a' + 10
		}

		// Decode bitfield into region codes
		// bit 0 = Japan, bit 2 = Americas, bit 3 = Europe
		if val&0x01 != 0 {
			regions = append(regions, 'J') // Japan
		}
		if val&0x04 != 0 {
			regions = append(regions, 'U') // Americas/USA
		}
		if val&0x08 != 0 {
			regions = append(regions, 'E') // Europe
		}
		return regions
	}

	// Old style: multiple characters, each representing a region
	for _, b := range chars {
		if isOldStyleRegion(b) {
			regions = append(regions, b)
		}
	}

	return regions
}

// isNewStyleRegionHex checks if a byte is a valid new-style region hex digit.
// Only digits 0-9 and A-F are valid. Single letters like J, U, E are ambiguous
// and should be treated as old-style when alone.
func isNewStyleRegionHex(b byte) bool {
	// Digits 0-9 are unambiguous new-style
	if b >= '0' && b <= '9' {
		return true
	}
	// A, B, C, D, F are unambiguous new-style (E is also used in old-style but
	// typically not alone - a single 'E' would be weird for Europe-only)
	// For safety, treat A-F as new-style when it's a single character
	if b >= 'A' && b <= 'F' {
		return true
	}
	if b >= 'a' && b <= 'f' {
		return true
	}
	return false
}

// isOldStyleRegion checks if a byte is a valid old-style region code.
func isOldStyleRegion(b byte) bool {
	switch b {
	case 'J', // Japan
		'U', // USA/Americas
		'E', // Europe
		'A', // Asia
		'B', // Brazil
		'K', // Korea
		'4', // Brazil (alternative)
		'1', // Japan (alternative)
		'8': // USA (alternative)
		return true
	default:
		return false
	}
}
