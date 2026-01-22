package chd

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz/lzma"
)

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
		return decompressZlib(compressedData, size)

	case CodecLZMA:
		return decompressLZMA(compressedData, size)

	case CodecHuff:
		return decompressHuffman(compressedData, size)

	case CodecZstd:
		return decompressZstd(compressedData, size)

	case CodecCDZlib:
		return decompressCDCodec(compressedData, hunkBytes, decompressZlib, "zlib")

	case CodecCDLZMA:
		return decompressCDCodec(compressedData, hunkBytes, decompressLZMA, "lzma")

	case CodecCDZstd:
		return decompressCDCodec(compressedData, hunkBytes, decompressZstd, "zstd")

	case CodecFLAC, CodecCDFLAC:
		// FLAC is for audio tracks - we don't need to decompress audio for identification
		return nil, fmt.Errorf("FLAC codec not supported (audio only)")

	default:
		return nil, fmt.Errorf("unknown codec: 0x%08x", codecID)
	}
}

// decompressZlib decompresses raw zlib/flate compressed data.
func decompressZlib(data []byte, outputSize int) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()

	result := make([]byte, outputSize)
	n, err := io.ReadFull(r, result)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return result[:n], nil
}

// decompressLZMA decompresses raw LZMA data (no header) as used by CHD.
// CHD stores raw LZMA compressed data without the standard 13-byte header.
func decompressLZMA(data []byte, outputSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("LZMA data empty")
	}

	// CHD LZMA uses default properties: lc=3, lp=0, pb=2
	// Properties byte = (pb * 5 + lp) * 9 + lc = (2*5 + 0)*9 + 3 = 93 = 0x5D
	const propsByte = 0x5D

	// Dictionary size: use output size or 64KB, whichever is larger.
	// LZMA decoders accept dictionaries larger than needed.
	dictSize := uint32(65536)
	if outputSize > 65536 {
		dictSize = uint32(outputSize)
	}

	// Build standard LZMA header:
	// [0]   Properties byte
	// [1-4] Dictionary size (little-endian)
	// [5-12] Uncompressed size (little-endian)
	header := make([]byte, 13)
	header[0] = propsByte
	binary.LittleEndian.PutUint32(header[1:5], dictSize)
	binary.LittleEndian.PutUint64(header[5:13], uint64(outputSize))

	r, err := lzma.NewReader(bytes.NewReader(append(header, data...)))
	if err != nil {
		return nil, err
	}

	result := make([]byte, outputSize)
	n, err := io.ReadFull(r, result)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return result[:n], nil
}

// decompressZstd decompresses Zstandard compressed data.
func decompressZstd(data []byte, outputSize int) ([]byte, error) {
	result, err := zstdDecoder.DecodeAll(data, make([]byte, 0, outputSize))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// decompressHuffman decompresses CHD Huffman-encoded data.
func decompressHuffman(data []byte, outputSize int) ([]byte, error) {
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
