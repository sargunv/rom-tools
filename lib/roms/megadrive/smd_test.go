package megadrive

import (
	"os"
	"testing"
)

func TestParseSMD(t *testing.T) {
	romPath := "testdata/Censor_Intro.smd"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseSMD(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseSMD() error = %v", err)
	}

	// Verify parsing was successful
	_ = info
}
