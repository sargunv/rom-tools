package n64

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestParseN64_Z64(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "flames.z64")

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open z64 file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat z64 file: %v", err)
	}

	info, err := ParseN64(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseN64() error = %v", err)
	}

	if info.ByteOrder != N64BigEndian {
		t.Errorf("Expected byte order %q, got %q", N64BigEndian, info.ByteOrder)
	}

	if info.Title != "Flame Demo 12/25/01" {
		t.Errorf("Expected title 'Flame Demo 12/25/01', got %q", info.Title)
	}
}

func TestParseN64_V64(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "flames.v64")

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open v64 file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat v64 file: %v", err)
	}

	info, err := ParseN64(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseN64() error = %v", err)
	}

	if info.ByteOrder != N64ByteSwapped {
		t.Errorf("Expected byte order %q, got %q", N64ByteSwapped, info.ByteOrder)
	}

	if info.Title != "Flame Demo 12/25/01" {
		t.Errorf("Expected title 'Flame Demo 12/25/01', got %q", info.Title)
	}
}

func TestParseN64_N64(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "flames.n64")

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open n64 file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat n64 file: %v", err)
	}

	info, err := ParseN64(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseN64() error = %v", err)
	}

	if info.ByteOrder != N64LittleEndian {
		t.Errorf("Expected byte order %q, got %q", N64LittleEndian, info.ByteOrder)
	}

	if info.Title != "Flame Demo 12/25/01" {
		t.Errorf("Expected title 'Flame Demo 12/25/01', got %q", info.Title)
	}
}

func TestParseN64_AllFormatsMatchTitle(t *testing.T) {
	romsPath := testutil.ROMsPath(t)
	formats := []struct {
		filename  string
		byteOrder N64ByteOrder
	}{
		{"flames.z64", N64BigEndian},
		{"flames.v64", N64ByteSwapped},
		{"flames.n64", N64LittleEndian},
	}

	var expectedTitle string

	for _, f := range formats {
		romPath := filepath.Join(romsPath, f.filename)

		file, err := os.Open(romPath)
		if err != nil {
			t.Fatalf("Failed to open %s: %v", f.filename, err)
		}

		stat, err := file.Stat()
		if err != nil {
			file.Close()
			t.Fatalf("Failed to stat %s: %v", f.filename, err)
		}

		info, err := ParseN64(file, stat.Size())
		file.Close()
		if err != nil {
			t.Fatalf("ParseN64(%s) error = %v", f.filename, err)
		}

		if info.ByteOrder != f.byteOrder {
			t.Errorf("%s: expected byte order %q, got %q", f.filename, f.byteOrder, info.ByteOrder)
		}

		if expectedTitle == "" {
			expectedTitle = info.Title
		} else if info.Title != expectedTitle {
			t.Errorf("%s: title mismatch - expected %q, got %q", f.filename, expectedTitle, info.Title)
		}
	}
}
