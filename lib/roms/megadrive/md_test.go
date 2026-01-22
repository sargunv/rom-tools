package megadrive

import (
	"os"
	"testing"
)

func TestParseRegionCodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Region
	}{
		{"empty", "", 0},
		{"J only", "J", RegionDomestic60Hz},
		{"U only", "U", RegionOverseas60Hz},
		{"E only", "E", RegionOverseas50Hz}, // must not be parsed as hex 0x0E
		{"JUE", "JUE", RegionDomestic60Hz | RegionOverseas60Hz | RegionOverseas50Hz},
		{"JU", "JU", RegionDomestic60Hz | RegionOverseas60Hz},
		{"new-style 0", "0", 0},
		{"new-style 1 (domestic 60Hz)", "1", RegionDomestic60Hz},
		{"new-style 4 (overseas 60Hz)", "4", RegionOverseas60Hz},
		{"new-style 5 (domestic+overseas 60Hz)", "5", RegionDomestic60Hz | RegionOverseas60Hz},
		{"new-style D (all regions)", "D", RegionDomestic60Hz | RegionOverseas60Hz | RegionOverseas50Hz},
		{"new-style F", "F", 0x0F},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pad input to 16 bytes like the real header field
			data := make([]byte, 16)
			copy(data, tt.input)
			got := parseRegionCodes(data)
			if got != tt.want {
				t.Errorf("parseRegionCodes(%q) = 0x%02X, want 0x%02X", tt.input, got, tt.want)
			}
		})
	}
}

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

	info, err := Parse(file, stat.Size())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify SourceFormat is set correctly for MD files
	if info.SourceFormat != FormatMD {
		t.Errorf("SourceFormat = %v, want FormatMD", info.SourceFormat)
	}

	// Verify basic parsing succeeded
	if info.SystemType == "" {
		t.Error("SystemType should not be empty")
	}
}
