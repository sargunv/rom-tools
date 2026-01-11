package gb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestParseGB(t *testing.T) {
	gbPath := filepath.Join(testutil.ROMsPath(t), "gbtictac.gb")

	file, err := os.Open(gbPath)
	if err != nil {
		t.Fatalf("Failed to open GB file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat GB file: %v", err)
	}

	info, err := ParseGB(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseGB() error = %v", err)
	}

	if info.Title != "TIC-TAC-TOE" {
		t.Errorf("Expected title 'TIC-TAC-TOE', got %q", info.Title)
	}

	if info.Platform != GBPlatformGB {
		t.Errorf("Expected platform %q, got %q", GBPlatformGB, info.Platform)
	}

	// Destination code 0x00 = Japanese
	if info.DestinationCode != 0x00 {
		t.Errorf("Expected destination code 0x00, got 0x%02X", info.DestinationCode)
	}

	// Should be original GB (no CGB flag)
	if info.CGBFlag != CGBFlagNone {
		t.Errorf("Expected CGB flag 0x00 (none), got 0x%02X", info.CGBFlag)
	}
}

func TestIsGBROM(t *testing.T) {
	gbPath := filepath.Join(testutil.ROMsPath(t), "gbtictac.gb")

	file, err := os.Open(gbPath)
	if err != nil {
		t.Fatalf("Failed to open GB file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat GB file: %v", err)
	}

	if !IsGBROM(file, stat.Size()) {
		t.Error("Expected IsGBROM to return true for gbtictac.gb")
	}
}

func TestParseGBC(t *testing.T) {
	gbcPath := filepath.Join(testutil.ROMsPath(t), "JUMPMAN86.GBC")

	file, err := os.Open(gbcPath)
	if err != nil {
		t.Fatalf("Failed to open GBC file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat GBC file: %v", err)
	}

	info, err := ParseGB(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseGB() error = %v", err)
	}

	if info.Title != "JUMPMAN 86" {
		t.Errorf("Expected title 'JUMPMAN 86', got %q", info.Title)
	}

	if info.Platform != GBPlatformGBC {
		t.Errorf("Expected platform %q, got %q", GBPlatformGBC, info.Platform)
	}

	// Should be GBC-only (0xC0) or GBC-compatible (0x80)
	if info.CGBFlag != CGBFlagSupported && info.CGBFlag != CGBFlagRequired {
		t.Errorf("Expected CGB flag 0x80 or 0xC0, got 0x%02X", info.CGBFlag)
	}
}

func TestIsGBROM_NotGB(t *testing.T) {
	// Test that a GBA ROM is not detected as GB
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

	if IsGBROM(file, stat.Size()) {
		t.Error("Expected IsGBROM to return false for GBA ROM")
	}
}
