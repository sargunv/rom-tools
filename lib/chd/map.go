package chd

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/lib/chd/internal/codec"
)

// V5 map compression types (from libchdr)
const (
	compressionType0      = 0  // Use compressor 0
	compressionType1      = 1  // Use compressor 1
	compressionType2      = 2  // Use compressor 2
	compressionType3      = 3  // Use compressor 3
	compressionNone       = 4  // Uncompressed
	compressionSelf       = 5  // Reference to another hunk in this file
	compressionParent     = 6  // Reference to hunk in parent file
	compressionRLESmall   = 7  // RLE with small count (Huffman symbol)
	compressionRLELarge   = 8  // RLE with large count (Huffman symbol)
	compressionSelf0      = 9  // Self reference, offset +0
	compressionSelf1      = 10 // Self reference, offset +1
	compressionParentSelf = 11 // Parent reference, same offset as self
	compressionParent0    = 12 // Parent reference, offset +0
	compressionParent1    = 13 // Parent reference, offset +1
)

// mapEntry represents a single hunk's location and compression info.
type mapEntry struct {
	compression uint8  // Compression type (0-6)
	length      uint32 // Compressed length in bytes
	offset      uint64 // Offset in file (or hunk number for self-reference)
	crc16       uint16 // CRC of uncompressed data
}

// chdMap contains the decoded hunk map for a CHD file.
type chdMap struct {
	entries []mapEntry
}

// V5 map header structure (16 bytes):
//
//	[0-3]   uint32 length      - compressed map data size
//	[4-9]   uint48 firstoffs   - offset of first block
//	[10-11] uint16 mapcrc      - CRC16 of decompressed map
//	[12]    uint8  lengthbits  - bits for length field
//	[13]    uint8  selfbits    - bits for self-reference offset
//	[14]    uint8  parentbits  - bits for parent reference offset
//	[15]    uint8  reserved
const mapHeaderSize = 16

// decodeMap reads and decompresses the V5 hunk map.
func decodeMap(r io.ReaderAt, header *Header) (*chdMap, error) {
	if header.TotalHunks == 0 {
		return &chdMap{entries: []mapEntry{}}, nil
	}

	// Read map header
	mapHeader := make([]byte, mapHeaderSize)
	if _, err := r.ReadAt(mapHeader, int64(header.MapOffset)); err != nil {
		return nil, fmt.Errorf("failed to read map header: %w", err)
	}

	compressedLen := binary.BigEndian.Uint32(mapHeader[0:4])
	firstOffset := readUint48BE(mapHeader[4:10])
	mapCRC := binary.BigEndian.Uint16(mapHeader[10:12])
	lengthBits := mapHeader[12]
	selfBits := mapHeader[13]
	parentBits := mapHeader[14]

	// Read compressed map data
	compressedData := make([]byte, compressedLen)
	if _, err := r.ReadAt(compressedData, int64(header.MapOffset)+mapHeaderSize); err != nil {
		return nil, fmt.Errorf("failed to read compressed map: %w", err)
	}

	// Decode using Huffman + RLE
	entries, err := decodeMapEntries(compressedData, header, lengthBits, selfBits, parentBits, firstOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to decode map entries: %w", err)
	}

	// Verify CRC
	if crc := calculateMapCRC(entries); crc != mapCRC {
		return nil, fmt.Errorf("map CRC mismatch: got %04x, want %04x", crc, mapCRC)
	}

	return &chdMap{entries: entries}, nil
}

