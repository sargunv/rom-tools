package nes

import (
	"bytes"
	"os"
	"testing"
)

func TestParseNES_INES_BombSweeper(t *testing.T) {
	romPath := "testdata/BombSweeper.nes"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseNES(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseNES() error = %v", err)
	}

	// BombSweeper.nes: iNES 1.0, PRG=1 (16KB), CHR=1 (8KB), Mapper=0, NROM
	// Header: 4E 45 53 1A 01 01 00 00 00 00 00 00 00 00 00 00
	if info.PRGROMSize != 16*1024 {
		t.Errorf("PRGROMSize = %d, want %d", info.PRGROMSize, 16*1024)
	}
	if info.CHRROMSize != 8*1024 {
		t.Errorf("CHRROMSize = %d, want %d", info.CHRROMSize, 8*1024)
	}
	if info.Mapper != 0 {
		t.Errorf("Mapper = %d, want 0", info.Mapper)
	}
	if info.Mirroring != NESMirroringHorizontal {
		t.Errorf("Mirroring = %d, want %d (Horizontal)", info.Mirroring, NESMirroringHorizontal)
	}
	if info.ConsoleType != NESConsoleNES {
		t.Errorf("ConsoleType = %d, want %d (NES)", info.ConsoleType, NESConsoleNES)
	}
	if info.TimingMode != NESTimingNTSC {
		t.Errorf("TimingMode = %d, want %d (NTSC)", info.TimingMode, NESTimingNTSC)
	}
	if info.HasBattery {
		t.Errorf("HasBattery = true, want false")
	}
	if info.HasTrainer {
		t.Errorf("HasTrainer = true, want false")
	}
	if info.FourScreen {
		t.Errorf("FourScreen = true, want false")
	}
	if info.IsNES20 {
		t.Errorf("IsNES20 = true, want false")
	}
	// iNES 1.0 defaults PRG-RAM to 8KB when byte 8 is 0
	if info.PRGRAMSize != 8*1024 {
		t.Errorf("PRGRAMSize = %d, want %d", info.PRGRAMSize, 8*1024)
	}
	if info.Submapper != 0 {
		t.Errorf("Submapper = %d, want 0 (iNES 1.0 has no submapper)", info.Submapper)
	}
}

func TestParseNES_NES20_SEROM(t *testing.T) {
	romPath := "testdata/serom.nes"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseNES(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseNES() error = %v", err)
	}

	// serom.nes: NES 2.0
	// Header: 4E 45 53 1A 02 01 10 08 50 00 00 00 00 00 00 00
	// Byte 4: PRG-ROM LSB = 2 -> 32KB
	// Byte 5: CHR-ROM LSB = 1 -> 8KB
	// Byte 6: 0x10 -> mapper low nibble = 1, horizontal mirroring
	// Byte 7: 0x08 -> NES 2.0 identifier, mapper mid = 0, console type = 0
	// Byte 8: 0x50 -> submapper = 5, mapper high = 0
	// Byte 9-15: all 0

	if !info.IsNES20 {
		t.Errorf("IsNES20 = false, want true")
	}
	if info.PRGROMSize != 32*1024 {
		t.Errorf("PRGROMSize = %d, want %d", info.PRGROMSize, 32*1024)
	}
	if info.CHRROMSize != 8*1024 {
		t.Errorf("CHRROMSize = %d, want %d", info.CHRROMSize, 8*1024)
	}
	if info.Mapper != 1 {
		t.Errorf("Mapper = %d, want 1 (MMC1)", info.Mapper)
	}
	if info.Submapper != 5 {
		t.Errorf("Submapper = %d, want 5", info.Submapper)
	}
	if info.Mirroring != NESMirroringHorizontal {
		t.Errorf("Mirroring = %d, want %d (Horizontal)", info.Mirroring, NESMirroringHorizontal)
	}
	if info.ConsoleType != NESConsoleNES {
		t.Errorf("ConsoleType = %d, want %d (NES)", info.ConsoleType, NESConsoleNES)
	}
	if info.TimingMode != NESTimingNTSC {
		t.Errorf("TimingMode = %d, want %d (NTSC)", info.TimingMode, NESTimingNTSC)
	}
	if info.HasBattery {
		t.Errorf("HasBattery = true, want false")
	}
	if info.HasTrainer {
		t.Errorf("HasTrainer = true, want false")
	}
	// All RAM sizes are 0 in serom.nes
	if info.PRGRAMSize != 0 {
		t.Errorf("PRGRAMSize = %d, want 0", info.PRGRAMSize)
	}
	if info.PRGNVRAMSize != 0 {
		t.Errorf("PRGNVRAMSize = %d, want 0", info.PRGNVRAMSize)
	}
	if info.CHRRAMSize != 0 {
		t.Errorf("CHRRAMSize = %d, want 0", info.CHRRAMSize)
	}
	if info.CHRNVRAMSize != 0 {
		t.Errorf("CHRNVRAMSize = %d, want 0", info.CHRNVRAMSize)
	}
}

