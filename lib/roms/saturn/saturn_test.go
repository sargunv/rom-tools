package saturn

import (
	"bytes"
	"testing"
)

func TestParseSaturn(t *testing.T) {
	// Create a synthetic valid Saturn header
	data := make([]byte, 256)

	// Hardware ID (magic)
	copy(data[0x00:], "SEGA SEGASATURN ")
	// Maker ID
	copy(data[0x10:], "SEGA ENTERPRISES")
	// Product Number
	copy(data[0x20:], "MK-81022  ")
	// Version
	copy(data[0x2A:], "V1.000")
	// Release Date
	copy(data[0x30:], "19961122")
	// Device Info
	copy(data[0x38:], "CD-1/1  ")
	// Area Symbols (Japan + USA + Europe)
	copy(data[0x40:], "JUE             ")
	// Peripherals
	copy(data[0x50:], "J               ")
	// Title
	copy(data[0x60:], "NIGHTS INTO DREAMS...")

	info, err := ParseSaturn(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ParseSaturn failed: %v", err)
	}

	// Verify all fields
	if info.Title != "NIGHTS INTO DREAMS..." {
		t.Errorf("Title = %q, want %q", info.Title, "NIGHTS INTO DREAMS...")
	}
	if info.MakerID != "SEGA ENTERPRISES" {
		t.Errorf("MakerID = %q, want %q", info.MakerID, "SEGA ENTERPRISES")
	}
	if info.ProductNumber != "MK-81022" {
		t.Errorf("ProductNumber = %q, want %q", info.ProductNumber, "MK-81022")
	}
	if info.Version != "V1.000" {
		t.Errorf("Version = %q, want %q", info.Version, "V1.000")
	}
	if info.ReleaseDate.Year() != 1996 || info.ReleaseDate.Month() != 11 || info.ReleaseDate.Day() != 22 {
		t.Errorf("ReleaseDate = %v, want 1996-11-22", info.ReleaseDate)
	}
	if info.DeviceInfo != "CD-1/1" {
		t.Errorf("DeviceInfo = %q, want %q", info.DeviceInfo, "CD-1/1")
	}
	expectedArea := AreaJapan | AreaNorthAmerica | AreaEurope
	if info.Area != expectedArea {
		t.Errorf("Area = %d, want %d (JUE)", info.Area, expectedArea)
	}
	if info.Peripherals != "J" {
		t.Errorf("Peripherals = %q, want %q", info.Peripherals, "J")
	}
}

func TestParseSaturn_InvalidMagic(t *testing.T) {
	data := make([]byte, 256)
	copy(data, "INVALID MAGIC   ")

	_, err := ParseSaturn(bytes.NewReader(data), int64(len(data)))
	if err == nil {
		t.Error("expected error for invalid magic, got nil")
	}
}

func TestParseSaturn_TooSmall(t *testing.T) {
	data := make([]byte, 100)

	_, err := ParseSaturn(bytes.NewReader(data), int64(len(data)))
	if err == nil {
		t.Error("expected error for too-small input, got nil")
	}
}

func TestParseSaturn_InvalidDate(t *testing.T) {
	// Create a valid header but with invalid date
	data := make([]byte, 256)
	copy(data[0:16], "SEGA SEGASATURN ")
	copy(data[0x30:0x38], "BADDATE!") // Invalid date format

	info, err := ParseSaturn(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ParseSaturn failed: %v", err)
	}

	// Invalid date should result in zero time
	if !info.ReleaseDate.IsZero() {
		t.Errorf("ReleaseDate = %v, want zero time for invalid date", info.ReleaseDate)
	}
}

func TestParseSaturn_AllAreas(t *testing.T) {
	// Test all area codes
	data := make([]byte, 256)
	copy(data[0:16], "SEGA SEGASATURN ")
	copy(data[0x40:], "JTUBKAEL        ") // All areas

	info, err := ParseSaturn(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ParseSaturn failed: %v", err)
	}

	expectedArea := AreaJapan | AreaAsiaNTSC | AreaNorthAmerica | AreaBrazil |
		AreaKorea | AreaAsiaPAL | AreaEurope | AreaLatinAmerica
	if info.Area != expectedArea {
		t.Errorf("Area = %d, want %d (all areas)", info.Area, expectedArea)
	}
}
