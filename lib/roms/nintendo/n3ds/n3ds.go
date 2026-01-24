package n3ds

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
)

// Nintendo 3DS CCI/NCSD format parsing.
//
// 3DS CCI (CTR Cart Image) files use the NCSD container format.
// NCSD (Nintendo CTR System Data) contains one or more NCCH partitions.
// https://www.3dbrew.org/wiki/NCSD
// https://www.3dbrew.org/wiki/NCCH
//
// NCSD Header layout (0x200 bytes at file offset 0x000):
//
//	Offset  Size  Description
//	0x000   256   RSA-2048 SHA-256 signature
//	0x100   4     Magic "NCSD" (0x4453434E little-endian)
//	0x104   4     Image size in media units (1 media unit = 0x200 bytes)
//	0x108   8     Media ID
//	0x110   8     Partition FS types (one byte per partition)
//	0x118   8     Partition crypto types (one byte per partition)
//	0x120   64    Partition table (8 entries × 8 bytes: offset + size in media units)
//	0x160   32    Extended header hash (SHA-256)
//	0x180   4     Additional header size
//	0x184   4     Sector zero offset
//	0x188   8     Partition flags
//	0x190   64    Partition ID table
//	0x1D0   48    Reserved
//
// NCCH Header layout (0x200 bytes at partition 0 offset):
//
//	Offset  Size  Description
//	0x000   256   RSA-2048 SHA-256 signature
//	0x100   4     Magic "NCCH" (0x4843434E little-endian)
//	0x104   4     Content size in media units
//	0x108   8     Partition/Title ID
//	0x110   2     Maker code (ASCII)
//	0x112   2     Version
//	0x114   4     Seed hash verification
//	0x118   8     Program ID / Title ID
//	0x120   16    Reserved
//	0x130   32    Logo region hash
//	0x150   16    Product code (ASCII, e.g., "CTR-P-ALGE")
//	0x160   32    Extended header hash
//	0x180   4     Extended header size
//	0x184   4     Reserved
//	0x188   8     Flags (content type, platform, crypto)
//	0x190   112   Region offsets/sizes (plain, logo, exefs, romfs)
//
// Product code format (16 bytes at NCCH+0x150):
//
//	Format: XXX-Y-ZZZZ
//	  - CTR = 3DS, KTR = New 3DS exclusive
//	  - P = retail game, N = demo/special
//	  - ZZZZ = 4-char game code, last char is region (J/E/P/U/C/K/T)
//
// NCCH Flags (8 bytes at NCCH+0x188):
//
//	Index  Description
//	0-2    Reserved
//	3      Crypto method (0=fixed key)
//	4      Content platform (bit 1 = New 3DS exclusive)
//	5      Content type (bits 0-2: 0=App, 1=SysUpdate, 2=Manual, 3=DLP, 4=Trial)
//	6      Content unit size (log2)
//	7      Fixed crypto key flag

const (
	ncsdHeaderSize       = 0x200 // 512 bytes
	ncsdMagicOffset      = 0x100
	ncsdMagic            = "NCSD"
	ncsdImageSizeOffset  = 0x104
	ncsdMediaIDOffset    = 0x108
	ncsdPartTableOffset  = 0x120
	ncsdPartTableEntries = 8
	ncsdPartEntrySize    = 8 // 4 bytes offset + 4 bytes size

	ncchMagicOffset       = 0x100
	ncchMagic             = "NCCH"
	ncchMakerCodeOffset   = 0x110
	ncchMakerCodeLen      = 2
	ncchVersionOffset     = 0x112
	ncchTitleIDOffset     = 0x118
	ncchProductCodeOffset = 0x150
	ncchProductCodeLen    = 16
	ncchFlagsOffset       = 0x188
	ncchMinHeaderSize     = 0x200 // Minimum NCCH header to read

	mediaUnitSize = 0x200 // 512 bytes per media unit
)

// N3DSContentType represents the type of NCCH content.
type N3DSContentType byte

