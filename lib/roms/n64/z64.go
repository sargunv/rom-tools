package n64

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
)

// N64 ROM format parsing (supports Z64, V64, and N64 byte orderings).
//
// N64 ROM header specification:
// https://n64brew.dev/wiki/ROM_Header
//
// N64 ROMs come in three byte orderings:
//   - Big Endian (.z64): 0x80 at position 0 - native N64 format
//   - Byte-Swapped (.v64): 0x80 at position 1 - pairs of bytes swapped
//   - Little Endian (.n64): 0x80 at position 3 - 32-bit words reversed
//
// N64 header layout (64 bytes, in native big-endian format):
//
//	Offset  Size  Description
//	0x00    1     Reserved (0x80 for all known commercial games)
//	0x01    3     PI BSD DOM1 Configuration Flags
//	0x04    4     Clock Rate
//	0x08    4     Boot Address (entry point)
//	0x0C    4     Libultra Version
//	0x10    8     Check Code (64-bit integrity checksum)
//	0x18    8     Reserved
//	0x20    20    Game Title (ASCII, space-padded)
//	0x34    7     Reserved (homebrew header uses some of this)
//	0x3B    4     Game Code (category, unique code, destination)
//	0x3F    1     ROM Version
//
// Game Code breakdown (4 bytes at 0x3B):
//   - Byte 0 (0x3B): Category Code - N=GamePak, D=64DD, C=Expandable, etc.
//   - Bytes 1-2 (0x3C-0x3D): Unique Code - 2-character game identifier
//   - Byte 3 (0x3E): Destination - J=Japan, E=USA, P=Europe, etc.

const (
	N64HeaderSize      = 0x40 // 64 bytes
	n64ReservedByte    = 0x80
	n64CheckCodeOffset = 0x10
	n64TitleOffset     = 0x20
	n64TitleLen        = 20
	n64GameCodeOffset  = 0x3B
	n64GameCodeLen     = 4
	n64VersionOffset   = 0x3F
)

// N64ByteOrder represents the byte ordering of an N64 ROM.
type N64ByteOrder string

const (
	N64BigEndian    N64ByteOrder = "z64" // Native format, 0x80 at position 0
	N64ByteSwapped  N64ByteOrder = "v64" // Byte-swapped pairs, 0x80 at position 1
	N64LittleEndian N64ByteOrder = "n64" // Word-swapped, 0x80 at position 3
	N64Unknown      N64ByteOrder = "unknown"
)

// N64CategoryCode represents the media type from the first byte of the game code.
type N64CategoryCode byte

// N64CategoryCode values per n64brew wiki.
const (
	N64CategoryGamePak      N64CategoryCode = 'N' // Standard cartridge
	N64Category64DD         N64CategoryCode = 'D' // 64DD Disk
	N64CategoryExpandable   N64CategoryCode = 'C' // Expandable: Game Pak Part
	N64CategoryExpandableDD N64CategoryCode = 'E' // Expandable: 64DD Part
	N64CategoryAleck64      N64CategoryCode = 'Z' // Aleck64 arcade
)

// N64Destination represents the target region from the game code.
type N64Destination byte

// N64Destination values per n64brew wiki.
const (
	N64DestinationAll          N64Destination = 'A'
	N64DestinationBrazil       N64Destination = 'B'
	N64DestinationChina        N64Destination = 'C'
	N64DestinationGermany      N64Destination = 'D'
	N64DestinationNorthAmerica N64Destination = 'E'
	N64DestinationFrance       N64Destination = 'F'
	N64DestinationGatewayNTSC  N64Destination = 'G'
	N64DestinationNetherlands  N64Destination = 'H'
	N64DestinationItaly        N64Destination = 'I'
	N64DestinationJapan        N64Destination = 'J'
	N64DestinationKorea        N64Destination = 'K'
	N64DestinationGatewayPAL   N64Destination = 'L'
	N64DestinationCanada       N64Destination = 'N'
	N64DestinationEurope       N64Destination = 'P'
	N64DestinationSpain        N64Destination = 'S'
	N64DestinationAustralia    N64Destination = 'U'
	N64DestinationScandinavia  N64Destination = 'W'
	N64DestinationEuropeX      N64Destination = 'X'
	N64DestinationEuropeY      N64Destination = 'Y'
	N64DestinationEuropeZ      N64Destination = 'Z'
)

// N64Info contains metadata extracted from an N64 ROM file.
type N64Info struct {
	// Title is the game title (space-padded ASCII, up to 20 characters).
	Title string
	// GameCode is the full 4-character game code.
	GameCode string
	// CategoryCode is the media type: N=GamePak, D=64DD, C=Expandable, etc.
	CategoryCode N64CategoryCode
	// UniqueCode is the 2-char unique game identifier.
	UniqueCode string
	// Destination is the target region: J=Japan, E=USA, P=Europe, etc.
	Destination N64Destination
	// Version is the ROM version number.
	Version int
	// ByteOrder is the detected byte ordering of the ROM.
	ByteOrder N64ByteOrder
	// CheckCode is the 64-bit integrity check value (big-endian at offset 0x10).
	CheckCode uint64
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
	// Extract check code (64-bit big-endian at offset 0x10)
	checkCode := binary.BigEndian.Uint64(header[n64CheckCodeOffset:])

	// Extract title (space-padded ASCII)
	title := util.ExtractASCII(header[n64TitleOffset : n64TitleOffset+n64TitleLen])

	// Extract game code
	gameCode := util.ExtractASCII(header[n64GameCodeOffset : n64GameCodeOffset+n64GameCodeLen])

	// Parse game code components
	var categoryCode N64CategoryCode
	var uniqueCode string
	var destination N64Destination
	if len(gameCode) >= 4 {
		categoryCode = N64CategoryCode(gameCode[0])
		uniqueCode = gameCode[1:3]
		destination = N64Destination(gameCode[3])
	}

	// Extract ROM version
	version := int(header[n64VersionOffset])

	return &N64Info{
		Title:        title,
		GameCode:     gameCode,
		CategoryCode: categoryCode,
		UniqueCode:   uniqueCode,
		Destination:  destination,
		Version:      version,
		ByteOrder:    byteOrder,
		CheckCode:    checkCode,
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
