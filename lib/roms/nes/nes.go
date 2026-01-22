package nes

import (
	"fmt"
	"io"
)

// NES ROM format parsing (iNES and NES 2.0).
//
// iNES format specification:
// https://www.nesdev.org/wiki/INES
//
// NES 2.0 format specification:
// https://www.nesdev.org/wiki/NES_2.0
//
// Header layout (16 bytes):
//
//	Offset  Size  Description
//	0x00    4     Magic: "NES" + 0x1A
//	0x04    1     PRG-ROM size (16 KB units)
//	0x05    1     CHR-ROM size (8 KB units, 0 = CHR-RAM)
//	0x06    1     Flags 6: Mapper low nibble, mirroring, battery, trainer
//	0x07    1     Flags 7: Mapper high nibble, console type, NES 2.0 identifier
//	0x08    1     PRG-RAM size (8 KB units, 0 = 8KB for compatibility)
//	0x09    1     Flags 9: TV system (0=NTSC, 1=PAL)
//	0x0A    6     Reserved (should be 0)

const (
	nesHeaderSize   = 16
	nesMagicOffset  = 0x00
	nesMagicLen     = 4
	nesPRGROMOffset = 0x04
	nesCHRROMOffset = 0x05
	nesFlags6Offset = 0x06
	nesFlags7Offset = 0x07
	nesPRGRAMOffset = 0x08
	nesFlags9Offset = 0x09
)

// iNES magic bytes: "NES" + 0x1A
var nesMagic = []byte{0x4E, 0x45, 0x53, 0x1A}

// NESConsoleType indicates the console type from flags 7.
type NESConsoleType byte

const (
	NESConsoleNES        NESConsoleType = 0 // NES/Famicom
	NESConsoleVsSystem   NESConsoleType = 1 // Vs. System
	NESConsolePlayChoice NESConsoleType = 2 // PlayChoice-10
	NESConsoleExtended   NESConsoleType = 3 // Extended (NES 2.0)
)

// NESTVSystem indicates NTSC or PAL.
type NESTVSystem byte

const (
	NESTVSystemNTSC NESTVSystem = 0
	NESTVSystemPAL  NESTVSystem = 1
)

// NESMirroring indicates the nametable mirroring type.
type NESMirroring byte

const (
	NESMirroringHorizontal NESMirroring = 0
	NESMirroringVertical   NESMirroring = 1
)

// NESInfo contains metadata extracted from an NES ROM file.
type NESInfo struct {
	PRGROMSize  int          // PRG-ROM size in bytes
	CHRROMSize  int          // CHR-ROM size in bytes (0 = uses CHR-RAM)
	PRGRAMSize  int          // PRG-RAM size in bytes
	Mapper      int          // Mapper number
	Mirroring   NESMirroring // Nametable mirroring
	HasBattery  bool         // Battery-backed save RAM
	HasTrainer  bool         // 512-byte trainer present
	FourScreen  bool         // Four-screen VRAM layout
	ConsoleType NESConsoleType
	TVSystem    NESTVSystem
	IsNES20     bool // True if NES 2.0 format
}

// ParseNES extracts information from an NES ROM file (iNES or NES 2.0 format).
func ParseNES(r io.ReaderAt, size int64) (*NESInfo, error) {
	if size < nesHeaderSize {
		return nil, fmt.Errorf("file too small for NES header: %d bytes", size)
	}

	header := make([]byte, nesHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read NES header: %w", err)
	}

	// Verify magic bytes
	for i := 0; i < nesMagicLen; i++ {
		if header[nesMagicOffset+i] != nesMagic[i] {
			return nil, fmt.Errorf("not a valid NES ROM: magic mismatch")
		}
	}

	flags6 := header[nesFlags6Offset]
	flags7 := header[nesFlags7Offset]
	flags9 := header[nesFlags9Offset]

	// Check for NES 2.0 format: bits 2-3 of flags7 == 2
	isNES20 := (flags7 & 0x0C) == 0x08

	// Extract mapper number
	mapperLow := (flags6 >> 4) & 0x0F
	mapperHigh := flags7 & 0xF0
	mapper := int(mapperHigh) | int(mapperLow)

	// PRG-ROM size (16 KB units)
	prgROMSize := int(header[nesPRGROMOffset]) * 16 * 1024

	// CHR-ROM size (8 KB units, 0 means CHR-RAM is used)
	chrROMSize := int(header[nesCHRROMOffset]) * 8 * 1024

	// PRG-RAM size (8 KB units, 0 means 8 KB for compatibility)
	prgRAMSize := int(header[nesPRGRAMOffset]) * 8 * 1024
	if prgRAMSize == 0 {
		prgRAMSize = 8 * 1024 // Default to 8 KB
	}

	// Extract flags
	mirroring := NESMirroring(flags6 & 0x01)
	hasBattery := (flags6 & 0x02) != 0
	hasTrainer := (flags6 & 0x04) != 0
	fourScreen := (flags6 & 0x08) != 0

	// Console type (bits 0-1 of flags7)
	consoleType := NESConsoleType(flags7 & 0x03)

	// TV system (bit 0 of flags9)
	tvSystem := NESTVSystem(flags9 & 0x01)

	return &NESInfo{
		PRGROMSize:  prgROMSize,
		CHRROMSize:  chrROMSize,
		PRGRAMSize:  prgRAMSize,
		Mapper:      mapper,
		Mirroring:   mirroring,
		HasBattery:  hasBattery,
		HasTrainer:  hasTrainer,
		FourScreen:  fourScreen,
		ConsoleType: consoleType,
		TVSystem:    tvSystem,
		IsNES20:     isNES20,
	}, nil
}
