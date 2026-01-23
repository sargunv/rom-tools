package nes

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/lib/core"
)

// NES ROM format parsing (iNES and NES 2.0).
//
// NES 2.0 format specification:
// https://www.nesdev.org/wiki/NES_2.0
//
// iNES format specification:
// https://www.nesdev.org/wiki/INES
//
// Header layout (16 bytes):
//
//	Offset  Size  Description
//	0x00    4     Magic: "NES" + 0x1A
//	0x04    1     PRG-ROM size LSB (16 KB units for iNES 1.0)
//	0x05    1     CHR-ROM size LSB (8 KB units for iNES 1.0, 0 = CHR-RAM)
//	0x06    1     Flags 6: Mapper low nibble, mirroring, battery, trainer, four-screen
//	0x07    1     Flags 7: Mapper high nibble, console type, NES 2.0 identifier
//	0x08    1     NES 2.0: Mapper MSB + Submapper / iNES 1.0: PRG-RAM size (8 KB units)
//	0x09    1     NES 2.0: PRG/CHR-ROM size MSB / iNES 1.0: TV system
//	0x0A    1     NES 2.0: PRG-RAM/NVRAM sizes (shift counts)
//	0x0B    1     NES 2.0: CHR-RAM/NVRAM sizes (shift counts)
//	0x0C    1     NES 2.0: CPU/PPU timing mode
//	0x0D    1     NES 2.0: Vs. System type or extended console type
//	0x0E    1     NES 2.0: Miscellaneous ROMs
//	0x0F    1     NES 2.0: Default expansion device

const (
	nesHeaderSize = 16
)

// iNES magic bytes: "NES" + 0x1A
var nesMagic = []byte{0x4E, 0x45, 0x53, 0x1A}

// NESMirroring indicates the nametable mirroring mode.
type NESMirroring byte

const (
	// NESMirroringHorizontal indicates horizontal nametable arrangement (vertical mirroring).
	NESMirroringHorizontal NESMirroring = 0
	// NESMirroringVertical indicates vertical nametable arrangement (horizontal mirroring).
	NESMirroringVertical NESMirroring = 1
)

// NESConsoleType indicates the target console from flags 7 bits 0-1.
type NESConsoleType byte

const (
	// NESConsoleNES indicates NES/Famicom console.
	NESConsoleNES NESConsoleType = 0
	// NESConsoleVsSystem indicates Vs. System arcade hardware.
	NESConsoleVsSystem NESConsoleType = 1
	// NESConsolePlayChoice indicates PlayChoice-10 arcade hardware.
	NESConsolePlayChoice NESConsoleType = 2
	// NESConsoleExtended indicates extended console type (check ExtendedConsoleType field).
	NESConsoleExtended NESConsoleType = 3
)

// NESTimingMode indicates the CPU/PPU timing region.
type NESTimingMode byte

const (
	// NESTimingNTSC indicates RP2C02 PPU (NTSC NES).
	NESTimingNTSC NESTimingMode = 0
	// NESTimingPAL indicates RP2C07 PPU (licensed PAL NES).
	NESTimingPAL NESTimingMode = 1
	// NESTimingMulti indicates multiple-region support.
	NESTimingMulti NESTimingMode = 2
	// NESTimingDendy indicates UA6538 PPU (Dendy clone).
	NESTimingDendy NESTimingMode = 3
)

// NESVsPPUType indicates the Vs. System PPU variant.
type NESVsPPUType byte

const (
	NESVsPPURP2C03B     NESVsPPUType = 0x00 // RP2C03B
	NESVsPPURP2C03G     NESVsPPUType = 0x01 // RP2C03G
	NESVsPPURP2C04_0001 NESVsPPUType = 0x02 // RP2C04-0001
	NESVsPPURP2C04_0002 NESVsPPUType = 0x03 // RP2C04-0002
	NESVsPPURP2C04_0003 NESVsPPUType = 0x04 // RP2C04-0003
	NESVsPPURP2C04_0004 NESVsPPUType = 0x05 // RP2C04-0004
	NESVsPPURC2C03B     NESVsPPUType = 0x06 // RC2C03B
	NESVsPPURC2C03C     NESVsPPUType = 0x07 // RC2C03C
	NESVsPPURC2C05_01   NESVsPPUType = 0x08 // RC2C05-01
	NESVsPPURC2C05_02   NESVsPPUType = 0x09 // RC2C05-02
	NESVsPPURC2C05_03   NESVsPPUType = 0x0A // RC2C05-03
	NESVsPPURC2C05_04   NESVsPPUType = 0x0B // RC2C05-04
	NESVsPPURC2C05_05   NESVsPPUType = 0x0C // RC2C05-05
)

