package nes

import (
	"os"
	"testing"
)

func TestParseNES(t *testing.T) {
	romPath := "testdata/BombSweeper.nes"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseNES(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseNES() error = %v", err)
	}

	// Verify the file was parsed without error - NES format doesn't include title
	if info.PRGROMSize <= 0 {
		t.Errorf("Expected positive PRG ROM size, got %d", info.PRGROMSize)
	}
}
