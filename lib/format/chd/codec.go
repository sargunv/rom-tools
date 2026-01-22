package chd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz/lzma"
)

// decompressHunk decompresses a single hunk using the appropriate codec.
func decompressHunk(compressedData []byte, codecID Codec, hunkBytes uint32) ([]byte, error) {
	switch codecID {
	case CodecNone:
		// Uncompressed - just return a copy
		result := make([]byte, hunkBytes)
		copy(result, compressedData)
		return result, nil

	case CodecZlib:
		return decompressZlib(compressedData, hunkBytes)

	case CodecLZMA:
		return decompressLZMA(compressedData, hunkBytes)

	case CodecHuff:
		return decompressHuffman(compressedData, hunkBytes)

	case CodecZstd:
		return decompressZstd(compressedData, hunkBytes)

	case CodecCDZlib:
		return decompressCDZlib(compressedData, hunkBytes)

	case CodecCDLZMA:
		return decompressCDLZMA(compressedData, hunkBytes)

	case CodecCDZstd:
		return decompressCDZstd(compressedData, hunkBytes)

	case CodecFLAC, CodecCDFLAC:
		// FLAC is for audio tracks - we don't need to decompress audio for identification
		return nil, fmt.Errorf("FLAC codec not supported (audio only)")

	default:
		return nil, fmt.Errorf("unknown codec: 0x%08x", codecID)
	}
}

// decompressZlib decompresses raw zlib/deflate data.
func decompressZlib(data []byte, outputSize uint32) ([]byte, error) {
	return decompressZlibRaw(data, int(outputSize))
}

// decompressLZMA decompresses raw LZMA data (no header) as used by CHD.
// CHD stores raw LZMA compressed data without the standard 13-byte header.
// The decoder properties are derived from compression level 9 with
// reduceSize = hunkbytes, which gives us: lc=3, lp=0, pb=2 (defaults).
func decompressLZMA(data []byte, outputSize uint32) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("LZMA data empty")
	}

	// CHD LZMA uses default properties: lc=3, lp=0, pb=2
	// Properties byte = (pb * 5 + lp) * 9 + lc = (2*5 + 0)*9 + 3 = 93 = 0x5D
	propsByte := byte(0x5D)

	// Dictionary size: CHD uses reduceSize=hunkbytes which limits dictionary.
	// For small hunks, use the output size as dictionary. For larger, use 64KB.
	dictSize := uint32(65536)
	if outputSize < dictSize {
		// Round up to next power of 2
		dictSize = outputSize
		dictSize--
		dictSize |= dictSize >> 1
		dictSize |= dictSize >> 2
		dictSize |= dictSize >> 4
		dictSize |= dictSize >> 8
		dictSize |= dictSize >> 16
		dictSize++
		if dictSize < 4096 {
			dictSize = 4096 // LZMA minimum
		}
	}

	// Build standard LZMA header:
	// [0]   Properties byte
	// [1-4] Dictionary size (little-endian)
	// [5-12] Uncompressed size (little-endian), -1 if unknown
	header := make([]byte, 13)
	header[0] = propsByte
	binary.LittleEndian.PutUint32(header[1:5], dictSize)
	binary.LittleEndian.PutUint64(header[5:13], uint64(outputSize))

	// Prepend header to compressed data
	fullData := append(header, data...)

	r, err := lzma.NewReader(bytes.NewReader(fullData))
	if err != nil {
		return nil, fmt.Errorf("LZMA reader: %w", err)
	}

	result := make([]byte, outputSize)
	n, err := io.ReadFull(r, result)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("LZMA decompress: %w", err)
	}
	return result[:n], nil
}

// decompressHuffman decompresses CHD Huffman-encoded data.
func decompressHuffman(data []byte, outputSize uint32) ([]byte, error) {
	// CHD Huffman uses 8-bit symbols (256 codes)
	hd := newHuffmanDecoder(256, huffmanMaxBits)
	br := newBitReader(data)

	// Import the Huffman tree from RLE-encoded data at start of stream
	if err := hd.importTreeRLE(br); err != nil {
		return nil, fmt.Errorf("huffman tree import: %w", err)
	}

	// Decode the data
	result := make([]byte, outputSize)
	for i := range outputSize {
		sym, err := hd.decode(br)
		if err != nil {
			return nil, fmt.Errorf("huffman decode at %d: %w", i, err)
		}
		result[i] = byte(sym)
	}

	return result, nil
}

var zstdDecoder *zstd.Decoder

func init() {
	var err error
	zstdDecoder, err = zstd.NewReader(nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create zstd decoder: %v", err))
	}
}

// decompressZstd decompresses Zstandard data.
func decompressZstd(data []byte, outputSize uint32) ([]byte, error) {
	return decompressZstdRaw(data, int(outputSize))
}