// NESVsHardwareType indicates the Vs. System hardware configuration.
type NESVsHardwareType byte

const (
	NESVsHardwareUnisystemNormal      NESVsHardwareType = 0x00 // Vs. Unisystem (normal)
	NESVsHardwareUnisystemRBI         NESVsHardwareType = 0x01 // Vs. Unisystem (RBI Baseball protection)
	NESVsHardwareUnisystemTKO         NESVsHardwareType = 0x02 // Vs. Unisystem (TKO Boxing protection)
	NESVsHardwareUnisystemSuperXevius NESVsHardwareType = 0x03 // Vs. Unisystem (Super Xevious protection)
	NESVsHardwareUnisystemIceClimber  NESVsHardwareType = 0x04 // Vs. Unisystem (Vs. Ice Climber Japan protection)
	NESVsHardwareDualsystemNormal     NESVsHardwareType = 0x05 // Vs. Dualsystem (normal)
	NESVsHardwareDualsystemRaid       NESVsHardwareType = 0x06 // Vs. Dualsystem (Raid on Bungeling Bay protection)
)

// NESExtendedConsoleType indicates extended console types (NES 2.0 only).
type NESExtendedConsoleType byte

const (
	NESExtendedRegularNES     NESExtendedConsoleType = 0x00 // Regular NES/Famicom/Dendy
	NESExtendedVsSystem       NESExtendedConsoleType = 0x01 // Vs. System
	NESExtendedPlayChoice     NESExtendedConsoleType = 0x02 // PlayChoice-10
	NESExtendedFamiclone      NESExtendedConsoleType = 0x03 // Famiclone with Decimal Mode
	NESExtendedNESEPSM        NESExtendedConsoleType = 0x04 // NES/Famicom with EPSM module
	NESExtendedVT01           NESExtendedConsoleType = 0x05 // VT01 with red/cyan STN palette
	NESExtendedVT02           NESExtendedConsoleType = 0x06 // VT02
	NESExtendedVT03           NESExtendedConsoleType = 0x07 // VT03
	NESExtendedVT09           NESExtendedConsoleType = 0x08 // VT09
	NESExtendedVT32           NESExtendedConsoleType = 0x09 // VT32
	NESExtendedVT369          NESExtendedConsoleType = 0x0A // VT369
	NESExtendedUM6578         NESExtendedConsoleType = 0x0B // UM6578
	NESExtendedFamicomNetwork NESExtendedConsoleType = 0x0C // Famicom Network System
)

