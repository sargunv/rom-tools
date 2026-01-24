package psnpkg

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

// makeTestPKG creates a synthetic PKG file for testing.
func makeTestPKG(contentID string, pkgType PKGType, contentType ContentType, includeSFO bool) []byte {
	// Calculate sizes
	metadataOffset := uint32(pkgHeaderSize + pkgPS3DigestSize) // After header + digest
	if pkgType == PKGTypePSPVita {
		metadataOffset += 0x40 // Extended header
	}

	// Metadata entries
	var metadata []byte

	// Content type entry: ID (4) + Size (4) + Data (4)
	ctEntry := make([]byte, 12)
	binary.BigEndian.PutUint32(ctEntry[0:4], metadataContentType)
	binary.BigEndian.PutUint32(ctEntry[4:8], 4)
	binary.BigEndian.PutUint32(ctEntry[8:12], uint32(contentType))
	metadata = append(metadata, ctEntry...)

	metadataCount := uint32(1)

	// SFO info entry if requested
	var sfoOffset uint32
	var sfoData []byte
	if includeSFO {
		// Create a minimal valid SFO
		sfoData = makeMinimalSFO("Test Game", "NPUA80472")
		sfoOffset = metadataOffset + uint32(len(metadata)) + 8 + 0x38 // After metadata entries

		// SFO info entry: ID (4) + Size (4) + Data (0x38)
		sfoInfoEntry := make([]byte, 8+0x38)
		binary.BigEndian.PutUint32(sfoInfoEntry[0:4], metadataSFOInfo)
		binary.BigEndian.PutUint32(sfoInfoEntry[4:8], 0x38)
		binary.BigEndian.PutUint32(sfoInfoEntry[8:12], sfoOffset)
		binary.BigEndian.PutUint32(sfoInfoEntry[12:16], uint32(len(sfoData)))
		// Rest is padding/sha256
		metadata = append(metadata, sfoInfoEntry...)
		metadataCount++
	}

	// Build the full PKG
	totalSize := int(metadataOffset) + len(metadata)
	if includeSFO {
		totalSize = int(sfoOffset) + len(sfoData)
	}
	pkg := make([]byte, totalSize)

	// Header
	copy(pkg[0:4], pkgMagic)
	binary.BigEndian.PutUint16(pkg[pkgRevisionOffset:], 0x8000) // Finalized
	binary.BigEndian.PutUint16(pkg[pkgTypeOffset:], uint16(pkgType))
	binary.BigEndian.PutUint32(pkg[pkgMetadataOffOffset:], metadataOffset)
	binary.BigEndian.PutUint32(pkg[pkgMetadataCountOff:], metadataCount)
	binary.BigEndian.PutUint32(pkg[pkgItemCountOffset:], 10)
	binary.BigEndian.PutUint64(pkg[pkgTotalSizeOffset:], uint64(totalSize))

	// Content ID
	copy(pkg[pkgContentIDOffset:], contentID)

	// Extended header for PSP/Vita
	if pkgType == PKGTypePSPVita {
		extOffset := pkgHeaderSize + pkgPS3DigestSize
		copy(pkg[extOffset:extOffset+4], pkgExtMagic)
		// Key ID at offset 0x24 within extended header
		binary.BigEndian.PutUint32(pkg[extOffset+pkgExtKeyIDOffset:], 0x00000002) // Vita key
	}

	// Metadata
	copy(pkg[metadataOffset:], metadata)

	// SFO data
	if includeSFO && len(sfoData) > 0 {
		copy(pkg[sfoOffset:], sfoData)
	}

	return pkg
}