// N3DSContentType values per 3dbrew.
const (
	N3DSContentTypeApplication  N3DSContentType = 0x00 // Application (game)
	N3DSContentTypeSystemUpdate N3DSContentType = 0x01 // System update
	N3DSContentTypeManual       N3DSContentType = 0x02 // Manual
	N3DSContentTypeDLPChild     N3DSContentType = 0x03 // Download Play child
	N3DSContentTypeTrial        N3DSContentType = 0x04 // Trial
)

// N3DSRegion represents the target region from the product code.
type N3DSRegion byte

// N3DSRegion values from product code.
const (
	N3DSRegionJapan     N3DSRegion = 'J'
	N3DSRegionUSA       N3DSRegion = 'E'
	N3DSRegionEurope    N3DSRegion = 'P'
	N3DSRegionAustralia N3DSRegion = 'U'
	N3DSRegionChina     N3DSRegion = 'C'
	N3DSRegionKorea     N3DSRegion = 'K'
	N3DSRegionTaiwan    N3DSRegion = 'T'
)

// N3DSInfo contains metadata extracted from a 3DS CCI/NCSD file.
type N3DSInfo struct {
	// MediaID is the unique media identifier from the NCSD header (0x108).
	MediaID uint64 `json:"media_id"`
	// ImageSize is the total image size in bytes.
	ImageSize int64 `json:"image_size"`
	// PartitionCount is the number of valid partitions.
	PartitionCount int `json:"partition_count"`

	// TitleID is the 64-bit title identifier from the NCCH header (0x118).
	TitleID uint64 `json:"title_id"`
	// ProductCode is the game identifier (e.g., "CTR-P-ALGE") from NCCH (0x150).
	ProductCode string `json:"product_code,omitempty"`
	// MakerCode is the 2-character publisher code from NCCH (0x110).
	MakerCode string `json:"maker_code,omitempty"`
	// Version is the NCCH version number (0x112).
	Version uint16 `json:"version"`

	// ContentType indicates the type of content from NCCH flags[5].
	ContentType N3DSContentType `json:"content_type"`
	// IsNew3DSExclusive indicates if this is a New 3DS exclusive title (NCCH flags[4] bit 1).
	IsNew3DSExclusive bool `json:"is_new3ds_exclusive"`
	// Region is the target region from the product code.
	Region N3DSRegion `json:"region"`

	// platform is the target platform (internal, used by GamePlatform).
	platform core.Platform
}

// GamePlatform implements identify.GameInfo.
func (i *N3DSInfo) GamePlatform() core.Platform { return i.platform }

// GameTitle implements identify.GameInfo.
// 3DS CCI format does not contain a title field in the header.
func (i *N3DSInfo) GameTitle() string { return "" }

// GameSerial implements identify.GameInfo.
// Returns the product code (e.g., "CTR-P-ALGE").
func (i *N3DSInfo) GameSerial() string { return i.ProductCode }

