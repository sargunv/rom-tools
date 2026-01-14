package z64

import (
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/romident/core"
)

// N64 ROM format parsing (Z64 - big-endian native format).
//
// N64 ROM header specification:
// https://n64brew.dev/wiki/ROM_Header
//
// N64 ROMs come in three byte orderings:
//   - Big Endian (.z64): 0x80 at position 0 - native N64 format (this package)
//   - Byte-Swapped (.v64): 0x80 at position 1 - pairs of bytes swapped
//   - Little Endian (.n64): 0x80 at position 3 - 32-bit words reversed
//
// N64 header layout (in native big-endian, relevant fields):
//
//	Offset  Size  Description
//	0x00    1     Reserved (0x80 for all known commercial games)
//	0x01    3     PI BSD DOM1 Configuration Flags
//	0x04    4     Clock Rate
//	0x08    4     Boot Address (entry point)
//	0x0C    4     Libultra Version
//	0x10    8     Check Code (64-bit checksum)
//	0x18    8     Reserved
//	0x20    20    Game Title (ASCII, space-padded)
//	0x34    7     Reserved (homebrew header uses some of this)
//	0x3B    4     Game Code (category, unique code, destination)
//	0x3F    1     ROM Version

const (
	N64HeaderSize     = 0x40 // 64 bytes
	N64ReservedByte   = 0x80
	n64TitleOffset    = 0x20
	n64TitleLen       = 20
	n64GameCodeOffset = 0x3B
	n64GameCodeLen    = 4
	n64VersionOffset  = 0x3F
)

// N64ByteOrder represents the byte ordering of an N64 ROM.
type N64ByteOrder string

const (
	N64BigEndian    N64ByteOrder = "z64" // Native format, 0x80 at position 0
	N64ByteSwapped  N64ByteOrder = "v64" // Byte-swapped pairs, 0x80 at position 1
	N64LittleEndian N64ByteOrder = "n64" // Word-swapped, 0x80 at position 3
	N64Unknown      N64ByteOrder = "unknown"
)

// N64Info contains metadata extracted from an N64 ROM file.
type N64Info struct {
	Title        string
	GameCode     string // Full 4-character game code
	CategoryCode byte   // 1st char: N=Game Pak, D=64DD, C=Expandable, etc.
	UniqueCode   string // 2nd-3rd chars: unique game identifier
	RegionCode   byte   // 4th char: J=Japan, E=USA, P=Europe, etc.
	Version      int
	ByteOrder    N64ByteOrder // Detected byte ordering
}

// ParseN64Header parses an N64 header from big-endian (z64) format bytes.
// The header must already be in native big-endian format.
func ParseN64Header(header []byte, byteOrder N64ByteOrder) (*N64Info, error) {
	if len(header) < N64HeaderSize {
		return nil, fmt.Errorf("header too small for N64: need %d bytes, got %d", N64HeaderSize, len(header))
	}

	// Extract title (space-padded ASCII)
	title := util.ExtractASCII(header[n64TitleOffset : n64TitleOffset+n64TitleLen])

	// Extract game code
	gameCode := util.ExtractASCII(header[n64GameCodeOffset : n64GameCodeOffset+n64GameCodeLen])

	// Parse game code components
	var categoryCode byte
	var uniqueCode string
	var regionCode byte
	if len(gameCode) >= 4 {
		categoryCode = gameCode[0]
		uniqueCode = gameCode[1:3]
		regionCode = gameCode[3]
	}

	// Extract ROM version
	version := int(header[n64VersionOffset])

	return &N64Info{
		Title:        title,
		GameCode:     gameCode,
		CategoryCode: categoryCode,
		UniqueCode:   uniqueCode,
		RegionCode:   regionCode,
		Version:      version,
		ByteOrder:    byteOrder,
	}, nil
}

// DetectByteOrder determines the byte ordering by finding where 0x80 is located.
func DetectByteOrder(first4 []byte) N64ByteOrder {
	switch {
	case first4[0] == N64ReservedByte:
		return N64BigEndian
	case first4[1] == N64ReservedByte:
		return N64ByteSwapped
	case first4[3] == N64ReservedByte:
		return N64LittleEndian
	default:
		return N64Unknown
	}
}

// Identify verifies the format and extracts game identification from a Z64 ROM.
func Identify(r io.ReaderAt, size int64) (*core.GameIdent, error) {
	if size < N64HeaderSize {
		return nil, fmt.Errorf("file too small for N64 header: %d bytes", size)
	}

	// Read first 4 bytes to check byte order
	first4 := make([]byte, 4)
	if _, err := r.ReadAt(first4, 0); err != nil {
		return nil, fmt.Errorf("failed to read N64 header: %w", err)
	}

	actualOrder := DetectByteOrder(first4)
	if actualOrder != N64BigEndian {
		return nil, fmt.Errorf("byte order mismatch: expected z64, got %s", actualOrder)
	}

	// Read full header
	header := make([]byte, N64HeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read N64 header: %w", err)
	}

	info, err := ParseN64Header(header, N64BigEndian)
	if err != nil {
		return nil, err
	}

	return N64InfoToGameIdent(info), nil
}

// N64InfoToGameIdent converts N64Info to GameIdent.
func N64InfoToGameIdent(info *N64Info) *core.GameIdent {
	version := info.Version

	return &core.GameIdent{
		Platform: core.PlatformN64,
		TitleID:  info.GameCode,
		Title:    info.Title,
		Version:  &version,
		Extra:    info,
	}
}

// SwapBytes16 converts v64 (byte-swapped) to z64 (big-endian) format in place.
// Each pair of bytes is swapped: AB CD -> BA DC
func SwapBytes16(data []byte) {
	for i := 0; i < len(data)-1; i += 2 {
		data[i], data[i+1] = data[i+1], data[i]
	}
}

// SwapBytes32 converts n64 (word-swapped/little-endian) to z64 (big-endian) format in place.
// Each 4-byte word is reversed: ABCD -> DCBA
func SwapBytes32(data []byte) {
	for i := 0; i < len(data)-3; i += 4 {
		data[i], data[i+1], data[i+2], data[i+3] = data[i+3], data[i+2], data[i+1], data[i]
	}
}
