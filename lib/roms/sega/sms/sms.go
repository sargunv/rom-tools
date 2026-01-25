package sms

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/lib/core"
)

// Sega Master System / Game Gear ROM format parsing.
//
// Both SMS and GG share the same "TMR SEGA" header format at offset 0x7FF0.
// The platform is determined by the region code in the header.
//
// Header specification:
// https://www.smspower.org/Development/ROMHeader
//
// Header layout (16 bytes at offset $7FF0):
//
//	Offset  Size  Description
//	$00     8     Magic string "TMR SEGA"
//	$08     2     Reserved
//	$0A     2     Checksum (little-endian)
//	$0C     2     Product code (BCD, big-endian)
//	$0E     1     Product code high nibble (bits 4-7) + Version (bits 0-3)
//	$0F     1     Region code (bits 4-7) + ROM size code (bits 0-3)
//
// Region codes (upper nibble of byte $0F):
//
//	3 = Japan (SMS)
//	4 = Export (SMS)
//	5 = Japan (GG)
//	6 = Export (GG)
//	7 = International (GG)

const (
	smsHeaderOffset     = 0x7FF0
	smsHeaderSize       = 16
	smsMinROMSize       = smsHeaderOffset + smsHeaderSize
	smsMagicOffset      = 0x00
	smsMagicSize        = 8
	smsChecksumOffset   = 0x0A
	smsProductLowOffset = 0x0C
	smsProductVerOffset = 0x0E
	smsRegionSizeOffset = 0x0F
)

var smsMagic = []byte("TMR SEGA")

// Region represents the region code from the SMS/GG header.
type Region byte

// Region values
const (
	RegionJapanSMS  Region = 0x3
	RegionExportSMS Region = 0x4
	RegionJapanGG   Region = 0x5
	RegionExportGG  Region = 0x6
	RegionIntlGG    Region = 0x7
)

// ROMSize represents the ROM size code from the SMS/GG header.
type ROMSize byte

// ROMSize values
const (
	ROMSize8KB   ROMSize = 0xA
	ROMSize16KB  ROMSize = 0xB
	ROMSize32KB  ROMSize = 0xC
	ROMSize48KB  ROMSize = 0xD
	ROMSize64KB  ROMSize = 0xE
	ROMSize128KB ROMSize = 0xF
	ROMSize256KB ROMSize = 0x0
	ROMSize512KB ROMSize = 0x1
	ROMSize1MB   ROMSize = 0x2
)

// Info contains metadata extracted from a Master System or Game Gear ROM file.
type Info struct {
	// ProductCode is the BCD-decoded product code (e.g., "7670").
	ProductCode string `json:"product_code,omitempty"`
	// Version is the version number (0-15).
	Version int `json:"version"`
	// Region is the region code indicating platform and region.
	Region Region `json:"region"`
	// ROMSize is the ROM size code.
	ROMSize ROMSize `json:"rom_size"`
	// Checksum is the ROM checksum (little-endian).
	Checksum uint16 `json:"checksum"`
	// platform is the detected platform (SMS or Game Gear) based on region code (internal, used by GamePlatform).
	platform core.Platform
}

// GamePlatform implements core.GameInfo.
func (i *Info) GamePlatform() core.Platform { return i.platform }

// GameTitle implements core.GameInfo. SMS/GG ROMs don't have embedded titles.
func (i *Info) GameTitle() string { return "" }

// GameSerial implements core.GameInfo.
func (i *Info) GameSerial() string { return i.ProductCode }

// GameRegions implements core.GameInfo.
func (i *Info) GameRegions() []core.Region {
	switch i.Region {
	case RegionJapanSMS, RegionJapanGG:
		return []core.Region{core.RegionJapan}
	case RegionExportSMS, RegionExportGG:
		return []core.Region{core.RegionWorld}
	case RegionIntlGG:
		return []core.Region{core.RegionWorld}
	default:
		return []core.Region{}
	}
}

// Parse extracts game information from a Master System or Game Gear ROM file.
func Parse(r io.ReaderAt, size int64) (*Info, error) {
	if size < smsMinROMSize {
		return nil, fmt.Errorf("file too small for SMS/GG header: %d bytes (need at least %d)", size, smsMinROMSize)
	}

	// Read header at 0x7FF0
	header := make([]byte, smsHeaderSize)
	if _, err := r.ReadAt(header, smsHeaderOffset); err != nil {
		return nil, fmt.Errorf("failed to read SMS/GG header: %w", err)
	}

	// Verify magic bytes
	if !bytes.Equal(header[smsMagicOffset:smsMagicOffset+smsMagicSize], smsMagic) {
		return nil, fmt.Errorf("not a valid SMS/GG ROM: invalid magic bytes")
	}

	// Extract checksum (little-endian)
	checksum := binary.LittleEndian.Uint16(header[smsChecksumOffset:])

	// Extract product code (BCD encoded in bytes 0x0C-0x0D, high nibble in 0x0E)
	productCode := decodeBCDProductCode(
		header[smsProductLowOffset],
		header[smsProductLowOffset+1],
		header[smsProductVerOffset]>>4,
	)

	// Extract version (lower nibble of byte 0x0E)
	version := int(header[smsProductVerOffset] & 0x0F)

	// Extract region code (upper nibble of byte 0x0F)
	region := Region(header[smsRegionSizeOffset] >> 4)

	// Extract ROM size code (lower nibble of byte 0x0F)
	romSize := ROMSize(header[smsRegionSizeOffset] & 0x0F)

	// Determine platform from region code
	platform := determinePlatform(region)

	return &Info{
		ProductCode: productCode,
		Version:     version,
		Region:      region,
		ROMSize:     romSize,
		Checksum:    checksum,
		platform:    platform,
	}, nil
}

// decodeBCDProductCode decodes the BCD-encoded product code.
// The product code is stored as BCD in 2.5 bytes (low byte, high byte, and high nibble).
func decodeBCDProductCode(low, high, extra byte) string {
	// Format as hex digits (BCD uses hex representation of decimal digits)
	if extra == 0 {
		return fmt.Sprintf("%02X%02X", low, high)
	}
	return fmt.Sprintf("%X%02X%02X", extra, low, high)
}

// determinePlatform returns the platform based on the region code.
func determinePlatform(region Region) core.Platform {
	switch region {
	case RegionJapanSMS, RegionExportSMS:
		return core.PlatformMS
	case RegionJapanGG, RegionExportGG, RegionIntlGG:
		return core.PlatformGameGear
	default:
		// Unknown region code - default to Master System
		return core.PlatformMS
	}
}
