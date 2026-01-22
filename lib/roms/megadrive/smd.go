package megadrive

import (
	"fmt"
	"io"
)

// SMD (Super Magic Drive Interleaved) ROM format parsing.
//
// SMD format specification:
// https://raw.githubusercontent.com/franckverrot/EmulationResources/refs/heads/master/consoles/megadrive/genesis_rom.txt
//
// SMD files have a 512-byte header followed by interleaved 16KB blocks.
//
// Header layout:
//
//	Offset  Size  Description
//	0       1     Number of 16KB blocks (0x00 if > 255)
//	1       1     0x03 (fixed magic byte)
//	2       1     Split flag (0x00 = last/only part, 0x40 = more parts follow)
//	3-7     5     0x00 (padding)
//	8       1     0xAA (fixed magic byte)
//	9       1     0xBB (fixed magic byte)
//	10-511  502   0x00 (padding)
//
// Data interleaving: Each 16KB block has even-position bytes in the first 8KB,
// odd-position bytes in the second 8KB.

const (
	smdHeaderSize = 512
	smdBlockSize  = 16384 // 16KB blocks
	smdHalfBlock  = 8192  // Half of a block (for interleaving)
	smdMagicByte1 = 0x03  // Fixed value at offset 1
	smdMagicByte8 = 0xAA  // Fixed value at offset 8
	smdMagicByte9 = 0xBB  // Fixed value at offset 9
)

// isSMDROM checks if the file has an SMD header.
// SMD files have a 512-byte header with specific magic bytes.
func isSMDROM(r io.ReaderAt, size int64) bool {
	if size < smdHeaderSize+smdBlockSize {
		return false
	}

	header := make([]byte, smdHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return false
	}

	// Check magic bytes
	if header[1] != smdMagicByte1 {
		return false
	}
	if header[8] != smdMagicByte8 {
		return false
	}
	if header[9] != smdMagicByte9 {
		return false
	}

	// Additional validation: most header bytes should be 0x00
	// Check bytes 3-7 (skip 0=block count, 1=magic, 2=split flag)
	for i := 3; i < 8; i++ {
		if header[i] != 0x00 {
			return false
		}
	}

	return true
}

// deinterleaveSMDBlock de-interleaves a single 16KB SMD block.
// In SMD format, even bytes are in the first half, odd bytes in the second half.
func deinterleaveSMDBlock(block []byte) []byte {
	if len(block) != smdBlockSize {
		return block
	}

	result := make([]byte, smdBlockSize)
	for i := 0; i < smdHalfBlock; i++ {
		result[i*2] = block[smdHalfBlock+i] // Odd bytes from second half
		result[i*2+1] = block[i]            // Even bytes from first half
	}
	return result
}

// deinterleaveSMD de-interleaves an entire SMD ROM (excluding header).
func deinterleaveSMD(data []byte) []byte {
	numBlocks := len(data) / smdBlockSize
	result := make([]byte, 0, len(data))

	for i := 0; i < numBlocks; i++ {
		start := i * smdBlockSize
		end := start + smdBlockSize
		block := deinterleaveSMDBlock(data[start:end])
		result = append(result, block...)
	}

	// Handle any remaining bytes (shouldn't happen with valid SMD files)
	remaining := len(data) % smdBlockSize
	if remaining > 0 {
		result = append(result, data[numBlocks*smdBlockSize:]...)
	}

	return result
}

// parseSMD extracts game information from an SMD (Super Magic Drive) ROM file.
// SMD files have a 512-byte header and interleaved data that needs de-interleaving.
func parseSMD(r io.ReaderAt, size int64) (*MDInfo, error) {
	if size < smdHeaderSize+smdBlockSize {
		return nil, fmt.Errorf("file too small for SMD format: %d bytes", size)
	}

	// Validate SMD format
	if !isSMDROM(r, size) {
		return nil, fmt.Errorf("not a valid SMD ROM")
	}

	// Read ROM data (after header)
	romSize := size - smdHeaderSize
	romData := make([]byte, romSize)
	if _, err := r.ReadAt(romData, smdHeaderSize); err != nil {
		return nil, fmt.Errorf("failed to read SMD ROM data: %w", err)
	}

	// De-interleave the ROM data
	deinterleaved := deinterleaveSMD(romData)

	// Parse as a regular MD ROM using the de-interleaved data
	info, err := parseMDBytes(deinterleaved)
	if err != nil {
		return nil, err
	}
	info.SourceFormat = FormatSMD
	return info, nil
}
