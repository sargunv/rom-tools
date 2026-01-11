package xbox

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestParseXISO(t *testing.T) {
	xisoPath := filepath.Join(testutil.ROMsPath(t), "xromwell.xiso.iso")

	file, err := os.Open(xisoPath)
	if err != nil {
		t.Fatalf("Failed to open XISO file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat XISO file: %v", err)
	}

	info, err := ParseXISO(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseXISO() error = %v", err)
	}

	if info.Title != "Xromwell" {
		t.Errorf("Expected title 'Xromwell', got '%s'", info.Title)
	}

	if info.TitleIDHex != "00000000" {
		t.Errorf("Expected title ID hex '00000000', got '%s'", info.TitleIDHex)
	}

	// RegionFlags should include NA (0x1), JP (0x2), EU/AU (0x4), and DEBUG (0x80000000)
	expectedFlags := uint32(XboxRegionNA | XboxRegionJapan | XboxRegionEUAU | XboxRegionDebug)
	if info.RegionFlags != expectedFlags {
		t.Errorf("Expected region flags 0x%08X, got 0x%08X", expectedFlags, info.RegionFlags)
	}
}
