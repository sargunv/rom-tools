// Package chd provides support for reading CHD (Compressed Hunks of Data) files.
// CHD is MAME's compressed disc image format.
//
// Format specification: https://github.com/mamedev/mame/blob/master/src/lib/util/chd.h
package chd

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
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

// CHDInfo contains metadata extracted from a CHD file header.
type CHDInfo struct {
	// Version is the CHD format version.
	Version uint32
	// Compressors contains up to 4 compression codecs.
	Compressors [4]Codec
	// LogicalBytes is the total uncompressed size.
	LogicalBytes uint64
	// MapOffset is the offset to hunk map.
	MapOffset uint64
	// MetaOffset is the offset to metadata.
	MetaOffset uint64
	// HunkBytes is the bytes per hunk.
	HunkBytes uint32
	// UnitBytes is the bytes per unit (sector size).
	UnitBytes uint32
	// TotalHunks is calculated as LogicalBytes / HunkBytes.
	TotalHunks uint32
	// RawSHA1 is the SHA1 of the raw (uncompressed) data.
	RawSHA1 string
	// SHA1 is the SHA1 of the compressed data.
	SHA1 string
	// ParentSHA1 is the SHA1 of parent CHD (for diffs), empty if standalone.
	ParentSHA1 string
}

// ParseCHD reads and parses a CHD file header.
func ParseCHD(r io.ReaderAt, size int64) (*CHDInfo, error) {
	if size < headerSize {
		return nil, fmt.Errorf("file too small for CHD header: need %d bytes, got %d", headerSize, size)
	}

	header := make([]byte, headerSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read CHD header: %w", err)
	}

	// Verify magic
	if string(header[0:8]) != "MComprHD" {
		return nil, fmt.Errorf("not a valid CHD file: invalid magic")
	}

	// Header length (big-endian uint32 at offset 8)
	headerLen := binary.BigEndian.Uint32(header[8:12])

	// Version (big-endian uint32 at offset 12)
	version := binary.BigEndian.Uint32(header[12:16])

	if version < 5 {
		return nil, fmt.Errorf("CHD version %d not supported (only v5+ supported)", version)
	}

	if headerLen < headerSize {
		return nil, fmt.Errorf("CHD header too small: %d bytes", headerLen)
	}

	// Compressors (4 x uint32 at offsets 16-31)
	var compressors [4]Codec
	for i := range 4 {
		compressors[i] = Codec(binary.BigEndian.Uint32(header[16+i*4:]))
	}

	// Logical bytes (uint64 at offset 32)
	logicalBytes := binary.BigEndian.Uint64(header[32:40])

	// Map offset (uint64 at offset 40)
	mapOffset := binary.BigEndian.Uint64(header[40:48])

	// Metadata offset (uint64 at offset 48)
	metaOffset := binary.BigEndian.Uint64(header[48:56])

	// Hunk bytes (uint32 at offset 56)
	hunkBytes := binary.BigEndian.Uint32(header[56:60])

	// Unit bytes (uint32 at offset 60)
	unitBytes := binary.BigEndian.Uint32(header[60:64])

	// Calculate total hunks
	var totalHunks uint32
	if hunkBytes > 0 {
		totalHunks = uint32((logicalBytes + uint64(hunkBytes) - 1) / uint64(hunkBytes))
	}

	// Raw SHA1 at offset 64 (20 bytes) - SHA1 of the uncompressed data
	rawSHA1 := hex.EncodeToString(header[rawSHA1Offset : rawSHA1Offset+sha1Size])

	// SHA1 at offset 84 (20 bytes) - SHA1 of the compressed data
	sha1 := hex.EncodeToString(header[sha1Offset : sha1Offset+sha1Size])

	// Parent SHA1 at offset 104 (20 bytes) - all zeros if no parent
	parentSHA1Bytes := header[parentSHA1Offset : parentSHA1Offset+sha1Size]
	parentSHA1 := ""
	for _, b := range parentSHA1Bytes {
		if b != 0 {
			parentSHA1 = hex.EncodeToString(parentSHA1Bytes)
			break
		}
	}

	return &CHDInfo{
		Version:      version,
		Compressors:  compressors,
		LogicalBytes: logicalBytes,
		MapOffset:    mapOffset,
		MetaOffset:   metaOffset,
		HunkBytes:    hunkBytes,
		UnitBytes:    unitBytes,
		TotalHunks:   totalHunks,
		RawSHA1:      rawSHA1,
		SHA1:         sha1,
		ParentSHA1:   parentSHA1,
	}, nil
}

// IsCDROM returns true if this appears to be a CD-ROM image based on unit size.
func (h *CHDInfo) IsCDROM() bool {
	// CD-ROM CHDs typically use 2448 bytes per unit (2352 sector + 96 subcode)
	// or 2352 bytes per unit
	return h.UnitBytes == 2448 || h.UnitBytes == 2352
}

// Reader provides sector-level access to CHD files.
type Reader struct {
	file      io.ReaderAt
	header    *CHDInfo
	hunkMap   *chdMap
	hunkCache map[uint32][]byte
	cacheMu   sync.RWMutex
}

