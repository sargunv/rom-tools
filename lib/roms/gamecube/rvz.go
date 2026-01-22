package gamecube

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

// RVZ/WIA (Dolphin compressed disc image) format parsing.
//
// Format specification:
// https://raw.githubusercontent.com/dolphin-emu/dolphin/refs/heads/master/docs/WiaAndRvz.md
//
// wia_file_head_t (offset 0x0, 0x48 bytes):
//
//	Offset  Size  Description
//	0x00    4     Magic: "WIA\x1"
//	0x04    4     Version
//	0x08    4     Compatible version
//	0x0C    4     Disc struct size
//	0x10    20    SHA-1 hash of disc struct
//	0x24    8     ISO file size (original disc size)
//	0x2C    8     WIA/RVZ file size
//	0x34    20    SHA-1 hash of file header
//
// wia_disc_t (offset 0x48):
//
//	Offset  Size  Description
//	0x00    4     Disc type (1=GameCube, 2=Wii)
//	0x04    4     Compression method (0=NONE, 2=BZIP2, 3=LZMA, 4=LZMA2, 5=Zstandard)
//	0x08    4     Compression level (signed for Zstandard)
//	0x0C    4     Chunk size
//	0x10    128   dhead[0x80] - First 128 bytes of disc (UNCOMPRESSED!)

const (
	fileHeadSize   = 0x48
	discStructBase = 0x48

	// wia_file_head_t offsets
	magicOffset        = 0x00
	versionOffset      = 0x04
	compatVerOffset    = 0x08
	discSizeOffset     = 0x0C
	discHashOffset     = 0x10
	isoFileSizeOffset  = 0x24
	wiaFileSizeOffset  = 0x2C
	fileHeadHashOffset = 0x34

	// wia_disc_t offsets (relative to discStructBase)
	discTypeOffset    = 0x00
	compressionOffset = 0x04
	comprLevelOffset  = 0x08
	chunkSizeOffset   = 0x0C
	dheadOffset       = 0x10
	dheadSize         = 0x80

	// Total header size we need to read
	totalHeaderSize = discStructBase + dheadOffset + dheadSize
)

// Compression indicates the compression method used in an RVZ/WIA file.
type Compression uint32

// Compression method constants.
const (
	CompressionNone      Compression = 0
	CompressionPurge     Compression = 1 // Not used in RVZ
	CompressionBZIP2     Compression = 2
	CompressionLZMA      Compression = 3
	CompressionLZMA2     Compression = 4
	CompressionZstandard Compression = 5
)

// DiscType indicates the disc type (GameCube or Wii) in an RVZ/WIA file.
type DiscType uint32

// Disc type constants.
const (
	DiscTypeUnknown  DiscType = 0
	DiscTypeGameCube DiscType = 1
	DiscTypeWii      DiscType = 2
)

// RVZInfo contains metadata extracted from an RVZ/WIA file header.
type RVZInfo struct {
	// Version is the WIA format version.
	Version uint32
	// CompatibleVersion is the minimum compatible version.
	CompatibleVersion uint32
	// DiscType indicates the disc type (GameCube or Wii).
	DiscType DiscType
	// Compression is the compression method.
	Compression Compression
	// CompressionLevel is the compression level (signed for Zstandard).
	CompressionLevel int32
	// ChunkSize is the chunk size for compressed data.
	ChunkSize uint32
	// ISOFileSize is the original uncompressed disc size.
	ISOFileSize uint64
	// WIAFileSize is the compressed file size.
	WIAFileSize uint64
	// DiscHash is the SHA-1 hash of disc struct (hex).
	DiscHash string
	// FileHeadHash is the SHA-1 hash of file header (hex).
	FileHeadHash string
	// DiscHeader contains the first 128 bytes of disc (uncompressed).
	DiscHeader []byte
}

// ParseRVZ reads and parses an RVZ/WIA file header.
func ParseRVZ(r io.ReaderAt, size int64) (*RVZInfo, error) {
	if size < totalHeaderSize {
		return nil, fmt.Errorf("file too small for RVZ header: need %d bytes, got %d", totalHeaderSize, size)
	}

	header := make([]byte, totalHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read RVZ header: %w", err)
	}

	// Verify magic bytes "WIA\x1" or "RVZ\x1"
	magic := string(header[magicOffset : magicOffset+4])
	if magic != "WIA\x01" && magic != "RVZ\x01" {
		return nil, fmt.Errorf("not a valid RVZ/WIA file: invalid magic (got %q)", magic)
	}

	// Parse wia_file_head_t
	version := binary.BigEndian.Uint32(header[versionOffset:])
	compatVer := binary.BigEndian.Uint32(header[compatVerOffset:])
	discHash := hex.EncodeToString(header[discHashOffset : discHashOffset+20])
	isoFileSize := binary.BigEndian.Uint64(header[isoFileSizeOffset:])
	wiaFileSize := binary.BigEndian.Uint64(header[wiaFileSizeOffset:])
	fileHeadHash := hex.EncodeToString(header[fileHeadHashOffset : fileHeadHashOffset+20])

	// Parse wia_disc_t
	discType := DiscType(binary.BigEndian.Uint32(header[discStructBase+discTypeOffset:]))
	compression := Compression(binary.BigEndian.Uint32(header[discStructBase+compressionOffset:]))
	comprLevel := int32(binary.BigEndian.Uint32(header[discStructBase+comprLevelOffset:]))
	chunkSize := binary.BigEndian.Uint32(header[discStructBase+chunkSizeOffset:])

	// Extract dhead (first 128 bytes of disc, uncompressed)
	dhead := make([]byte, dheadSize)
	copy(dhead, header[discStructBase+dheadOffset:])

	if len(dhead) < discHeaderSize {
		return nil, fmt.Errorf("RVZ disc header too small: %d bytes (need %d)", len(dhead), discHeaderSize)
	}

	return &RVZInfo{
		Version:           version,
		CompatibleVersion: compatVer,
		DiscType:          discType,
		Compression:       compression,
		CompressionLevel:  comprLevel,
		ChunkSize:         chunkSize,
		ISOFileSize:       isoFileSize,
		WIAFileSize:       wiaFileSize,
		DiscHash:          discHash,
		FileHeadHash:      fileHeadHash,
		DiscHeader:        dhead,
	}, nil
}
