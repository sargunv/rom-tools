package gamecube

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
)

// GameCube/Wii disc header format parsing.
//
// Both GameCube and Wii use the same disc header format.
// Per WiiBrew wiki: "The first 0x400 bytes are like the GameCube disc header format."
// https://wiibrew.org/wiki/Wii_disc
//
// Disc header layout (relevant fields):
//
//	Offset  Size  Description
//	0x000   1     Disc ID (G=GameCube, R/S=Wii, D=GC Demo)
//	0x001   2     Game code
//	0x003   1     Region code (E=USA, P=PAL, J=Japan, etc.)
//	0x004   2     Maker code
//	0x006   1     Disc number
//	0x007   1     Disc version
//	0x018   4     Wii magic word (0x5D1C9EA3 for Wii, 0x00000000 for GameCube)
//	0x01C   4     GameCube magic word (0xC2339F3D for GameCube, 0x00000000 for Wii)
//	0x020   64    Game title (ASCII, null-terminated)

const (
	discHeaderSize = 0x60 // We only need first 96 bytes for identification

	discIDOffset = 0x000
	gameCodeOffset    = 0x001
	gameCodeLen       = 2
	regionCodeOffset  = 0x003
	makerCodeOffset   = 0x004
	makerCodeLen      = 2
	discNumberOffset  = 0x006
	discVersionOffset = 0x007
	wiiMagicOffset    = 0x018
	gcMagicOffset     = 0x01C
	titleOffset       = 0x020
	titleLen          = 64

	wiiMagicWord = 0x5D1C9EA3
	gcMagicWord  = 0xC2339F3D
)

// GCMInfo contains metadata extracted from a GameCube/Wii disc header.
type GCMInfo struct {
	// DiscID is the console identifier (G=GameCube, R/S=Wii, D=GC Demo).
	DiscID byte
	// GameCode is the 2-character unique game identifier.
	GameCode string
	// RegionCode is the region code (E=USA, P=PAL, J=Japan, etc.).
	RegionCode byte
	// MakerCode is the 2-character publisher identifier.
	MakerCode string
	// DiscNumber is the disc number for multi-disc games.
	DiscNumber int
	// Version is the disc version/revision.
	Version int
	// Title is the game title.
	Title string
	// Platform is the target platform (GameCube or Wii).
	Platform core.Platform
}

// ParseGCM parses a GameCube/Wii disc header from a reader.
func ParseGCM(r io.ReaderAt, size int64) (*GCMInfo, error) {
	if size < discHeaderSize {
		return nil, fmt.Errorf("file too small for disc header: need %d bytes, got %d", discHeaderSize, size)
	}

	header := make([]byte, discHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read disc header: %w", err)
	}

	return parseGCMBytes(header)
}

func parseGCMBytes(header []byte) (*GCMInfo, error) {
	// Check magic words to determine platform and validate
	wiiMagic := binary.BigEndian.Uint32(header[wiiMagicOffset:])
	gcMagic := binary.BigEndian.Uint32(header[gcMagicOffset:])

	isWii := wiiMagic == wiiMagicWord
	isGC := gcMagic == gcMagicWord

	if !isWii && !isGC {
		return nil, fmt.Errorf("not a valid GameCube/Wii disc: no magic word found (Wii: 0x%08X, GC: 0x%08X)",
			wiiMagic, gcMagic)
	}

	// Determine platform
	var platform core.Platform
	if isWii {
		platform = core.PlatformWii
	} else {
		platform = core.PlatformGC
	}

	// Extract fields
	discID := header[discIDOffset]
	gameCode := util.ExtractASCII(header[gameCodeOffset : gameCodeOffset+gameCodeLen])
	regionCode := header[regionCodeOffset]
	makerCode := util.ExtractASCII(header[makerCodeOffset : makerCodeOffset+makerCodeLen])
	discNumber := int(header[discNumberOffset])
	version := int(header[discVersionOffset])
	title := util.ExtractASCII(header[titleOffset : titleOffset+titleLen])

	return &GCMInfo{
		DiscID:     discID,
		GameCode:   gameCode,
		RegionCode: regionCode,
		MakerCode:  makerCode,
		DiscNumber: discNumber,
		Version:    version,
		Title:      title,
		Platform:   platform,
	}, nil
}
