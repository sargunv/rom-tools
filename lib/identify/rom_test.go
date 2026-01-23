package identify

import (
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

func TestIdentifyZIPSlowMode(t *testing.T) {
	romPath := "testdata/AGB_Rogue.gba.zip"

	rom, err := IdentifyROM(romPath, Options{HashMode: HashModeSlow})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Type != ROMTypeZIP {
		t.Errorf("Expected type %s, got %s", ROMTypeZIP, rom.Type)
	}

	if rom.Info == nil {
		t.Fatal("Expected game identification in slow mode, got nil")
	}

	if rom.Info.GamePlatform() != core.PlatformGBA {
		t.Errorf("Expected platform %s, got %s", core.PlatformGBA, rom.Info.GamePlatform())
	}

	if rom.Info.GameTitle() != "ROGUE" {
		t.Errorf("Expected title 'ROGUE', got '%s'", rom.Info.GameTitle())
	}
}

func TestIdentifyFolder(t *testing.T) {
	romPath := "testdata/xromwell"

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Type != ROMTypeFolder {
		t.Errorf("Expected type %s, got %s", ROMTypeFolder, rom.Type)
	}

	if rom.Info == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Info.GamePlatform() != core.PlatformXbox {
		t.Errorf("Expected platform %s, got %s", core.PlatformXbox, rom.Info.GamePlatform())
	}
}

func TestIdentifyLooseFile_Hashing(t *testing.T) {
	romPath := "testdata/gbtictac.gb"

	rom, err := IdentifyROM(romPath, Options{HashMode: HashModeDefault})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	file, ok := rom.Files["gbtictac.gb"]
	if !ok {
		t.Fatal("Expected file 'gbtictac.gb' not found")
	}

	if file.Size != 32768 {
		t.Errorf("Expected size 32768, got %d", file.Size)
	}

	if len(file.Hashes) != 3 {
		t.Fatalf("Expected 3 hashes, got %d", len(file.Hashes))
	}

	// Verify SHA1 hash
	var sha1Hash *Hash
	for i := range file.Hashes {
		if file.Hashes[i].Algorithm == HashSHA1 {
			sha1Hash = &file.Hashes[i]
			break
		}
	}
	if sha1Hash == nil {
		t.Fatal("SHA1 hash not found")
	}
	if sha1Hash.Value != "48a59d5b31e374731ece4d9eb33679d38143495e" {
		t.Errorf("Expected SHA1 '48a59d5b31e374731ece4d9eb33679d38143495e', got '%s'", sha1Hash.Value)
	}
}