// NESInfo contains metadata extracted from an NES ROM file.
// Designed for NES 2.0 headers; iNES 1.0 headers populate a subset of fields.
type NESInfo struct {
	// PRGROMSize is the PRG-ROM size in bytes.
	PRGROMSize int `json:"prg_rom_size"`
	// CHRROMSize is the CHR-ROM size in bytes. Zero indicates CHR-RAM.
	CHRROMSize int `json:"chr_rom_size"`

	// PRGRAMSize is the volatile PRG-RAM size in bytes.
	PRGRAMSize int `json:"prg_ram_size"`
	// PRGNVRAMSize is the non-volatile (battery-backed) PRG-RAM size in bytes.
	PRGNVRAMSize int `json:"prg_nvram_size"`
	// CHRRAMSize is the volatile CHR-RAM size in bytes.
	CHRRAMSize int `json:"chr_ram_size"`
	// CHRNVRAMSize is the non-volatile CHR-RAM size in bytes.
	CHRNVRAMSize int `json:"chr_nvram_size"`

	// Mapper is the mapper number (0-4095 for NES 2.0, 0-255 for iNES 1.0).
	Mapper int `json:"mapper"`
	// Submapper disambiguates mapper variants (NES 2.0 only, 0-15).
	Submapper int `json:"submapper"`

	// Mirroring indicates the nametable mirroring mode.
	Mirroring NESMirroring `json:"mirroring"`
	// FourScreen indicates four-screen VRAM layout (overrides Mirroring).
	FourScreen bool `json:"four_screen"`

	// HasBattery indicates battery-backed save RAM is present.
	HasBattery bool `json:"has_battery"`
	// HasTrainer indicates a 512-byte trainer is present before PRG-ROM.
	HasTrainer bool `json:"has_trainer"`

	// ConsoleType indicates the target console.
	ConsoleType NESConsoleType `json:"console_type"`
	// TimingMode indicates the CPU/PPU timing region.
	TimingMode NESTimingMode `json:"timing_mode"`
	// ExpansionDevice is the default expansion device (NES 2.0 only, raw byte).
	ExpansionDevice byte `json:"expansion_device"`

	// VsPPUType indicates the Vs. System PPU variant (only valid when ConsoleType == NESConsoleVsSystem).
	VsPPUType NESVsPPUType `json:"vs_ppu_type"`
	// VsHardwareType indicates the Vs. System hardware configuration (only valid when ConsoleType == NESConsoleVsSystem).
	VsHardwareType NESVsHardwareType `json:"vs_hardware_type"`

	// ExtendedConsoleType indicates the extended console variant (only valid when ConsoleType == NESConsoleExtended).
	ExtendedConsoleType NESExtendedConsoleType `json:"extended_console_type"`

	// MiscROMs indicates the number of miscellaneous ROM chips (NES 2.0 only).
	MiscROMs int `json:"misc_roms"`

	// IsNES20 is true if the header is NES 2.0 format.
	IsNES20 bool `json:"is_nes20"`
}

// GamePlatform implements identify.GameInfo.
func (i *NESInfo) GamePlatform() core.Platform { return core.PlatformNES }

// GameTitle implements identify.GameInfo. NES ROMs don't have embedded titles.
func (i *NESInfo) GameTitle() string { return "" }

// GameSerial implements identify.GameInfo. NES ROMs don't have serial numbers.
func (i *NESInfo) GameSerial() string { return "" }

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
	if !bytes.Equal(header[0:4], nesMagic) {
		return nil, fmt.Errorf("not a valid NES ROM: magic mismatch")
	}

	flags6 := header[6]
	flags7 := header[7]

	// Check for NES 2.0 format: bits 2-3 of flags7 == 0b10
	isNES20 := (flags7 & 0x0C) == 0x08

	// Extract common fields from flags 6
	mirroring := NESMirroring(flags6 & 0x01)
	hasBattery := (flags6 & 0x02) != 0
	hasTrainer := (flags6 & 0x04) != 0
	fourScreen := (flags6 & 0x08) != 0

	// Console type (bits 0-1 of flags7)
	consoleType := NESConsoleType(flags7 & 0x03)

	info := &NESInfo{
		Mirroring:   mirroring,
		HasBattery:  hasBattery,
		HasTrainer:  hasTrainer,
		FourScreen:  fourScreen,
		ConsoleType: consoleType,
		IsNES20:     isNES20,
	}

	if isNES20 {
		parseNES20(header, info)
	} else {
		parseINES(header, info)
	}

	return info, nil
}