// Open creates a new CHD Reader for raw sector access.
func Open(r io.ReaderAt, size int64) (*Reader, error) {
	// Parse header
	header, err := ParseCHD(r, size)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	// Check for unsupported features
	if header.ParentSHA1 != "" {
		return nil, fmt.Errorf("parent CHD references not supported")
	}

	// Decode hunk map
	hunkMap, err := decodeMap(r, header)
	if err != nil {
		return nil, fmt.Errorf("decode hunk map: %w", err)
	}

	return &Reader{
		file:      r,
		header:    header,
		hunkMap:   hunkMap,
		hunkCache: make(map[uint32][]byte),
	}, nil
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
		// Calculate which hunk contains this position
		hunkNum := uint32(pos / hunkBytes)
		hunkOffset := int(pos % hunkBytes)

		// Get the decompressed hunk
		hunkData, err := r.readHunk(hunkNum)
		if err != nil {
			if n > 0 {
				return n, nil // Return partial read
			}
			return 0, fmt.Errorf("read hunk %d: %w", hunkNum, err)
		}

		// Copy data from this hunk
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
	// Check cache first
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
		// Uncompressed - read directly from file
		data = make([]byte, hunkBytes)
		_, err = r.file.ReadAt(data, int64(entry.offset))
		if err != nil {
			return nil, fmt.Errorf("read uncompressed hunk: %w", err)
		}

	case compressionType0, compressionType1, compressionType2, compressionType3:
		// Compressed - use the corresponding compressor
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
		// Self-reference - copy from another hunk
		refHunk := uint32(entry.offset)
		if refHunk >= hunkNum {
			return nil, fmt.Errorf("self-reference to hunk %d from hunk %d (forward reference)", refHunk, hunkNum)
		}
		data, err = r.readHunk(refHunk)
		if err != nil {
			return nil, fmt.Errorf("read self-referenced hunk %d: %w", refHunk, err)
		}
		// Make a copy so cache entries don't share backing arrays
		data = append([]byte(nil), data...)

	case compressionParent:
		return nil, fmt.Errorf("parent CHD references not supported")

	default:
		return nil, fmt.Errorf("unknown compression type: %d", entry.compression)
	}

	// Cache the decompressed hunk (limit cache size to avoid memory issues)
	r.cacheMu.Lock()
	if len(r.hunkCache) < 32 { // Cache up to 32 hunks
		r.hunkCache[hunkNum] = data
	}
	r.cacheMu.Unlock()

	return data, nil
}

// readSector reads a single sector (unit) from the CHD.
// For CD-ROM, this returns 2352 bytes (raw frame) or 2448 bytes (frame + subcode).
// For DVD/other, this returns the unit size specified in the header.
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

// Header returns the CHD header information.
func (r *Reader) Header() *CHDInfo {
	return r.header
}

// Size returns the logical (uncompressed) size in bytes.
func (r *Reader) Size() int64 {
	return int64(r.header.LogicalBytes)
}

// OpenUserData returns an io.ReaderAt suitable for filesystem parsing
// (ISO9660, UDF, etc.) from a CHD file. For CD-ROM CHDs, it handles sector
// translation, extracting 2048-byte user data from raw 2352-byte sectors.
// For DVD/other CHDs, it returns the raw sector data directly.
func OpenUserData(r io.ReaderAt, size int64) (io.ReaderAt, int64, error) {
	reader, err := Open(r, size)
	if err != nil {
		return nil, 0, err
	}

	if reader.header.IsCDROM() {
		// CD-ROM CHD: sectors are 2448 bytes (2352 data + 96 subcode)
		// User data is at offset 16 within the 2352-byte sector data
		isoReader := &sectorReader{
			reader:     reader,
			sectorSize: int64(reader.header.UnitBytes),
			dataOffset: 16, // Mode 1 data offset within sector
		}
		// Calculate logical size (number of sectors * 2048)
		numSectors := int64(reader.header.LogicalBytes) / int64(reader.header.UnitBytes)
		return isoReader, numSectors * 2048, nil
	}

	// DVD/other CHD: sectors are typically 2048 bytes (no translation needed)
	return reader, int64(reader.header.LogicalBytes), nil
}

// sectorReader translates logical 2048-byte sector reads to CHD raw sector reads.
// For CD-ROM CHDs, it extracts user data from raw 2352-byte sectors.
type sectorReader struct {
	reader     *Reader
	sectorSize int64 // Physical sector size (typically 2448 for CD-ROM)
	dataOffset int64 // Offset to user data within physical sector (16 for Mode 1)
}

// ReadAt implements io.ReaderAt, reading logical data from CHD sectors.
func (c *sectorReader) ReadAt(p []byte, off int64) (int, error) {
	n := 0
	for n < len(p) {
		logicalOffset := off + int64(n)

		// Which logical sector (2048-byte)?
		logicalSector := logicalOffset / 2048
		offsetInSector := logicalOffset % 2048

		// Read the physical sector from CHD
		sectorData, err := c.reader.readSector(uint64(logicalSector))
		if err != nil {
			if n > 0 {
				return n, nil
			}
			return 0, err
		}

		// Extract user data starting at dataOffset
		// For CD-ROM Mode 1: bytes 16-2063 of the 2352-byte sector
		dataStart := int(c.dataOffset + offsetInSector)
		dataEnd := int(c.dataOffset) + 2048

		if dataStart >= len(sectorData) {
			if n > 0 {
				return n, nil
			}
			return 0, io.EOF
		}

		if dataEnd > len(sectorData) {
			dataEnd = len(sectorData)
		}

		// Copy available data
		bytesToCopy := min(dataEnd-dataStart, len(p)-n)

		copy(p[n:n+bytesToCopy], sectorData[dataStart:dataStart+bytesToCopy])
		n += bytesToCopy
	}

	return n, nil
}
