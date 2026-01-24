package n3ds

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

// makeSyntheticNCSD creates a synthetic 3DS CCI/NCSD file for testing.
// NCSD header at 0x000, NCCH header at 0x200 (partition 0 at media unit 1).
func makeSyntheticNCSD(productCode, makerCode string, titleID, mediaID uint64, isNew3DS bool) []byte {
	// NCSD header (0x200 bytes) + NCCH header at partition 0 (0x200 bytes)
	// Partition 0 starts at media unit 1 (offset 0x200)
	data := make([]byte, 0x400)

	// NCSD Header
	// Magic at 0x100
	copy(data[ncsdMagicOffset:], ncsdMagic)

	// Image size (2 media units = 0x400 bytes)
	binary.LittleEndian.PutUint32(data[ncsdImageSizeOffset:], 2)

	// Media ID
	binary.LittleEndian.PutUint64(data[ncsdMediaIDOffset:], mediaID)

	// Partition table entry 0: offset=1 (media unit), size=1 (media unit)
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset:], 1)   // offset
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+4:], 1) // size

	// NCCH Header at offset 0x200
	ncchBase := 0x200

	// NCCH Magic at ncchBase + 0x100
	copy(data[ncchBase+ncchMagicOffset:], ncchMagic)

	// Title ID
	binary.LittleEndian.PutUint64(data[ncchBase+ncchTitleIDOffset:], titleID)

	// Maker code
	copy(data[ncchBase+ncchMakerCodeOffset:], makerCode)

	// Version
	binary.LittleEndian.PutUint16(data[ncchBase+ncchVersionOffset:], 0x0001)

	// Product code
	copy(data[ncchBase+ncchProductCodeOffset:], productCode)

	// Flags - set New 3DS exclusive flag if needed
	if isNew3DS {
		data[ncchBase+ncchFlagsOffset+4] = 0x02 // Bit 1 = New 3DS exclusive
	}

	return data
}

func TestParseN3DS_Standard3DS(t *testing.T) {
	productCode := "CTR-P-ALGE"
	makerCode := "00"
	titleID := uint64(0x0004000000123400)
	mediaID := uint64(0x0000000000000001)

	data := makeSyntheticNCSD(productCode, makerCode, titleID, mediaID, false)
	reader := bytes.NewReader(data)

	info, err := ParseN3DS(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("ParseN3DS() error = %v", err)
	}

	if info.GamePlatform() != core.Platform3DS {
		t.Errorf("GamePlatform() = %v, want %v", info.GamePlatform(), core.Platform3DS)
	}
	if info.ProductCode != productCode {
		t.Errorf("ProductCode = %q, want %q", info.ProductCode, productCode)
	}
	if info.GameSerial() != productCode {
		t.Errorf("GameSerial() = %q, want %q", info.GameSerial(), productCode)
	}
	if info.MakerCode != makerCode {
		t.Errorf("MakerCode = %q, want %q", info.MakerCode, makerCode)
	}
	if info.TitleID != titleID {
		t.Errorf("TitleID = %016X, want %016X", info.TitleID, titleID)
	}
	if info.MediaID != mediaID {
		t.Errorf("MediaID = %016X, want %016X", info.MediaID, mediaID)
	}
	if info.IsNew3DSExclusive {
		t.Error("IsNew3DSExclusive = true, want false")
	}
	if info.Region != N3DSRegionUSA {
		t.Errorf("Region = %c, want %c", info.Region, N3DSRegionUSA)
	}
	if info.PartitionCount != 1 {
		t.Errorf("PartitionCount = %d, want 1", info.PartitionCount)
	}
}

func TestParseN3DS_New3DSExclusive(t *testing.T) {
	productCode := "KTR-P-BNEJ" // New 3DS exclusive, Japan
	makerCode := "01"
	titleID := uint64(0x0004000000456700)
	mediaID := uint64(0x0000000000000002)

	data := makeSyntheticNCSD(productCode, makerCode, titleID, mediaID, true)
	reader := bytes.NewReader(data)

	info, err := ParseN3DS(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("ParseN3DS() error = %v", err)
	}

	if info.GamePlatform() != core.PlatformNew3DS {
		t.Errorf("GamePlatform() = %v, want %v", info.GamePlatform(), core.PlatformNew3DS)
	}
	if !info.IsNew3DSExclusive {
		t.Error("IsNew3DSExclusive = false, want true")
	}
	if info.Region != N3DSRegionJapan {
		t.Errorf("Region = %c, want %c", info.Region, N3DSRegionJapan)
	}
}