// makeSyntheticNES creates a synthetic NES ROM header with specified parameters.
func makeSyntheticNES(header []byte) []byte {
	if len(header) < nesHeaderSize {
		h := make([]byte, nesHeaderSize)
		copy(h, header)
		return h
	}
	return header
}

func TestParseNES_Synthetic_INES(t *testing.T) {
	tests := []struct {
		name        string
		header      []byte
		wantPRGSize int
		wantCHRSize int
		wantMapper  int
		wantMirror  NESMirroring
		wantTiming  NESTimingMode
		wantNES20   bool
	}{
		{
			name: "basic NROM",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A, // Magic
				0x02, // PRG-ROM: 2 * 16KB = 32KB
				0x01, // CHR-ROM: 1 * 8KB = 8KB
				0x00, // Flags 6: mapper 0, horizontal mirroring
				0x00, // Flags 7: iNES 1.0
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			wantPRGSize: 32 * 1024,
			wantCHRSize: 8 * 1024,
			wantMapper:  0,
			wantMirror:  NESMirroringHorizontal,
			wantTiming:  NESTimingNTSC,
			wantNES20:   false,
		},
		{
			name: "mapper 1 vertical mirroring",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x08,       // PRG-ROM: 8 * 16KB = 128KB
				0x04,       // CHR-ROM: 4 * 8KB = 32KB
				0x11,       // Flags 6: mapper low=1, vertical mirroring
				0x00,       // Flags 7: mapper high=0
				0x00, 0x00, // byte 8-9
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			wantPRGSize: 128 * 1024,
			wantCHRSize: 32 * 1024,
			wantMapper:  1,
			wantMirror:  NESMirroringVertical,
			wantTiming:  NESTimingNTSC,
			wantNES20:   false,
		},
		{
			name: "PAL timing",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x01, 0x01, // PRG/CHR
				0x00, 0x00, // Flags 6-7
				0x00, // PRG-RAM
				0x01, // Flags 9: PAL
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			wantPRGSize: 16 * 1024,
			wantCHRSize: 8 * 1024,
			wantMapper:  0,
			wantMirror:  NESMirroringHorizontal,
			wantTiming:  NESTimingPAL,
			wantNES20:   false,
		},
		{
			name: "high mapper number (MMC3 = 4)",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x10, 0x10, // PRG/CHR: 256KB each
				0x40, // Flags 6: mapper low=4
				0x00, // Flags 7: mapper high=0
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			wantPRGSize: 256 * 1024,
			wantCHRSize: 128 * 1024,
			wantMapper:  4,
			wantMirror:  NESMirroringHorizontal,
			wantTiming:  NESTimingNTSC,
			wantNES20:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rom := makeSyntheticNES(tc.header)
			reader := bytes.NewReader(rom)

			info, err := ParseNES(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("ParseNES() error = %v", err)
			}

			if info.PRGROMSize != tc.wantPRGSize {
				t.Errorf("PRGROMSize = %d, want %d", info.PRGROMSize, tc.wantPRGSize)
			}
			if info.CHRROMSize != tc.wantCHRSize {
				t.Errorf("CHRROMSize = %d, want %d", info.CHRROMSize, tc.wantCHRSize)
			}
			if info.Mapper != tc.wantMapper {
				t.Errorf("Mapper = %d, want %d", info.Mapper, tc.wantMapper)
			}
			if info.Mirroring != tc.wantMirror {
				t.Errorf("Mirroring = %d, want %d", info.Mirroring, tc.wantMirror)
			}
			if info.TimingMode != tc.wantTiming {
				t.Errorf("TimingMode = %d, want %d", info.TimingMode, tc.wantTiming)
			}
			if info.IsNES20 != tc.wantNES20 {
				t.Errorf("IsNES20 = %v, want %v", info.IsNES20, tc.wantNES20)
			}
		})
	}
}

