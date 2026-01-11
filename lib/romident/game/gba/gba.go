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
// GBA header layout (relevant fields):
//
//	Offset  Size  Description
//	0x00    4     ROM Entry Point (32bit ARM branch opcode)
//	0x04    156   Nintendo Logo (compressed bitmap, required)
//	0xA0    12    Game Title (uppercase ASCII, max 12 characters)
//	0xAC    4     Game Code (uppercase ASCII, 4 characters)
//	0xB0    2     Maker Code (uppercase ASCII, 2 characters)
//	0xB2    1     Fixed value (must be 0x96, required)
//	0xBC    1     Software version
//	0xBD    1     Complement check (header checksum)

const (
	gbaHeaderSize     = 0xC0 // 192 bytes
	gbaTitleOffset    = 0xA0
	gbaTitleLen       = 12
	gbaGameCodeOffset = 0xAC
	gbaGameCodeLen    = 4
	gbaMakerOffset    = 0xB0
	gbaMakerLen       = 2
	gbaFixedOffset    = 0xB2
	gbaFixedValue     = 0x96
	gbaVersionOffset  = 0xBC
)

// GBAInfo contains metadata extracted from a GBA ROM file.
type GBAInfo struct {
	Title      string
	GameCode   string
	MakerCode  string
	RegionCode byte // 4th character of game code (J, E, P, etc.)
	Version    int
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

	// Extract maker code
	makerCode := util.ExtractASCII(header[gbaMakerOffset : gbaMakerOffset+gbaMakerLen])

	// Extract region code (4th character of game code)
	var regionCode byte
	if len(gameCode) >= 4 {
		regionCode = gameCode[3]
	}

	// Extract software version
	version := int(header[gbaVersionOffset])

	return &GBAInfo{
		Title:      title,
		GameCode:   gameCode,
		MakerCode:  makerCode,
		RegionCode: regionCode,
		Version:    version,
	}, nil
}
