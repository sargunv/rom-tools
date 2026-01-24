package n64

import (
	"os"
	"testing"
)

func TestParseV64(t *testing.T) {
	romPath := "testdata/flames.v64"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseN64(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseN64() error = %v", err)
	}

	if info.ByteOrder != N64ByteSwapped {
		t.Errorf("Expected byte order v64, got %s", info.ByteOrder)
	}
}
