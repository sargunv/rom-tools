package sms

import (
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
//	$0A     2     Checksum (big-endian)
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
	smsReservedOffset   = 0x08
	smsChecksumOffset   = 0x0A
	smsProductLowOffset = 0x0C
	smsProductVerOffset = 0x0E
	smsRegionSizeOffset = 0x0F
)

var smsMagic = []byte("TMR SEGA")

// Region code constants
const (
	regionJapanSMS  = 0x3
	regionExportSMS = 0x4
	regionJapanGG   = 0x5
	regionExportGG  = 0x6
	regionIntlGG    = 0x7
)

// SMSInfo contains metadata extracted from a Master System or Game Gear ROM file.
type SMSInfo struct {
	// ProductCode is the BCD-decoded product code (e.g., "7670").
	ProductCode string
	// Version is the version number (0-15).
	Version int
	// RegionCode is the raw region code (3-7).
	RegionCode byte
	// ROMSizeCode is the ROM size code (0-C).
	ROMSizeCode byte
	// Checksum is the ROM checksum.
	Checksum uint16
	// Reserved contains the reserved bytes.
	Reserved [2]byte
	// Platform is the detected platform (SMS or Game Gear) based on region code.
	Platform core.Platform
}

// ParseSMS extracts game information from a Master System or Game Gear ROM file.
func ParseSMS(r io.ReaderAt, size int64) (*SMSInfo, error) {
	if size < smsMinROMSize {
		return nil, fmt.Errorf("file too small for SMS/GG header: %d bytes (need at least %d)", size, smsMinROMSize)
	}

	// Read header at 0x7FF0
	header := make([]byte, smsHeaderSize)
	if _, err := r.ReadAt(header, smsHeaderOffset); err != nil {
		return nil, fmt.Errorf("failed to read SMS/GG header: %w", err)
	}

	// Verify magic bytes
	for i := 0; i < smsMagicSize; i++ {
		if header[smsMagicOffset+i] != smsMagic[i] {
			return nil, fmt.Errorf("not a valid SMS/GG ROM: invalid magic bytes")
		}
	}

	// Extract reserved bytes
	var reserved [2]byte
	reserved[0] = header[smsReservedOffset]
	reserved[1] = header[smsReservedOffset+1]

	// Extract checksum (big-endian)
	checksum := uint16(header[smsChecksumOffset])<<8 | uint16(header[smsChecksumOffset+1])

	// Extract product code (BCD encoded in bytes 0x0C-0x0D, high nibble in 0x0E)
	productCode := decodeBCDProductCode(
		header[smsProductLowOffset],
		header[smsProductLowOffset+1],
		header[smsProductVerOffset]>>4,
	)

	// Extract version (lower nibble of byte 0x0E)
	version := int(header[smsProductVerOffset] & 0x0F)

	// Extract region code (upper nibble of byte 0x0F)
	regionCode := header[smsRegionSizeOffset] >> 4

	// Extract ROM size code (lower nibble of byte 0x0F)
	romSizeCode := header[smsRegionSizeOffset] & 0x0F

	// Determine platform from region code
	platform := determinePlatform(regionCode)

	return &SMSInfo{
		ProductCode: productCode,
		Version:     version,
		RegionCode:  regionCode,
		ROMSizeCode: romSizeCode,
		Checksum:    checksum,
		Reserved:    reserved,
		Platform:    platform,
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
func determinePlatform(regionCode byte) core.Platform {
	switch regionCode {
	case regionJapanSMS, regionExportSMS:
		return core.PlatformMS
	case regionJapanGG, regionExportGG, regionIntlGG:
		return core.PlatformGameGear
	default:
		// Unknown region code - default to Master System
		return core.PlatformMS
	}
}
