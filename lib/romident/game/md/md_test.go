package md

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestParseMD(t *testing.T) {
	mdPath := filepath.Join(testutil.ROMsPath(t), "Censor_Intro.md")

	file, err := os.Open(mdPath)
	if err != nil {
		t.Fatalf("Failed to open MD file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat MD file: %v", err)
	}

	info, err := ParseMD(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseMD() error = %v", err)
	}

	if info.SystemType != "SEGA MEGA DRIVE" {
		t.Errorf("Expected system type 'SEGA MEGA DRIVE', got %q", info.SystemType)
	}

	if info.OverseasTitle != "CENSOR INTRO 1" {
		t.Errorf("Expected overseas title 'CENSOR INTRO 1', got %q", info.OverseasTitle)
	}

	if info.DomesticTitle != "CENSOR INTRO 1" {
		t.Errorf("Expected domestic title 'CENSOR INTRO 1', got %q", info.DomesticTitle)
	}

	if info.SerialNumber != "" {
		t.Errorf("Expected serial number '', got %q", info.SerialNumber)
	}

	if info.Copyright != "(C)CEN! 1992.OCT" {
		t.Errorf("Expected copyright '(C)CEN! 1992.OCT', got %q", info.Copyright)
	}

	// Check regions (should have J, U, E)
	expectedRegions := map[byte]bool{'J': true, 'U': true, 'E': true}
	if len(info.Regions) != 3 {
		t.Errorf("Expected 3 regions, got %d: %v", len(info.Regions), info.Regions)
	}
	for _, r := range info.Regions {
		if !expectedRegions[r] {
			t.Errorf("Unexpected region: %c", r)
		}
	}
}

func TestIsMDROM(t *testing.T) {
	mdPath := filepath.Join(testutil.ROMsPath(t), "Censor_Intro.md")

	file, err := os.Open(mdPath)
	if err != nil {
		t.Fatalf("Failed to open MD file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat MD file: %v", err)
	}

	if !IsMDROM(file, stat.Size()) {
		t.Error("Expected IsMDROM to return true for testrom.md")
	}
}

func TestIsMDROM_NotMD(t *testing.T) {
	// Test that a GBA ROM is not detected as MD
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

	if IsMDROM(file, stat.Size()) {
		t.Error("Expected IsMDROM to return false for GBA ROM")
	}
}

func TestParseMDRegionNewStyle(t *testing.T) {
	// Test new-style hex digit region parsing
	// F = 1111 = Japan (bit 0) + Americas (bit 2) + Europe (bit 3)
	regions := parseRegionCodes([]byte{'F'})

	if len(regions) != 3 {
		t.Fatalf("Expected 3 regions for 'F', got %d: %v", len(regions), regions)
	}

	expectedRegions := map[byte]bool{'J': true, 'U': true, 'E': true}
	for _, r := range regions {
		if !expectedRegions[r] {
			t.Errorf("Unexpected region: %c", r)
		}
	}
}
