package md

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestParseSMD(t *testing.T) {
	smdPath := filepath.Join(testutil.ROMsPath(t), "Censor_Intro.smd")

	file, err := os.Open(smdPath)
	if err != nil {
		t.Fatalf("Failed to open SMD file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat SMD file: %v", err)
	}

	info, err := ParseSMD(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseSMD() error = %v", err)
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

func TestIsSMDROM(t *testing.T) {
	smdPath := filepath.Join(testutil.ROMsPath(t), "Censor_Intro.smd")

	file, err := os.Open(smdPath)
	if err != nil {
		t.Fatalf("Failed to open SMD file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat SMD file: %v", err)
	}

	if !IsSMDROM(file, stat.Size()) {
		t.Error("Expected IsSMDROM to return true for Censor_Intro.smd")
	}
}

func TestIsSMDROM_NotSMD(t *testing.T) {
	// Test that a raw MD ROM is not detected as SMD
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

	if IsSMDROM(file, stat.Size()) {
		t.Error("Expected IsSMDROM to return false for raw MD ROM")
	}
}
