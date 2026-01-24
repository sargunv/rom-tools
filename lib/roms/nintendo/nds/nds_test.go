package nds

import (
	"bytes"
	"os"
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

func TestParseNDS(t *testing.T) {
	romPath := "testdata/MixedCubes.nds"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseNDS(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseNDS() error = %v", err)
	}

	// Verify platform
	if info.GamePlatform() != core.PlatformNDS {
		t.Errorf("GamePlatform() = %s, want %s", info.GamePlatform(), core.PlatformNDS)
	}

	// Verify game code and components
	if info.GameCode != "AXXE" {
		t.Errorf("GameCode = %q, want %q", info.GameCode, "AXXE")
	}
	if info.GameType != NDSGameTypeNDS {
		t.Errorf("GameType = %c, want %c", info.GameType, NDSGameTypeNDS)
	}
	if info.UniqueCode != "XX" {
		t.Errorf("UniqueCode = %q, want %q", info.UniqueCode, "XX")
	}
	if info.Destination != NDSDestinationUSA {
		t.Errorf("Destination = %c, want %c", info.Destination, NDSDestinationUSA)
	}

	// Verify serial (GameSerial returns GameCode)
	if info.GameSerial() != "AXXE" {
		t.Errorf("GameSerial() = %q, want %q", info.GameSerial(), "AXXE")
	}

	// Verify unit code (NDS only)
	if info.UnitCode != NDSUnitCodeNDS {
		t.Errorf("UnitCode = %d, want %d", info.UnitCode, NDSUnitCodeNDS)
	}

	// Verify NDS region (Normal/worldwide)
	if info.NDSRegion != NDSRegionNormal {
		t.Errorf("NDSRegion = %d, want %d", info.NDSRegion, NDSRegionNormal)
	}

	// Verify ROM size calculation (DeviceCapacity=0 means 128KB)
	if info.DeviceCapacity != 0 {
		t.Errorf("DeviceCapacity = %d, want %d", info.DeviceCapacity, 0)
	}
	expectedROMSize := 128 * 1024 // 128KB
	if info.ROMSize != expectedROMSize {
		t.Errorf("ROMSize = %d, want %d", info.ROMSize, expectedROMSize)
	}

	// Verify version
	if info.Version != 0 {
		t.Errorf("Version = %d, want %d", info.Version, 0)
	}
}

func TestParseNDS_TooSmall(t *testing.T) {
	// Create a reader with less than header size
	data := make([]byte, 100)
	r := bytes.NewReader(data)

	_, err := ParseNDS(r, int64(len(data)))
	if err == nil {
		t.Error("ParseNDS() expected error for file too small, got nil")
	}
}
