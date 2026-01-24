package xiso

import (
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

func TestParseXISO(t *testing.T) {
	romPath := "testdata/xromwell.xiso.iso"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseXISO(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseXISO() error = %v", err)
	}

	if info.Title != "Xromwell" {
		t.Errorf("Expected title 'Xromwell', got '%s'", info.Title)
	}
}

func TestParseXISO_TooSmall(t *testing.T) {
	// XISO needs at least xisoVolumeDescOffset + 32 bytes
	rom := make([]byte, xisoVolumeDescOffset)

	_, err := ParseXISO(readerAt(rom), int64(len(rom)))
	if err == nil {
		t.Error("ParseXISO() expected error for small file, got nil")
	}
}

func TestParseXISO_InvalidMagic(t *testing.T) {
	// Create a buffer large enough but with invalid magic
	rom := make([]byte, xisoVolumeDescOffset+64)
	copy(rom[xisoVolumeDescOffset:], "INVALID*MAGIC*HERE!!")

	_, err := ParseXISO(readerAt(rom), int64(len(rom)))
	if err == nil {
		t.Error("ParseXISO() expected error for invalid magic, got nil")
	}
}

func TestParseXISO_MissingDefaultXBE(t *testing.T) {
	// Create a minimal valid XISO structure but with empty directory
	rom := make([]byte, xisoVolumeDescOffset+2048+128)

	// Write valid XISO magic
	copy(rom[xisoVolumeDescOffset:], "MICROSOFT*XBOX*MEDIA")

	// Point to a valid but empty directory (sector 1, offset 2048)
	// Root directory sector at offset 0x14
	rom[xisoVolumeDescOffset+xisoRootDirOffset] = 1
	// Root directory size at offset 0x18
	rom[xisoVolumeDescOffset+xisoRootDirSizeOff] = 14 // Minimum entry size

	// Write a minimal directory entry that is NOT default.xbe
	dirOffset := 2048
	// Left child = 0
	rom[dirOffset] = 0
	rom[dirOffset+1] = 0
	// Right child = 0
	rom[dirOffset+2] = 0
	rom[dirOffset+3] = 0
	// File sector = 0
	rom[dirOffset+4] = 0
	rom[dirOffset+5] = 0
	rom[dirOffset+6] = 0
	rom[dirOffset+7] = 0
	// File size = 0
	rom[dirOffset+8] = 0
	rom[dirOffset+9] = 0
	rom[dirOffset+10] = 0
	rom[dirOffset+11] = 0
	// Attributes = 0
	rom[dirOffset+12] = 0
	// Filename length = 8
	rom[dirOffset+13] = 8
	// Filename = "test.txt"
	copy(rom[dirOffset+14:], "test.txt")

	_, err := ParseXISO(readerAt(rom), int64(len(rom)))
	if err == nil {
		t.Error("ParseXISO() expected error for missing default.xbe, got nil")
	}
}
