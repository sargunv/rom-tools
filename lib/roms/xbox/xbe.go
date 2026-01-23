package xbox

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"

	"github.com/sargunv/rom-tools/lib/core"
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

	xbeCertTimestampOff     = 0x04
	xbeCertTitleIDOff       = 0x08
	xbeCertTitleNameOff     = 0x0C
	xbeCertTitleNameLen     = 80
	xbeCertAltTitleIDsOff   = 0x5C
	xbeCertAltTitleIDsCount = 16
	xbeCertMediaTypesOff    = 0x9C
	xbeCertRegionOff        = 0xA0
	xbeCertRatingsOff       = 0xA4
	xbeCertDiscNumOff       = 0xA8
	xbeCertVersionOff       = 0xAC
	xbeCertSize             = 0x1D0
)

// XboxRegion represents Xbox region flags.
type XboxRegion uint32

const (
	XboxRegionNA    XboxRegion = 0x00000001 // North America (US + Canada)
	XboxRegionJapan XboxRegion = 0x00000002 // Japan
	XboxRegionEUAU  XboxRegion = 0x00000004 // Europe and Australia combined
	XboxRegionDebug XboxRegion = 0x80000000 // Debug/development region
)

// XboxMediaType represents Xbox allowed media type flags.
type XboxMediaType uint32

const (
	XboxMediaHardDisk      XboxMediaType = 0x00000001
	XboxMediaDVDX2         XboxMediaType = 0x00000002
	XboxMediaDVDCD         XboxMediaType = 0x00000004
	XboxMediaCD            XboxMediaType = 0x00000008
	XboxMediaDVD5RO        XboxMediaType = 0x00000010
	XboxMediaDVD9RO        XboxMediaType = 0x00000020
	XboxMediaDVD5RW        XboxMediaType = 0x00000040
	XboxMediaDVD9RW        XboxMediaType = 0x00000080
	XboxMediaUSB           XboxMediaType = 0x00000100
	XboxMediaMemoryUnit    XboxMediaType = 0x00000200
	XboxMediaOnlineContent XboxMediaType = 0x00000400
)

// XBEInfo contains metadata extracted from an Xbox XBE file.
type XBEInfo struct {
	// TitleID is the numeric title ID.
	TitleID uint32 `json:"title_id"`
	// TitleIDHex is the title ID as an 8-character hex string.
	TitleIDHex string `json:"title_id_hex,omitempty"`
	// PublisherCode is the 2-character publisher code from title ID.
	PublisherCode string `json:"publisher_code,omitempty"`
	// GameNumber is the game number from title ID.
	GameNumber uint16 `json:"game_number"`
	// Title is the game title.
	Title string `json:"title,omitempty"`
	// Timestamp is the certificate timestamp (seconds since 2000-01-01).
	Timestamp uint32 `json:"timestamp"`
	// AlternateTitleIDs contains alternate title IDs for region variants.
	AlternateTitleIDs []uint32 `json:"alternate_title_ids,omitempty"`
	// AllowedMediaTypes is a bitmask of allowed media types.
	AllowedMediaTypes XboxMediaType `json:"allowed_media_types"`
	// RegionFlags is the bitmask of XboxRegion values.
	RegionFlags XboxRegion `json:"region_flags"`
	// GameRatings contains game rating flags.
	GameRatings uint32 `json:"game_ratings"`
	// DiscNumber is the disc number for multi-disc games.
	DiscNumber uint32 `json:"disc_number"`
	// Version is the game version.
	Version uint32 `json:"version"`
}

// GamePlatform implements identify.GameInfo.
func (i *XBEInfo) GamePlatform() core.Platform { return core.PlatformXbox }

// GameTitle implements identify.GameInfo.
func (i *XBEInfo) GameTitle() string { return i.Title }

// GameSerial implements identify.GameInfo. Returns serial in "XX-###" format.
func (i *XBEInfo) GameSerial() string {
	return fmt.Sprintf("%s-%03d", i.PublisherCode, i.GameNumber)
}

// ParseXBE extracts game information from an XBE file.
func ParseXBE(r io.ReaderAt, size int64) (*XBEInfo, error) {
	return parseXBEAt(r, 0, size)
}

// parseXBEAt extracts game information from an XBE file at the given offset.
// This is useful when the XBE is embedded within another file (e.g., in an XISO).
func parseXBEAt(r io.ReaderAt, xbeOffset int64, size int64) (*XBEInfo, error) {
	// Validate minimum size
	if size < xbeHeaderSize {
		return nil, fmt.Errorf("file too small for XBE header: %d bytes (need at least %d)", size, xbeHeaderSize)
	}

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
	timestamp := binary.LittleEndian.Uint32(cert[xbeCertTimestampOff:])
	titleID := binary.LittleEndian.Uint32(cert[xbeCertTitleIDOff:])
	mediaTypes := binary.LittleEndian.Uint32(cert[xbeCertMediaTypesOff:])
	regionFlags := binary.LittleEndian.Uint32(cert[xbeCertRegionOff:])
	gameRatings := binary.LittleEndian.Uint32(cert[xbeCertRatingsOff:])
	discNumber := binary.LittleEndian.Uint32(cert[xbeCertDiscNumOff:])
	version := binary.LittleEndian.Uint32(cert[xbeCertVersionOff:])

	// Parse title name (UTF-16LE)
	titleBytes := cert[xbeCertTitleNameOff : xbeCertTitleNameOff+xbeCertTitleNameLen]
	title := decodeUTF16LE(titleBytes)

	// Parse alternate title IDs
	altTitleIDs := make([]uint32, 0, xbeCertAltTitleIDsCount)
	for i := 0; i < xbeCertAltTitleIDsCount; i++ {
		off := xbeCertAltTitleIDsOff + i*4
		altID := binary.LittleEndian.Uint32(cert[off:])
		if altID != 0 {
			altTitleIDs = append(altTitleIDs, altID)
		}
	}

	// Decode title ID into publisher code and game number
	publisherCode, gameNumber := decodeTitleID(titleID)

	return &XBEInfo{
		TitleID:           titleID,
		TitleIDHex:        fmt.Sprintf("%08X", titleID),
		PublisherCode:     publisherCode,
		GameNumber:        gameNumber,
		Title:             title,
		Timestamp:         timestamp,
		AlternateTitleIDs: altTitleIDs,
		AllowedMediaTypes: XboxMediaType(mediaTypes),
		RegionFlags:       XboxRegion(regionFlags),
		GameRatings:       gameRatings,
		DiscNumber:        discNumber,
		Version:           version,
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
