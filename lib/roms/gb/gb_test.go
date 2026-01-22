package gb

import (
	"os"
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

func TestParseGB_GB(t *testing.T) {
	romPath := "testdata/gbtictac.gb"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseGB(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseGB() error = %v", err)
	}

	if info.Platform != core.PlatformGB {
		t.Errorf("Expected platform %s, got %s", core.PlatformGB, info.Platform)
	}

	if info.Title != "TIC-TAC-TOE" {
		t.Errorf("Expected title 'TIC-TAC-TOE', got '%s'", info.Title)
	}
}

func TestParseGB_GBC(t *testing.T) {
	romPath := "testdata/JUMPMAN86.GBC"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseGB(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseGB() error = %v", err)
	}

	if info.Platform != core.PlatformGBC {
		t.Errorf("Expected platform %s, got %s", core.PlatformGBC, info.Platform)
	}
}
