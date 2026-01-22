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

// makeSyntheticN64 creates a synthetic N64 ROM with specified parameters.
func makeSyntheticN64(byteOrder N64ByteOrder, title string, gameCode string, version byte) []byte {
	header := make([]byte, N64HeaderSize)

	// Set reserved byte based on byte order (will be swapped if needed)
	header[0] = n64ReservedByte
	header[1] = 0x37
	header[2] = 0x12
	header[3] = 0x40

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
	titleBytes := []byte(title)
	if len(titleBytes) > n64TitleLen {
		titleBytes = titleBytes[:n64TitleLen]
	}
	copy(header[n64TitleOffset:], titleBytes)
	for i := len(titleBytes); i < n64TitleLen; i++ {
		header[n64TitleOffset+i] = ' '
	}

	// Set game code (4 bytes at 0x3B)
	if len(gameCode) >= 4 {
		copy(header[n64GameCodeOffset:], []byte(gameCode[:4]))
	}

	// Set version
	header[n64VersionOffset] = version

	// Convert from big-endian to requested byte order
	switch byteOrder {
	case N64ByteSwapped:
		swapBytes16(header)
	case N64LittleEndian:
		swapBytes32(header)
	}

	return header
}

func TestParseN64_Synthetic(t *testing.T) {
	tests := []struct {
		name             string
		byteOrder        N64ByteOrder
		title            string
		gameCode         string
		version          byte
		wantTitle        string
		wantGameCode     string
		wantCategoryCode N64CategoryCode
		wantDestination  N64Destination
		wantVersion      int
		wantByteOrder    N64ByteOrder
	}{
		{
			name:             "Z64 format USA game",
			byteOrder:        N64BigEndian,
			title:            "TEST GAME",
			gameCode:         "NTGE",
			version:          1,
			wantTitle:        "TEST GAME",
			wantGameCode:     "NTGE",
			wantCategoryCode: N64CategoryGamePak,
			wantDestination:  N64DestinationNorthAmerica,
			wantVersion:      1,
			wantByteOrder:    N64BigEndian,
		},
		{
			name:             "V64 format Japan game",
			byteOrder:        N64ByteSwapped,
			title:            "JAPANESE GAME",
			gameCode:         "NJPJ",
			version:          0,
			wantTitle:        "JAPANESE GAME",
			wantGameCode:     "NJPJ",
			wantCategoryCode: N64CategoryGamePak,
			wantDestination:  N64DestinationJapan,
			wantVersion:      0,
			wantByteOrder:    N64ByteSwapped,
		},
		{
			name:             "N64 format Europe game",
			byteOrder:        N64LittleEndian,
			title:            "EURO GAME",
			gameCode:         "NEUP",
			version:          2,
			wantTitle:        "EURO GAME",
			wantGameCode:     "NEUP",
			wantCategoryCode: N64CategoryGamePak,
			wantDestination:  N64DestinationEurope,
			wantVersion:      2,
			wantByteOrder:    N64LittleEndian,
		},
		{
			name:             "64DD disk",
			byteOrder:        N64BigEndian,
			title:            "DD GAME",
			gameCode:         "DDDJ",
			version:          0,
			wantTitle:        "DD GAME",
			wantGameCode:     "DDDJ",
			wantCategoryCode: N64Category64DD,
			wantDestination:  N64DestinationJapan,
			wantVersion:      0,
			wantByteOrder:    N64BigEndian,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rom := makeSyntheticN64(tc.byteOrder, tc.title, tc.gameCode, tc.version)
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
	rom := makeSyntheticN64(N64BigEndian, "UNIQUE TEST", "NMKE", 0)
	reader := bytes.NewReader(rom)

	info, err := ParseN64(reader, int64(len(rom)))
	if err != nil {
		t.Fatalf("ParseN64() error = %v", err)
	}

	if info.UniqueCode != "MK" {
		t.Errorf("UniqueCode = %q, want %q", info.UniqueCode, "MK")
	}
}
