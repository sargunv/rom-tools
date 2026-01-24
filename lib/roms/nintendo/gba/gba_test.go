package gba

import (
	"bytes"
	"os"
	"testing"
)

func TestParseGBA(t *testing.T) {
	romPath := "testdata/AGB_Rogue.gba"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseGBA(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseGBA() error = %v", err)
	}

	if info.Title != "ROGUE" {
		t.Errorf("Expected title 'ROGUE', got '%s'", info.Title)
	}

	if info.GameCode != "AAAA" {
		t.Errorf("Expected game code 'AAAA', got '%s'", info.GameCode)
	}

	if info.GameType != GBAGameTypeNormalOld {
		t.Errorf("Expected game type 'A' (0x%02X), got 0x%02X", GBAGameTypeNormalOld, info.GameType)
	}

	if info.Destination != GBADestination('A') {
		t.Errorf("Expected destination 'A' (0x%02X), got 0x%02X", 'A', info.Destination)
	}

	if info.MakerCode != "AA" {
		t.Errorf("Expected maker code 'AA', got '%s'", info.MakerCode)
	}

	if info.MainUnitCode != 0x00 {
		t.Errorf("Expected main unit code 0x00, got 0x%02X", info.MainUnitCode)
	}

	if info.DeviceType != 0x80 {
		t.Errorf("Expected device type 0x80, got 0x%02X", info.DeviceType)
	}

	if info.Version != 1 {
		t.Errorf("Expected version 1, got %d", info.Version)
	}

	if info.HeaderChecksum != 0xC8 {
		t.Errorf("Expected header checksum 0xC8, got 0x%02X", info.HeaderChecksum)
	}
}

// makeSyntheticGBA creates a minimal valid GBA ROM header for testing.
func makeSyntheticGBA(title, gameCode, makerCode string, mainUnit, deviceType, version, checksum byte) []byte {
	header := make([]byte, gbaHeaderSize)

	// Copy title (up to 12 bytes)
	copy(header[gbaTitleOffset:gbaTitleOffset+gbaTitleLen], title)

	// Copy game code (4 bytes)
	copy(header[gbaGameCodeOffset:gbaGameCodeOffset+gbaGameCodeLen], gameCode)

	// Copy maker code (2 bytes)
	copy(header[gbaMakerOffset:gbaMakerOffset+gbaMakerLen], makerCode)

	// Set fixed value
	header[gbaFixedOffset] = gbaFixedValue

	// Set hardware fields
	header[gbaMainUnitOffset] = mainUnit
	header[gbaDeviceTypeOffset] = deviceType

	// Set version and checksum
	header[gbaVersionOffset] = version
	header[gbaChecksumOffset] = checksum

	return header
}

func TestParseGBASyntheticGameTypes(t *testing.T) {
	tests := []struct {
		name         string
		gameCode     string
		wantGameType GBAGameType
		wantDest     GBADestination
	}{
		{
			name:         "Normal old (A)",
			gameCode:     "AMGE",
			wantGameType: GBAGameTypeNormalOld,
			wantDest:     GBADestinationUSA,
		},
		{
			name:         "Normal new (B)",
			gameCode:     "BPEJ",
			wantGameType: GBAGameTypeNormalNew,
			wantDest:     GBADestinationJapan,
		},
		{
			name:         "Famicom Mini (F)",
			gameCode:     "FMRP",
			wantGameType: GBAGameTypeFamicom,
			wantDest:     GBADestinationEurope,
		},
		{
			name:         "Acceleration sensor (K)",
			gameCode:     "KYGE",
			wantGameType: GBAGameTypeAcceleration,
			wantDest:     GBADestinationUSA,
		},
		{
			name:         "e-Reader (P)",
			gameCode:     "PSAE",
			wantGameType: GBAGameTypeEReader,
			wantDest:     GBADestinationUSA,
		},
		{
			name:         "Rumble + gyro (R)",
			gameCode:     "RZWJ",
			wantGameType: GBAGameTypeRumbleGyro,
			wantDest:     GBADestinationJapan,
		},
		{
			name:         "RTC + solar (U)",
			gameCode:     "U3IJ",
			wantGameType: GBAGameTypeRTCSolar,
			wantDest:     GBADestinationJapan,
		},
		{
			name:         "Rumble only (V)",
			gameCode:     "V49J",
			wantGameType: GBAGameTypeRumble,
			wantDest:     GBADestinationJapan,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rom := makeSyntheticGBA("TEST", tt.gameCode, "01", 0x00, 0x00, 0, 0)
			reader := bytes.NewReader(rom)

			info, err := ParseGBA(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("ParseGBA() error = %v", err)
			}

			if info.GameType != tt.wantGameType {
				t.Errorf("GameType = 0x%02X, want 0x%02X", info.GameType, tt.wantGameType)
			}
			if info.Destination != tt.wantDest {
				t.Errorf("Destination = 0x%02X, want 0x%02X", info.Destination, tt.wantDest)
			}
		})
	}
}

