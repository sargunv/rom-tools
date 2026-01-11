package nes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestParseNES(t *testing.T) {
	nesPath := filepath.Join(testutil.ROMsPath(t), "BombSweeper.nes")

	file, err := os.Open(nesPath)
	if err != nil {
		t.Fatalf("Failed to open NES file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat NES file: %v", err)
	}

	info, err := ParseNES(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseNES() error = %v", err)
	}

	// BombSweeper uses mapper 0 (NROM)
	if info.Mapper != 0 {
		t.Errorf("Expected mapper 0, got %d", info.Mapper)
	}

	// 1 PRG-ROM bank = 16 KB
	expectedPRGSize := 16 * 1024
	if info.PRGROMSize != expectedPRGSize {
		t.Errorf("Expected PRG-ROM size %d, got %d", expectedPRGSize, info.PRGROMSize)
	}

	// 1 CHR-ROM bank = 8 KB
	expectedCHRSize := 8 * 1024
	if info.CHRROMSize != expectedCHRSize {
		t.Errorf("Expected CHR-ROM size %d, got %d", expectedCHRSize, info.CHRROMSize)
	}

	// Should be NTSC
	if info.TVSystem != NESTVSystemNTSC {
		t.Errorf("Expected NTSC TV system, got %d", info.TVSystem)
	}

	// Should not be NES 2.0
	if info.IsNES20 {
		t.Error("Expected iNES format, got NES 2.0")
	}
}

func TestIsNESROM(t *testing.T) {
	nesPath := filepath.Join(testutil.ROMsPath(t), "BombSweeper.nes")

	file, err := os.Open(nesPath)
	if err != nil {
		t.Fatalf("Failed to open NES file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat NES file: %v", err)
	}

	if !IsNESROM(file, stat.Size()) {
		t.Error("Expected IsNESROM to return true for BombSweeper.nes")
	}
}

func TestIsNESROM_NotNES(t *testing.T) {
	// Test that a GBA ROM is not detected as NES
	gbaPath := filepath.Join(testutil.ROMsPath(t), "AGB_Rogue.gba")

	file, err := os.Open(gbaPath)
	if err != nil {
		t.Fatalf("Failed to open GBA file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat GBA file: %v", err)
	}

	if IsNESROM(file, stat.Size()) {
		t.Error("Expected IsNESROM to return false for GBA ROM")
	}
}
