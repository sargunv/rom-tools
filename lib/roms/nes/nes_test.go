package nes

import (
	"bytes"
	"os"
	"testing"
)

func TestParseNES(t *testing.T) {
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

	// Verify the file was parsed without error - NES format doesn't include title
	if info.PRGROMSize <= 0 {
		t.Errorf("Expected positive PRG ROM size, got %d", info.PRGROMSize)
	}
}

func TestParseNES_Fields(t *testing.T) {
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

	// BombSweeper.nes has: PRG=1 (16KB), CHR=1 (8KB), Mapper=0, NROM
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
	if info.TVSystem != NESTVSystemNTSC {
		t.Errorf("TVSystem = %d, want %d (NTSC)", info.TVSystem, NESTVSystemNTSC)
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
}

// makeSyntheticNES creates a synthetic NES ROM with specified parameters.
func makeSyntheticNES(prgBanks, chrBanks byte, flags6, flags7, flags9 byte) []byte {
	header := make([]byte, nesHeaderSize)
	copy(header[0:4], nesMagic)
	header[nesPRGROMOffset] = prgBanks
	header[nesCHRROMOffset] = chrBanks
	header[nesFlags6Offset] = flags6
	header[nesFlags7Offset] = flags7
	header[nesPRGRAMOffset] = 0 // Will default to 8KB
	header[nesFlags9Offset] = flags9
	return header
}

func TestParseNES_Synthetic(t *testing.T) {
	tests := []struct {
		name        string
		prgBanks    byte
		chrBanks    byte
		flags6      byte
		flags7      byte
		flags9      byte
		wantPRGSize int
		wantCHRSize int
		wantMapper  int
		wantMirror  NESMirroring
		wantNES20   bool
	}{
		{
			name:        "basic NROM",
			prgBanks:    2,
			chrBanks:    1,
			flags6:      0x00,
			flags7:      0x00,
			flags9:      0x00,
			wantPRGSize: 32 * 1024,
			wantCHRSize: 8 * 1024,
			wantMapper:  0,
			wantMirror:  NESMirroringHorizontal,
			wantNES20:   false,
		},
		{
			name:        "mapper 1 vertical mirroring",
			prgBanks:    8,
			chrBanks:    4,
			flags6:      0x11, // mapper 1 low nibble, vertical mirroring
			flags7:      0x00, // mapper 1 high nibble
			flags9:      0x00,
			wantPRGSize: 128 * 1024,
			wantCHRSize: 32 * 1024,
			wantMapper:  1,
			wantMirror:  NESMirroringVertical,
			wantNES20:   false,
		},
		{
			name:        "NES 2.0 format",
			prgBanks:    4,
			chrBanks:    2,
			flags6:      0x00,
			flags7:      0x08, // NES 2.0 identifier (bits 2-3 = 2)
			flags9:      0x00,
			wantPRGSize: 64 * 1024,
			wantCHRSize: 16 * 1024,
			wantMapper:  0,
			wantMirror:  NESMirroringHorizontal,
			wantNES20:   true,
		},
		{
			name:        "PAL TV system",
			prgBanks:    1,
			chrBanks:    1,
			flags6:      0x00,
			flags7:      0x00,
			flags9:      0x01,
			wantPRGSize: 16 * 1024,
			wantCHRSize: 8 * 1024,
			wantMapper:  0,
			wantMirror:  NESMirroringHorizontal,
			wantNES20:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rom := makeSyntheticNES(tc.prgBanks, tc.chrBanks, tc.flags6, tc.flags7, tc.flags9)
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
			if info.IsNES20 != tc.wantNES20 {
				t.Errorf("IsNES20 = %v, want %v", info.IsNES20, tc.wantNES20)
			}
		})
	}
}

func TestParseNES_TooSmall(t *testing.T) {
	// File smaller than header
	data := []byte{0x4E, 0x45, 0x53, 0x1A, 0x01}
	reader := bytes.NewReader(data)

	_, err := ParseNES(reader, int64(len(data)))
	if err == nil {
		t.Error("ParseNES() expected error for too small file, got nil")
	}
}

func TestParseNES_InvalidMagic(t *testing.T) {
	// Valid size but invalid magic
	header := make([]byte, nesHeaderSize)
	header[0] = 'X' // Invalid magic
	header[1] = 'E'
	header[2] = 'S'
	header[3] = 0x1A
	reader := bytes.NewReader(header)

	_, err := ParseNES(reader, int64(len(header)))
	if err == nil {
		t.Error("ParseNES() expected error for invalid magic, got nil")
	}
}
