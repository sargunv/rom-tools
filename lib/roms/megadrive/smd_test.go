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

	info, err := Parse(file, stat.Size())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify SourceFormat is set correctly for SMD files
	if info.SourceFormat != FormatSMD {
		t.Errorf("SourceFormat = %v, want FormatSMD", info.SourceFormat)
	}

	// Verify basic parsing succeeded
	if info.SystemType == "" {
		t.Error("SystemType should not be empty")
	}
}
