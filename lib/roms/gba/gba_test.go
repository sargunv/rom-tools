package gba

import (
	"os"
	"testing"
)

func TestParseGBA(t *testing.T) {
	romPath := "testdata/AGB_Rogue.gba"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseGBA(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseGBA() error = %v", err)
	}

	if info.Title != "ROGUE" {
		t.Errorf("Expected title 'ROGUE', got '%s'", info.Title)
	}

	if info.GameCode != "AAAA" {
		t.Errorf("Expected game code 'AAAA', got '%s'", info.GameCode)
	}

	if info.MakerCode != "AA" {
		t.Errorf("Expected maker code 'AA', got '%s'", info.MakerCode)
	}
}
