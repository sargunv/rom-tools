package n64

import (
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
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
	n64ReservedByte   = 0x80
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
	// Title is the game title (space-padded ASCII, up to 20 characters).
	Title string
	// GameCode is the full 4-character game code.
	GameCode string
	// CategoryCode is the 1st character: N=Game Pak, D=64DD, C=Expandable, etc.
	CategoryCode byte
	// UniqueCode is the 2nd-3rd characters: unique game identifier.
	UniqueCode string
	// RegionCode is the 4th character: J=Japan, E=USA, P=Europe, etc.
	RegionCode byte
	// Version is the ROM version number.
	Version int
	// ByteOrder is the detected byte ordering of the ROM.
	ByteOrder N64ByteOrder
}

// ParseN64 extracts game information from an N64 ROM file, auto-detecting byte order.
func ParseN64(r io.ReaderAt, size int64) (*N64Info, error) {
	if size < N64HeaderSize {
		return nil, fmt.Errorf("file too small for N64 header: %d bytes", size)
	}

	// Read first 4 bytes to detect byte order
	first4 := make([]byte, 4)
	if _, err := r.ReadAt(first4, 0); err != nil {
		return nil, fmt.Errorf("failed to read N64 header: %w", err)
	}

	byteOrder := detectByteOrder(first4)
	if byteOrder == N64Unknown {
		return nil, fmt.Errorf("not a valid N64 ROM: could not detect byte order")
	}

	// Read full header
	header := make([]byte, N64HeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read N64 header: %w", err)
	}

	// Convert to big-endian if needed
	switch byteOrder {
	case N64ByteSwapped:
		swapBytes16(header)
	case N64LittleEndian:
		swapBytes32(header)
	}

	return parseN64Header(header, byteOrder)
}

// parseN64Header parses an N64 header from big-endian (z64) format bytes.
func parseN64Header(header []byte, byteOrder N64ByteOrder) (*N64Info, error) {
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

// detectByteOrder determines the byte ordering by finding where 0x80 is located.
func detectByteOrder(first4 []byte) N64ByteOrder {
	switch {
	case first4[0] == n64ReservedByte:
		return N64BigEndian
	case first4[1] == n64ReservedByte:
		return N64ByteSwapped
	case first4[3] == n64ReservedByte:
		return N64LittleEndian
	default:
		return N64Unknown
	}
}

// swapBytes16 converts v64 (byte-swapped) to z64 (big-endian) format in place.
// Each pair of bytes is swapped: AB CD -> BA DC
func swapBytes16(data []byte) {
	for i := 0; i < len(data)-1; i += 2 {
		data[i], data[i+1] = data[i+1], data[i]
	}
}

// swapBytes32 converts n64 (word-swapped/little-endian) to z64 (big-endian) format in place.
// Each 4-byte word is reversed: ABCD -> DCBA
func swapBytes32(data []byte) {
	for i := 0; i < len(data)-3; i += 4 {
		data[i], data[i+1], data[i+2], data[i+3] = data[i+3], data[i+2], data[i+1], data[i]
	}
}
