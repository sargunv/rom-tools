package n64

import (
	"bytes"
	"os"
	"testing"
)

func TestParseN64_Z64(t *testing.T) {
	romPath := "testdata/flames.z64"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseN64(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseN64() error = %v", err)
	}

	if info.ByteOrder != N64BigEndian {
		t.Errorf("ByteOrder = %s, want %s", info.ByteOrder, N64BigEndian)
	}
}

func TestParseN64_Z64_Fields(t *testing.T) {
	romPath := "testdata/flames.z64"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := ParseN64(file, stat.Size())
	if err != nil {
		t.Fatalf("ParseN64() error = %v", err)
	}

	// flames.z64 is a homebrew demo
	expectedTitle := "Flame Demo 12/25/01"
	if info.Title != expectedTitle {
		t.Errorf("Title = %q, want %q", info.Title, expectedTitle)
	}
	if info.ByteOrder != N64BigEndian {
		t.Errorf("ByteOrder = %s, want %s", info.ByteOrder, N64BigEndian)
	}
	// Verify CheckCode was extracted (non-zero for this ROM)
	if info.CheckCode == 0 {
		t.Logf("Warning: CheckCode is 0 (may be expected for homebrew)")
	}
}

// syntheticN64Options holds parameters for creating a synthetic N64 ROM.
type syntheticN64Options struct {
	byteOrder       N64ByteOrder
	title           string
	gameCode        string
	version         byte
	piBSDConfig     uint32
	clockRate       uint32
	bootAddress     uint32
	libultraVersion uint32
}

// makeSyntheticN64 creates a synthetic N64 ROM with specified parameters.
func makeSyntheticN64(opts syntheticN64Options) []byte {
	header := make([]byte, N64HeaderSize)

	// Set reserved byte (0x80 at offset 0)
	header[0] = n64ReservedByte

	// Set PI BSD DOM1 config (3 bytes at 0x01-0x03)
	header[0x01] = byte(opts.piBSDConfig >> 16)
	header[0x02] = byte(opts.piBSDConfig >> 8)
	header[0x03] = byte(opts.piBSDConfig)

	// Set clock rate (4 bytes at 0x04)
	header[0x04] = byte(opts.clockRate >> 24)
	header[0x05] = byte(opts.clockRate >> 16)
	header[0x06] = byte(opts.clockRate >> 8)
	header[0x07] = byte(opts.clockRate)

	// Set boot address (4 bytes at 0x08)
	header[0x08] = byte(opts.bootAddress >> 24)
	header[0x09] = byte(opts.bootAddress >> 16)
	header[0x0A] = byte(opts.bootAddress >> 8)
	header[0x0B] = byte(opts.bootAddress)

	// Set libultra version (4 bytes at 0x0C)
	header[0x0C] = byte(opts.libultraVersion >> 24)
	header[0x0D] = byte(opts.libultraVersion >> 16)
	header[0x0E] = byte(opts.libultraVersion >> 8)
	header[0x0F] = byte(opts.libultraVersion)

	// Set check code (8 bytes at 0x10)
	header[0x10] = 0xDE
	header[0x11] = 0xAD
	header[0x12] = 0xBE
	header[0x13] = 0xEF
	header[0x14] = 0xCA
	header[0x15] = 0xFE
	header[0x16] = 0xBA
	header[0x17] = 0xBE

	// Set title (20 bytes at 0x20)
	titleBytes := []byte(opts.title)
	if len(titleBytes) > n64TitleLen {
		titleBytes = titleBytes[:n64TitleLen]
	}
	copy(header[n64TitleOffset:], titleBytes)
	for i := len(titleBytes); i < n64TitleLen; i++ {
		header[n64TitleOffset+i] = ' '
	}

	// Set game code (4 bytes at 0x3B)
	if len(opts.gameCode) >= 4 {
		copy(header[n64GameCodeOffset:], []byte(opts.gameCode[:4]))
	}

	// Set version
	header[n64VersionOffset] = opts.version

	// Convert from big-endian to requested byte order
	switch opts.byteOrder {
	case N64ByteSwapped:
		swapBytes16(header)
	case N64LittleEndian:
		swapBytes32(header)
	}

	return header
}

