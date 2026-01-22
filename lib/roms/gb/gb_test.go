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
		t.Errorf("Platform: expected %s, got %s", core.PlatformGB, info.Platform)
	}

	// Title (old format, no manufacturer code)
	if info.Title != "TIC-TAC-TOE" {
		t.Errorf("Title: expected 'TIC-TAC-TOE', got '%s'", info.Title)
	}

	// Manufacturer code should be empty for old-style GB games
	if info.ManufacturerCode != "" {
		t.Errorf("ManufacturerCode: expected empty, got '%s'", info.ManufacturerCode)
	}

	// CGB flag
	if info.CGBFlag != GBCGBFlagNone {
		t.Errorf("CGBFlag: expected %#x, got %#x", GBCGBFlagNone, info.CGBFlag)
	}

	// SGB flag
	if info.SGBFlag != GBSGBFlagNone {
		t.Errorf("SGBFlag: expected %#x, got %#x", GBSGBFlagNone, info.SGBFlag)
	}

	// Cartridge type (ROM only)
	if info.CartridgeType != 0x00 {
		t.Errorf("CartridgeType: expected 0x00, got %#x", info.CartridgeType)
	}

	// ROM size (32KB)
	if info.ROMSize != GBROMSize32KB {
		t.Errorf("ROMSize: expected %#x (32KB), got %#x", GBROMSize32KB, info.ROMSize)
	}

	// RAM size (none)
	if info.RAMSize != GBRAMSizeNone {
		t.Errorf("RAMSize: expected %#x (none), got %#x", GBRAMSizeNone, info.RAMSize)
	}

	// Destination code (Japan)
	if info.DestinationCode != 0x00 {
		t.Errorf("DestinationCode: expected 0x00 (Japan), got %#x", info.DestinationCode)
	}

	// Licensee code (old format, 0x14B = 0x00)
	if info.LicenseeCode != "00" {
		t.Errorf("LicenseeCode: expected '00', got '%s'", info.LicenseeCode)
	}

	// Version
	if info.Version != 1 {
		t.Errorf("Version: expected 1, got %d", info.Version)
	}

	// Header checksum
	if info.HeaderChecksum != 0x00 {
		t.Errorf("HeaderChecksum: expected 0x00, got %#x", info.HeaderChecksum)
	}

	// Global checksum
	if info.GlobalChecksum != 0xa9e1 {
		t.Errorf("GlobalChecksum: expected 0xa9e1, got %#x", info.GlobalChecksum)
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
		t.Errorf("Platform: expected %s, got %s", core.PlatformGBC, info.Platform)
	}

	// Title - early CGB game without manufacturer code (spaces in mfg bytes)
	// Should parse full title area
	if info.Title != "JUMPMAN 86" {
		t.Errorf("Title: expected 'JUMPMAN 86', got '%s'", info.Title)
	}

	// Manufacturer code should be empty (bytes are spaces, not uppercase ASCII)
	if info.ManufacturerCode != "" {
		t.Errorf("ManufacturerCode: expected empty, got '%s'", info.ManufacturerCode)
	}

	// CGB flag (CGB only)
	if info.CGBFlag != GBCGBFlagRequired {
		t.Errorf("CGBFlag: expected %#x (required), got %#x", GBCGBFlagRequired, info.CGBFlag)
	}

	// SGB flag
	if info.SGBFlag != GBSGBFlagNone {
		t.Errorf("SGBFlag: expected %#x, got %#x", GBSGBFlagNone, info.SGBFlag)
	}

	// Cartridge type (MBC5)
	if info.CartridgeType != 0x19 {
		t.Errorf("CartridgeType: expected 0x19 (MBC5), got %#x", info.CartridgeType)
	}

	// ROM size (64KB)
	if info.ROMSize != GBROMSize64KB {
		t.Errorf("ROMSize: expected %#x (64KB), got %#x", GBROMSize64KB, info.ROMSize)
	}

	// RAM size (none)
	if info.RAMSize != GBRAMSizeNone {
		t.Errorf("RAMSize: expected %#x (none), got %#x", GBRAMSizeNone, info.RAMSize)
	}

	// Destination code (Overseas)
	if info.DestinationCode != 0x01 {
		t.Errorf("DestinationCode: expected 0x01 (Overseas), got %#x", info.DestinationCode)
	}

	// Licensee code (new format, old licensee = 0x33)
	// The bytes at 0x144-0x145 are 0xb1 0xb0
	if info.LicenseeCode != "\xb1\xb0" {
		t.Errorf("LicenseeCode: expected new format bytes, got '%s' (%#x)", info.LicenseeCode, []byte(info.LicenseeCode))
	}

	// Version
	if info.Version != 0 {
		t.Errorf("Version: expected 0, got %d", info.Version)
	}

	// Header checksum
	if info.HeaderChecksum != 0x32 {
		t.Errorf("HeaderChecksum: expected 0x32, got %#x", info.HeaderChecksum)
	}

	// Global checksum
	if info.GlobalChecksum != 0x4e46 {
		t.Errorf("GlobalChecksum: expected 0x4e46, got %#x", info.GlobalChecksum)
	}
}

func TestParseGB_FileTooSmall(t *testing.T) {
	// Create a file that's too small for a valid GB header
	tmpFile, err := os.CreateTemp("", "small*.gb")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write less than the required header size
	if _, err := tmpFile.Write(make([]byte, 0x100)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	stat, _ := tmpFile.Stat()
	_, err = ParseGB(tmpFile, stat.Size())
	if err == nil {
		t.Error("Expected error for file too small, got nil")
	}
}
