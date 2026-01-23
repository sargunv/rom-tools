package snes

import (
	"fmt"
	"io"
)

// SNES ROM format parsing.
//
// SNES ROM header specification:
// https://sneslab.net/wiki/SNES_ROM_Header
// https://snes.nesdev.org/wiki/ROM_header
//
// The internal header is located at different offsets depending on mapping mode:
// - LoROM: 0x7FC0
// - HiROM: 0xFFC0
// - ExHiROM: 0x40FFC0
//
// Extended header layout (16 bytes at FFB0-FFBF, relative -0x10 from standard header):
//
//	Offset  Size  Description
//	FFB0    2     Maker Code (ASCII, when old maker code is $33)
//	FFB2    4     Game Code (ASCII)
//	FFB6    6     Reserved
//	FFBC    1     Reserved
//	FFBD    1     Expansion RAM Size (log2 of kilobytes)
//	FFBE    1     Special Version
//	FFBF    1     Cartridge Sub-Type
//
// Standard header layout (32 bytes at FFC0-FFDF):
//
//	Offset  Size  Description
//	FFC0    21    Game title (ASCII, space-padded)
//	FFD5    1     Map mode
//	FFD6    1     Cartridge type (chipset info)
//	FFD7    1     ROM size (log2 of kilobytes)
//	FFD8    1     RAM size (log2 of kilobytes)
//	FFD9    1     Destination code (region)
//	FFDA    1     Old maker code (0x33 = use extended header maker code)
//	FFDB    1     Mask ROM version
//	FFDC    2     Checksum complement
//	FFDE    2     Checksum

const (
	snesHeaderSize      = 32
	snesTitleOffset     = 0x00
	snesTitleLen        = 21
	snesMapModeOffset   = 0x15
	snesCartTypeOffset  = 0x16
	snesROMSizeOffset   = 0x17
	snesRAMSizeOffset   = 0x18
	snesDestCodeOffset  = 0x19
	snesMakerOldOffset  = 0x1A
	snesVersionOffset   = 0x1B
	snesChecksumCOffset = 0x1C
	snesChecksumOffset  = 0x1E

	// Extended header offsets (relative to standard header start, i.e., -0x10)
	snesMakerCodeOffset      = -0x10 // FFB0, 2 bytes
	snesGameCodeOffset       = -0x0E // FFB2, 4 bytes
	snesExpansionRAMOffset   = -0x03 // FFBD, 1 byte
	snesSpecialVersionOffset = -0x02 // FFBE, 1 byte
	snesCartSubTypeOffset    = -0x01 // FFBF, 1 byte

	// Header offsets for different mapping modes (without copier header)
	snesLoROMOffset   = 0x7FC0
	snesHiROMOffset   = 0xFFC0
	snesExHiROMOffset = 0x40FFC0

	// Copier header size (some ROMs have this prepended)
	snesCopierHeaderSize = 512
)

// SNESMapMode indicates the memory mapping mode.
type SNESMapMode byte

// SNESMapMode values per sneslab wiki.
const (
	SNESMapModeLoROM          SNESMapMode = 0x20 // 2.68MHz LoROM
	SNESMapModeHiROM          SNESMapMode = 0x21 // 2.68MHz HiROM
	SNESMapModeSA1            SNESMapMode = 0x23 // SA-1
	SNESMapModeExHiROM        SNESMapMode = 0x25 // 2.68MHz ExHiROM
	SNESMapModeFastROMLoROM   SNESMapMode = 0x30 // 3.58MHz LoROM
	SNESMapModeFastROMHiROM   SNESMapMode = 0x31 // 3.58MHz HiROM
	SNESMapModeFastROMExHiROM SNESMapMode = 0x35 // 3.58MHz ExHiROM
	SNESMapModeSPC7110        SNESMapMode = 0x3A // SPC7110 variant (Tengai Makyou Zero)
)

// SNESDestination indicates the destination/region code.
type SNESDestination byte

// SNESDestination values per SNES Development Manual.
const (
	SNESDestinationJapan       SNESDestination = 0x00
	SNESDestinationUSA         SNESDestination = 0x01
	SNESDestinationEurope      SNESDestination = 0x02
	SNESDestinationScandinavia SNESDestination = 0x03
	SNESDestinationFrench      SNESDestination = 0x06
	SNESDestinationDutch       SNESDestination = 0x07
	SNESDestinationSpanish     SNESDestination = 0x08
	SNESDestinationGerman      SNESDestination = 0x09
	SNESDestinationItalian     SNESDestination = 0x0A
	SNESDestinationChinese     SNESDestination = 0x0B
	SNESDestinationKorean      SNESDestination = 0x0D
	SNESDestinationCommon      SNESDestination = 0x0E
	SNESDestinationCanada      SNESDestination = 0x0F
	SNESDestinationBrazil      SNESDestination = 0x10
	SNESDestinationAustralia   SNESDestination = 0x11
)