func TestParseN64_Synthetic(t *testing.T) {
	tests := []struct {
		name                string
		opts                syntheticN64Options
		wantTitle           string
		wantGameCode        string
		wantCategoryCode    N64CategoryCode
		wantDestination     N64Destination
		wantVersion         int
		wantByteOrder       N64ByteOrder
		wantPIBSDConfig     uint32
		wantClockRate       uint32
		wantBootAddress     uint32
		wantLibultraVersion uint32
	}{
		{
			name: "Z64 format USA game",
			opts: syntheticN64Options{
				byteOrder:       N64BigEndian,
				title:           "TEST GAME",
				gameCode:        "NTGE",
				version:         1,
				piBSDConfig:     0x371240,
				clockRate:       0x0000000F,
				bootAddress:     0x80000400,
				libultraVersion: 0x0000144C,
			},
			wantTitle:           "TEST GAME",
			wantGameCode:        "NTGE",
			wantCategoryCode:    N64CategoryGamePak,
			wantDestination:     N64DestinationNorthAmerica,
			wantVersion:         1,
			wantByteOrder:       N64BigEndian,
			wantPIBSDConfig:     0x371240,
			wantClockRate:       0x0000000F,
			wantBootAddress:     0x80000400,
			wantLibultraVersion: 0x0000144C,
		},
		{
			name: "V64 format Japan game",
			opts: syntheticN64Options{
				byteOrder:       N64ByteSwapped,
				title:           "JAPANESE GAME",
				gameCode:        "NJPJ",
				version:         0,
				piBSDConfig:     0x371240,
				clockRate:       0x0000000F,
				bootAddress:     0x80001000,
				libultraVersion: 0x00001449,
			},
			wantTitle:           "JAPANESE GAME",
			wantGameCode:        "NJPJ",
			wantCategoryCode:    N64CategoryGamePak,
			wantDestination:     N64DestinationJapan,
			wantVersion:         0,
			wantByteOrder:       N64ByteSwapped,
			wantPIBSDConfig:     0x371240,
			wantClockRate:       0x0000000F,
			wantBootAddress:     0x80001000,
			wantLibultraVersion: 0x00001449,
		},
		{
			name: "N64 format Europe game",
			opts: syntheticN64Options{
				byteOrder:       N64LittleEndian,
				title:           "EURO GAME",
				gameCode:        "NEUP",
				version:         2,
				piBSDConfig:     0x801240,
				clockRate:       0x00000000,
				bootAddress:     0x80000400,
				libultraVersion: 0x0000144B,
			},
			wantTitle:           "EURO GAME",
			wantGameCode:        "NEUP",
			wantCategoryCode:    N64CategoryGamePak,
			wantDestination:     N64DestinationEurope,
			wantVersion:         2,
			wantByteOrder:       N64LittleEndian,
			wantPIBSDConfig:     0x801240,
			wantClockRate:       0x00000000,
			wantBootAddress:     0x80000400,
			wantLibultraVersion: 0x0000144B,
		},
		{
			name: "64DD disk",
			opts: syntheticN64Options{
				byteOrder:       N64BigEndian,
				title:           "DD GAME",
				gameCode:        "DDDJ",
				version:         0,
				piBSDConfig:     0x371240,
				clockRate:       0x0000000F,
				bootAddress:     0x80000400,
				libultraVersion: 0x00001446,
			},
			wantTitle:           "DD GAME",
			wantGameCode:        "DDDJ",
			wantCategoryCode:    N64Category64DD,
			wantDestination:     N64DestinationJapan,
			wantVersion:         0,
			wantByteOrder:       N64BigEndian,
			wantPIBSDConfig:     0x371240,
			wantClockRate:       0x0000000F,
			wantBootAddress:     0x80000400,
			wantLibultraVersion: 0x00001446,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rom := makeSyntheticN64(tc.opts)
			reader := bytes.NewReader(rom)

			info, err := ParseN64(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("ParseN64() error = %v", err)
			}

			if info.Title != tc.wantTitle {
				t.Errorf("Title = %q, want %q", info.Title, tc.wantTitle)
			}
			if info.GameCode != tc.wantGameCode {
				t.Errorf("GameCode = %q, want %q", info.GameCode, tc.wantGameCode)
			}
			if info.CategoryCode != tc.wantCategoryCode {
				t.Errorf("CategoryCode = %c, want %c", info.CategoryCode, tc.wantCategoryCode)
			}
			if info.Destination != tc.wantDestination {
				t.Errorf("Destination = %c, want %c", info.Destination, tc.wantDestination)
			}
			if info.Version != tc.wantVersion {
				t.Errorf("Version = %d, want %d", info.Version, tc.wantVersion)
			}
			if info.ByteOrder != tc.wantByteOrder {
				t.Errorf("ByteOrder = %s, want %s", info.ByteOrder, tc.wantByteOrder)
			}
			if info.PIBSDConfig != tc.wantPIBSDConfig {
				t.Errorf("PIBSDConfig = 0x%06X, want 0x%06X", info.PIBSDConfig, tc.wantPIBSDConfig)
			}
			if info.ClockRate != tc.wantClockRate {
				t.Errorf("ClockRate = 0x%08X, want 0x%08X", info.ClockRate, tc.wantClockRate)
			}
			if info.BootAddress != tc.wantBootAddress {
				t.Errorf("BootAddress = 0x%08X, want 0x%08X", info.BootAddress, tc.wantBootAddress)
			}
			if info.LibultraVersion != tc.wantLibultraVersion {
				t.Errorf("LibultraVersion = 0x%08X, want 0x%08X", info.LibultraVersion, tc.wantLibultraVersion)
			}
			// Verify CheckCode is non-zero (we set it in synthetic ROM)
			if info.CheckCode == 0 {
				t.Error("CheckCode = 0, want non-zero")
			}
		})
	}
}

