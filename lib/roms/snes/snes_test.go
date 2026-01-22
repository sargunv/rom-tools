package snes

import (
	"os"
	"testing"
)

func TestParseSNES(t *testing.T) {
	romPath := "testdata/col15.sfc"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseSNES(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseSNES() error = %v", err)
	}

	// Verify the file was parsed without error
	if info.Title == "" {
		t.Errorf("Expected non-empty title")
	}
}
