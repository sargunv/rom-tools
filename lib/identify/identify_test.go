package identify

import (
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

func TestIdentifyZIP(t *testing.T) {
	romPath := "testdata/AGB_Rogue.gba.zip"

	result, err := Identify(romPath, DefaultOptions())
	if err != nil {
		t.Fatalf("Identify() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	// Check item details
	item := result.Items[0]
	if item.Name != "AGB_Rogue.gba" {
		t.Errorf("Expected item name 'AGB_Rogue.gba', got '%s'", item.Name)
	}

	// Game should be identified
	if item.Game == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if item.Game.GamePlatform() != core.PlatformGBA {
		t.Errorf("Expected platform %s, got %s", core.PlatformGBA, item.Game.GamePlatform())
	}

	if item.Game.GameTitle() != "ROGUE" {
		t.Errorf("Expected title 'ROGUE', got '%s'", item.Game.GameTitle())
	}

	// Should use ZIP metadata hash (never calculate hashes for containers with metadata)
	if len(item.Hashes) != 1 {
		t.Fatalf("Expected 1 hash (zip-crc32 from metadata), got %d", len(item.Hashes))
	}

	_, ok := item.Hashes[core.HashZipCRC32]
	if !ok {
		t.Error("Expected zip-crc32 hash from ZIP metadata")
	}
}

func TestIdentifyFolder(t *testing.T) {
	romPath := "testdata/xromwell"

	result, err := Identify(romPath, DefaultOptions())
	if err != nil {
		t.Fatalf("Identify() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	// Check item details
	item := result.Items[0]
	if item.Name != "default.xbe" {
		t.Errorf("Expected item name 'default.xbe', got '%s'", item.Name)
	}

	if item.Game == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if item.Game.GamePlatform() != core.PlatformXbox {
		t.Errorf("Expected platform %s, got %s", core.PlatformXbox, item.Game.GamePlatform())
	}

	// Folders don't have pre-computed metadata hashes like ZIP,
	// so hashes should be calculated (SHA1, MD5, CRC32)
	if len(item.Hashes) != 3 {
		t.Fatalf("Expected 3 calculated hashes, got %d", len(item.Hashes))
	}

	if _, ok := item.Hashes[core.HashSHA1]; !ok {
		t.Error("Expected SHA1 hash")
	}
	if _, ok := item.Hashes[core.HashMD5]; !ok {
		t.Error("Expected MD5 hash")
	}
	if _, ok := item.Hashes[core.HashCRC32]; !ok {
		t.Error("Expected CRC32 hash")
	}
}

func TestIdentifyLooseFile(t *testing.T) {
	romPath := "testdata/gbtictac.gb"

	result, err := Identify(romPath, DefaultOptions())
	if err != nil {
		t.Fatalf("Identify() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]

	if item.Size != 32768 {
		t.Errorf("Expected size 32768, got %d", item.Size)
	}

	if len(item.Hashes) != 3 {
		t.Fatalf("Expected 3 hashes, got %d", len(item.Hashes))
	}

	// Verify SHA1 hash
	sha1Value, ok := item.Hashes[core.HashSHA1]
	if !ok {
		t.Fatal("SHA1 hash not found")
	}
	if sha1Value != "48a59d5b31e374731ece4d9eb33679d38143495e" {
		t.Errorf("Expected SHA1 '48a59d5b31e374731ece4d9eb33679d38143495e', got '%s'", sha1Value)
	}

	// Verify MD5 hash
	md5Value, ok := item.Hashes[core.HashMD5]
	if !ok {
		t.Fatal("MD5 hash not found")
	}
	if md5Value != "ab37d2fbe51e62215975d6e8354dd071" {
		t.Errorf("Expected MD5 'ab37d2fbe51e62215975d6e8354dd071', got '%s'", md5Value)
	}

	// Verify CRC32 hash
	crc32Value, ok := item.Hashes[core.HashCRC32]
	if !ok {
		t.Fatal("CRC32 hash not found")
	}
	if crc32Value != "775ae755" {
		t.Errorf("Expected CRC32 '775ae755', got '%s'", crc32Value)
	}
}

func TestIdentifyLooseFileSkipsHashForLargeFiles(t *testing.T) {
	romPath := "testdata/gbtictac.gb"

	// Set max hash size to 0 bytes, which should skip hashing
	opts := Options{MaxHashSize: 0}

	result, err := Identify(romPath, opts)
	if err != nil {
		t.Fatalf("Identify() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]

	// Should have no hashes since file (32KB) exceeds MaxHashSize (0)
	if len(item.Hashes) != 0 {
		t.Errorf("Expected 0 hashes with MaxHashSize=0, got %d", len(item.Hashes))
	}

	// Game identification should still work
	if item.Game == nil {
		t.Fatal("Expected game identification even with MaxHashSize=0")
	}
}

func TestIdentifyLooseFileNoLimitHashes(t *testing.T) {
	romPath := "testdata/gbtictac.gb"

	// Set max hash size to -1 (no limit)
	opts := Options{MaxHashSize: -1}

	result, err := Identify(romPath, opts)
	if err != nil {
		t.Fatalf("Identify() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]

	// Should have all 3 hashes with no limit
	if len(item.Hashes) != 3 {
		t.Errorf("Expected 3 hashes with MaxHashSize=-1, got %d", len(item.Hashes))
	}
}
