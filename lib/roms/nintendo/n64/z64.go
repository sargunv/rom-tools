package n64

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
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
	N64HeaderSize        = 0x40 // 64 bytes
	n64ReservedByte      = 0x80
	n64PIBSDConfigOffset = 0x01 // 3 bytes (0x01-0x03)
	n64ClockRateOffset   = 0x04 // 4 bytes (0x04-0x07)
	n64BootAddressOffset = 0x08 // 4 bytes (0x08-0x0B)
	n64LibultraOffset    = 0x0C // 4 bytes (0x0C-0x0F)
	n64CheckCodeOffset   = 0x10 // 8 bytes (0x10-0x17)
	n64TitleOffset       = 0x20 // 20 bytes (0x20-0x33)
	n64TitleLen          = 20
	n64GameCodeOffset    = 0x3B // 4 bytes (0x3B-0x3E)
	n64GameCodeLen       = 4
	n64VersionOffset     = 0x3F // 1 byte
)

// ByteOrder represents the byte ordering of an N64 ROM.
// N64 ROMs are distributed in three byte orderings, identifiable by where 0x80 appears
// in the first 4 bytes. All commercial N64 ROMs begin with 0x80371240 in native format.
type ByteOrder string

const (
	// ByteOrderBigEndian (.z64) is the native N64 format with no byte reordering.
	// Bytes appear as: [0x80, 0x37, 0x12, 0x40] - 0x80 at position 0.
	ByteOrderBigEndian ByteOrder = "z64"
	// ByteOrderByteSwapped (.v64) has each pair of bytes swapped (16-bit byte swap).
	// Bytes appear as: [0x37, 0x80, 0x40, 0x12] - 0x80 at position 1.
	// Conversion: swap adjacent bytes (AB CD -> BA DC).
	ByteOrderByteSwapped ByteOrder = "v64"
	// ByteOrderLittleEndian (.n64) has each 32-bit word byte-reversed.
	// Bytes appear as: [0x40, 0x12, 0x37, 0x80] - 0x80 at position 3.
	// Conversion: reverse each 4-byte group (ABCD -> DCBA).
	ByteOrderLittleEndian ByteOrder = "n64"
	// ByteOrderUnknown indicates the byte order could not be detected.
	ByteOrderUnknown ByteOrder = "unknown"
)

// CategoryCode represents the media type from the first byte of the game code.
type CategoryCode byte

// CategoryCode values per n64brew wiki.
const (
	CategoryGamePak      CategoryCode = 'N' // Standard cartridge
	Category64DD         CategoryCode = 'D' // 64DD Disk
	CategoryExpandable   CategoryCode = 'C' // Expandable: Game Pak Part
	CategoryExpandableDD CategoryCode = 'E' // Expandable: 64DD Part
	CategoryAleck64      CategoryCode = 'Z' // Aleck64 arcade
)

// Destination represents the target region from the game code.
type Destination byte

// Destination values per n64brew wiki.
const (
	DestinationAll          Destination = 'A'
	DestinationBrazil       Destination = 'B'
	DestinationChina        Destination = 'C'
	DestinationGermany      Destination = 'D'
	DestinationNorthAmerica Destination = 'E'
	DestinationFrance       Destination = 'F'
	DestinationGatewayNTSC  Destination = 'G'
	DestinationNetherlands  Destination = 'H'
	DestinationItaly        Destination = 'I'
	DestinationJapan        Destination = 'J'
	DestinationKorea        Destination = 'K'
	DestinationGatewayPAL   Destination = 'L'
	DestinationCanada       Destination = 'N'
	DestinationEurope       Destination = 'P'
	DestinationSpain        Destination = 'S'
	DestinationAustralia    Destination = 'U'
	DestinationScandinavia  Destination = 'W'
	DestinationEuropeX      Destination = 'X'
	DestinationEuropeY      Destination = 'Y'
	DestinationEuropeZ      Destination = 'Z'
)

// Info contains metadata extracted from an N64 ROM file.
type Info struct {
	// PIBSDConfig is the PI BSD DOM1 configuration flags (0x01-0x03, 24-bit).
	// Controls ROM access timing for the Parallel Interface.
	PIBSDConfig uint32 `json:"pi_bsd_config"`
	// ClockRate is the clock rate override value (0x04-0x07).
	// Used by libultra for timing calculations; 0 means use default.
	ClockRate uint32 `json:"clock_rate"`
	// BootAddress is the program counter entry point in RDRAM (0x08-0x0B).
	BootAddress uint32 `json:"boot_address"`
	// LibultraVersion is the SDK version used to build the ROM (0x0C-0x0F).
	LibultraVersion uint32 `json:"libultra_version"`
	// CheckCode is the 64-bit integrity check value (0x10-0x17).
	CheckCode uint64 `json:"check_code"`
	// Title is the game title (0x20-0x33, space-padded ASCII, up to 20 characters).
	Title string `json:"title,omitempty"`
	// GameCode is the full 4-character game code (0x3B-0x3E).
	GameCode string `json:"game_code,omitempty"`
	// CategoryCode is the media type: N=GamePak, D=64DD, C=Expandable, etc.
	CategoryCode CategoryCode `json:"category_code"`
	// UniqueCode is the 2-char unique game identifier.
	UniqueCode string `json:"unique_code,omitempty"`
	// Destination is the target region: J=Japan, E=USA, P=Europe, etc.
	Destination Destination `json:"destination"`
	// Version is the ROM version number (0x3F).
	Version int `json:"version"`
	// ByteOrder is the detected byte ordering of the ROM.
	ByteOrder ByteOrder `json:"byte_order"`
}