// makeMinimalSFO creates a minimal valid SFO for testing.
func makeMinimalSFO(title, discID string) []byte {
	// SFO format: header (20 bytes) + index entries + key table + data table
	// Minimal: 2 entries (DISC_ID and TITLE)

	keyTable := "DISC_ID\x00TITLE\x00"
	dataTable := discID + "\x00" + title + "\x00"

	// Pad data table to align
	for len(dataTable)%4 != 0 {
		dataTable += "\x00"
	}

	headerSize := 20
	indexSize := 2 * 16 // 2 entries, 16 bytes each
	keyTableOffset := headerSize + indexSize
	dataTableOffset := keyTableOffset + len(keyTable)

	totalSize := dataTableOffset + len(dataTable)
	sfo := make([]byte, totalSize)

	// Header
	copy(sfo[0:4], "\x00PSF")                                           // Magic
	binary.LittleEndian.PutUint32(sfo[4:8], 0x00000101)                  // Version
	binary.LittleEndian.PutUint32(sfo[8:12], uint32(keyTableOffset))     // Key table offset
	binary.LittleEndian.PutUint32(sfo[12:16], uint32(dataTableOffset))   // Data table offset
	binary.LittleEndian.PutUint32(sfo[16:20], 2)                         // Entry count

	// Index entry 1: DISC_ID
	idx1 := sfo[headerSize : headerSize+16]
	binary.LittleEndian.PutUint16(idx1[0:2], 0)                         // Key offset
	binary.LittleEndian.PutUint16(idx1[2:4], 0x0204)                    // Format (UTF-8)
	binary.LittleEndian.PutUint32(idx1[4:8], uint32(len(discID)+1))     // Data len
	binary.LittleEndian.PutUint32(idx1[8:12], uint32(len(discID)+1))    // Max len
	binary.LittleEndian.PutUint32(idx1[12:16], 0)                       // Data offset

	// Index entry 2: TITLE
	idx2 := sfo[headerSize+16 : headerSize+32]
	binary.LittleEndian.PutUint16(idx2[0:2], 8)                         // Key offset ("DISC_ID\x00" = 8 bytes)
	binary.LittleEndian.PutUint16(idx2[2:4], 0x0204)                    // Format (UTF-8)
	binary.LittleEndian.PutUint32(idx2[4:8], uint32(len(title)+1))      // Data len
	binary.LittleEndian.PutUint32(idx2[8:12], uint32(len(title)+1))     // Max len
	binary.LittleEndian.PutUint32(idx2[12:16], uint32(len(discID)+1))   // Data offset

	// Key table
	copy(sfo[keyTableOffset:], keyTable)

	// Data table
	copy(sfo[dataTableOffset:], dataTable)

	return sfo
}

func TestParsePKG(t *testing.T) {
	contentID := "UP0001-NPUA80472_00-LITTLEBIGPLAN001"
	pkg := makeTestPKG(contentID, PKGTypePS3, ContentTypeGameData, false)

	info, err := ParsePKG(bytes.NewReader(pkg), int64(len(pkg)))
	if err != nil {
		t.Fatalf("ParsePKG failed: %v", err)
	}

	if info.ContentID != contentID {
		t.Errorf("ContentID = %q, want %q", info.ContentID, contentID)
	}
	if info.Type != PKGTypePS3 {
		t.Errorf("Type = %v, want %v", info.Type, PKGTypePS3)
	}
	if info.ContentType != ContentTypeGameData {
		t.Errorf("ContentType = %v, want %v", info.ContentType, ContentTypeGameData)
	}
	if info.Platform != core.PlatformPS3 {
		t.Errorf("Platform = %v, want %v", info.Platform, core.PlatformPS3)
	}
}

func TestParsePKGWithSFO(t *testing.T) {
	contentID := "UP0001-NPUA80472_00-LITTLEBIGPLAN001"
	pkg := makeTestPKG(contentID, PKGTypePS3, ContentTypeGameData, true)

	info, err := ParsePKG(bytes.NewReader(pkg), int64(len(pkg)))
	if err != nil {
		t.Fatalf("ParsePKG failed: %v", err)
	}

	if info.SFO == nil {
		t.Fatal("SFO is nil, expected parsed SFO data")
	}
	if info.SFO.Title != "Test Game" {
		t.Errorf("SFO.Title = %q, want %q", info.SFO.Title, "Test Game")
	}
	if info.Title != "Test Game" {
		t.Errorf("Title = %q, want %q", info.Title, "Test Game")
	}
	if info.GameTitle() != "Test Game" {
		t.Errorf("GameTitle() = %q, want %q", info.GameTitle(), "Test Game")
	}
}

