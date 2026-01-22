// Package chd provides support for reading CHD (Compressed Hunks of Data) files.
// CHD is MAME's compressed disc image format.
//
// The API mirrors archive/zip: use NewReader to open a CHD, then access
// individual tracks via the Tracks slice.
//
// Format specification: https://github.com/mamedev/mame/blob/master/src/lib/util/chd.h
package chd

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/sargunv/rom-tools/lib/format/chd/internal/codec"
)

// V5 header layout (124 bytes):
//
//	Offset  Size  Description
//	0       8     Magic ("MComprHD")
//	8       4     Header length (big-endian)
//	12      4     Version (big-endian)
//	16      4     Compressors[0]
//	20      4     Compressors[1]
//	24      4     Compressors[2]
//	28      4     Compressors[3]
//	32      8     Logical bytes (big-endian)
//	40      8     Map offset (big-endian)
//	48      8     Metadata offset (big-endian)
//	56      4     Hunk bytes (big-endian)
//	60      4     Unit bytes (big-endian)
//	64      20    Raw SHA1 (of the raw, uncompressed data)
//	84      20    SHA1 (of the compressed data)
//	104     20    Parent SHA1 (all zeros if no parent)
const (
	headerSize       = 124
	rawSHA1Offset    = 64
	sha1Offset       = 84
	parentSHA1Offset = 104
	sha1Size         = 20
)

// Codec represents a CHD compression codec ID (4-byte ASCII code stored as big-endian uint32).
type Codec uint32

// Codec IDs
const (
	CodecNone   Codec = 0
	CodecZlib   Codec = 0x7a6c6962 // 'zlib'
	CodecLZMA   Codec = 0x6c7a6d61 // 'lzma'
	CodecHuff   Codec = 0x68756666 // 'huff'
	CodecFLAC   Codec = 0x666c6163 // 'flac'
	CodecZstd   Codec = 0x7a737464 // 'zstd'
	CodecCDZlib Codec = 0x63647a6c // 'cdzl'
	CodecCDLZMA Codec = 0x63646c7a // 'cdlz'
	CodecCDFLAC Codec = 0x6364666c // 'cdfl'
	CodecCDZstd Codec = 0x63647a73 // 'cdzs'
)

// Header contains metadata extracted from a CHD file header.
type Header struct {
	Version      uint32
	Compressors  [4]Codec
	LogicalBytes uint64
	MapOffset    uint64
	HunkBytes    uint32
	UnitBytes    uint32
	TotalHunks   uint32
	RawSHA1      string
	SHA1         string
	ParentSHA1   string
}

// Reader provides access to a CHD file's contents.
// It mirrors the archive/zip.Reader pattern.
type Reader struct {
	// Tracks contains all tracks in the CHD (like zip.Reader.File).
	// For single-track images, this will have one entry.
	// For multi-track CDs (e.g., with audio tracks), iterate to find data tracks.
	Tracks []*Track

	file      io.ReaderAt
	header    *Header
	hunkMap   *chdMap
	hunkCache map[uint32][]byte
	cacheMu   sync.RWMutex
}

// NewReader creates a Reader reading from r, which must be an io.ReaderAt.
// This mirrors the archive/zip.NewReader pattern.
func NewReader(r io.ReaderAt, size int64) (*Reader, error) {
	header, err := parseHeader(r, size)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	if header.ParentSHA1 != "" {
		return nil, fmt.Errorf("parent CHD references not supported")
	}

	hunkMap, err := decodeMap(r, header)
	if err != nil {
		return nil, fmt.Errorf("decode hunk map: %w", err)
	}

	reader := &Reader{
		file:      r,
		header:    header,
		hunkMap:   hunkMap,
		hunkCache: make(map[uint32][]byte),
	}

	// Parse track metadata
	reader.Tracks, err = parseTrackMetadata(r, header, reader)
	if err != nil {
		return nil, fmt.Errorf("parse tracks: %w", err)
	}

	return reader, nil
}

