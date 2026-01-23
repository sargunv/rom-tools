package gba

import (
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
)

// GBA (Game Boy Advance) ROM format parsing.
//
// GBA ROM header specification:
// https://problemkaputt.de/gbatek-gba-cartridge-header.htm
//
// GBA header layout:
//
//	Offset  Size  Description
//	0x00    4     ROM Entry Point (32bit ARM branch opcode)
//	0x04    156   Nintendo Logo (compressed bitmap, required)
//	0xA0    12    Game Title (uppercase ASCII, max 12 characters)
//	0xAC    4     Game Code (uppercase ASCII, 4 characters)
//	0xB0    2     Maker Code (uppercase ASCII, 2 characters)
//	0xB2    1     Fixed value (must be 0x96, required)
//	0xB3    1     Main unit code (0x00 for GBA)
//	0xB4    1     Device type (bit 7 = debug DACS)
//	0xB5    7     Reserved (should be zero)
//	0xBC    1     Software version
//	0xBD    1     Header checksum (complement check)
//	0xBE    2     Reserved (should be zero)
//
// Game Code breakdown (4 bytes at 0xAC):
//   - Byte 0: Game Type - hardware/cartridge type indicator
//   - Bytes 1-2: Unique Code - 2-character game identifier
//   - Byte 3: Destination - target region code

const (
	gbaHeaderSize       = 0xC0 // 192 bytes
	gbaTitleOffset      = 0xA0
	gbaTitleLen         = 12
	gbaGameCodeOffset   = 0xAC
	gbaGameCodeLen      = 4
	gbaMakerOffset      = 0xB0
	gbaMakerLen         = 2
	gbaFixedOffset      = 0xB2
	gbaFixedValue       = 0x96
	gbaMainUnitOffset   = 0xB3
	gbaDeviceTypeOffset = 0xB4
	gbaVersionOffset    = 0xBC
	gbaChecksumOffset   = 0xBD
)

// GBAGameType represents the cartridge/hardware type from the first byte of the game code.
type GBAGameType byte

// GBAGameType values indicate cartridge features and hardware generation.
const (
	GBAGameTypeNormalOld    GBAGameType = 'A' // Normal game (2001-2003)
	GBAGameTypeNormalNew    GBAGameType = 'B' // Normal game (2003+)
	GBAGameTypeNormalUnused GBAGameType = 'C' // Normal game (unused)
	GBAGameTypeFamicom      GBAGameType = 'F' // Classic NES Series (Famicom Mini)
	GBAGameTypeAcceleration GBAGameType = 'K' // Acceleration sensor (tilt)
	GBAGameTypeEReader      GBAGameType = 'P' // e-Reader
	GBAGameTypeRumbleGyro   GBAGameType = 'R' // Rumble + gyro sensor (WarioWare Twisted)
	GBAGameTypeRTCSolar     GBAGameType = 'U' // RTC + solar sensor (Boktai series)
	GBAGameTypeRumble       GBAGameType = 'V' // Rumble only (Drill Dozer)
)

// GBADestination represents the target region from the fourth byte of the game code.
type GBADestination byte

// GBADestination values indicate the target region for the game.
const (
	GBADestinationJapan   GBADestination = 'J'
	GBADestinationUSA     GBADestination = 'E'
	GBADestinationEurope  GBADestination = 'P'
	GBADestinationFrance  GBADestination = 'F'
	GBADestinationSpain   GBADestination = 'S'
	GBADestinationGermany GBADestination = 'D'
	GBADestinationItaly   GBADestination = 'I'
)

// GBAInfo contains metadata extracted from a GBA ROM file.
type GBAInfo struct {
	// Title is the game title (0xA0-0xAB, up to 12 uppercase ASCII characters).
	Title string
	// GameCode is the full 4-character game code (0xAC-0xAF).
	GameCode string
	// GameType is the cartridge/hardware type from byte 0 of GameCode.
	GameType GBAGameType
	// Destination is the target region from byte 3 of GameCode.
	Destination GBADestination
	// MakerCode is the 2-character manufacturer code (0xB0-0xB1).
	MakerCode string
	// MainUnitCode indicates the target hardware (0xB3, 0x00 for GBA).
	MainUnitCode byte
	// DeviceType indicates debug hardware (0xB4, bit 7 = debug DACS enabled).
	DeviceType byte
	// Version is the software version number (0xBC).
	Version int
	// HeaderChecksum is the complement check value (0xBD).
	HeaderChecksum byte
}

// ParseGBA extracts game information from a GBA ROM file.
func ParseGBA(r io.ReaderAt, size int64) (*GBAInfo, error) {
	if size < gbaHeaderSize {
		return nil, fmt.Errorf("file too small for GBA header: %d bytes", size)
	}

	header := make([]byte, gbaHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read GBA header: %w", err)
	}

	// Verify fixed value at 0xB2
	if header[gbaFixedOffset] != gbaFixedValue {
		return nil, fmt.Errorf("not a valid GBA ROM: invalid fixed byte (got 0x%02X, expected 0x%02X)",
			header[gbaFixedOffset], gbaFixedValue)
	}

	// Extract title (null-terminated ASCII)
	title := util.ExtractASCII(header[gbaTitleOffset : gbaTitleOffset+gbaTitleLen])

	// Extract game code
	gameCode := util.ExtractASCII(header[gbaGameCodeOffset : gbaGameCodeOffset+gbaGameCodeLen])

	// Parse game code components
	var gameType GBAGameType
	var destination GBADestination
	if len(gameCode) >= 4 {
		gameType = GBAGameType(gameCode[0])
		destination = GBADestination(gameCode[3])
	}

	// Extract maker code
	makerCode := util.ExtractASCII(header[gbaMakerOffset : gbaMakerOffset+gbaMakerLen])

	// Extract hardware fields
	mainUnitCode := header[gbaMainUnitOffset]
	deviceType := header[gbaDeviceTypeOffset]

	// Extract software version
	version := int(header[gbaVersionOffset])

	// Extract header checksum
	headerChecksum := header[gbaChecksumOffset]

	return &GBAInfo{
		Title:          title,
		GameCode:       gameCode,
		GameType:       gameType,
		Destination:    destination,
		MakerCode:      makerCode,
		MainUnitCode:   mainUnitCode,
		DeviceType:     deviceType,
		Version:        version,
		HeaderChecksum: headerChecksum,
	}, nil
}