// GamePlatform implements core.GameInfo.
func (i *Info) GamePlatform() core.Platform { return core.PlatformN64 }

// GameTitle implements core.GameInfo.
func (i *Info) GameTitle() string { return i.Title }

// GameSerial implements core.GameInfo.
func (i *Info) GameSerial() string { return i.GameCode }

// GameRegions implements core.GameInfo.
func (i *Info) GameRegions() []core.Region {
	switch i.Destination {
	case DestinationJapan:
		return []core.Region{core.RegionJapan}
	case DestinationNorthAmerica:
		return []core.Region{core.RegionUSA}
	case DestinationEurope, DestinationEuropeX, DestinationEuropeY, DestinationEuropeZ:
		return []core.Region{core.RegionEurope}
	case DestinationGermany:
		return []core.Region{core.RegionGermany}
	case DestinationFrance:
		return []core.Region{core.RegionFrance}
	case DestinationItaly:
		return []core.Region{core.RegionItaly}
	case DestinationSpain:
		return []core.Region{core.RegionSpain}
	case DestinationAustralia:
		return []core.Region{core.RegionAustralia}
	case DestinationBrazil:
		return []core.Region{core.RegionBrazil}
	case DestinationCanada:
		return []core.Region{core.RegionCanada}
	case DestinationChina:
		return []core.Region{core.RegionChina}
	case DestinationKorea:
		return []core.Region{core.RegionKorea}
	case DestinationNetherlands:
		return []core.Region{core.RegionNetherlands}
	case DestinationScandinavia:
		return []core.Region{core.RegionDenmark, core.RegionNorway, core.RegionSweden}
	case DestinationAll:
		return []core.Region{core.RegionWorld}
	case DestinationGatewayNTSC:
		return []core.Region{core.RegionAmerica}
	case DestinationGatewayPAL:
		return []core.Region{core.RegionEurope, core.RegionAsia}
	default:
		return []core.Region{}
	}
}

// Parse extracts game information from an N64 ROM file, auto-detecting byte order.
func Parse(r io.ReaderAt, size int64) (*Info, error) {
	if size < N64HeaderSize {
		return nil, fmt.Errorf("file too small for N64 header: %d bytes", size)
	}

	// Read first 4 bytes to detect byte order
	first4 := make([]byte, 4)
	if _, err := r.ReadAt(first4, 0); err != nil {
		return nil, fmt.Errorf("failed to read N64 header: %w", err)
	}

	byteOrder := detectByteOrder(first4)
	if byteOrder == ByteOrderUnknown {
		return nil, fmt.Errorf("not a valid N64 ROM: could not detect byte order")
	}

	// Read full header
	header := make([]byte, N64HeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read N64 header: %w", err)
	}

	// Convert to big-endian if needed
	switch byteOrder {
	case ByteOrderByteSwapped:
		swapBytes16(header)
	case ByteOrderLittleEndian:
		swapBytes32(header)
	}

	return parseN64Header(header, byteOrder)
}

// parseN64Header parses an N64 header from big-endian (z64) format bytes.
func parseN64Header(header []byte, byteOrder ByteOrder) (*Info, error) {
	// Extract PI BSD DOM1 config (24-bit at 0x01-0x03, stored in low 3 bytes)
	piBSDConfig := uint32(header[n64PIBSDConfigOffset])<<16 |
		uint32(header[n64PIBSDConfigOffset+1])<<8 |
		uint32(header[n64PIBSDConfigOffset+2])

	// Extract clock rate (32-bit at 0x04)
	clockRate := binary.BigEndian.Uint32(header[n64ClockRateOffset:])

	// Extract boot address (32-bit at 0x08)
	bootAddress := binary.BigEndian.Uint32(header[n64BootAddressOffset:])

	// Extract libultra version (32-bit at 0x0C)
	libultraVersion := binary.BigEndian.Uint32(header[n64LibultraOffset:])

	// Extract check code (64-bit at 0x10)
	checkCode := binary.BigEndian.Uint64(header[n64CheckCodeOffset:])

	// Extract title (space-padded ASCII at 0x20)
	title := util.ExtractASCII(header[n64TitleOffset : n64TitleOffset+n64TitleLen])

	// Extract game code (4 bytes at 0x3B)
	gameCode := util.ExtractASCII(header[n64GameCodeOffset : n64GameCodeOffset+n64GameCodeLen])

	// Parse game code components
	var categoryCode CategoryCode
	var uniqueCode string
	var destination Destination
	if len(gameCode) >= 4 {
		categoryCode = CategoryCode(gameCode[0])
		uniqueCode = gameCode[1:3]
		destination = Destination(gameCode[3])
	}

	// Extract ROM version (1 byte at 0x3F)
	version := int(header[n64VersionOffset])

	return &Info{
		PIBSDConfig:     piBSDConfig,
		ClockRate:       clockRate,
		BootAddress:     bootAddress,
		LibultraVersion: libultraVersion,
		CheckCode:       checkCode,
		Title:           title,
		GameCode:        gameCode,
		CategoryCode:    categoryCode,
		UniqueCode:      uniqueCode,
		Destination:     destination,
		Version:         version,
		ByteOrder:       byteOrder,
	}, nil
}

// detectByteOrder determines the byte ordering by finding where 0x80 is located.
func detectByteOrder(first4 []byte) ByteOrder {
	switch {
	case first4[0] == n64ReservedByte:
		return ByteOrderBigEndian
	case first4[1] == n64ReservedByte:
		return ByteOrderByteSwapped
	case first4[3] == n64ReservedByte:
		return ByteOrderLittleEndian
	default:
		return ByteOrderUnknown
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
