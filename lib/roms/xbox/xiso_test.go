package xbox

import (
	"os"
	"testing"
)

func TestParseXISO(t *testing.T) {
	romPath := "testdata/xromwell.xiso.iso"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseXISO(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseXISO() error = %v", err)
	}

	if info.Title != "Xromwell" {
		t.Errorf("Expected title 'Xromwell', got '%s'", info.Title)
	}
}
