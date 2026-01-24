package rvz

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/gcm"
)

// GCM header constants for test data generation
const (
	testGCMHeaderSize    = 0x60
	testSystemCodeOffset = 0x000
	testGameCodeOffset   = 0x001
	testRegionOffset     = 0x003
	testMakerCodeOffset  = 0x004
	testWiiMagicOffset   = 0x018
	testGCMagicOffset    = 0x01C
	testTitleOffset      = 0x020
	testTitleLen         = 64
	testWiiMagicWord     = 0x5D1C9EA3
	testGCMagicWord      = 0xC2339F3D
)

// makeSyntheticGCMData creates a synthetic GameCube/Wii disc header for testing.
func makeSyntheticGCMData(system gcm.SystemCode, gameCode string, region gcm.Region, title string, isWii bool) []byte {
	header := make([]byte, testGCMHeaderSize)

	header[testSystemCodeOffset] = byte(system)
	if len(gameCode) >= 2 {
		copy(header[testGameCodeOffset:], gameCode[:2])
	}
	header[testRegionOffset] = byte(region)
	copy(header[testMakerCodeOffset:], "01")

	if isWii {
		binary.BigEndian.PutUint32(header[testWiiMagicOffset:], testWiiMagicWord)
	} else {
		binary.BigEndian.PutUint32(header[testGCMagicOffset:], testGCMagicWord)
	}

	if len(title) > testTitleLen {
		title = title[:testTitleLen]
	}
	copy(header[testTitleOffset:], title)

	return header
}

// makeSyntheticRVZ creates a synthetic RVZ/WIA file header for testing.
func makeSyntheticRVZ(magic string, gcmData []byte, discType DiscType, compression Compression) []byte {
	header := make([]byte, totalHeaderSize)

	// Magic at offset 0x00
	copy(header[magicOffset:], magic)

	// Version and compatible version
	binary.BigEndian.PutUint32(header[versionOffset:], 1)
	binary.BigEndian.PutUint32(header[compatVerOffset:], 1)

	// ISO and WIA file sizes
	binary.BigEndian.PutUint64(header[isoFileSizeOffset:], 1459978240) // ~1.4GB typical GC disc
	binary.BigEndian.PutUint64(header[wiaFileSizeOffset:], 500000000)  // compressed size

	// Disc type at discStructBase + 0x00
	binary.BigEndian.PutUint32(header[discStructBase+discTypeOffset:], uint32(discType))

	// Compression at discStructBase + 0x04
	binary.BigEndian.PutUint32(header[discStructBase+compressionOffset:], uint32(compression))

	// Compression level at discStructBase + 0x08
	binary.BigEndian.PutUint32(header[discStructBase+comprLevelOffset:], 5)

	// Chunk size at discStructBase + 0x0C
	binary.BigEndian.PutUint32(header[discStructBase+chunkSizeOffset:], 2097152) // 2MB chunks

	// Copy GCM data into dhead at discStructBase + 0x10
	if len(gcmData) > dheadSize {
		gcmData = gcmData[:dheadSize]
	}
	copy(header[discStructBase+dheadOffset:], gcmData)

	return header
}

func TestParseRVZ_WIA(t *testing.T) {
	gcmData := makeSyntheticGCMData(gcm.SystemCodeGameCube, "MK", gcm.RegionNorthAmerica, "Test Game", false)
	header := makeSyntheticRVZ("WIA\x01", gcmData, DiscTypeGameCube, CompressionZstandard)
	reader := bytes.NewReader(header)

	info, err := ParseRVZ(reader, int64(len(header)))
	if err != nil {
		t.Fatalf("ParseRVZ() error = %v", err)
	}

	if info.GCM == nil {
		t.Fatal("GCM is nil")
	}
	if info.GCM.GamePlatform() != core.PlatformGC {
		t.Errorf("GCM.GamePlatform() = %v, want %v", info.GCM.GamePlatform(), core.PlatformGC)
	}
	if info.DiscType != DiscTypeGameCube {
		t.Errorf("DiscType = %v, want %v", info.DiscType, DiscTypeGameCube)
	}
	if info.Compression != CompressionZstandard {
		t.Errorf("Compression = %v, want %v", info.Compression, CompressionZstandard)
	}
}

func TestParseRVZ_RVZ(t *testing.T) {
	gcmData := makeSyntheticGCMData(gcm.SystemCodeWii, "SM", gcm.RegionJapan, "Wii Game", true)
	header := makeSyntheticRVZ("RVZ\x01", gcmData, DiscTypeWii, CompressionLZMA2)
	reader := bytes.NewReader(header)

	info, err := ParseRVZ(reader, int64(len(header)))
	if err != nil {
		t.Fatalf("ParseRVZ() error = %v", err)
	}

	if info.GCM == nil {
		t.Fatal("GCM is nil")
	}
	if info.GCM.GamePlatform() != core.PlatformWii {
		t.Errorf("GCM.GamePlatform() = %v, want %v", info.GCM.GamePlatform(), core.PlatformWii)
	}
	if info.DiscType != DiscTypeWii {
		t.Errorf("DiscType = %v, want %v", info.DiscType, DiscTypeWii)
	}
	if info.Compression != CompressionLZMA2 {
		t.Errorf("Compression = %v, want %v", info.Compression, CompressionLZMA2)
	}
}

func TestParseRVZ_InvalidMagic(t *testing.T) {
	gcmData := makeSyntheticGCMData(gcm.SystemCodeGameCube, "MK", gcm.RegionNorthAmerica, "Test", false)
	header := makeSyntheticRVZ("BAD\x01", gcmData, DiscTypeGameCube, CompressionNone)
	reader := bytes.NewReader(header)

	_, err := ParseRVZ(reader, int64(len(header)))
	if err == nil {
		t.Error("ParseRVZ() expected error for invalid magic, got nil")
	}
}

func TestParseRVZ_TooSmall(t *testing.T) {
	header := make([]byte, 64) // Less than totalHeaderSize
	reader := bytes.NewReader(header)

	_, err := ParseRVZ(reader, int64(len(header)))
	if err == nil {
		t.Error("ParseRVZ() expected error for too small file, got nil")
	}
}

func TestRVZInfo_GameInfo(t *testing.T) {
	gcmData := makeSyntheticGCMData(gcm.SystemCodeGameCube, "MK", gcm.RegionNorthAmerica, "Test Title", false)
	header := makeSyntheticRVZ("RVZ\x01", gcmData, DiscTypeGameCube, CompressionZstandard)
	reader := bytes.NewReader(header)

	info, err := ParseRVZ(reader, int64(len(header)))
	if err != nil {
		t.Fatalf("ParseRVZ() error = %v", err)
	}

	// Verify GameInfo interface methods delegate to GCM
	if info.GamePlatform() != core.PlatformGC {
		t.Errorf("GamePlatform() = %v, want %v", info.GamePlatform(), core.PlatformGC)
	}
	if info.GameTitle() != "Test Title" {
		t.Errorf("GameTitle() = %q, want %q", info.GameTitle(), "Test Title")
	}
	if info.GameSerial() != "GMKE" {
		t.Errorf("GameSerial() = %q, want %q", info.GameSerial(), "GMKE")
	}
}
