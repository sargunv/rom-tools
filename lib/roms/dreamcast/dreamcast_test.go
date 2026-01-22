package dreamcast

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestParseDreamcast(t *testing.T) {
	data, err := os.ReadFile("testdata/jet_set_radio_jp.bin")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	info, err := ParseDreamcast(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ParseDreamcast failed: %v", err)
	}

	// Verify all fields
	if info.Title != "JET SET RADIO" {
		t.Errorf("Title = %q, want %q", info.Title, "JET SET RADIO")
	}
	if info.ProductNumber != "HDR-0078" {
		t.Errorf("ProductNumber = %q, want %q", info.ProductNumber, "HDR-0078")
	}
	if info.MakerID != "SEGA ENTERPRISES" {
		t.Errorf("MakerID = %q, want %q", info.MakerID, "SEGA ENTERPRISES")
	}
	if info.DeviceInfo != "BCA4 GD-ROM1/1" {
		t.Errorf("DeviceInfo = %q, want %q", info.DeviceInfo, "BCA4 GD-ROM1/1")
	}
	if info.AreaSymbols != "J       " {
		t.Errorf("AreaSymbols = %q, want %q", info.AreaSymbols, "J       ")
	}
	if info.Peripherals != "0799A10" {
		t.Errorf("Peripherals = %q, want %q", info.Peripherals, "0799A10")
	}
	if info.Version != "V1.006" {
		t.Errorf("Version = %q, want %q", info.Version, "V1.006")
	}
	expectedDate := time.Date(2000, 6, 1, 0, 0, 0, 0, time.UTC)
	if !info.ReleaseDate.Equal(expectedDate) {
		t.Errorf("ReleaseDate = %v, want %v", info.ReleaseDate, expectedDate)
	}
	if info.BootFilename != "1ST_READ.BIN" {
		t.Errorf("BootFilename = %q, want %q", info.BootFilename, "1ST_READ.BIN")
	}
	if info.SWMakerName != "SEGA ENTERPRISES" {
		t.Errorf("SWMakerName = %q, want %q", info.SWMakerName, "SEGA ENTERPRISES")
	}
}

func TestParseDreamcast_InvalidMagic(t *testing.T) {
	data := make([]byte, 256)
	copy(data, "INVALID MAGIC   ")

	_, err := ParseDreamcast(bytes.NewReader(data), int64(len(data)))
	if err == nil {
		t.Error("expected error for invalid magic, got nil")
	}
}

func TestParseDreamcast_TooSmall(t *testing.T) {
	data := make([]byte, 100)

	_, err := ParseDreamcast(bytes.NewReader(data), int64(len(data)))
	if err == nil {
		t.Error("expected error for too-small input, got nil")
	}
}

func TestParseDreamcast_InvalidDate(t *testing.T) {
	// Create a valid header but with invalid date
	data := make([]byte, 256)
	copy(data[0:16], "SEGA SEGAKATANA ")
	copy(data[0x50:0x58], "BADDATE!") // Invalid date format

	info, err := ParseDreamcast(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ParseDreamcast failed: %v", err)
	}

	// Invalid date should result in zero time
	if !info.ReleaseDate.IsZero() {
		t.Errorf("ReleaseDate = %v, want zero time for invalid date", info.ReleaseDate)
	}
}
