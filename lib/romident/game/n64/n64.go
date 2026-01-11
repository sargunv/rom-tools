package n64

import (
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
)

// N64 ROM format parsing.
//
// N64 ROM header specification:
// https://n64brew.dev/wiki/ROM_Header
//
// N64 ROMs come in three byte orderings:
//   - Big Endian (.z64): 0x80 at position 0 - native N64 format
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
	n64HeaderSize     = 0x40 // 64 bytes
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
	Title        string
	GameCode     string // Full 4-character game code
	CategoryCode byte   // 1st char: N=Game Pak, D=64DD, C=Expandable, etc.
	UniqueCode   string // 2nd-3rd chars: unique game identifier
	RegionCode   byte   // 4th char: J=Japan, E=USA, P=Europe, etc.
	Version      int
	ByteOrder    N64ByteOrder // Detected byte ordering
}

// ParseN64 extracts game information from an N64 ROM file.
func ParseN64(r io.ReaderAt, size int64) (*N64Info, error) {
	if size < n64HeaderSize {
		return nil, fmt.Errorf("file too small for N64 header: %d bytes", size)
	}

	// Read first 4 bytes to detect byte order
	first4 := make([]byte, 4)
	if _, err := r.ReadAt(first4, 0); err != nil {
		return nil, fmt.Errorf("failed to read N64 header start: %w", err)
	}

	byteOrder, err := detectN64ByteOrder(first4)
	if err != nil {
		return nil, err
	}

	// Read full header
	header := make([]byte, n64HeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read N64 header: %w", err)
	}

	// Convert to big-endian (native format) if needed
	switch byteOrder {
	case N64ByteSwapped:
		swapBytes16(header)
	case N64LittleEndian:
		swapBytes32(header)
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

// detectN64ByteOrder determines the byte ordering by finding where 0x80 is located.
func detectN64ByteOrder(first4 []byte) (N64ByteOrder, error) {
	switch {
	case first4[0] == n64ReservedByte:
		return N64BigEndian, nil
	case first4[1] == n64ReservedByte:
		return N64ByteSwapped, nil
	case first4[3] == n64ReservedByte:
		return N64LittleEndian, nil
	default:
		return "", fmt.Errorf("not a valid N64 ROM: could not detect byte order (first 4 bytes: %02X %02X %02X %02X)",
			first4[0], first4[1], first4[2], first4[3])
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

// IsN64ROM checks if the first 4 bytes indicate an N64 ROM (any byte order).
func IsN64ROM(first4 []byte) bool {
	if len(first4) < 4 {
		return false
	}
	return first4[0] == n64ReservedByte ||
		first4[1] == n64ReservedByte ||
		first4[3] == n64ReservedByte
}

// DetectN64ByteOrder returns the specific N64 format (Z64, V64, or N64) based on byte order.
func DetectN64ByteOrder(first4 []byte) N64ByteOrder {
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
