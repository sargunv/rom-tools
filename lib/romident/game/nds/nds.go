package nds

import (
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
)

// NDS (Nintendo DS) ROM format parsing.
//
// NDS/DSi cartridge header specification:
// https://dsibrew.org/wiki/DSi_cartridge_header
//
// Header layout (512 bytes):
//
//	Offset  Size  Description
//	0x000   12    Game Title (uppercase ASCII, null-padded)
//	0x00C   4     Game Code (e.g., "AMFE")
//	0x010   2     Maker Code (e.g., "01" for Nintendo)
//	0x012   1     Unit Code (0x00=DS, 0x02=DS+DSi, 0x03=DSi only)
//	0x013   1     Encryption seed select
//	0x014   1     Device capacity (ROM size = 2^(20+n) bytes)
//	0x01E   1     ROM Version
//	0x0C0   156   Nintendo Logo
//	0x15E   2     Header CRC

const (
	ndsHeaderSize      = 0x200 // 512 bytes
	ndsTitleOffset     = 0x000
	ndsTitleLen        = 12
	ndsGameCodeOffset  = 0x00C
	ndsGameCodeLen     = 4
	ndsMakerCodeOffset = 0x010
	ndsMakerCodeLen    = 2
	ndsUnitCodeOffset  = 0x012
	ndsVersionOffset   = 0x01E
	ndsARM9OffsetPos   = 0x020
	ndsARM7OffsetPos   = 0x030
)

// NDSUnitCode indicates the target platform for the ROM
type NDSUnitCode byte

const (
	NDSUnitCodeDS      NDSUnitCode = 0x00 // Original Nintendo DS
	NDSUnitCodeDSiDual NDSUnitCode = 0x02 // DS + DSi enhanced
	NDSUnitCodeDSi     NDSUnitCode = 0x03 // DSi only
)

// NDSPlatform indicates whether this is a DS, DSi-enhanced, or DSi-only game
type NDSPlatform string

const (
	NDSPlatformDS      NDSPlatform = "nds"     // Original Nintendo DS
	NDSPlatformDSiDual NDSPlatform = "nds+dsi" // DS with DSi enhancements
	NDSPlatformDSi     NDSPlatform = "dsi"     // DSi only
)

// NDSInfo contains metadata extracted from an NDS ROM file.
type NDSInfo struct {
	Title      string
	GameCode   string
	MakerCode  string
	RegionCode byte // 4th character of game code (J, E, P, etc.)
	Version    int
	UnitCode   NDSUnitCode
	Platform   NDSPlatform
}

// ParseNDS extracts game information from an NDS ROM file.
func ParseNDS(r io.ReaderAt, size int64) (*NDSInfo, error) {
	if size < ndsHeaderSize {
		return nil, fmt.Errorf("file too small for NDS header: %d bytes", size)
	}

	header := make([]byte, ndsHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read NDS header: %w", err)
	}

	// Extract title (null-terminated ASCII)
	title := util.ExtractASCII(header[ndsTitleOffset : ndsTitleOffset+ndsTitleLen])

	// Extract game code
	gameCode := util.ExtractASCII(header[ndsGameCodeOffset : ndsGameCodeOffset+ndsGameCodeLen])

	// Extract maker code
	makerCode := util.ExtractASCII(header[ndsMakerCodeOffset : ndsMakerCodeOffset+ndsMakerCodeLen])

	// Extract region code (4th character of game code)
	var regionCode byte
	if len(gameCode) >= 4 {
		regionCode = gameCode[3]
	}

	// Extract unit code
	unitCode := NDSUnitCode(header[ndsUnitCodeOffset])

	// Determine platform based on unit code
	var platform NDSPlatform
	switch unitCode {
	case NDSUnitCodeDSiDual:
		platform = NDSPlatformDSiDual
	case NDSUnitCodeDSi:
		platform = NDSPlatformDSi
	default:
		platform = NDSPlatformDS
	}

	// Extract software version
	version := int(header[ndsVersionOffset])

	return &NDSInfo{
		Title:      title,
		GameCode:   gameCode,
		MakerCode:  makerCode,
		RegionCode: regionCode,
		Version:    version,
		UnitCode:   unitCode,
		Platform:   platform,
	}, nil
}

// IsNDSROM checks if the data has a valid NDS ROM structure.
// We validate the ARM9 and ARM7 ROM offsets which must be present in all NDS ROMs.
func IsNDSROM(r io.ReaderAt, size int64) bool {
	if size < ndsHeaderSize {
		return false
	}

	header := make([]byte, ndsHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return false
	}

	// Read ARM9 ROM offset (little-endian u32 at 0x20)
	arm9Offset := uint32(header[ndsARM9OffsetPos]) |
		uint32(header[ndsARM9OffsetPos+1])<<8 |
		uint32(header[ndsARM9OffsetPos+2])<<16 |
		uint32(header[ndsARM9OffsetPos+3])<<24

	// Read ARM7 ROM offset (little-endian u32 at 0x30)
	arm7Offset := uint32(header[ndsARM7OffsetPos]) |
		uint32(header[ndsARM7OffsetPos+1])<<8 |
		uint32(header[ndsARM7OffsetPos+2])<<16 |
		uint32(header[ndsARM7OffsetPos+3])<<24

	// ARM9 offset must be at least 0x200 (after header) and word-aligned
	if arm9Offset < ndsHeaderSize || arm9Offset%4 != 0 {
		return false
	}

	// ARM7 offset must be at least 0x200 (after header) and word-aligned
	if arm7Offset < ndsHeaderSize || arm7Offset%4 != 0 {
		return false
	}

	// Both offsets must be within the file
	if uint32(size) < arm9Offset || uint32(size) < arm7Offset {
		return false
	}

	return true
}