// Header returns the CHD header information.
func (r *Reader) Header() *Header {
	return r.header
}

// Size returns the logical (uncompressed) size in bytes.
func (r *Reader) Size() int64 {
	return int64(r.header.LogicalBytes)
}

// ReadAt implements io.ReaderAt, reading from the logical (uncompressed) data.
func (r *Reader) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, fmt.Errorf("negative offset")
	}
	if off >= int64(r.header.LogicalBytes) {
		return 0, io.EOF
	}

	hunkBytes := int64(r.header.HunkBytes)
	remaining := len(p)
	pos := off

	for remaining > 0 && pos < int64(r.header.LogicalBytes) {
		hunkNum := uint32(pos / hunkBytes)
		hunkOffset := int(pos % hunkBytes)

		hunkData, err := r.readHunk(hunkNum)
		if err != nil {
			if n > 0 {
				return n, nil
			}
			return 0, fmt.Errorf("read hunk %d: %w", hunkNum, err)
		}

		available := len(hunkData) - hunkOffset
		if available <= 0 {
			break
		}
		toCopy := min(remaining, available)

		copy(p[n:n+toCopy], hunkData[hunkOffset:hunkOffset+toCopy])
		n += toCopy
		remaining -= toCopy
		pos += int64(toCopy)
	}

	if n == 0 && remaining > 0 {
		return 0, io.EOF
	}
	return n, nil
}

// readHunk reads and decompresses a single hunk.
func (r *Reader) readHunk(hunkNum uint32) ([]byte, error) {
	r.cacheMu.RLock()
	if cached, ok := r.hunkCache[hunkNum]; ok {
		r.cacheMu.RUnlock()
		return cached, nil
	}
	r.cacheMu.RUnlock()

	if int(hunkNum) >= len(r.hunkMap.entries) {
		return nil, fmt.Errorf("hunk %d out of range (total: %d)", hunkNum, len(r.hunkMap.entries))
	}

	entry := r.hunkMap.entries[hunkNum]
	hunkBytes := r.header.HunkBytes

	var data []byte
	var err error

	switch entry.compression {
	case compressionNone:
		data = make([]byte, hunkBytes)
		_, err = r.file.ReadAt(data, int64(entry.offset))
		if err != nil {
			return nil, fmt.Errorf("read uncompressed hunk: %w", err)
		}

	case compressionType0, compressionType1, compressionType2, compressionType3:
		codecID := r.header.Compressors[entry.compression]

		compressed := make([]byte, entry.length)
		_, err = r.file.ReadAt(compressed, int64(entry.offset))
		if err != nil {
			return nil, fmt.Errorf("read compressed data: %w", err)
		}

		data, err = decompressHunk(compressed, codecID, hunkBytes)
		if err != nil {
			return nil, fmt.Errorf("decompress hunk (codec 0x%08x): %w", codecID, err)
		}

	case compressionSelf:
		refHunk := uint32(entry.offset)
		if refHunk >= hunkNum {
			return nil, fmt.Errorf("self-reference to hunk %d from hunk %d (forward reference)", refHunk, hunkNum)
		}
		data, err = r.readHunk(refHunk)
		if err != nil {
			return nil, fmt.Errorf("read self-referenced hunk %d: %w", refHunk, err)
		}
		data = append([]byte(nil), data...)

	case compressionParent:
		return nil, fmt.Errorf("parent CHD references not supported")

	default:
		return nil, fmt.Errorf("unknown compression type: %d", entry.compression)
	}

	r.cacheMu.Lock()
	if len(r.hunkCache) < 32 {
		r.hunkCache[hunkNum] = data
	}
	r.cacheMu.Unlock()

	return data, nil
}

// readSector reads a single sector (unit) from the CHD.
func (r *Reader) readSector(sectorNum uint64) ([]byte, error) {
	unitBytes := uint64(r.header.UnitBytes)
	offset := int64(sectorNum * unitBytes)

	data := make([]byte, unitBytes)
	n, err := r.ReadAt(data, offset)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return data[:n], nil
}

