package xbe

import (
	"encoding/binary"
	"os"
	"testing"
)

// readerAt wraps a byte slice to implement io.ReaderAt
type readerAt []byte

func (r readerAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(r)) {
		return 0, nil
	}
	n = copy(p, r[off:])
	return n, nil
}

// makeTestXBE creates a minimal XBE file with a certificate.
// The XBE has a base address of 0x10000 and certificate at offset xbeHeaderSize.
func makeTestXBE(title string, titleID uint32, regionFlags XboxRegion, discNumber, version uint32) readerAt {
	// Create buffer large enough for header + certificate
	bufSize := xbeHeaderSize + xbeCertSize
	buf := make([]byte, bufSize)

	// Write magic
	copy(buf[0:], "XBEH")

	// Base address at 0x104
	baseAddr := uint32(0x10000)
	binary.LittleEndian.PutUint32(buf[xbeBaseAddrOffset:], baseAddr)

	// Certificate address at 0x118 - points to right after header in virtual memory
	certAddr := baseAddr + uint32(xbeHeaderSize)
	binary.LittleEndian.PutUint32(buf[xbeCertAddrOffset:], certAddr)

	// Now write certificate at offset xbeHeaderSize
	certOffset := xbeHeaderSize

	// Timestamp at 0x04
	binary.LittleEndian.PutUint32(buf[certOffset+xbeCertTimestampOff:], 0x12345678)

	// Title ID at 0x08
	binary.LittleEndian.PutUint32(buf[certOffset+xbeCertTitleIDOff:], titleID)

	// Title name at 0x0C (UTF-16LE, null-terminated)
	titleU16 := make([]uint16, len(title)+1)
	for i, r := range title {
		titleU16[i] = uint16(r)
	}
	for i, u := range titleU16 {
		binary.LittleEndian.PutUint16(buf[certOffset+xbeCertTitleNameOff+i*2:], u)
	}

	// Media types at 0x9C
	binary.LittleEndian.PutUint32(buf[certOffset+xbeCertMediaTypesOff:], uint32(XboxMediaDVDX2|XboxMediaDVD5RO))

	// Region flags at 0xA0
	binary.LittleEndian.PutUint32(buf[certOffset+xbeCertRegionOff:], uint32(regionFlags))

	// Game ratings at 0xA4
	binary.LittleEndian.PutUint32(buf[certOffset+xbeCertRatingsOff:], 0)

	// Disc number at 0xA8
	binary.LittleEndian.PutUint32(buf[certOffset+xbeCertDiscNumOff:], discNumber)

	// Version at 0xAC
	binary.LittleEndian.PutUint32(buf[certOffset+xbeCertVersionOff:], version)

	return buf
}

func TestParseXBE(t *testing.T) {
	romPath := "testdata/default.xbe"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseXBE(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseXBE() error = %v", err)
	}

	if info.Title != "Xromwell" {
		t.Errorf("Expected title 'Xromwell', got '%s'", info.Title)
	}
}

func TestParseXBE_Synthetic(t *testing.T) {
	rom := makeTestXBE("Test Game", 0x4D530001, XboxRegionNA, 1, 2)

	info, err := ParseXBE(rom, int64(len(rom)))
	if err != nil {
		t.Fatalf("ParseXBE() error = %v", err)
	}

	if info.Title != "Test Game" {
		t.Errorf("Title = %q, want %q", info.Title, "Test Game")
	}
	if info.TitleID != 0x4D530001 {
		t.Errorf("TitleID = 0x%08X, want 0x4D530001", info.TitleID)
	}
	if info.TitleIDHex != "4D530001" {
		t.Errorf("TitleIDHex = %q, want %q", info.TitleIDHex, "4D530001")
	}
	if info.PublisherCode != "MS" {
		t.Errorf("PublisherCode = %q, want %q", info.PublisherCode, "MS")
	}
	if info.GameNumber != 1 {
		t.Errorf("GameNumber = %d, want 1", info.GameNumber)
	}
	if info.RegionFlags != XboxRegionNA {
		t.Errorf("RegionFlags = %v, want %v", info.RegionFlags, XboxRegionNA)
	}
	if info.DiscNumber != 1 {
		t.Errorf("DiscNumber = %d, want 1", info.DiscNumber)
	}
	if info.Version != 2 {
		t.Errorf("Version = %d, want 2", info.Version)
	}
	if info.Timestamp != 0x12345678 {
		t.Errorf("Timestamp = 0x%08X, want 0x12345678", info.Timestamp)
	}
	expectedMedia := XboxMediaDVDX2 | XboxMediaDVD5RO
	if info.AllowedMediaTypes != expectedMedia {
		t.Errorf("AllowedMediaTypes = 0x%08X, want 0x%08X", info.AllowedMediaTypes, expectedMedia)
	}
}

func TestParseXBE_InvalidMagic(t *testing.T) {
	rom := make([]byte, xbeHeaderSize+xbeCertSize)
	copy(rom[0:], "NOTX")

	_, err := ParseXBE(readerAt(rom), int64(len(rom)))
	if err == nil {
		t.Error("ParseXBE() expected error for invalid magic, got nil")
	}
}

func TestParseXBE_TooSmall(t *testing.T) {
	rom := make([]byte, xbeHeaderSize-1)

	_, err := ParseXBE(readerAt(rom), int64(len(rom)))
	if err == nil {
		t.Error("ParseXBE() expected error for small file, got nil")
	}
}

func TestDecodeTitleID(t *testing.T) {
	tests := []struct {
		titleID       uint32
		publisherCode string
		gameNumber    uint16
	}{
		{0x4D530001, "MS", 1},   // Microsoft, game 1
		{0x45410010, "EA", 16},  // EA, game 16
		{0x53450100, "SE", 256}, // Square Enix, game 256
		{0x00000000, "\x00\x00", 0},
		{0xFFFFFFFF, "\xFF\xFF", 65535},
	}

	for _, tt := range tests {
		publisherCode, gameNumber := decodeTitleID(tt.titleID)
		if publisherCode != tt.publisherCode {
			t.Errorf("decodeTitleID(0x%08X) publisherCode = %q, want %q",
				tt.titleID, publisherCode, tt.publisherCode)
		}
		if gameNumber != tt.gameNumber {
			t.Errorf("decodeTitleID(0x%08X) gameNumber = %d, want %d",
				tt.titleID, gameNumber, tt.gameNumber)
		}
	}
}

func TestDecodeUTF16LE(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "simple ASCII",
			input:    []byte{'H', 0, 'i', 0, 0, 0},
			expected: "Hi",
		},
		{
			name:     "empty string",
			input:    []byte{0, 0},
			expected: "",
		},
		{
			name:     "unicode characters",
			input:    []byte{0xE9, 0x00, 0, 0}, // é (U+00E9)
			expected: "é",
		},
		{
			name:     "no null terminator",
			input:    []byte{'A', 0, 'B', 0},
			expected: "AB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeUTF16LE(tt.input)
			if got != tt.expected {
				t.Errorf("decodeUTF16LE(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