// SNESCartridgeType indicates the cartridge chipset type.
type SNESCartridgeType byte

// SNESCartridgeType common values.
const (
	SNESCartridgeROMOnly       SNESCartridgeType = 0x00
	SNESCartridgeROMRAM        SNESCartridgeType = 0x01
	SNESCartridgeROMRAMBattery SNESCartridgeType = 0x02
	SNESCartridgeSA1           SNESCartridgeType = 0x33
	SNESCartridgeSA1RAM        SNESCartridgeType = 0x34
	SNESCartridgeSA1RAMBattery SNESCartridgeType = 0x35
)

// SNESInfo contains metadata extracted from a SNES ROM file.
type SNESInfo struct {
	// Extended header (FFB0-FFBF) - may not be present in older ROMs
	// MakerCode is the 2-char maker code (FFB0), present when MakerCodeOld is 0x33.
	MakerCode string
	// GameCode is the 4-char game code (FFB2).
	GameCode string
	// ExpansionRAMSize is the expansion RAM size in bytes (FFBD).
	ExpansionRAMSize int
	// SpecialVersion is the special version byte (FFBE).
	SpecialVersion byte
	// CartridgeSubType is the cartridge sub-type (FFBF).
	CartridgeSubType byte

	// Standard header (FFC0-FFDF)
	// Title is the game title (21 chars max, space-padded).
	Title string
	// MapMode is the memory mapping mode (FFD5).
	MapMode SNESMapMode
	// CartridgeType is the chipset info (FFD6).
	CartridgeType SNESCartridgeType
	// ROMSize is the ROM size in bytes (FFD7).
	ROMSize int
	// RAMSize is the RAM/SRAM size in bytes (FFD8).
	RAMSize int
	// Destination is the region code (FFD9).
	Destination SNESDestination
	// MakerCodeOld is the old maker code (FFDA) - 0x33 means use MakerCode.
	MakerCodeOld byte
	// MaskROMVersion is the ROM version number (FFDB).
	MaskROMVersion int
	// ComplementCheck is the checksum complement (FFDC).
	ComplementCheck uint16
	// Checksum is the ROM checksum (FFDE).
	Checksum uint16
	// HasCopierHeader is true if 512-byte copier header detected.
	HasCopierHeader bool
}

// ParseSNES extracts information from a SNES ROM file.
func ParseSNES(r io.ReaderAt, size int64) (*SNESInfo, error) {
	// Determine if there's a copier header (file size % 1024 == 512)
	hasCopierHeader := (size % 1024) == snesCopierHeaderSize
	copierOffset := int64(0)
	if hasCopierHeader {
		copierOffset = snesCopierHeaderSize
	}

	// Calculate all three possible header offsets
	offsets := []int64{
		copierOffset + snesLoROMOffset,   // LoROM
		copierOffset + snesHiROMOffset,   // HiROM
		copierOffset + snesExHiROMOffset, // ExHiROM
	}

	// Try each offset and return the first valid header
	for _, offset := range offsets {
		if offset+snesHeaderSize <= size {
			info, err := parseSNESHeader(r, offset, size, hasCopierHeader)
			if err == nil && isValidSNESHeader(info, size) {
				return info, nil
			}
		}
	}

	return nil, fmt.Errorf("could not find valid SNES header")
}

