package codec

import (
	"fmt"
)

// CD-ROM frame constants (from libchdr/cdrom.h)
const (
	cdMaxSectorData  = 2352 // Raw sector data size
	cdMaxSubcodeData = 96   // Subcode data per frame
	cdFrameSize      = 2448 // cdMaxSectorData + cdMaxSubcodeData
)

// CDZLIB decompresses CD-ROM data using zlib for the base codec.
func CDZLIB(data []byte, hunkBytes uint32) ([]byte, error) {
	return decompressCDCodec(data, hunkBytes, Zlib, "zlib")
}

// CDLZMA decompresses CD-ROM data using LZMA for the base codec.
func CDLZMA(data []byte, hunkBytes uint32) ([]byte, error) {
	return decompressCDCodec(data, hunkBytes, LZMA, "lzma")
}

// CDZstd decompresses CD-ROM data using Zstd for the base codec.
func CDZstd(data []byte, hunkBytes uint32) ([]byte, error) {
	return decompressCDCodec(data, hunkBytes, Zstd, "zstd")
}

// decompressCDCodec is the common implementation for CD codecs.
// Format: [ECC bitmap] [compressed base length] [base data (sector)] [subcode data (zlib)]
func decompressCDCodec(data []byte, hunkBytes uint32, baseDecompress func([]byte, int) ([]byte, error), codecName string) ([]byte, error) {
	// Calculate frame count
	frames := int(hunkBytes / cdFrameSize)
	if frames == 0 {
		return nil, fmt.Errorf("CD codec: invalid hunk size %d", hunkBytes)
	}

	// Header format:
	// - ECC bitmap: (frames + 7) / 8 bytes
	// - Compressed base length: 2 bytes if hunkBytes < 65536, else 3 bytes
	eccBytes := (frames + 7) / 8
	complenBytes := 2
	if hunkBytes >= 65536 {
		complenBytes = 3
	}
	headerBytes := eccBytes + complenBytes

	if len(data) < headerBytes {
		return nil, fmt.Errorf("CD codec: data too short for header (need %d, have %d)", headerBytes, len(data))
	}

	// Extract compressed base length
	var complenBase int
	if complenBytes == 2 {
		complenBase = int(data[eccBytes])<<8 | int(data[eccBytes+1])
	} else {
		complenBase = int(data[eccBytes])<<16 | int(data[eccBytes+1])<<8 | int(data[eccBytes+2])
	}

	if len(data) < headerBytes+complenBase {
		return nil, fmt.Errorf("CD codec: data too short for base (need %d, have %d)", headerBytes+complenBase, len(data))
	}

	// Decompress base (sector) data - outputs cdMaxSectorData (2352) bytes per frame
	baseCompressed := data[headerBytes : headerBytes+complenBase]
	expectedBaseSize := frames * cdMaxSectorData
	baseData, err := baseDecompress(baseCompressed, expectedBaseSize)
	if err != nil {
		return nil, fmt.Errorf("CD codec base decompress (%s): %w", codecName, err)
	}

	// Decompress subcode data (always zlib) - outputs cdMaxSubcodeData (96) bytes per frame
	subcodeCompressed := data[headerBytes+complenBase:]
	expectedSubcodeSize := frames * cdMaxSubcodeData
	var subcodeData []byte
	if len(subcodeCompressed) > 0 {
		subcodeData, err = Zlib(subcodeCompressed, expectedSubcodeSize)
		if err != nil {
			return nil, fmt.Errorf("CD codec subcode decompress: %w", err)
		}
	} else {
		// No subcode data - fill with zeros
		subcodeData = make([]byte, expectedSubcodeSize)
	}

	// Reconstruct frames: interleave sector data and subcode data
	result := make([]byte, hunkBytes)
	for i := range frames {
		// Copy sector data (2352 bytes)
		srcOffset := i * cdMaxSectorData
		dstOffset := i * cdFrameSize
		if srcOffset+cdMaxSectorData <= len(baseData) {
			copy(result[dstOffset:], baseData[srcOffset:srcOffset+cdMaxSectorData])
		}

		// Copy subcode data (96 bytes)
		srcSubOffset := i * cdMaxSubcodeData
		dstSubOffset := dstOffset + cdMaxSectorData
		if srcSubOffset+cdMaxSubcodeData <= len(subcodeData) {
			copy(result[dstSubOffset:], subcodeData[srcSubOffset:srcSubOffset+cdMaxSubcodeData])
		}
	}

	return result, nil
}
