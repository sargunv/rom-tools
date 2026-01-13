package romident

import (
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
	"github.com/sargunv/rom-tools/lib/romident/game"
)

// Test GBA identification
func TestIdentifyGBA(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "AGB_Rogue.gba")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Type != ROMTypeFile {
		t.Errorf("Expected type %s, got %s", ROMTypeFile, rom.Type)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformGBA {
		t.Errorf("Expected platform %s, got %s", game.PlatformGBA, rom.Ident.Platform)
	}

	if rom.Ident.TitleID != "AAAA" {
		t.Errorf("Expected title ID 'AAAA', got '%s'", rom.Ident.TitleID)
	}

	if rom.Ident.Title != "ROGUE" {
		t.Errorf("Expected title 'ROGUE', got '%s'", rom.Ident.Title)
	}

	if rom.Ident.MakerCode != "AA" {
		t.Errorf("Expected maker code 'AA', got '%s'", rom.Ident.MakerCode)
	}
}

// Test NDS identification
func TestIdentifyNDS(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "MixedCubes.nds")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformNDS {
		t.Errorf("Expected platform %s, got %s", game.PlatformNDS, rom.Ident.Platform)
	}

	// Title may be empty for some homebrew ROMs
	// Just verify we got a valid identification
}

// Test NES identification
func TestIdentifyNES(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "BombSweeper.nes")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformNES {
		t.Errorf("Expected platform %s, got %s", game.PlatformNES, rom.Ident.Platform)
	}
}

// Test SNES identification
func TestIdentifySNES(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "col15.sfc")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformSNES {
		t.Errorf("Expected platform %s, got %s", game.PlatformSNES, rom.Ident.Platform)
	}
}

// Test GB identification
func TestIdentifyGB(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "gbtictac.gb")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformGB {
		t.Errorf("Expected platform %s, got %s", game.PlatformGB, rom.Ident.Platform)
	}

	if rom.Ident.Title != "TIC-TAC-TOE" {
		t.Errorf("Expected title 'TIC-TAC-TOE', got '%s'", rom.Ident.Title)
	}
}

// Test GBC identification
func TestIdentifyGBC(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "JUMPMAN86.GBC")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformGBC {
		t.Errorf("Expected platform %s, got %s", game.PlatformGBC, rom.Ident.Platform)
	}
}

// Test N64 Z64 identification
func TestIdentifyN64_Z64(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "flames.z64")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformN64 {
		t.Errorf("Expected platform %s, got %s", game.PlatformN64, rom.Ident.Platform)
	}
}

// Test N64 V64 identification
func TestIdentifyN64_V64(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "flames.v64")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformN64 {
		t.Errorf("Expected platform %s, got %s", game.PlatformN64, rom.Ident.Platform)
	}
}

// Test N64 N64 identification (word-swapped)
func TestIdentifyN64_N64(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "flames.n64")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformN64 {
		t.Errorf("Expected platform %s, got %s", game.PlatformN64, rom.Ident.Platform)
	}
}

// Test Mega Drive identification
func TestIdentifyMD(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "Censor_Intro.md")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformMD {
		t.Errorf("Expected platform %s, got %s", game.PlatformMD, rom.Ident.Platform)
	}
}

// Test SMD identification
func TestIdentifySMD(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "Censor_Intro.smd")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformMD {
		t.Errorf("Expected platform %s, got %s", game.PlatformMD, rom.Ident.Platform)
	}
}

// Test Xbox XBE identification
func TestIdentifyXBE(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "xromwell", "default.xbe")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformXbox {
		t.Errorf("Expected platform %s, got %s", game.PlatformXbox, rom.Ident.Platform)
	}

	if rom.Ident.Title != "Xromwell" {
		t.Errorf("Expected title 'Xromwell', got '%s'", rom.Ident.Title)
	}
}

// Test Xbox XISO identification
func TestIdentifyXISO(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "xromwell.xiso.iso")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformXbox {
		t.Errorf("Expected platform %s, got %s", game.PlatformXbox, rom.Ident.Platform)
	}

	if rom.Ident.Title != "Xromwell" {
		t.Errorf("Expected title 'Xromwell', got '%s'", rom.Ident.Title)
	}
}

// Test ZIP identification preserves game ident in slow mode
func TestIdentifyZIPSlowMode(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "AGB_Rogue.gba.zip")

	rom, err := IdentifyROM(romPath, Options{HashMode: HashModeSlow})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Type != ROMTypeZIP {
		t.Errorf("Expected type %s, got %s", ROMTypeZIP, rom.Type)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification in slow mode, got nil")
	}

	if rom.Ident.Platform != game.PlatformGBA {
		t.Errorf("Expected platform %s, got %s", game.PlatformGBA, rom.Ident.Platform)
	}

	if rom.Ident.Title != "ROGUE" {
		t.Errorf("Expected title 'ROGUE', got '%s'", rom.Ident.Title)
	}
}

// Test folder identification
func TestIdentifyFolder(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "xromwell")

	rom, err := IdentifyROM(romPath, Options{})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Type != ROMTypeFolder {
		t.Errorf("Expected type %s, got %s", ROMTypeFolder, rom.Type)
	}

	if rom.Ident == nil {
		t.Fatal("Expected game identification, got nil")
	}

	if rom.Ident.Platform != game.PlatformXbox {
		t.Errorf("Expected platform %s, got %s", game.PlatformXbox, rom.Ident.Platform)
	}
}

// Test hashing (moved from rom_test.go)
func TestIdentifyLooseFile_Hashing(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "gbtictac.gb")

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

// Test ZIP metadata hashing
func TestIdentifyZIP_Hashing(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "gbtictac.gb.zip")

	rom, err := IdentifyROM(romPath, Options{HashMode: HashModeDefault})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	file, ok := rom.Files["gbtictac.gb"]
	if !ok {
		t.Fatal("Expected file 'gbtictac.gb' not found")
	}

	// In default mode, ZIP should only have CRC32 from metadata
	if len(file.Hashes) != 1 {
		t.Fatalf("Expected 1 hash (CRC32 from metadata), got %d", len(file.Hashes))
	}

	crc32Hash := file.Hashes[0]
	if crc32Hash.Algorithm != HashCRC32 {
		t.Errorf("Expected CRC32 hash, got %s", crc32Hash.Algorithm)
	}
	if crc32Hash.Value != "775ae755" {
		t.Errorf("Expected CRC32 '775ae755', got '%s'", crc32Hash.Value)
	}
	if crc32Hash.Source != "zip-metadata" {
		t.Errorf("Expected source 'zip-metadata', got '%s'", crc32Hash.Source)
	}
}
