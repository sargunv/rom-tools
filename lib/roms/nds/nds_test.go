package nds

import (
	"os"
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

func TestParseNDS(t *testing.T) {
	romPath := "testdata/MixedCubes.nds"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseNDS(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseNDS() error = %v", err)
	}

	if info.Platform != core.PlatformNDS {
		t.Errorf("Expected platform %s, got %s", core.PlatformNDS, info.Platform)
	}
}
