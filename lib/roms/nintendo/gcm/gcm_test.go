package gcm

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

// makeSyntheticGCM creates a synthetic GameCube/Wii disc header for testing.
func makeSyntheticGCM(system SystemCode, gameCode string, region Region, title string, isWii bool) []byte {
	header := make([]byte, DiscHeaderSize)

	// System code at offset 0x00
	header[systemCodeOffset] = byte(system)

	// Game code at offset 0x01 (2 bytes)
	if len(gameCode) >= 2 {
		copy(header[gameCodeOffset:], gameCode[:2])
	}

	// Region at offset 0x03
	header[regionOffset] = byte(region)

	// Maker code at offset 0x04 (2 bytes)
	copy(header[makerCodeOffset:], "01") // Nintendo

	// Disc number and version
	header[discNumberOffset] = 0
	header[discVersionOffset] = 0

	// Magic words
	if isWii {
		binary.BigEndian.PutUint32(header[wiiMagicOffset:], wiiMagicWord)
		binary.BigEndian.PutUint32(header[gcMagicOffset:], 0)
	} else {
		binary.BigEndian.PutUint32(header[wiiMagicOffset:], 0)
		binary.BigEndian.PutUint32(header[gcMagicOffset:], gcMagicWord)
	}

	// Title at offset 0x20 (64 bytes max)
	if len(title) > titleLen {
		title = title[:titleLen]
	}
	copy(header[titleOffset:], title)

	return header
}

func TestParseGCM_GameCube(t *testing.T) {
	header := makeSyntheticGCM(SystemCodeGameCube, "MK", RegionNorthAmerica, "Test GameCube Game", false)
	reader := bytes.NewReader(header)

	info, err := ParseGCM(reader, int64(len(header)))
	if err != nil {
		t.Fatalf("ParseGCM() error = %v", err)
	}

	if info.platform != core.PlatformGC {
		t.Errorf("platform = %v, want %v", info.platform, core.PlatformGC)
	}
	if info.GamePlatform() != core.PlatformGC {
		t.Errorf("GamePlatform() = %v, want %v", info.GamePlatform(), core.PlatformGC)
	}
	if info.SystemCode != SystemCodeGameCube {
		t.Errorf("SystemCode = %c, want %c", info.SystemCode, SystemCodeGameCube)
	}
	if info.GameCode != "MK" {
		t.Errorf("GameCode = %q, want %q", info.GameCode, "MK")
	}
	if info.Region != RegionNorthAmerica {
		t.Errorf("Region = %c, want %c", info.Region, RegionNorthAmerica)
	}
	if info.Title != "Test GameCube Game" {
		t.Errorf("Title = %q, want %q", info.Title, "Test GameCube Game")
	}
}

func TestParseGCM_Wii(t *testing.T) {
	header := makeSyntheticGCM(SystemCodeWii, "SM", RegionJapan, "Test Wii Game", true)
	reader := bytes.NewReader(header)

	info, err := ParseGCM(reader, int64(len(header)))
	if err != nil {
		t.Fatalf("ParseGCM() error = %v", err)
	}

	if info.platform != core.PlatformWii {
		t.Errorf("platform = %v, want %v", info.platform, core.PlatformWii)
	}
	if info.GamePlatform() != core.PlatformWii {
		t.Errorf("GamePlatform() = %v, want %v", info.GamePlatform(), core.PlatformWii)
	}
	if info.SystemCode != SystemCodeWii {
		t.Errorf("SystemCode = %c, want %c", info.SystemCode, SystemCodeWii)
	}
	if info.Region != RegionJapan {
		t.Errorf("Region = %c, want %c", info.Region, RegionJapan)
	}
}

func TestParseGCM_InvalidMagic(t *testing.T) {
	header := make([]byte, DiscHeaderSize)
	// No magic words set - both are zero
	reader := bytes.NewReader(header)

	_, err := ParseGCM(reader, int64(len(header)))
	if err == nil {
		t.Error("ParseGCM() expected error for invalid magic, got nil")
	}
}

func TestParseGCM_TooSmall(t *testing.T) {
	header := make([]byte, 32) // Less than DiscHeaderSize (0x60 = 96)
	reader := bytes.NewReader(header)

	_, err := ParseGCM(reader, int64(len(header)))
	if err == nil {
		t.Error("ParseGCM() expected error for too small file, got nil")
	}
}

func TestGCMInfo_GameSerial(t *testing.T) {
	tests := []struct {
		name       string
		systemCode SystemCode
		gameCode   string
		region     Region
		wantSerial string
	}{
		{
			name:       "GameCube USA",
			systemCode: SystemCodeGameCube,
			gameCode:   "MK",
			region:     RegionNorthAmerica,
			wantSerial: "GMKE",
		},
		{
			name:       "Wii Japan",
			systemCode: SystemCodeWii,
			gameCode:   "SM",
			region:     RegionJapan,
			wantSerial: "RSMJ",
		},
		{
			name:       "Wii Europe",
			systemCode: SystemCodeWiiNew,
			gameCode:   "AB",
			region:     RegionEurope,
			wantSerial: "SABP",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			header := makeSyntheticGCM(tc.systemCode, tc.gameCode, tc.region, "Test", tc.systemCode == SystemCodeWii || tc.systemCode == SystemCodeWiiNew)
			reader := bytes.NewReader(header)

			info, err := ParseGCM(reader, int64(len(header)))
			if err != nil {
				t.Fatalf("ParseGCM() error = %v", err)
			}

			if info.GameSerial() != tc.wantSerial {
				t.Errorf("GameSerial() = %q, want %q", info.GameSerial(), tc.wantSerial)
			}
		})
	}
}
