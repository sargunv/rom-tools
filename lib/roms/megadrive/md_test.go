package megadrive

import (
	"os"
	"testing"
)

func TestParseMD(t *testing.T) {
	romPath := "testdata/Censor_Intro.md"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseMD(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseMD() error = %v", err)
	}

	// Verify parsing was successful
	if stat.Size() < mdHeaderStart+mdHeaderSize {
		t.Errorf("File too small for MD header")
	}

	_ = info // Ensure info is used
}
