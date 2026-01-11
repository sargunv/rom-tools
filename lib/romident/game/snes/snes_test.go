package snes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestParseSNES(t *testing.T) {
	snesPath := filepath.Join(testutil.ROMsPath(t), "col15.sfc")

	file, err := os.Open(snesPath)
	if err != nil {
		t.Fatalf("Failed to open SNES file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat SNES file: %v", err)
	}

	info, err := ParseSNES(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseSNES() error = %v", err)
	}

	// Game title
	expectedTitle := "32,768 color demo"
	if info.Title != expectedTitle {
		t.Errorf("Expected title %q, got %q", expectedTitle, info.Title)
	}

	// Should be LoROM
	if info.MapMode != SNESMapModeLoROM {
		t.Errorf("Expected map mode 0x20 (LoROM), got 0x%02X", info.MapMode)
	}

	// Destination code JP
	if info.DestinationCode != 0x00 {
		t.Errorf("Expected destination code 0x00, got 0x%02X", info.DestinationCode)
	}

	// Checksum should validate (checksum + complement = 0xFFFF)
	if info.Checksum+info.ChecksumComplement != 0xFFFF {
		t.Errorf("Checksum validation failed: 0x%04X + 0x%04X != 0xFFFF",
			info.Checksum, info.ChecksumComplement)
	}
}

func TestIsSNESROM(t *testing.T) {
	snesPath := filepath.Join(testutil.ROMsPath(t), "col15.sfc")

	file, err := os.Open(snesPath)
	if err != nil {
		t.Fatalf("Failed to open SNES file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat SNES file: %v", err)
	}

	if !IsSNESROM(file, stat.Size()) {
		t.Error("Expected IsSNESROM to return true for blt.sfc")
	}
}

func TestIsSNESROM_NotSNES(t *testing.T) {
	// Test that an NES ROM is not detected as SNES
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

	if IsSNESROM(file, stat.Size()) {
		t.Error("Expected IsSNESROM to return false for NES ROM")
	}
}