func TestParseNES_Synthetic_NES20(t *testing.T) {
	tests := []struct {
		name           string
		header         []byte
		wantPRGSize    int
		wantCHRSize    int
		wantMapper     int
		wantSubmapper  int
		wantPRGRAM     int
		wantPRGNVRAM   int
		wantCHRRAM     int
		wantCHRNVRAM   int
		wantTiming     NESTimingMode
		wantConsole    NESConsoleType
		wantVsPPU      NESVsPPUType
		wantVsHardware NESVsHardwareType
		wantExtConsole NESExtendedConsoleType
		wantMiscROMs   int
		wantExpansion  byte
	}{
		{
			name: "NES 2.0 basic",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x04,       // PRG-ROM LSB
				0x02,       // CHR-ROM LSB
				0x00,       // Flags 6
				0x08,       // Flags 7: NES 2.0
				0x00,       // Mapper/Submapper
				0x00,       // ROM size MSB
				0x00, 0x00, // RAM sizes
				0x00, // Timing
				0x00, // Vs/Extended
				0x00, // Misc ROMs
				0x00, // Expansion device
			},
			wantPRGSize:   64 * 1024,
			wantCHRSize:   16 * 1024,
			wantMapper:    0,
			wantSubmapper: 0,
			wantTiming:    NESTimingNTSC,
		},
		{
			name: "NES 2.0 with submapper",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x02, 0x01, // PRG=32KB, CHR=8KB
				0x10, // Flags 6: mapper low=1
				0x08, // Flags 7: NES 2.0, mapper mid=0
				0x50, // Submapper=5, mapper high=0
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			wantPRGSize:   32 * 1024,
			wantCHRSize:   8 * 1024,
			wantMapper:    1,
			wantSubmapper: 5,
			wantTiming:    NESTimingNTSC,
		},
		{
			name: "NES 2.0 extended mapper",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x02, 0x01, // PRG=32KB, CHR=8KB
				0x00, // Flags 6: mapper low=0
				0x08, // Flags 7: NES 2.0, mapper mid=0
				0x01, // Mapper high=1 (mapper 256)
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			wantPRGSize:   32 * 1024,
			wantCHRSize:   8 * 1024,
			wantMapper:    256,
			wantSubmapper: 0,
			wantTiming:    NESTimingNTSC,
		},
		{
			name: "NES 2.0 all timing modes - Multi",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x01, 0x01, 0x00, 0x08,
				0x00, 0x00, 0x00, 0x00,
				0x02, // Timing: Multi
				0x00, 0x00, 0x00,
			},
			wantPRGSize: 16 * 1024,
			wantCHRSize: 8 * 1024,
			wantTiming:  NESTimingMulti,
		},
		{
			name: "NES 2.0 all timing modes - Dendy",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x01, 0x01, 0x00, 0x08,
				0x00, 0x00, 0x00, 0x00,
				0x03, // Timing: Dendy
				0x00, 0x00, 0x00,
			},
			wantPRGSize: 16 * 1024,
			wantCHRSize: 8 * 1024,
			wantTiming:  NESTimingDendy,
		},
		{
			name: "NES 2.0 RAM sizes",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x01, 0x00, // PRG=16KB, CHR=0 (uses CHR-RAM)
				0x00, 0x08, // NES 2.0
				0x00, 0x00,
				0x75, // PRG-RAM: shift=5 (2KB), PRG-NVRAM: shift=7 (8KB)
				0x87, // CHR-RAM: shift=7 (8KB), CHR-NVRAM: shift=8 (16KB)
				0x00, 0x00, 0x00, 0x00,
			},
			wantPRGSize:  16 * 1024,
			wantCHRSize:  0,
			wantPRGRAM:   64 << 5, // 2KB
			wantPRGNVRAM: 64 << 7, // 8KB
			wantCHRRAM:   64 << 7, // 8KB
			wantCHRNVRAM: 64 << 8, // 16KB
			wantTiming:   NESTimingNTSC,
		},
		{
			name: "NES 2.0 Vs. System",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x01, 0x01,
				0x00, 0x09, // Flags 7: NES 2.0 + console type = 1 (Vs. System)
				0x00, 0x00, 0x00, 0x00, 0x00,
				0x53, // Vs. PPU type = 3 (RP2C04-0002), Vs. hardware = 5 (Dualsystem normal)
				0x00, 0x00,
			},
			wantPRGSize:    16 * 1024,
			wantCHRSize:    8 * 1024,
			wantConsole:    NESConsoleVsSystem,
			wantVsPPU:      NESVsPPURP2C04_0002,
			wantVsHardware: NESVsHardwareDualsystemNormal,
			wantTiming:     NESTimingNTSC,
		},
		{
			name: "NES 2.0 Extended Console",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x01, 0x01,
				0x00, 0x0B, // Flags 7: NES 2.0 + console type = 3 (Extended)
				0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, // Extended console type = 5 (VT01)
				0x00, 0x00,
			},
			wantPRGSize:    16 * 1024,
			wantCHRSize:    8 * 1024,
			wantConsole:    NESConsoleExtended,
			wantExtConsole: NESExtendedVT01,
			wantTiming:     NESTimingNTSC,
		},
		{
			name: "NES 2.0 Misc ROMs and Expansion",
			header: []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x01, 0x01, 0x00, 0x08,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x02, // Misc ROMs = 2
				0x15, // Expansion device = 0x15 (21)
			},
			wantPRGSize:   16 * 1024,
			wantCHRSize:   8 * 1024,
			wantMiscROMs:  2,
			wantExpansion: 0x15,
			wantTiming:    NESTimingNTSC,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rom := makeSyntheticNES(tc.header)
			reader := bytes.NewReader(rom)

			info, err := ParseNES(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("ParseNES() error = %v", err)
			}

			if !info.IsNES20 {
				t.Errorf("IsNES20 = false, want true")
			}
			if info.PRGROMSize != tc.wantPRGSize {
				t.Errorf("PRGROMSize = %d, want %d", info.PRGROMSize, tc.wantPRGSize)
			}
			if info.CHRROMSize != tc.wantCHRSize {
				t.Errorf("CHRROMSize = %d, want %d", info.CHRROMSize, tc.wantCHRSize)
			}
			if tc.wantMapper != 0 && info.Mapper != tc.wantMapper {
				t.Errorf("Mapper = %d, want %d", info.Mapper, tc.wantMapper)
			}
			if tc.wantSubmapper != 0 && info.Submapper != tc.wantSubmapper {
				t.Errorf("Submapper = %d, want %d", info.Submapper, tc.wantSubmapper)
			}
			if info.TimingMode != tc.wantTiming {
				t.Errorf("TimingMode = %d, want %d", info.TimingMode, tc.wantTiming)
			}
			if tc.wantPRGRAM != 0 && info.PRGRAMSize != tc.wantPRGRAM {
				t.Errorf("PRGRAMSize = %d, want %d", info.PRGRAMSize, tc.wantPRGRAM)
			}
			if tc.wantPRGNVRAM != 0 && info.PRGNVRAMSize != tc.wantPRGNVRAM {
				t.Errorf("PRGNVRAMSize = %d, want %d", info.PRGNVRAMSize, tc.wantPRGNVRAM)
			}
			if tc.wantCHRRAM != 0 && info.CHRRAMSize != tc.wantCHRRAM {
				t.Errorf("CHRRAMSize = %d, want %d", info.CHRRAMSize, tc.wantCHRRAM)
			}
			if tc.wantCHRNVRAM != 0 && info.CHRNVRAMSize != tc.wantCHRNVRAM {
				t.Errorf("CHRNVRAMSize = %d, want %d", info.CHRNVRAMSize, tc.wantCHRNVRAM)
			}
			if tc.wantConsole != 0 && info.ConsoleType != tc.wantConsole {
				t.Errorf("ConsoleType = %d, want %d", info.ConsoleType, tc.wantConsole)
			}
			if tc.wantVsPPU != 0 && info.VsPPUType != tc.wantVsPPU {
				t.Errorf("VsPPUType = %d, want %d", info.VsPPUType, tc.wantVsPPU)
			}
			if tc.wantVsHardware != 0 && info.VsHardwareType != tc.wantVsHardware {
				t.Errorf("VsHardwareType = %d, want %d", info.VsHardwareType, tc.wantVsHardware)
			}
			if tc.wantExtConsole != 0 && info.ExtendedConsoleType != tc.wantExtConsole {
				t.Errorf("ExtendedConsoleType = %d, want %d", info.ExtendedConsoleType, tc.wantExtConsole)
			}
			if tc.wantMiscROMs != 0 && info.MiscROMs != tc.wantMiscROMs {
				t.Errorf("MiscROMs = %d, want %d", info.MiscROMs, tc.wantMiscROMs)
			}
			if tc.wantExpansion != 0 && info.ExpansionDevice != tc.wantExpansion {
				t.Errorf("ExpansionDevice = %d, want %d", info.ExpansionDevice, tc.wantExpansion)
			}
		})
	}
}