// parseHeader reads and parses a CHD file header.
func parseHeader(r io.ReaderAt, size int64) (*Header, error) {
	if size < headerSize {
		return nil, fmt.Errorf("file too small for CHD header: need %d bytes, got %d", headerSize, size)
	}

	buf := make([]byte, headerSize)
	if _, err := r.ReadAt(buf, 0); err != nil {
		return nil, fmt.Errorf("failed to read CHD header: %w", err)
	}

	if string(buf[0:8]) != "MComprHD" {
		return nil, fmt.Errorf("not a valid CHD file: invalid magic")
	}

	headerLen := binary.BigEndian.Uint32(buf[8:12])
	version := binary.BigEndian.Uint32(buf[12:16])

	if version < 5 {
		return nil, fmt.Errorf("CHD version %d not supported (only v5+ supported)", version)
	}
	if headerLen < headerSize {
		return nil, fmt.Errorf("CHD header too small: %d bytes", headerLen)
	}

	var compressors [4]Codec
	for i := range 4 {
		compressors[i] = Codec(binary.BigEndian.Uint32(buf[16+i*4:]))
	}

	logicalBytes := binary.BigEndian.Uint64(buf[32:40])
	mapOffset := binary.BigEndian.Uint64(buf[40:48])
	hunkBytes := binary.BigEndian.Uint32(buf[56:60])
	unitBytes := binary.BigEndian.Uint32(buf[60:64])

	var totalHunks uint32
	if hunkBytes > 0 {
		totalHunks = uint32((logicalBytes + uint64(hunkBytes) - 1) / uint64(hunkBytes))
	}

	rawSHA1 := hex.EncodeToString(buf[rawSHA1Offset : rawSHA1Offset+sha1Size])
	sha1 := hex.EncodeToString(buf[sha1Offset : sha1Offset+sha1Size])

	parentSHA1Bytes := buf[parentSHA1Offset : parentSHA1Offset+sha1Size]
	parentSHA1 := ""
	for _, b := range parentSHA1Bytes {
		if b != 0 {
			parentSHA1 = hex.EncodeToString(parentSHA1Bytes)
			break
		}
	}

	return &Header{
		Version:      version,
		Compressors:  compressors,
		LogicalBytes: logicalBytes,
		MapOffset:    mapOffset,
		HunkBytes:    hunkBytes,
		UnitBytes:    unitBytes,
		TotalHunks:   totalHunks,
		RawSHA1:      rawSHA1,
		SHA1:         sha1,
		ParentSHA1:   parentSHA1,
	}, nil
}

// decompressHunk decompresses a single hunk using the appropriate codec.
func decompressHunk(compressedData []byte, codecID Codec, hunkBytes uint32) ([]byte, error) {
	size := int(hunkBytes)

	switch codecID {
	case CodecNone:
		// Uncompressed - just return a copy
		result := make([]byte, size)
		copy(result, compressedData)
		return result, nil

	case CodecZlib:
		return codec.Zlib(compressedData, size)

	case CodecLZMA:
		return codec.LZMA(compressedData, size)

	case CodecHuff:
		return codec.Huffman(compressedData, size)

	case CodecZstd:
		return codec.Zstd(compressedData, size)

	case CodecCDZlib:
		return codec.CDZLIB(compressedData, hunkBytes)

	case CodecCDLZMA:
		return codec.CDLZMA(compressedData, hunkBytes)

	case CodecCDZstd:
		return codec.CDZstd(compressedData, hunkBytes)

	case CodecFLAC, CodecCDFLAC:
		// FLAC is for audio tracks - we don't need to decompress audio for identification
		return nil, fmt.Errorf("FLAC codec not supported (audio only)")

	default:
		return nil, fmt.Errorf("unknown codec: 0x%08x", codecID)
	}
}
