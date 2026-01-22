package gb

import (
	"os"
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

func TestParseGB_GB(t *testing.T) {
	romPath := "testdata/gbtictac.gb"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseGB(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseGB() error = %v", err)
	}

	// Platform detection
	if info.Platform != core.PlatformGB {
		t.Errorf("Expected platform %s, got %s", core.PlatformGB, info.Platform)
	}

	// Title
	if info.Title != "TIC-TAC-TOE" {
		t.Errorf("Expected title 'TIC-TAC-TOE', got '%s'", info.Title)
	}

	// CGB flag (original GB - no CGB support)
	if info.CGBFlag != CGBFlagNone {
		t.Errorf("Expected CGBFlag %#02x, got %#02x", CGBFlagNone, info.CGBFlag)
	}

	// SGB flag (no SGB support)
	if info.SGBFlag != SGBFlagNone {
		t.Errorf("Expected SGBFlag %#02x, got %#02x", SGBFlagNone, info.SGBFlag)
	}

	// Cartridge type (ROM only)
	if info.CartridgeType != 0x00 {
		t.Errorf("Expected CartridgeType 0x00, got %#02x", info.CartridgeType)
	}

	// ROM size (32KB)
	if info.ROMSize != ROMSize32KB {
		t.Errorf("Expected ROMSize %#02x, got %#02x", ROMSize32KB, info.ROMSize)
	}

	// RAM size (none)
	if info.RAMSize != RAMSizeNone {
		t.Errorf("Expected RAMSize %#02x, got %#02x", RAMSizeNone, info.RAMSize)
	}

	// Destination (Japan)
	if info.Destination != DestinationJapan {
		t.Errorf("Expected Destination %#02x, got %#02x", DestinationJapan, info.Destination)
	}

	// Licensee code (old format 0x00)
	if info.LicenseeCode != "00" {
		t.Errorf("Expected LicenseeCode '00', got '%s'", info.LicenseeCode)
	}

	// Version
	if info.Version != 1 {
		t.Errorf("Expected Version 1, got %d", info.Version)
	}

	// Header checksum
	if info.HeaderChecksum != 0x00 {
		t.Errorf("Expected HeaderChecksum 0x00, got %#02x", info.HeaderChecksum)
	}

	// Global checksum
	if info.GlobalChecksum != 0xA9E1 {
		t.Errorf("Expected GlobalChecksum 0xA9E1, got %#04x", info.GlobalChecksum)
	}

	// ManufacturerCode should be empty for original GB games
	if info.ManufacturerCode != "" {
		t.Errorf("Expected empty ManufacturerCode for GB game, got '%s'", info.ManufacturerCode)
	}
}

func TestParseGB_GBC(t *testing.T) {
	romPath := "testdata/JUMPMAN86.GBC"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseGB(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseGB() error = %v", err)
	}

	// Platform detection
	if info.Platform != core.PlatformGBC {
		t.Errorf("Expected platform %s, got %s", core.PlatformGBC, info.Platform)
	}

	// Title (11 chars for GBC format)
	if info.Title != "JUMPMAN 86" {
		t.Errorf("Expected title 'JUMPMAN 86', got '%s'", info.Title)
	}

	// CGB flag (CGB only)
	if info.CGBFlag != CGBFlagRequired {
		t.Errorf("Expected CGBFlag %#02x, got %#02x", CGBFlagRequired, info.CGBFlag)
	}

	// Cartridge type (MBC5+RAM+BATTERY = 0x19)
	if info.CartridgeType != 0x19 {
		t.Errorf("Expected CartridgeType 0x19, got %#02x", info.CartridgeType)
	}

	// ROM size (64KB = 0x01)
	if info.ROMSize != ROMSize64KB {
		t.Errorf("Expected ROMSize %#02x, got %#02x", ROMSize64KB, info.ROMSize)
	}

	// RAM size (none)
	if info.RAMSize != RAMSizeNone {
		t.Errorf("Expected RAMSize %#02x, got %#02x", RAMSizeNone, info.RAMSize)
	}

	// Destination (Overseas)
	if info.Destination != DestinationOverseas {
		t.Errorf("Expected Destination %#02x, got %#02x", DestinationOverseas, info.Destination)
	}

	// Licensee code (new format - 0x33 triggers new licensee lookup, "±°" from header)
	// The new licensee code is 2 bytes at 0x144-0x145: 0xB1, 0xB0
	if info.LicenseeCode != "\xb1\xb0" {
		t.Errorf("Expected LicenseeCode '\\xb1\\xb0', got %q", info.LicenseeCode)
	}

	// Version
	if info.Version != 0 {
		t.Errorf("Expected Version 0, got %d", info.Version)
	}

	// Global checksum should be non-zero
	if info.GlobalChecksum == 0 {
		t.Errorf("Expected non-zero GlobalChecksum, got %#04x", info.GlobalChecksum)
	}
}

func TestParseGB_FileTooSmall(t *testing.T) {
	// Create a file that's too small for a GB header
	data := make([]byte, 0x100) // Header starts at 0x100, so this is too small

	_, err := ParseGB(&mockReaderAt{data: data}, int64(len(data)))
	if err == nil {
		t.Error("Expected error for file too small, got nil")
	}
}

// mockReaderAt implements io.ReaderAt for testing
type mockReaderAt struct {
	data []byte
}

func (m *mockReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.data)) {
		return 0, nil
	}
	n = copy(p, m.data[off:])
	return n, nil
}