func parseSNESHeader(r io.ReaderAt, offset int64, fileSize int64, hasCopierHeader bool) (*SNESInfo, error) {
	header := make([]byte, snesHeaderSize)
	if _, err := r.ReadAt(header, offset); err != nil {
		return nil, fmt.Errorf("failed to read SNES header: %w", err)
	}

	// Extract raw title bytes for validation
	titleBytes := header[snesTitleOffset : snesTitleOffset+snesTitleLen]

	// Extract title (ASCII, space-padded)
	title := extractSNESTitle(titleBytes)

	// Map mode
	mapMode := SNESMapMode(header[snesMapModeOffset])

	// Cartridge type
	cartType := SNESCartridgeType(header[snesCartTypeOffset])

	// ROM size: 1 << (header value) kilobytes
	romSizeExp := header[snesROMSizeOffset]
	romSize := 0
	if romSizeExp > 0 && romSizeExp < 16 {
		romSize = (1 << romSizeExp) * 1024
	}

	// RAM size: 1 << (header value) kilobytes (0 = no RAM)
	ramSizeExp := header[snesRAMSizeOffset]
	ramSize := 0
	if ramSizeExp > 0 && ramSizeExp < 16 {
		ramSize = (1 << ramSizeExp) * 1024
	}

	// Destination code
	destination := SNESDestination(header[snesDestCodeOffset])

	// Old maker code
	makerCodeOld := header[snesMakerOldOffset]

	// Mask ROM version
	maskROMVersion := int(header[snesVersionOffset])

	// Checksum complement and checksum (little-endian)
	complementCheck := uint16(header[snesChecksumCOffset]) | uint16(header[snesChecksumCOffset+1])<<8
	checksum := uint16(header[snesChecksumOffset]) | uint16(header[snesChecksumOffset+1])<<8

	// Extended header fields (only read if offset allows and maker code is 0x33)
	var makerCode, gameCode string
	var expansionRAMSize int
	var specialVersion, cartSubType byte

	extOffset := offset + int64(snesMakerCodeOffset)
	if makerCodeOld == 0x33 && extOffset >= 0 {
		extHeader := make([]byte, 16)
		if _, err := r.ReadAt(extHeader, extOffset); err == nil {
			// Maker code (2 bytes at offset 0 = FFB0)
			makerCode = extractSNESTitle(extHeader[0:2])
			// Game code (4 bytes at offset 2 = FFB2)
			gameCode = extractSNESTitle(extHeader[2:6])
			// Expansion RAM size (byte at offset 13 = FFBD)
			expRAMExp := extHeader[13]
			if expRAMExp > 0 && expRAMExp < 16 {
				expansionRAMSize = (1 << expRAMExp) * 1024
			}
			// Special version (byte at offset 14 = FFBE)
			specialVersion = extHeader[14]
			// Cartridge sub-type (byte at offset 15 = FFBF)
			cartSubType = extHeader[15]
		}
	}

	return &SNESInfo{
		// Extended header
		MakerCode:        makerCode,
		GameCode:         gameCode,
		ExpansionRAMSize: expansionRAMSize,
		SpecialVersion:   specialVersion,
		CartridgeSubType: cartSubType,
		// Standard header
		Title:           title,
		MapMode:         mapMode,
		CartridgeType:   cartType,
		ROMSize:         romSize,
		RAMSize:         ramSize,
		Destination:     destination,
		MakerCodeOld:    makerCodeOld,
		MaskROMVersion:  maskROMVersion,
		ComplementCheck: complementCheck,
		Checksum:        checksum,
		HasCopierHeader: hasCopierHeader,
	}, nil
}

// isValidSNESHeader checks if the header looks valid using multiple heuristics.
func isValidSNESHeader(info *SNESInfo, fileSize int64) bool {
	// 1. Checksum validation: checksum + complement should equal 0xFFFF
	//    This is the strongest signal (1 in 65,536 chance of random data passing)
	if info.Checksum+info.ComplementCheck != 0xFFFF {
		return false
	}

	// 2. Map mode should be in valid range (0x20-0x3F)
	//    All known map modes have high nibble 0x2 or 0x3
	if info.MapMode < 0x20 || info.MapMode > 0x3F {
		// Check for known games with header bugs before rejecting
		if !isKnownHeaderBug(info) {
			return false
		}
	}

	// 3. Title should have at least some printable ASCII characters
	//    Random code/data rarely looks like text
	printableCount := 0
	for _, c := range info.Title {
		if c >= 0x20 && c <= 0x7E {
			printableCount++
		}
	}
	if printableCount < 2 {
		return false
	}

	// 4. Declared ROM size should be reasonable compared to file size
	//    Allow declared size to be up to 2x file size (some headers are inaccurate)
	if info.ROMSize > 0 && int64(info.ROMSize) > fileSize*2 {
		return false
	}

	return true
}

// isKnownHeaderBug checks if the ROM matches a known game with a header manufacturing defect.
// These are original cartridge bugs, not ROM corruption.
func isKnownHeaderBug(info *SNESInfo) bool {
	// Contra III: The Alien Wars (USA)
	// The title "CONTRA3 THE ALIEN WARS" is 22 chars but the header only has 21 bytes,
	// causing the 'S' (0x53) to overflow into the map mode byte at 0x7FD5.
	// The Japanese version has the correct header (0x20 = LoROM + SlowROM).
	// Reference: https://datacrystal.tcrf.net/wiki/Contra_III:_The_Alien_Wars
	if info.Title == "CONTRA3 THE ALIEN WAR" &&
		info.MapMode == 0x53 && // 'S' from title overflow
		info.Checksum == 0x0C3C &&
		info.ComplementCheck == 0xF3C3 {
		return true
	}

	return false
}

// extractSNESTitle extracts and cleans a SNES title string.
func extractSNESTitle(data []byte) string {
	// Find end of valid ASCII characters
	end := len(data)
	for i, b := range data {
		// SNES titles should be printable ASCII (0x20-0x7E) or space-padded
		if b == 0 || b < 0x20 || b > 0x7E {
			end = i
			break
		}
	}

	// Trim trailing spaces
	for end > 0 && data[end-1] == ' ' {
		end--
	}

	return string(data[:end])
}
