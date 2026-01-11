package chd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestParseCHDHeader(t *testing.T) {
	chdPath := filepath.Join(testutil.ROMsPath(t), "empty.chd")

	file, err := os.Open(chdPath)
	if err != nil {
		t.Fatalf("Failed to open CHD file: %v", err)
	}
	defer file.Close()

	info, err := ParseCHDHeader(file)
	if err != nil {
		t.Fatalf("ParseCHDHeader() error = %v", err)
	}

	if info.Version < 5 {
		t.Errorf("Expected version >= 5, got %d", info.Version)
	}

	if info.RawSHA1 != "f6348f85d8487e7aff1fa54e5987b172bce2a3a6" {
		t.Errorf("Expected raw SHA1 'f6348f85d8487e7aff1fa54e5987b172bce2a3a6', got '%s'", info.RawSHA1)
	}

	if info.SHA1 != "cdd8baa51e7b84bb11037fb3415d698d011fe40a" {
		t.Errorf("Expected compressed SHA1 'cdd8baa51e7b84bb11037fb3415d698d011fe40a', got '%s'", info.SHA1)
	}

	// Parent SHA1 should be empty for standalone CHD
	if info.ParentSHA1 != "" {
		t.Errorf("Expected empty parent SHA1, got '%s'", info.ParentSHA1)
	}
}
