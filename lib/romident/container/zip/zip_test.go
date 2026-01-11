package zip

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestZIPArchive(t *testing.T) {
	zipPath := filepath.Join(testutil.ROMsPath(t), "gbtictac.gb.zip")

	handler := NewZIPHandler()
	archive, err := handler.Open(zipPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer archive.Close()

	entries := archive.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Name != "gbtictac.gb" {
		t.Errorf("Expected entry name 'gbtictac.gb', got '%s'", entry.Name)
	}

	if entry.Size != 32768 {
		t.Errorf("Expected size 32768, got %d", entry.Size)
	}

	// ZIP should have pre-computed CRC32
	if entry.CRC32 != 0x775ae755 {
		t.Errorf("Expected CRC32 0x775ae755, got 0x%08x", entry.CRC32)
	}
}

func TestZIPArchiveOpenFileAt(t *testing.T) {
	zipPath := filepath.Join(testutil.ROMsPath(t), "xromwell.xiso.iso.zip")

	handler := NewZIPHandler()
	archive, err := handler.Open(zipPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer archive.Close()

	reader, err := archive.OpenFileAt("xbox.xiso.iso")
	if err != nil {
		t.Fatalf("OpenFileAt() error = %v", err)
	}
	defer reader.Close()

	// Verify we can read the XISO magic bytes from inside the ZIP
	// XISO files start with "MICROSOFT*XBOX*MEDIA" at offset 0x10000
	xisoMagic := make([]byte, 25)
	n, err := reader.ReadAt(xisoMagic, 0x10000)
	if err != nil && err != io.EOF {
		t.Fatalf("ReadAt() error = %v", err)
	}
	if n < 20 {
		t.Fatalf("Expected to read at least 20 bytes at offset 0x10000, got %d", n)
	}
	// "MICROSOFT*XBOX*MEDIA" is 20 characters
	expectedMagic := "MICROSOFT*XBOX*MEDIA"
	if string(xisoMagic[:20]) != expectedMagic {
		t.Errorf("Expected XISO magic '%s', got '%s'", expectedMagic, string(xisoMagic[:20]))
	}

	// Verify Seek works
	pos, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("Seek() error = %v", err)
	}
	if pos != 0 {
		t.Errorf("Expected position 0, got %d", pos)
	}
}
