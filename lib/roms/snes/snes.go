package snes

import (
	"fmt"
	"io"
)

// SNES ROM format parsing.
//
// SNES ROM header specification:
// https://snes.nesdev.org/wiki/ROM_header
//
// The internal header is located at different offsets depending on mapping mode:
// - LoROM: 0x7FC0
// - HiROM: 0xFFC0
// - ExHiROM: 0x40FFC0
//
// Internal header layout (32 bytes):
//
//	Offset  Size  Description
//	0x00    21    Game title (ASCII, space-padded)
//	0x15    1     Map mode
//	0x16    1     ROM type (cartridge type)
//	0x17    1     ROM size (log2 of kilobytes)
//	0x18    1     SRAM size (log2 of kilobytes)
//	0x19    1     Destination code (region)
//	0x1A    1     Old licensee code (0x33 = use extended header)
//	0x1B    1     Version number
//	0x1C    2     Checksum complement
//	0x1E    2     Checksum

const (
	snesHeaderSize      = 32
	snesTitleOffset     = 0x00
	snesTitleLen        = 21
	snesMapModeOffset   = 0x15
	snesROMTypeOffset   = 0x16
	snesROMSizeOffset   = 0x17
	snesSRAMSizeOffset  = 0x18
	snesDestCodeOffset  = 0x19
	snesLicenseeOffset  = 0x1A
	snesVersionOffset   = 0x1B
	snesChecksumCOffset = 0x1C
	snesChecksumOffset  = 0x1E

	// Header offsets for different mapping modes (without copier header)
	snesLoROMOffset   = 0x7FC0
	snesHiROMOffset   = 0xFFC0
	snesExHiROMOffset = 0x40FFC0

	// Copier header size (some ROMs have this prepended)
	snesCopierHeaderSize = 512
)

// SNESMapMode indicates the memory mapping mode.
type SNESMapMode byte

const (
	SNESMapModeLoROM     SNESMapMode = 0x20
	SNESMapModeHiROM     SNESMapMode = 0x21
	SNESMapModeLoROMSA1  SNESMapMode = 0x23 // SA-1
	SNESMapModeExLoROM   SNESMapMode = 0x30 // ExLoROM
	SNESMapModeExHiROM   SNESMapMode = 0x31 // ExHiROM
	SNESMapModeHiROMSPC  SNESMapMode = 0x35 // SPC7110
	SNESMapModeHiROMSPC2 SNESMapMode = 0x3A // SPC7110 variant (Tengai Makyou Zero)
)

// SNESInfo contains metadata extracted from a SNES ROM file.
type SNESInfo struct {
	Title              string
	MapMode            SNESMapMode
	ROMType            byte // Cartridge type (coprocessor info)
	ROMSize            int  // ROM size in bytes
	SRAMSize           int  // SRAM size in bytes
	DestinationCode    byte // Region code
	LicenseeCode       byte
	Version            int
	Checksum           uint16
	ChecksumComplement uint16
	HasCopierHeader    bool // True if 512-byte copier header was detected
}

// ParseSNES extracts information from a SNES ROM file.
func ParseSNES(r io.ReaderAt, size int64) (*SNESInfo, error) {
	// Determine if there's a copier header (file size % 1024 == 512)
	hasCopierHeader := (size % 1024) == snesCopierHeaderSize
	copierOffset := int64(0)
	if hasCopierHeader {
		copierOffset = snesCopierHeaderSize
	}

	// Calculate all three possible header offsets
	offsets := []int64{
		copierOffset + snesLoROMOffset,   // LoROM
		copierOffset + snesHiROMOffset,   // HiROM
		copierOffset + snesExHiROMOffset, // ExHiROM
	}

	// Try each offset and return the first valid header
	for _, offset := range offsets {
		if offset+snesHeaderSize <= size {
			info, err := parseSNESHeader(r, offset, size, hasCopierHeader)
			if err == nil && isValidSNESHeader(info, size) {
				return info, nil
			}
		}
	}

	return nil, fmt.Errorf("could not find valid SNES header")
}

