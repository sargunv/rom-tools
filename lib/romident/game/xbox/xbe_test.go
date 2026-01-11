package xbox

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestParseXBE(t *testing.T) {
	xbePath := filepath.Join(testutil.ROMsPath(t), "xromwell", "default.xbe")

	file, err := os.Open(xbePath)
	if err != nil {
		t.Fatalf("Failed to open XBE file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat XBE file: %v", err)
	}

	info, err := ParseXBE(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseXBE() error = %v", err)
	}

	if info.Title == "" {
		t.Error("Expected non-empty title")
	}

	if info.TitleIDHex == "" {
		t.Error("Expected non-empty title ID hex")
	}
}
