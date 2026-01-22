package xbox

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"
)

// XBE (Xbox Executable) format parsing.
//
// XBE specification:
// https://www.caustik.com/cxbx/download/xbe.htm
//
// XBE header layout (relevant fields):
//
//	Offset  Size  Description
//	0x0000  4     Magic ("XBEH")
//	0x0104  4     Base address (virtual memory base)
//	0x0118  4     Certificate address (virtual)
//
// XBE certificate layout (at certificate address - base address):
//
//	Offset  Size  Description
//	0x0000  4     Certificate size
//	0x0004  4     Timestamp
//	0x0008  4     Title ID
//	0x000C  80    Title name (UTF-16LE, null-terminated)
//	0x005C  64    Alternate title IDs
//	0x009C  4     Allowed media types
//	0x00A0  4     Game region flags
//	0x00A4  4     Game ratings
//	0x00A8  4     Disc number
//	0x00AC  4     Version

const (
	xbeMagicSize      = 4
	xbeBaseAddrOffset = 0x104
	xbeCertAddrOffset = 0x118
	xbeHeaderSize     = 0x178

	xbeCertTitleIDOff   = 0x08
	xbeCertTitleNameOff = 0x0C
	xbeCertTitleNameLen = 80
	xbeCertRegionOff    = 0xA0
	xbeCertDiscNumOff   = 0xA8
	xbeCertVersionOff   = 0xAC
	xbeCertSize         = 0x1D0
)

// XboxRegion represents Xbox region flags.
type XboxRegion uint32

const (
	XboxRegionNA    XboxRegion = 0x00000001 // North America (US + Canada)
	XboxRegionJapan XboxRegion = 0x00000002 // Japan
	XboxRegionEUAU  XboxRegion = 0x00000004 // Europe and Australia combined
	XboxRegionDebug XboxRegion = 0x80000000 // Debug/development region
)

// XBEInfo contains metadata extracted from an Xbox XBE file.
type XBEInfo struct {
	// TitleID is the numeric title ID.
	TitleID uint32
	// TitleIDHex is the title ID as an 8-character hex string.
	TitleIDHex string
	// PublisherCode is the 2-character publisher code from title ID.
	PublisherCode string
	// GameNumber is the game number from title ID.
	GameNumber uint16
	// Title is the game title.
	Title string
	// RegionFlags is the bitmask of XboxRegion values.
	RegionFlags uint32
	// DiscNumber is the disc number for multi-disc games.
	DiscNumber uint32
	// Version is the game version.
	Version uint32
}

// ParseXBE extracts game information from an XBE file.
func ParseXBE(r io.ReaderAt, size int64) (*XBEInfo, error) {
	return parseXBEAt(r, 0)
}

// parseXBEAt extracts game information from an XBE file at the given offset.
// This is useful when the XBE is embedded within another file (e.g., in an XISO).
func parseXBEAt(r io.ReaderAt, xbeOffset int64) (*XBEInfo, error) {
	// Read XBE header
	header := make([]byte, xbeHeaderSize)
	if _, err := r.ReadAt(header, xbeOffset); err != nil {
		return nil, fmt.Errorf("failed to read XBE header: %w", err)
	}

	// Verify magic (XBEH)
	if string(header[:xbeMagicSize]) != "XBEH" {
		return nil, fmt.Errorf("not a valid XBE: invalid magic")
	}

	// Get base address and certificate address
	baseAddr := binary.LittleEndian.Uint32(header[xbeBaseAddrOffset:])
	certAddr := binary.LittleEndian.Uint32(header[xbeCertAddrOffset:])

	// Calculate certificate file offset
	certFileOffset := xbeOffset + int64(certAddr-baseAddr)

	// Read certificate
	cert := make([]byte, xbeCertSize)
	if _, err := r.ReadAt(cert, certFileOffset); err != nil {
		return nil, fmt.Errorf("failed to read XBE certificate: %w", err)
	}

	// Parse certificate fields
	titleID := binary.LittleEndian.Uint32(cert[xbeCertTitleIDOff:])
	regionFlags := binary.LittleEndian.Uint32(cert[xbeCertRegionOff:])
	discNumber := binary.LittleEndian.Uint32(cert[xbeCertDiscNumOff:])
	version := binary.LittleEndian.Uint32(cert[xbeCertVersionOff:])

	// Parse title name (UTF-16LE)
	titleBytes := cert[xbeCertTitleNameOff : xbeCertTitleNameOff+xbeCertTitleNameLen]
	title := decodeUTF16LE(titleBytes)

	// Decode title ID into publisher code and game number
	publisherCode, gameNumber := decodeTitleID(titleID)

	return &XBEInfo{
		TitleID:       titleID,
		TitleIDHex:    fmt.Sprintf("%08X", titleID),
		PublisherCode: publisherCode,
		GameNumber:    gameNumber,
		Title:         title,
		RegionFlags:   regionFlags,
		DiscNumber:    discNumber,
		Version:       version,
	}, nil
}

// decodeTitleID extracts publisher code and game number from a title ID.
// Title ID format: high 16 bits = publisher code (2 ASCII chars), low 16 bits = game number
func decodeTitleID(titleID uint32) (string, uint16) {
	gameNumber := uint16(titleID & 0xFFFF)
	publisherCode := uint16((titleID >> 16) & 0xFFFF)

	// Publisher code is 2 ASCII chars (big-endian in the high word)
	char1 := byte((publisherCode >> 8) & 0xFF)
	char2 := byte(publisherCode & 0xFF)

	return string([]byte{char1, char2}), gameNumber
}

// decodeUTF16LE decodes a null-terminated UTF-16LE string.
func decodeUTF16LE(data []byte) string {
	// Convert bytes to uint16 slice
	u16s := make([]uint16, len(data)/2)
	for i := range u16s {
		u16s[i] = binary.LittleEndian.Uint16(data[i*2:])
	}

	// Find null terminator
	for i, v := range u16s {
		if v == 0 {
			u16s = u16s[:i]
			break
		}
	}

	return string(utf16.Decode(u16s))
}