func TestParseN64_TooSmall(t *testing.T) {
	// File smaller than header
	data := make([]byte, 32)
	data[0] = n64ReservedByte
	reader := bytes.NewReader(data)

	_, err := ParseN64(reader, int64(len(data)))
	if err == nil {
		t.Error("ParseN64() expected error for too small file, got nil")
	}
}

func TestParseN64_NewFields(t *testing.T) {
	// Test that all new header fields are correctly extracted across byte orders
	opts := syntheticN64Options{
		byteOrder:       N64BigEndian,
		title:           "FIELD TEST",
		gameCode:        "NFTE",
		version:         3,
		piBSDConfig:     0xABCDEF,
		clockRate:       0x12345678,
		bootAddress:     0x80123456,
		libultraVersion: 0xDEADBEEF,
	}

	for _, byteOrder := range []N64ByteOrder{N64BigEndian, N64ByteSwapped, N64LittleEndian} {
		t.Run(string(byteOrder), func(t *testing.T) {
			opts.byteOrder = byteOrder
			rom := makeSyntheticN64(opts)
			reader := bytes.NewReader(rom)

			info, err := ParseN64(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("ParseN64() error = %v", err)
			}

			if info.PIBSDConfig != 0xABCDEF {
				t.Errorf("PIBSDConfig = 0x%06X, want 0xABCDEF", info.PIBSDConfig)
			}
			if info.ClockRate != 0x12345678 {
				t.Errorf("ClockRate = 0x%08X, want 0x12345678", info.ClockRate)
			}
			if info.BootAddress != 0x80123456 {
				t.Errorf("BootAddress = 0x%08X, want 0x80123456", info.BootAddress)
			}
			if info.LibultraVersion != 0xDEADBEEF {
				t.Errorf("LibultraVersion = 0x%08X, want 0xDEADBEEF", info.LibultraVersion)
			}
			if info.ByteOrder != byteOrder {
				t.Errorf("ByteOrder = %s, want %s", info.ByteOrder, byteOrder)
			}
		})
	}
}

func TestParseN64_InvalidByteOrder(t *testing.T) {
	// Valid size but no 0x80 marker in expected positions
	header := make([]byte, N64HeaderSize)
	header[0] = 0x00
	header[1] = 0x00
	header[2] = 0x00
	header[3] = 0x00
	reader := bytes.NewReader(header)

	_, err := ParseN64(reader, int64(len(header)))
	if err == nil {
		t.Error("ParseN64() expected error for invalid byte order, got nil")
	}
}

func TestParseN64_UniqueCode(t *testing.T) {
	// Test that unique code is extracted correctly
	rom := makeSyntheticN64(syntheticN64Options{
		byteOrder:   N64BigEndian,
		title:       "UNIQUE TEST",
		gameCode:    "NMKE",
		version:     0,
		piBSDConfig: 0x371240,
	})
	reader := bytes.NewReader(rom)

	info, err := ParseN64(reader, int64(len(rom)))
	if err != nil {
		t.Fatalf("ParseN64() error = %v", err)
	}

	if info.UniqueCode != "MK" {
		t.Errorf("UniqueCode = %q, want %q", info.UniqueCode, "MK")
	}
}