func TestParseN3DS_Regions(t *testing.T) {
	tests := []struct {
		name        string
		productCode string
		wantRegion  N3DSRegion
	}{
		{"Japan", "CTR-P-AAAJ", N3DSRegionJapan},
		{"USA", "CTR-P-AAAE", N3DSRegionUSA},
		{"Europe", "CTR-P-AAAP", N3DSRegionEurope},
		{"Australia", "CTR-P-AAAU", N3DSRegionAustralia},
		{"China", "CTR-P-AAAC", N3DSRegionChina},
		{"Korea", "CTR-P-AAAK", N3DSRegionKorea},
		{"Taiwan", "CTR-P-AAAT", N3DSRegionTaiwan},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := makeSyntheticNCSD(tc.productCode, "00", 0, 0, false)
			reader := bytes.NewReader(data)

			info, err := ParseN3DS(reader, int64(len(data)))
			if err != nil {
				t.Fatalf("ParseN3DS() error = %v", err)
			}

			if info.Region != tc.wantRegion {
				t.Errorf("Region = %c, want %c", info.Region, tc.wantRegion)
			}
		})
	}
}

func TestParseN3DS_ContentTypes(t *testing.T) {
	tests := []struct {
		name            string
		contentTypeFlag byte
		wantContentType N3DSContentType
	}{
		{"Application", 0x00, N3DSContentTypeApplication},
		{"SystemUpdate", 0x01, N3DSContentTypeSystemUpdate},
		{"Manual", 0x02, N3DSContentTypeManual},
		{"DLPChild", 0x03, N3DSContentTypeDLPChild},
		{"Trial", 0x04, N3DSContentTypeTrial},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := makeSyntheticNCSD("CTR-P-TEST", "00", 0, 0, false)
			// Set content type in flags[5]
			data[0x200+ncchFlagsOffset+5] = tc.contentTypeFlag
			reader := bytes.NewReader(data)

			info, err := ParseN3DS(reader, int64(len(data)))
			if err != nil {
				t.Fatalf("ParseN3DS() error = %v", err)
			}

			if info.ContentType != tc.wantContentType {
				t.Errorf("ContentType = %d, want %d", info.ContentType, tc.wantContentType)
			}
		})
	}
}

func TestParseN3DS_Errors(t *testing.T) {
	t.Run("file too small", func(t *testing.T) {
		data := make([]byte, ncsdHeaderSize-1)
		reader := bytes.NewReader(data)

		_, err := ParseN3DS(reader, int64(len(data)))
		if err == nil {
			t.Error("ParseN3DS() expected error for file too small, got nil")
		}
	})

	t.Run("invalid NCSD magic", func(t *testing.T) {
		data := make([]byte, ncsdHeaderSize)
		copy(data[ncsdMagicOffset:], "XXXX") // Invalid magic
		reader := bytes.NewReader(data)

		_, err := ParseN3DS(reader, int64(len(data)))
		if err == nil {
			t.Error("ParseN3DS() expected error for invalid NCSD magic, got nil")
		}
	})

	t.Run("empty partition 0", func(t *testing.T) {
		data := make([]byte, ncsdHeaderSize)
		copy(data[ncsdMagicOffset:], ncsdMagic)
		// Partition table entries are all zero
		reader := bytes.NewReader(data)

		_, err := ParseN3DS(reader, int64(len(data)))
		if err == nil {
			t.Error("ParseN3DS() expected error for empty partition 0, got nil")
		}
	})

	t.Run("partition beyond file", func(t *testing.T) {
		data := make([]byte, ncsdHeaderSize)
		copy(data[ncsdMagicOffset:], ncsdMagic)
		// Set partition 0 to point beyond file
		binary.LittleEndian.PutUint32(data[ncsdPartTableOffset:], 100) // offset = 100 media units
		binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+4:], 1) // size = 1 media unit
		reader := bytes.NewReader(data)

		_, err := ParseN3DS(reader, int64(len(data)))
		if err == nil {
			t.Error("ParseN3DS() expected error for partition beyond file, got nil")
		}
	})

	t.Run("invalid NCCH magic", func(t *testing.T) {
		data := make([]byte, 0x400)
		copy(data[ncsdMagicOffset:], ncsdMagic)
		// Set partition 0 to point to valid location but with invalid NCCH magic
		binary.LittleEndian.PutUint32(data[ncsdPartTableOffset:], 1)
		binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+4:], 1)
		copy(data[0x200+ncchMagicOffset:], "XXXX") // Invalid NCCH magic
		reader := bytes.NewReader(data)

		_, err := ParseN3DS(reader, int64(len(data)))
		if err == nil {
			t.Error("ParseN3DS() expected error for invalid NCCH magic, got nil")
		}
	})
}