func TestParseNES_Flags(t *testing.T) {
	tests := []struct {
		name        string
		flags6      byte
		wantMirror  NESMirroring
		wantBattery bool
		wantTrainer bool
		wantFour    bool
	}{
		{
			name:        "all flags off",
			flags6:      0x00,
			wantMirror:  NESMirroringHorizontal,
			wantBattery: false,
			wantTrainer: false,
			wantFour:    false,
		},
		{
			name:        "vertical mirroring",
			flags6:      0x01,
			wantMirror:  NESMirroringVertical,
			wantBattery: false,
			wantTrainer: false,
			wantFour:    false,
		},
		{
			name:        "battery present",
			flags6:      0x02,
			wantMirror:  NESMirroringHorizontal,
			wantBattery: true,
			wantTrainer: false,
			wantFour:    false,
		},
		{
			name:        "trainer present",
			flags6:      0x04,
			wantMirror:  NESMirroringHorizontal,
			wantBattery: false,
			wantTrainer: true,
			wantFour:    false,
		},
		{
			name:        "four-screen VRAM",
			flags6:      0x08,
			wantMirror:  NESMirroringHorizontal,
			wantBattery: false,
			wantTrainer: false,
			wantFour:    true,
		},
		{
			name:        "all flags on",
			flags6:      0x0F,
			wantMirror:  NESMirroringVertical,
			wantBattery: true,
			wantTrainer: true,
			wantFour:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			header := []byte{
				0x4E, 0x45, 0x53, 0x1A,
				0x01, 0x01,
				tc.flags6,
				0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			}
			reader := bytes.NewReader(header)

			info, err := ParseNES(reader, int64(len(header)))
			if err != nil {
				t.Fatalf("ParseNES() error = %v", err)
			}

			if info.Mirroring != tc.wantMirror {
				t.Errorf("Mirroring = %d, want %d", info.Mirroring, tc.wantMirror)
			}
			if info.HasBattery != tc.wantBattery {
				t.Errorf("HasBattery = %v, want %v", info.HasBattery, tc.wantBattery)
			}
			if info.HasTrainer != tc.wantTrainer {
				t.Errorf("HasTrainer = %v, want %v", info.HasTrainer, tc.wantTrainer)
			}
			if info.FourScreen != tc.wantFour {
				t.Errorf("FourScreen = %v, want %v", info.FourScreen, tc.wantFour)
			}
		})
	}
}