// parseNES20 parses NES 2.0 specific fields.
func parseNES20(header []byte, info *NESInfo) {
	flags6 := header[6]
	flags7 := header[7]
	byte8 := header[8]
	byte9 := header[9]
	byte10 := header[10]
	byte11 := header[11]
	byte12 := header[12]
	byte13 := header[13]
	byte14 := header[14]
	byte15 := header[15]

	// Mapper number (12 bits): bits 4-7 of flags6 + bits 4-7 of flags7 + bits 0-3 of byte8
	mapperLow := (flags6 >> 4) & 0x0F
	mapperMid := flags7 & 0xF0
	mapperHigh := int(byte8&0x0F) << 8
	info.Mapper = mapperHigh | int(mapperMid) | int(mapperLow)

	// Submapper (4 bits): bits 4-7 of byte8
	info.Submapper = int(byte8 >> 4)

	// PRG-ROM size using exponent-multiplier notation
	prgROMSizeLSB := header[4]
	prgROMSizeMSB := byte9 & 0x0F
	info.PRGROMSize = calculateNES20ROMSize(prgROMSizeLSB, prgROMSizeMSB, 16*1024)

	// CHR-ROM size using exponent-multiplier notation
	chrROMSizeLSB := header[5]
	chrROMSizeMSB := (byte9 >> 4) & 0x0F
	info.CHRROMSize = calculateNES20ROMSize(chrROMSizeLSB, chrROMSizeMSB, 8*1024)

	// PRG-RAM sizes (byte 10)
	// Volatile PRG-RAM: bits 0-3 (shift count, 64 << value)
	// Non-volatile PRG-RAM: bits 4-7 (shift count, 64 << value)
	prgRAMShift := byte10 & 0x0F
	prgNVRAMShift := (byte10 >> 4) & 0x0F
	info.PRGRAMSize = calculateNES20RAMSize(prgRAMShift)
	info.PRGNVRAMSize = calculateNES20RAMSize(prgNVRAMShift)

	// CHR-RAM sizes (byte 11)
	// Volatile CHR-RAM: bits 0-3 (shift count, 64 << value)
	// Non-volatile CHR-RAM: bits 4-7 (shift count, 64 << value)
	chrRAMShift := byte11 & 0x0F
	chrNVRAMShift := (byte11 >> 4) & 0x0F
	info.CHRRAMSize = calculateNES20RAMSize(chrRAMShift)
	info.CHRNVRAMSize = calculateNES20RAMSize(chrNVRAMShift)

	// CPU/PPU timing mode (byte 12, bits 0-1)
	info.TimingMode = NESTimingMode(byte12 & 0x03)

	// System type specifics (byte 13)
	switch info.ConsoleType {
	case NESConsoleVsSystem:
		info.VsPPUType = NESVsPPUType(byte13 & 0x0F)
		info.VsHardwareType = NESVsHardwareType((byte13 >> 4) & 0x0F)
	case NESConsoleExtended:
		info.ExtendedConsoleType = NESExtendedConsoleType(byte13 & 0x0F)
	}

	// Miscellaneous ROMs (byte 14, bits 0-1)
	info.MiscROMs = int(byte14 & 0x03)

	// Default expansion device (byte 15, bits 0-5)
	info.ExpansionDevice = byte15 & 0x3F
}

// parseINES parses iNES 1.0 specific fields.
func parseINES(header []byte, info *NESInfo) {
	flags6 := header[6]
	flags7 := header[7]
	flags9 := header[9]

	// Mapper number (8 bits): bits 4-7 of flags6 + bits 4-7 of flags7
	mapperLow := (flags6 >> 4) & 0x0F
	mapperHigh := flags7 & 0xF0
	info.Mapper = int(mapperHigh) | int(mapperLow)

	// PRG-ROM size (16 KB units)
	info.PRGROMSize = int(header[4]) * 16 * 1024

	// CHR-ROM size (8 KB units, 0 = CHR-RAM)
	info.CHRROMSize = int(header[5]) * 8 * 1024

	// PRG-RAM size (8 KB units, 0 = infer 8KB for compatibility)
	prgRAMBanks := int(header[8])
	if prgRAMBanks == 0 {
		prgRAMBanks = 1 // Default to 8 KB for compatibility
	}
	info.PRGRAMSize = prgRAMBanks * 8 * 1024

	// TV system (bit 0 of flags9)
	// iNES 1.0 only supports NTSC/PAL distinction
	if (flags9 & 0x01) != 0 {
		info.TimingMode = NESTimingPAL
	} else {
		info.TimingMode = NESTimingNTSC
	}
}

// calculateNES20ROMSize calculates ROM size using NES 2.0 exponent-multiplier notation.
// If MSB < 0x0F, size = (MSB << 8 | LSB) * unit
// If MSB == 0x0F, size = (2^(LSB>>2)) * ((LSB&3)*2 + 1)
func calculateNES20ROMSize(lsb, msb byte, unit int) int {
	if msb < 0x0F {
		return (int(msb)<<8 | int(lsb)) * unit
	}
	// Exponent-multiplier form
	exponent := lsb >> 2
	multiplier := int(lsb&3)*2 + 1
	return (1 << exponent) * multiplier
}

// calculateNES20RAMSize calculates RAM size from NES 2.0 shift count.
// Size = 64 << shiftCount when shiftCount > 0, otherwise 0.
func calculateNES20RAMSize(shiftCount byte) int {
	if shiftCount == 0 {
		return 0
	}
	return 64 << shiftCount
}
