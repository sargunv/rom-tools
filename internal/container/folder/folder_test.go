package folder

import (
	"io"
	"testing"
)

func TestFolderContainer(t *testing.T) {
	folderPath := "testdata/xromwell"

	container, err := NewFolderContainer(folderPath)
	if err != nil {
		t.Fatalf("NewFolderContainer() error = %v", err)
	}
	defer container.Close()

	entries := container.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Name != "default.xbe" {
		t.Errorf("Expected entry name 'default.xbe', got '%s'", entry.Name)
	}

	if entry.Size != 290768 {
		t.Errorf("Expected size 290768, got %d", entry.Size)
	}

	// CRC32 should be 0 for folders (not pre-computed)
	if entry.CRC32 != 0 {
		t.Errorf("Expected CRC32 0 for folder entry, got %d", entry.CRC32)
	}
}

func TestFolderContainerOpenFileAt(t *testing.T) {
	folderPath := "testdata/xromwell"

	container, err := NewFolderContainer(folderPath)
	if err != nil {
		t.Fatalf("NewFolderContainer() error = %v", err)
	}
	defer container.Close()

	reader, err := container.OpenFileAt("default.xbe")
	if err != nil {
		t.Fatalf("OpenFileAt() error = %v", err)
	}
	defer reader.Close()

	// Verify Size()
	if reader.Size() != 290768 {
		t.Errorf("Expected size 290768, got %d", reader.Size())
	}

	// Verify we can read the XBE magic bytes
	magic := make([]byte, 4)
	n, err := reader.ReadAt(magic, 0)
	if err != nil {
		t.Fatalf("ReadAt() error = %v", err)
	}
	if n != 4 {
		t.Fatalf("Expected to read 4 bytes, got %d", n)
	}
	if string(magic) != "XBEH" {
		t.Errorf("Expected magic 'XBEH', got '%s'", string(magic))
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
