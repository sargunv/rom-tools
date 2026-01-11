package nds

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestParseNDS(t *testing.T) {
	ndsPath := filepath.Join(testutil.ROMsPath(t), "MixedCubes.nds")

	file, err := os.Open(ndsPath)
	if err != nil {
		t.Fatalf("Failed to open NDS file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat NDS file: %v", err)
	}

	info, err := ParseNDS(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseNDS() error = %v", err)
	}

	// MixedCubes is homebrew with minimal header
	// Game code should be AXXE (homebrew marker)
	if info.GameCode != "AXXE" {
		t.Errorf("Expected game code 'AXXE', got %q", info.GameCode)
	}

	// Unit code 0x00 = Original DS
	if info.UnitCode != NDSUnitCodeDS {
		t.Errorf("Expected unit code DS (0x00), got 0x%02X", info.UnitCode)
	}

	if info.Platform != NDSPlatformDS {
		t.Errorf("Expected platform %q, got %q", NDSPlatformDS, info.Platform)
	}
}

func TestIsNDSROM(t *testing.T) {
	ndsPath := filepath.Join(testutil.ROMsPath(t), "MixedCubes.nds")

	file, err := os.Open(ndsPath)
	if err != nil {
		t.Fatalf("Failed to open NDS file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat NDS file: %v", err)
	}

	if !IsNDSROM(file, stat.Size()) {
		t.Error("Expected IsNDSROM to return true for MixedCubes.nds")
	}
}

func TestParseNDS_Homebrew(t *testing.T) {
	// NitroTracker is a homebrew NDS ROM without the official Nintendo logo
	ndsPath := filepath.Join(testutil.ROMsPath(t), "NitroTracker.nds")

	file, err := os.Open(ndsPath)
	if err != nil {
		t.Fatalf("Failed to open NDS file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat NDS file: %v", err)
	}

	if !IsNDSROM(file, stat.Size()) {
		t.Error("Expected IsNDSROM to return true for NitroTracker.nds (homebrew)")
	}

	info, err := ParseNDS(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseNDS() error = %v", err)
	}

	// NitroTracker has game code "####" (homebrew marker)
	if info.GameCode != "####" {
		t.Errorf("Expected game code '####', got %q", info.GameCode)
	}

	if info.Platform != NDSPlatformDS {
		t.Errorf("Expected platform %q, got %q", NDSPlatformDS, info.Platform)
	}
}

func TestIsNDSROM_NotNDS(t *testing.T) {
	// Test that a GBA ROM is not detected as NDS
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

	if IsNDSROM(file, stat.Size()) {
		t.Error("Expected IsNDSROM to return false for GBA ROM")
	}
}