func TestContentTypePlatform(t *testing.T) {
	tests := []struct {
		name        string
		pkgType     PKGType
		contentType ContentType
		extKeyID    uint32
		want        core.Platform
	}{
		// PS3 content types
		{"PS3 GameData", PKGTypePS3, ContentTypeGameData, 0, core.PlatformPS3},
		{"PS3 GameExec", PKGTypePS3, ContentTypeGameExec, 0, core.PlatformPS3},
		{"PS3 Theme", PKGTypePS3, ContentTypeTheme, 0, core.PlatformPS3},
		{"PS3 Widget", PKGTypePS3, ContentTypeWidget, 0, core.PlatformPS3},
		{"PS3 License", PKGTypePS3, ContentTypeLicense, 0, core.PlatformPS3},

		// PSP content types
		{"PSP game", PKGTypePSPVita, ContentTypePSP, 1, core.PlatformPSP},
		{"PSPgo", PKGTypePSPVita, ContentTypePSPGo, 1, core.PlatformPSP},
		{"Minis", PKGTypePSPVita, ContentTypeMinis, 1, core.PlatformPSP},
		{"NeoGeo", PKGTypePSPVita, ContentTypeNeoGeo, 1, core.PlatformPSP},

		// Vita content types
		{"Vita game", PKGTypePSPVita, ContentTypePSVitaGame, 2, core.PlatformPSVita},
		{"Vita DLC", PKGTypePSPVita, ContentTypePSVitaDLC, 2, core.PlatformPSVita},
		{"Vita LiveArea", PKGTypePSPVita, ContentTypePSVitaLA, 2, core.PlatformPSVita},
		{"Vita Theme", PKGTypePSPVita, ContentTypePSVitaTheme, 2, core.PlatformPSVita},

		// PSM content types
		{"PSM", PKGTypePSPVita, ContentTypePSM, 4, core.PlatformPSM},
		{"PSM Unity", PKGTypePSPVita, ContentTypePSMUnity, 4, core.PlatformPSM},

		// PS1 emulator (platform depends on pkg_type)
		{"PS1 on PS3", PKGTypePS3, ContentTypePS1Emu, 0, core.PlatformPS3},
		{"PS1 on PSP", PKGTypePSPVita, ContentTypePS1Emu, 1, core.PlatformPSP},

		// Fallback to extended header key
		{"Unknown content, Vita key", PKGTypePSPVita, 0, 2, core.PlatformPSVita},
		{"Unknown content, PSP key", PKGTypePSPVita, 0, 1, core.PlatformPSP},
		{"Unknown content, PSM key", PKGTypePSPVita, 0, 4, core.PlatformPSM},

		// Fallback to pkg_type
		{"Unknown PS3", PKGTypePS3, 0, 0, core.PlatformPS3},
		{"Unknown PSP/Vita defaults to PSP", PKGTypePSPVita, 0, 0, core.PlatformPSP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectPlatform(tt.pkgType, tt.contentType, tt.extKeyID)
			if got != tt.want {
				t.Errorf("detectPlatform(%v, %v, %v) = %v, want %v",
					tt.pkgType, tt.contentType, tt.extKeyID, got, tt.want)
			}
		})
	}
}

func TestParsePKGErrors(t *testing.T) {
	t.Run("file too small", func(t *testing.T) {
		data := make([]byte, pkgHeaderSize-1)
		_, err := ParsePKG(bytes.NewReader(data), int64(len(data)))
		if err == nil {
			t.Error("expected error for file too small")
		}
	})

	t.Run("invalid magic", func(t *testing.T) {
		data := make([]byte, pkgHeaderSize)
		copy(data[0:4], "XXXX")
		_, err := ParsePKG(bytes.NewReader(data), int64(len(data)))
		if err == nil {
			t.Error("expected error for invalid magic")
		}
	})
}

func TestPKGGameInfo(t *testing.T) {
	contentID := "UP0001-NPUA80472_00-LITTLEBIGPLAN001"
	pkg := makeTestPKG(contentID, PKGTypePSPVita, ContentTypePSVitaGame, true)

	info, err := ParsePKG(bytes.NewReader(pkg), int64(len(pkg)))
	if err != nil {
		t.Fatalf("ParsePKG failed: %v", err)
	}

	// GameSerial should return the full Content ID
	if serial := info.GameSerial(); serial != contentID {
		t.Errorf("GameSerial() = %q, want %q", serial, contentID)
	}

	// GameTitle should return SFO title
	if title := info.GameTitle(); title != "Test Game" {
		t.Errorf("GameTitle() = %q, want %q", title, "Test Game")
	}

	// GamePlatform should return Vita (based on content type)
	if platform := info.GamePlatform(); platform != core.PlatformPSVita {
		t.Errorf("GamePlatform() = %v, want %v", platform, core.PlatformPSVita)
	}
}

func TestParsePKGVitaExtHeader(t *testing.T) {
	// Test that Vita packages use extended header for platform detection
	contentID := "JP0001-PCSG00001_00-TESTVITA00000001"
	pkg := makeTestPKG(contentID, PKGTypePSPVita, ContentTypePSVitaGame, false)

	info, err := ParsePKG(bytes.NewReader(pkg), int64(len(pkg)))
	if err != nil {
		t.Fatalf("ParsePKG failed: %v", err)
	}

	if info.Platform != core.PlatformPSVita {
		t.Errorf("Platform = %v, want %v", info.Platform, core.PlatformPSVita)
	}
}
