package identify

import (
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

func TestIdentifyZIPSlowMode(t *testing.T) {
	romPath := "testdata/AGB_Rogue.gba.zip"

	result, err := Identify(romPath, Options{HashMode: HashModeSlow})
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

	if item.Game == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if item.Game.GamePlatform() != core.PlatformGBA {
		t.Errorf("Expected platform %s, got %s", core.PlatformGBA, item.Game.GamePlatform())
	}

	if item.Game.GameTitle() != "ROGUE" {
		t.Errorf("Expected title 'ROGUE', got '%s'", item.Game.GameTitle())
	}
}

func TestIdentifyFolder(t *testing.T) {
	romPath := "testdata/xromwell"

	result, err := Identify(romPath, Options{})
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
}

func TestIdentifyLooseFile_Hashing(t *testing.T) {
	romPath := "testdata/gbtictac.gb"

	result, err := Identify(romPath, Options{HashMode: HashModeDefault})
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
	sha1Value, ok := item.Hashes[HashSHA1]
	if !ok {
		t.Fatal("SHA1 hash not found")
	}
	if sha1Value != "48a59d5b31e374731ece4d9eb33679d38143495e" {
		t.Errorf("Expected SHA1 '48a59d5b31e374731ece4d9eb33679d38143495e', got '%s'", sha1Value)
	}

	// Verify MD5 hash
	md5Value, ok := item.Hashes[HashMD5]
	if !ok {
		t.Fatal("MD5 hash not found")
	}
	if md5Value != "ab37d2fbe51e62215975d6e8354dd071" {
		t.Errorf("Expected MD5 'ab37d2fbe51e62215975d6e8354dd071', got '%s'", md5Value)
	}

	// Verify CRC32 hash
	crc32Value, ok := item.Hashes[HashCRC32]
	if !ok {
		t.Fatal("CRC32 hash not found")
	}
	if crc32Value != "775ae755" {
		t.Errorf("Expected CRC32 '775ae755', got '%s'", crc32Value)
	}
}

func TestIdentifyZIPFastMode(t *testing.T) {
	romPath := "testdata/AGB_Rogue.gba.zip"

	result, err := Identify(romPath, Options{HashMode: HashModeDefault})
	if err != nil {
		t.Fatalf("Identify() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]

	// In fast mode, we should only have ZIP CRC32
	if len(item.Hashes) != 1 {
		t.Fatalf("Expected 1 hash (zip-crc32), got %d", len(item.Hashes))
	}

	_, ok := item.Hashes[HashZipCRC32]
	if !ok {
		t.Error("Expected zip-crc32 hash in fast mode")
	}

	// No game identification in fast mode
	if item.Game != nil {
		t.Error("Expected no game identification in fast mode")
	}
}