func TestN3DSInfo_GameInfo(t *testing.T) {
	info := &N3DSInfo{
		ProductCode: "CTR-P-ALGE",
		platform:    core.Platform3DS,
	}

	// GameTitle should return empty string (3DS headers don't have title)
	if info.GameTitle() != "" {
		t.Errorf("GameTitle() = %q, want empty string", info.GameTitle())
	}

	// GameSerial should return product code
	if info.GameSerial() != "CTR-P-ALGE" {
		t.Errorf("GameSerial() = %q, want %q", info.GameSerial(), "CTR-P-ALGE")
	}

	// GamePlatform should return the set platform
	if info.GamePlatform() != core.Platform3DS {
		t.Errorf("GamePlatform() = %v, want %v", info.GamePlatform(), core.Platform3DS)
	}
}

func TestParseN3DS_MultiplePartitions(t *testing.T) {
	// Create synthetic data with multiple partitions
	data := make([]byte, 0x600) // 3 media units

	// NCSD Header
	copy(data[ncsdMagicOffset:], ncsdMagic)
	binary.LittleEndian.PutUint32(data[ncsdImageSizeOffset:], 3)

	// Partition table: 2 valid entries
	// Partition 0: offset=1, size=1
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset:], 1)
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+4:], 1)
	// Partition 1: offset=2, size=1
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+8:], 2)
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+12:], 1)

	// NCCH Header for partition 0 at 0x200
	copy(data[0x200+ncchMagicOffset:], ncchMagic)
	copy(data[0x200+ncchProductCodeOffset:], "CTR-P-TEST")

	reader := bytes.NewReader(data)
	info, err := ParseN3DS(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("ParseN3DS() error = %v", err)
	}

	if info.PartitionCount != 2 {
		t.Errorf("PartitionCount = %d, want 2", info.PartitionCount)
	}
}

func TestParseN3DS_OutOfBoundsPartitionsExcluded(t *testing.T) {
	// Create synthetic data where partition table has entries beyond file/image bounds
	data := make([]byte, 0x600) // 3 media units

	// NCSD Header
	copy(data[ncsdMagicOffset:], ncsdMagic)
	binary.LittleEndian.PutUint32(data[ncsdImageSizeOffset:], 3) // Image is 3 media units

	// Partition 0: valid, within bounds (offset=1, size=1)
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset:], 1)
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+4:], 1)

	// Partition 1: invalid, extends beyond image size (offset=2, size=5)
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+8:], 2)
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+12:], 5) // ends at media unit 7, beyond image size 3

	// Partition 2: invalid, starts beyond file (offset=100, size=1)
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+16:], 100)
	binary.LittleEndian.PutUint32(data[ncsdPartTableOffset+20:], 1)

	// NCCH Header for partition 0 at 0x200
	copy(data[0x200+ncchMagicOffset:], ncchMagic)
	copy(data[0x200+ncchProductCodeOffset:], "CTR-P-TEST")

	reader := bytes.NewReader(data)
	info, err := ParseN3DS(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("ParseN3DS() error = %v", err)
	}

	// Only partition 0 should be counted as valid
	if info.PartitionCount != 1 {
		t.Errorf("PartitionCount = %d, want 1 (out-of-bounds partitions should be excluded)", info.PartitionCount)
	}
}