func TestParseGBASyntheticDestinations(t *testing.T) {
	tests := []struct {
		name     string
		destCode byte
		wantDest GBADestination
	}{
		{"Japan", 'J', GBADestinationJapan},
		{"USA", 'E', GBADestinationUSA},
		{"Europe", 'P', GBADestinationEurope},
		{"France", 'F', GBADestinationFrance},
		{"Spain", 'S', GBADestinationSpain},
		{"Germany", 'D', GBADestinationGermany},
		{"Italy", 'I', GBADestinationItaly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameCode := "AXX" + string(tt.destCode)
			rom := makeSyntheticGBA("TEST", gameCode, "01", 0x00, 0x00, 0, 0)
			reader := bytes.NewReader(rom)

			info, err := ParseGBA(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("ParseGBA() error = %v", err)
			}

			if info.Destination != tt.wantDest {
				t.Errorf("Destination = 0x%02X, want 0x%02X", info.Destination, tt.wantDest)
			}
		})
	}
}

func TestParseGBASyntheticHardwareFields(t *testing.T) {
	tests := []struct {
		name           string
		mainUnit       byte
		deviceType     byte
		version        byte
		checksum       byte
		wantMainUnit   byte
		wantDeviceType byte
		wantVersion    int
		wantChecksum   byte
	}{
		{
			name:           "Standard GBA",
			mainUnit:       0x00,
			deviceType:     0x00,
			version:        0,
			checksum:       0x00,
			wantMainUnit:   0x00,
			wantDeviceType: 0x00,
			wantVersion:    0,
			wantChecksum:   0x00,
		},
		{
			name:           "Debug hardware (DACS)",
			mainUnit:       0x00,
			deviceType:     0x80,
			version:        1,
			checksum:       0xAB,
			wantMainUnit:   0x00,
			wantDeviceType: 0x80,
			wantVersion:    1,
			wantChecksum:   0xAB,
		},
		{
			name:           "High version number",
			mainUnit:       0x00,
			deviceType:     0x00,
			version:        255,
			checksum:       0xFF,
			wantMainUnit:   0x00,
			wantDeviceType: 0x00,
			wantVersion:    255,
			wantChecksum:   0xFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rom := makeSyntheticGBA("TEST", "AXXX", "01", tt.mainUnit, tt.deviceType, tt.version, tt.checksum)
			reader := bytes.NewReader(rom)

			info, err := ParseGBA(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("ParseGBA() error = %v", err)
			}

			if info.MainUnitCode != tt.wantMainUnit {
				t.Errorf("MainUnitCode = 0x%02X, want 0x%02X", info.MainUnitCode, tt.wantMainUnit)
			}
			if info.DeviceType != tt.wantDeviceType {
				t.Errorf("DeviceType = 0x%02X, want 0x%02X", info.DeviceType, tt.wantDeviceType)
			}
			if info.Version != tt.wantVersion {
				t.Errorf("Version = %d, want %d", info.Version, tt.wantVersion)
			}
			if info.HeaderChecksum != tt.wantChecksum {
				t.Errorf("HeaderChecksum = 0x%02X, want 0x%02X", info.HeaderChecksum, tt.wantChecksum)
			}
		})
	}
}

func TestParseGBAErrors(t *testing.T) {
	t.Run("file too small", func(t *testing.T) {
		rom := make([]byte, gbaHeaderSize-1)
		reader := bytes.NewReader(rom)

		_, err := ParseGBA(reader, int64(len(rom)))
		if err == nil {
			t.Error("Expected error for file too small, got nil")
		}
	})

	t.Run("invalid fixed byte", func(t *testing.T) {
		rom := make([]byte, gbaHeaderSize)
		rom[gbaFixedOffset] = 0x00 // Invalid, should be 0x96
		reader := bytes.NewReader(rom)

		_, err := ParseGBA(reader, int64(len(rom)))
		if err == nil {
			t.Error("Expected error for invalid fixed byte, got nil")
		}
	})
}