func parseSNESHeader(r io.ReaderAt, offset int64, fileSize int64, hasCopierHeader bool) (*SNESInfo, error) {
	header := make([]byte, snesHeaderSize)
	if _, err := r.ReadAt(header, offset); err != nil {
		return nil, fmt.Errorf("failed to read SNES header: %w", err)
	}

	// Extract raw title bytes for validation
	titleBytes := header[snesTitleOffset : snesTitleOffset+snesTitleLen]

	// Extract title (ASCII, space-padded)
	title := extractSNESTitle(titleBytes)

	// Map mode
	mapMode := SNESMapMode(header[snesMapModeOffset])

	// ROM type
	romType := header[snesROMTypeOffset]

	// ROM size: 1 << (header value) kilobytes
	romSizeExp := header[snesROMSizeOffset]
	romSize := 0
	if romSizeExp > 0 && romSizeExp < 16 {
		romSize = (1 << romSizeExp) * 1024
	}

	// SRAM size: 1 << (header value) kilobytes (0 = no SRAM)
	sramSizeExp := header[snesSRAMSizeOffset]
	sramSize := 0
	if sramSizeExp > 0 && sramSizeExp < 16 {
		sramSize = (1 << sramSizeExp) * 1024
	}

	// Destination code
	destCode := header[snesDestCodeOffset]

	// Licensee code
	licenseeCode := header[snesLicenseeOffset]

	// Version
	version := int(header[snesVersionOffset])

	// Checksum complement and checksum (little-endian)
	checksumC := uint16(header[snesChecksumCOffset]) | uint16(header[snesChecksumCOffset+1])<<8
	checksum := uint16(header[snesChecksumOffset]) | uint16(header[snesChecksumOffset+1])<<8

	return &SNESInfo{
		Title:              title,
		MapMode:            mapMode,
		ROMType:            romType,
		ROMSize:            romSize,
		SRAMSize:           sramSize,
		DestinationCode:    destCode,
		LicenseeCode:       licenseeCode,
		Version:            version,
		Checksum:           checksum,
		ChecksumComplement: checksumC,
		HasCopierHeader:    hasCopierHeader,
	}, nil
}

// isValidSNESHeader checks if the header looks valid using multiple heuristics.
func isValidSNESHeader(info *SNESInfo, fileSize int64) bool {
	// 1. Checksum validation: checksum + complement should equal 0xFFFF
	//    This is the strongest signal (1 in 65,536 chance of random data passing)
	if info.Checksum+info.ChecksumComplement != 0xFFFF {
		return false
	}

	// 2. Map mode should be in valid range (0x20-0x3F)
	//    All known map modes have high nibble 0x2 or 0x3
	if info.MapMode < 0x20 || info.MapMode > 0x3F {
		return false
	}

	// 3. Title should have at least some printable ASCII characters
	//    Random code/data rarely looks like text
	printableCount := 0
	for _, c := range info.Title {
		if c >= 0x20 && c <= 0x7E {
			printableCount++
		}
	}
	if printableCount < 2 {
		return false
	}

	// 4. Declared ROM size should be reasonable compared to file size
	//    Allow declared size to be up to 2x file size (some headers are inaccurate)
	if info.ROMSize > 0 && int64(info.ROMSize) > fileSize*2 {
		return false
	}

	return true
}

// extractSNESTitle extracts and cleans a SNES title string.
func extractSNESTitle(data []byte) string {
	// Find end of valid ASCII characters
	end := len(data)
	for i, b := range data {
		// SNES titles should be printable ASCII (0x20-0x7E) or space-padded
		if b == 0 || b < 0x20 || b > 0x7E {
			end = i
			break
		}
	}

	// Trim trailing spaces
	for end > 0 && data[end-1] == ' ' {
		end--
	}

	return string(data[:end])
}