func TestParseNES_TooSmall(t *testing.T) {
	data := []byte{0x4E, 0x45, 0x53, 0x1A, 0x01}
	reader := bytes.NewReader(data)

	_, err := ParseNES(reader, int64(len(data)))
	if err == nil {
		t.Error("ParseNES() expected error for too small file, got nil")
	}
}

func TestParseNES_InvalidMagic(t *testing.T) {
	header := make([]byte, nesHeaderSize)
	header[0] = 'X'
	header[1] = 'E'
	header[2] = 'S'
	header[3] = 0x1A
	reader := bytes.NewReader(header)

	_, err := ParseNES(reader, int64(len(header)))
	if err == nil {
		t.Error("ParseNES() expected error for invalid magic, got nil")
	}
}

func TestCalculateNES20ROMSize(t *testing.T) {
	tests := []struct {
		name string
		lsb  byte
		msb  byte
		unit int
		want int
	}{
		{
			name: "standard notation - 16KB unit",
			lsb:  0x02,
			msb:  0x00,
			unit: 16 * 1024,
			want: 32 * 1024,
		},
		{
			name: "standard notation - 8KB unit",
			lsb:  0x04,
			msb:  0x00,
			unit: 8 * 1024,
			want: 32 * 1024,
		},
		{
			name: "standard notation - large value",
			lsb:  0x00,
			msb:  0x01,
			unit: 16 * 1024,
			want: 256 * 16 * 1024, // (1<<8 | 0) * 16KB = 4MB
		},
		{
			name: "exponent-multiplier - 2^2 * 1",
			lsb:  0x08, // exponent=2, multiplier bits=0 -> multiplier=1
			msb:  0x0F,
			unit: 16 * 1024,
			want: (1 << 2) * 1, // 4 bytes
		},
		{
			name: "exponent-multiplier - 2^2 * 3",
			lsb:  0x09, // exponent=2, multiplier bits=1 -> multiplier=3
			msb:  0x0F,
			unit: 16 * 1024,
			want: (1 << 2) * 3, // 12 bytes
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calculateNES20ROMSize(tc.lsb, tc.msb, tc.unit)
			if got != tc.want {
				t.Errorf("calculateNES20ROMSize(%d, %d, %d) = %d, want %d",
					tc.lsb, tc.msb, tc.unit, got, tc.want)
			}
		})
	}
}

func TestCalculateNES20RAMSize(t *testing.T) {
	tests := []struct {
		shiftCount byte
		want       int
	}{
		{0, 0},
		{1, 128},      // 64 << 1 = 128
		{5, 2048},     // 64 << 5 = 2KB
		{7, 8192},     // 64 << 7 = 8KB
		{8, 16384},    // 64 << 8 = 16KB
		{10, 65536},   // 64 << 10 = 64KB
		{14, 1048576}, // 64 << 14 = 1MB
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			got := calculateNES20RAMSize(tc.shiftCount)
			if got != tc.want {
				t.Errorf("calculateNES20RAMSize(%d) = %d, want %d",
					tc.shiftCount, got, tc.want)
			}
		})
	}
}
