package sms

import (
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

// readerAt wraps a byte slice to implement io.ReaderAt
type readerAt []byte

func (r readerAt) ReadAt(p []byte, off int64) (n int, err error) {
	copy(p, r[off:])
	return len(p), nil
}

// makeTestROM creates a minimal SMS ROM with the given header fields
func makeTestROM(region SMSRegion, romSize SMSROMSize, checksum uint16, productLow, productHigh, productVer byte) readerAt {
	rom := make([]byte, smsMinROMSize)

	// Write magic at header offset
	copy(rom[smsHeaderOffset:], smsMagic)

	// Checksum (little-endian) at offset 0x0A
	rom[smsHeaderOffset+smsChecksumOffset] = byte(checksum)
	rom[smsHeaderOffset+smsChecksumOffset+1] = byte(checksum >> 8)

	// Product code low/high at 0x0C-0x0D
	rom[smsHeaderOffset+smsProductLowOffset] = productLow
	rom[smsHeaderOffset+smsProductLowOffset+1] = productHigh

	// Product code high nibble + version at 0x0E
	rom[smsHeaderOffset+smsProductVerOffset] = productVer

	// Region (high nibble) + ROM size (low nibble) at 0x0F
	rom[smsHeaderOffset+smsRegionSizeOffset] = byte(region)<<4 | byte(romSize)

	return rom
}

func TestParseSMS_MasterSystem(t *testing.T) {
	rom := makeTestROM(SMSRegionExportSMS, SMSROMSize256KB, 0x1234, 0x76, 0x70, 0x00)

	info, err := ParseSMS(rom, int64(len(rom)))
	if err != nil {
		t.Fatalf("ParseSMS() error = %v", err)
	}

	if info.GamePlatform() != core.PlatformMS {
		t.Errorf("Platform = %v, want %v", info.GamePlatform(), core.PlatformMS)
	}
	if info.Region != SMSRegionExportSMS {
		t.Errorf("Region = %v, want %v", info.Region, SMSRegionExportSMS)
	}
	if info.ROMSize != SMSROMSize256KB {
		t.Errorf("ROMSize = %v, want %v", info.ROMSize, SMSROMSize256KB)
	}
	if info.Checksum != 0x1234 {
		t.Errorf("Checksum = 0x%04X, want 0x1234", info.Checksum)
	}
	if info.ProductCode != "7670" {
		t.Errorf("ProductCode = %q, want %q", info.ProductCode, "7670")
	}
}

func TestParseSMS_GameGear(t *testing.T) {
	rom := makeTestROM(SMSRegionIntlGG, SMSROMSize512KB, 0xABCD, 0x12, 0x34, 0x52)

	info, err := ParseSMS(rom, int64(len(rom)))
	if err != nil {
		t.Fatalf("ParseSMS() error = %v", err)
	}

	if info.GamePlatform() != core.PlatformGameGear {
		t.Errorf("Platform = %v, want %v", info.GamePlatform(), core.PlatformGameGear)
	}
	if info.Region != SMSRegionIntlGG {
		t.Errorf("Region = %v, want %v", info.Region, SMSRegionIntlGG)
	}
	if info.Version != 2 {
		t.Errorf("Version = %d, want 2", info.Version)
	}
	if info.ProductCode != "51234" {
		t.Errorf("ProductCode = %q, want %q", info.ProductCode, "51234")
	}
}

func TestParseSMS_InvalidMagic(t *testing.T) {
	rom := make([]byte, smsMinROMSize)
	copy(rom[smsHeaderOffset:], []byte("NOTVALID"))

	_, err := ParseSMS(readerAt(rom), int64(len(rom)))
	if err == nil {
		t.Error("ParseSMS() expected error for invalid magic, got nil")
	}
}

func TestParseSMS_TooSmall(t *testing.T) {
	rom := make([]byte, smsMinROMSize-1)

	_, err := ParseSMS(readerAt(rom), int64(len(rom)))
	if err == nil {
		t.Error("ParseSMS() expected error for small file, got nil")
	}
}

func TestParseSMS_ChecksumEndianness(t *testing.T) {
	// Verify little-endian: byte at lower address is LSB
	rom := makeTestROM(SMSRegionExportSMS, SMSROMSize256KB, 0, 0, 0, 0)
	rom[smsHeaderOffset+smsChecksumOffset] = 0x34   // LSB
	rom[smsHeaderOffset+smsChecksumOffset+1] = 0x12 // MSB

	info, err := ParseSMS(rom, int64(len(rom)))
	if err != nil {
		t.Fatalf("ParseSMS() error = %v", err)
	}
	if info.Checksum != 0x1234 {
		t.Errorf("Checksum = 0x%04X, want 0x1234 (little-endian)", info.Checksum)
	}
}

func TestDeterminePlatform(t *testing.T) {
	tests := []struct {
		region   SMSRegion
		expected core.Platform
	}{
		{SMSRegionJapanSMS, core.PlatformMS},
		{SMSRegionExportSMS, core.PlatformMS},
		{SMSRegionJapanGG, core.PlatformGameGear},
		{SMSRegionExportGG, core.PlatformGameGear},
		{SMSRegionIntlGG, core.PlatformGameGear},
		{SMSRegion(0x0), core.PlatformMS}, // unknown defaults to SMS
	}

	for _, tt := range tests {
		got := determinePlatform(tt.region)
		if got != tt.expected {
			t.Errorf("determinePlatform(%v) = %v, want %v", tt.region, got, tt.expected)
		}
	}
}

func TestDecodeBCDProductCode(t *testing.T) {
	tests := []struct {
		low, high, extra byte
		expected         string
	}{
		{0x76, 0x70, 0x00, "7670"},
		{0x12, 0x34, 0x05, "51234"},
		{0x00, 0x00, 0x00, "0000"},
	}

	for _, tt := range tests {
		got := decodeBCDProductCode(tt.low, tt.high, tt.extra)
		if got != tt.expected {
			t.Errorf("decodeBCDProductCode(0x%02X, 0x%02X, 0x%02X) = %q, want %q",
				tt.low, tt.high, tt.extra, got, tt.expected)
		}
	}
}
