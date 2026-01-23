package megadrive

import (
	"bytes"
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

	// Verify Is32X is false for regular MD ROMs
	if info.Is32X {
		t.Error("Is32X should be false for regular MD ROMs")
	}
}

func TestParse32X(t *testing.T) {
	// Create synthetic 32X ROM data
	// Minimum size for 32X detection: md32XHeaderOffset + md32XMagicLen = 0x3C4
	data := make([]byte, 0x400)

	// Set up valid MD header at offset 0x100
	copy(data[mdSystemTypeOffset:], "SEGA 32X        ") // 16 bytes
	copy(data[mdCopyrightOffset:], "(C)TEST 2024.JAN")  // 16 bytes
	copy(data[mdDomesticTitleOff:], "TEST 32X GAME")    // Domestic title
	copy(data[mdOverseasTitleOff:], "TEST 32X GAME")    // Overseas title
	copy(data[mdSerialNumberOffset:], "GM 00000000-00") // Serial
	copy(data[mdDeviceSupportOff:], "J")                // Device support
	copy(data[mdRegionOffset:], "JUE")                  // Region

	// Set up MARS header at offset 0x3C0
	copy(data[md32XHeaderOffset:], "MARS CHECK MODE")

	reader := bytes.NewReader(data)
	info, err := Parse(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify Is32X is true
	if !info.Is32X {
		t.Error("Is32X should be true for 32X ROMs with MARS header")
	}

	// Verify other fields are still parsed correctly
	if info.SystemType != "SEGA 32X" {
		t.Errorf("SystemType = %q, want %q", info.SystemType, "SEGA 32X")
	}
}

func TestParse32X_NoMarsHeader(t *testing.T) {
	// Create synthetic MD ROM data without MARS header
	// Large enough for 32X detection, but without the magic string
	data := make([]byte, 0x400)

	// Set up valid MD header at offset 0x100
	copy(data[mdSystemTypeOffset:], "SEGA MEGA DRIVE ") // 16 bytes
	copy(data[mdCopyrightOffset:], "(C)TEST 2024.JAN")  // 16 bytes
	copy(data[mdDomesticTitleOff:], "TEST MD GAME")     // Domestic title
	copy(data[mdOverseasTitleOff:], "TEST MD GAME")     // Overseas title
	copy(data[mdSerialNumberOffset:], "GM 00000000-00") // Serial
	copy(data[mdDeviceSupportOff:], "J")                // Device support
	copy(data[mdRegionOffset:], "JUE")                  // Region

	// No MARS header - leave offset 0x3C0 as zeros

	reader := bytes.NewReader(data)
	info, err := Parse(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify Is32X is false (no MARS header)
	if info.Is32X {
		t.Error("Is32X should be false for MD ROMs without MARS header")
	}
}

func TestParse32X_SmallFile(t *testing.T) {
	// Create synthetic MD ROM that is too small for 32X detection
	// Just large enough for the MD header (0x200 bytes)
	data := make([]byte, 0x200)

	// Set up valid MD header at offset 0x100
	copy(data[mdSystemTypeOffset:], "SEGA MEGA DRIVE ") // 16 bytes
	copy(data[mdCopyrightOffset:], "(C)TEST 2024.JAN")  // 16 bytes
	copy(data[mdDomesticTitleOff:], "TEST SMALL")       // Domestic title
	copy(data[mdOverseasTitleOff:], "TEST SMALL")       // Overseas title
	copy(data[mdSerialNumberOffset:], "GM 00000000-00") // Serial
	copy(data[mdDeviceSupportOff:], "J")                // Device support
	copy(data[mdRegionOffset:], "JUE")                  // Region

	reader := bytes.NewReader(data)
	info, err := Parse(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify Is32X is false for small files (cannot check MARS header)
	if info.Is32X {
		t.Error("Is32X should be false for files too small to contain MARS header")
	}
}