// ParseN3DS extracts game information from a 3DS CCI/NCSD file.
func ParseN3DS(r io.ReaderAt, size int64) (*N3DSInfo, error) {
	if size < ncsdHeaderSize {
		return nil, fmt.Errorf("file too small for NCSD header: %d bytes", size)
	}

	// Read NCSD header
	ncsdHeader := make([]byte, ncsdHeaderSize)
	if _, err := r.ReadAt(ncsdHeader, 0); err != nil {
		return nil, fmt.Errorf("failed to read NCSD header: %w", err)
	}

	// Validate NCSD magic
	magic := string(ncsdHeader[ncsdMagicOffset : ncsdMagicOffset+4])
	if magic != ncsdMagic {
		return nil, fmt.Errorf("not a valid 3DS NCSD file: expected magic %q, got %q", ncsdMagic, magic)
	}

	// Parse NCSD header fields
	imageSizeUnits := binary.LittleEndian.Uint32(ncsdHeader[ncsdImageSizeOffset:])
	imageSize := int64(imageSizeUnits) * mediaUnitSize
	mediaID := binary.LittleEndian.Uint64(ncsdHeader[ncsdMediaIDOffset:])

	// Find partition 0 (main content)
	partOffset, partSize, err := parsePartitionEntry(ncsdHeader, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse partition 0: %w", err)
	}

	if partOffset == 0 || partSize == 0 {
		return nil, fmt.Errorf("partition 0 is empty or invalid")
	}

	// Count valid partitions (entries fully contained within the NCSD image and file)
	partCount := 0
	for i := 0; i < ncsdPartTableEntries; i++ {
		pOff, pSize, _ := parsePartitionEntry(ncsdHeader, i)
		if pOff > 0 && pSize > 0 {
			partEnd := int64(pOff+pSize) * mediaUnitSize
			if partEnd <= imageSize && partEnd <= size {
				partCount++
			}
		}
	}

	// Calculate NCCH offset in bytes
	ncchOffset := int64(partOffset) * mediaUnitSize
	if ncchOffset+ncchMinHeaderSize > size {
		return nil, fmt.Errorf("NCCH partition extends beyond file: offset %d, file size %d", ncchOffset, size)
	}

	// Read NCCH header
	ncchHeader := make([]byte, ncchMinHeaderSize)
	if _, err := r.ReadAt(ncchHeader, ncchOffset); err != nil {
		return nil, fmt.Errorf("failed to read NCCH header at offset %d: %w", ncchOffset, err)
	}

	// Validate NCCH magic
	ncchMagicVal := string(ncchHeader[ncchMagicOffset : ncchMagicOffset+4])
	if ncchMagicVal != ncchMagic {
		return nil, fmt.Errorf("not a valid NCCH partition: expected magic %q, got %q", ncchMagic, ncchMagicVal)
	}

	// Parse NCCH fields
	titleID := binary.LittleEndian.Uint64(ncchHeader[ncchTitleIDOffset:])
	makerCode := util.ExtractASCII(ncchHeader[ncchMakerCodeOffset : ncchMakerCodeOffset+ncchMakerCodeLen])
	version := binary.LittleEndian.Uint16(ncchHeader[ncchVersionOffset:])
	productCode := util.ExtractASCII(ncchHeader[ncchProductCodeOffset : ncchProductCodeOffset+ncchProductCodeLen])

	// Parse flags
	flags := ncchHeader[ncchFlagsOffset : ncchFlagsOffset+8]
	contentType := N3DSContentType(flags[5] & 0x07) // Lower 3 bits
	isNew3DSExclusive := (flags[4] & 0x02) != 0     // Bit 1 of flags[4]

	// Determine region from product code (last character of game code)
	// Format: CTR-P-XXXR where R is region
	var region N3DSRegion
	if len(productCode) >= 10 {
		region = N3DSRegion(productCode[9])
	}

	// Determine platform
	var platform core.Platform
	if isNew3DSExclusive {
		platform = core.PlatformNew3DS
	} else {
		platform = core.Platform3DS
	}

	return &N3DSInfo{
		MediaID:           mediaID,
		ImageSize:         imageSize,
		PartitionCount:    partCount,
		TitleID:           titleID,
		ProductCode:       productCode,
		MakerCode:         makerCode,
		Version:           version,
		ContentType:       contentType,
		IsNew3DSExclusive: isNew3DSExclusive,
		Region:            region,
		platform:          platform,
	}, nil
}

// parsePartitionEntry extracts offset and size for a partition from the NCSD table.
func parsePartitionEntry(header []byte, index int) (offset, size uint32, err error) {
	if index < 0 || index >= ncsdPartTableEntries {
		return 0, 0, fmt.Errorf("partition index %d out of range", index)
	}
	entryOffset := ncsdPartTableOffset + (index * ncsdPartEntrySize)
	offset = binary.LittleEndian.Uint32(header[entryOffset:])
	size = binary.LittleEndian.Uint32(header[entryOffset+4:])
	return offset, size, nil
}
