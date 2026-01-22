package xbox

import (
	"os"
	"testing"
)

func TestParseXBE(t *testing.T) {
	romPath := "testdata/default.xbe"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseXBE(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseXBE() error = %v", err)
	}

	if info.Title != "Xromwell" {
		t.Errorf("Expected title 'Xromwell', got '%s'", info.Title)
	}
}