func decodeMapEntries(data []byte, header *Header, lengthBits, selfBits, parentBits uint8, firstOffset uint64) ([]mapEntry, error) {
	br := codec.NewBitReader(data)
	numHunks := header.TotalHunks

	// Create Huffman decoder for compression types (16 symbols, max 8 bits)
	huffman := codec.NewHuffmanDecoder(16, 8)
	if err := huffman.ImportTreeRLE(br); err != nil {
		return nil, fmt.Errorf("failed to import huffman tree: %w", err)
	}

	// Phase 1: Decode compression types for all hunks
	compressionTypes := make([]uint8, numHunks)
	var lastComp uint8
	repCount := 0

	for hunkNum := range numHunks {
		if repCount > 0 {
			compressionTypes[hunkNum] = lastComp
			repCount--
			continue
		}

		val, err := huffman.Decode(br)
		if err != nil {
			return nil, fmt.Errorf("failed to decode compression type for hunk %d: %w", hunkNum, err)
		}

		switch val {
		case compressionRLESmall:
			// RLE small: repcount = 2 + huffman_decode_one()
			count, err := huffman.Decode(br)
			if err != nil {
				return nil, fmt.Errorf("failed to read RLE small count: %w", err)
			}
			compressionTypes[hunkNum] = lastComp
			repCount = int(2 + count)

		case compressionRLELarge:
			// RLE large: repcount = 2 + 16 + (huffman_decode_one() << 4) + huffman_decode_one()
			high, err := huffman.Decode(br)
			if err != nil {
				return nil, fmt.Errorf("failed to read RLE large high: %w", err)
			}
			low, err := huffman.Decode(br)
			if err != nil {
				return nil, fmt.Errorf("failed to read RLE large low: %w", err)
			}
			compressionTypes[hunkNum] = lastComp
			repCount = int(2 + 16 + (high << 4) + low)

		default:
			compressionTypes[hunkNum] = uint8(val)
			lastComp = uint8(val)
		}
	}

	// Phase 2: Extract offset/length/CRC data for each hunk
	entries := make([]mapEntry, numHunks)
	curOffset := firstOffset
	var lastSelf uint64
	var lastParent uint64

	for hunkNum := range numHunks {
		entry := &entries[hunkNum]
		compType := compressionTypes[hunkNum]

		switch compType {
		case compressionType0, compressionType1, compressionType2, compressionType3:
			// Compressed hunk: read length and CRC
			length, err := br.ReadBits(uint32(lengthBits))
			if err != nil {
				return nil, fmt.Errorf("failed to read length for hunk %d: %w", hunkNum, err)
			}
			crc, err := br.ReadBits(16)
			if err != nil {
				return nil, fmt.Errorf("failed to read CRC for hunk %d: %w", hunkNum, err)
			}
			entry.compression = compType
			entry.length = length
			entry.offset = curOffset
			entry.crc16 = uint16(crc)
			curOffset += uint64(length)

		case compressionNone:
			// Uncompressed hunk: length is hunkbytes
			crc, err := br.ReadBits(16)
			if err != nil {
				return nil, fmt.Errorf("failed to read CRC for hunk %d: %w", hunkNum, err)
			}
			entry.compression = compressionNone
			entry.length = header.HunkBytes
			entry.offset = curOffset
			entry.crc16 = uint16(crc)
			curOffset += uint64(header.HunkBytes)

		case compressionSelf:
			// Self reference: read hunk number
			selfOffset, err := br.ReadBits(uint32(selfBits))
			if err != nil {
				return nil, fmt.Errorf("failed to read self offset for hunk %d: %w", hunkNum, err)
			}
			entry.compression = compressionSelf
			entry.offset = uint64(selfOffset)
			lastSelf = uint64(selfOffset)

		case compressionSelf0:
			// Self reference with offset +0
			entry.compression = compressionSelf
			entry.offset = lastSelf

		case compressionSelf1:
			// Self reference with offset +1
			lastSelf++
			entry.compression = compressionSelf
			entry.offset = lastSelf

		case compressionParent:
			// Parent reference: read unit number
			parentOffset, err := br.ReadBits(uint32(parentBits))
			if err != nil {
				return nil, fmt.Errorf("failed to read parent offset for hunk %d: %w", hunkNum, err)
			}
			entry.compression = compressionParent
			entry.offset = uint64(parentOffset)
			lastParent = uint64(parentOffset)

		case compressionParent0:
			// Parent reference with offset +0
			entry.compression = compressionParent
			entry.offset = lastParent

		case compressionParent1:
			// Parent reference: offset += hunkbytes/unitbytes
			lastParent += uint64(header.HunkBytes / header.UnitBytes)
			entry.compression = compressionParent
			entry.offset = lastParent

		case compressionParentSelf:
			// Parent reference using current hunk's logical offset
			offset := (uint64(hunkNum) * uint64(header.HunkBytes)) / uint64(header.UnitBytes)
			entry.compression = compressionParent
			entry.offset = offset
			lastParent = offset

		default:
			return nil, fmt.Errorf("unknown compression type %d for hunk %d", compType, hunkNum)
		}
	}

	return entries, nil
}

// readUint48BE reads a 48-bit big-endian unsigned integer.
func readUint48BE(b []byte) uint64 {
	return uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 |
		uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5])
}

// calculateMapCRC calculates CRC16 of the decompressed map entries.
// Each entry is serialized to 12 bytes before CRC calculation.
func calculateMapCRC(entries []mapEntry) uint16 {
	// Serialize entries to bytes (12 bytes each)
	data := make([]byte, len(entries)*12)
	for i, e := range entries {
		off := i * 12
		data[off+0] = e.compression
		// Length as 24-bit big-endian
		data[off+1] = byte(e.length >> 16)
		data[off+2] = byte(e.length >> 8)
		data[off+3] = byte(e.length)
		// Offset as 48-bit big-endian
		data[off+4] = byte(e.offset >> 40)
		data[off+5] = byte(e.offset >> 32)
		data[off+6] = byte(e.offset >> 24)
		data[off+7] = byte(e.offset >> 16)
		data[off+8] = byte(e.offset >> 8)
		data[off+9] = byte(e.offset)
		// CRC as 16-bit big-endian
		data[off+10] = byte(e.crc16 >> 8)
		data[off+11] = byte(e.crc16)
	}
	return crc16(data)
}

// crc16 calculates CRC-16-CCITT (polynomial 0x1021).
func crc16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for range 8 {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
