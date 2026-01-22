package playstation_sfo

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

// makeTestSFO builds a valid SFO binary with the given key-value pairs.
// String values are stored as UTF-8, uint32 values as integers.
func makeTestSFO(entries map[string]any) []byte {
	// Sort keys for deterministic output (required by SFO format)
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	// Simple sort for test purposes
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// Calculate sizes
	numEntries := len(keys)
	indexTableSize := numEntries * 16
	keyTableStart := 20 + indexTableSize

	// Build key table and track offsets
	keyTable := make([]byte, 0)
	keyOffsets := make([]int, numEntries)
	for i, key := range keys {
		keyOffsets[i] = len(keyTable)
		keyTable = append(keyTable, []byte(key)...)
		keyTable = append(keyTable, 0) // null terminator
	}

	// Align key table to 4 bytes
	for len(keyTable)%4 != 0 {
		keyTable = append(keyTable, 0)
	}

	dataTableStart := keyTableStart + len(keyTable)

	// Build data table and index entries
	dataTable := make([]byte, 0)
	indexEntries := make([]byte, 0, indexTableSize)
	dataOffsets := make([]int, numEntries)

	for i, key := range keys {
		dataOffsets[i] = len(dataTable)
		val := entries[key]

		var dataFormat uint16
		var dataLen uint32
		var maxLen uint32

		switch v := val.(type) {
		case string:
			dataFormat = formatUTF8
			strBytes := append([]byte(v), 0) // null terminated
			dataLen = uint32(len(strBytes))
			maxLen = dataLen
			// Align to 4 bytes
			for len(strBytes)%4 != 0 {
				strBytes = append(strBytes, 0)
				maxLen++
			}
			dataTable = append(dataTable, strBytes...)
		case uint32:
			dataFormat = formatInt32
			dataLen = 4
			maxLen = 4
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, v)
			dataTable = append(dataTable, buf...)
		}

		// Build index entry (16 bytes)
		entry := make([]byte, 16)
		binary.LittleEndian.PutUint16(entry[0:2], uint16(keyOffsets[i]))
		binary.LittleEndian.PutUint16(entry[2:4], dataFormat)
		binary.LittleEndian.PutUint32(entry[4:8], dataLen)
		binary.LittleEndian.PutUint32(entry[8:12], maxLen)
		binary.LittleEndian.PutUint32(entry[12:16], uint32(dataOffsets[i]))
		indexEntries = append(indexEntries, entry...)
	}

	// Build header
	header := make([]byte, 20)
	copy(header[0:4], sfoMagic)
	binary.LittleEndian.PutUint32(header[4:8], 0x00000101)             // version 1.1
	binary.LittleEndian.PutUint32(header[8:12], uint32(keyTableStart)) // key table offset
	binary.LittleEndian.PutUint32(header[12:16], uint32(dataTableStart))
	binary.LittleEndian.PutUint32(header[16:20], uint32(numEntries))

	// Combine all parts
	result := make([]byte, 0, dataTableStart+len(dataTable))
	result = append(result, header...)
	result = append(result, indexEntries...)
	result = append(result, keyTable...)
	result = append(result, dataTable...)

	return result
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		entries  map[string]any
		want     *SFOInfo
		wantErr  bool
		errMatch string
	}{
		{
			name: "PSP UMD game",
			entries: map[string]any{
				"DISC_ID":        "ULUS10041",
				"TITLE":          "Test PSP Game",
				"CATEGORY":       "UG",
				"DISC_VERSION":   "1.00",
				"PSP_SYSTEM_VER": "3.00",
				"PARENTAL_LEVEL": uint32(3),
				"REGION":         uint32(32768),
			},
			want: &SFOInfo{
				Platform:      core.PlatformPSP,
				DiscID:        "ULUS10041",
				Title:         "Test PSP Game",
				Category:      "UG",
				Version:       "1.00",
				SystemVersion: "3.00",
				ParentalLevel: 3,
				Region:        32768,
			},
		},
		{
			name: "PS3 Blu-ray game",
			entries: map[string]any{
				"DISC_ID":        "BLUS30001",
				"TITLE":          "Test PS3 Game",
				"CATEGORY":       "DG",
				"DISC_VERSION":   "01.00",
				"APP_VER":        "01.00",
				"PS3_SYSTEM_VER": "02.00",
				"PARENTAL_LEVEL": uint32(5),
			},
			want: &SFOInfo{
				Platform:      core.PlatformPS3,
				DiscID:        "BLUS30001",
				Title:         "Test PS3 Game",
				Category:      "DG",
				Version:       "01.00",
				AppVersion:    "01.00",
				SystemVersion: "02.00",
				ParentalLevel: 5,
			},
		},
		{
			name: "PS3 with TITLE_ID fallback",
			entries: map[string]any{
				"TITLE_ID": "NPUB30001",
				"TITLE":    "Test PS3 Digital",
				"CATEGORY": "HG",
			},
			want: &SFOInfo{
				Platform: core.PlatformPS3,
				DiscID:   "NPUB30001",
				Title:    "Test PS3 Digital",
				Category: "HG",
			},
		},
		{
			name: "PS Vita game",
			entries: map[string]any{
				"DISC_ID":  "PCSA00001",
				"TITLE":    "Test Vita Game",
				"CATEGORY": "gd",
			},
			want: &SFOInfo{
				Platform: core.PlatformPSVita,
				DiscID:   "PCSA00001",
				Title:    "Test Vita Game",
				Category: "gd",
			},
		},
		{
			name: "PS4 game",
			entries: map[string]any{
				"DISC_ID":        "CUSA00001",
				"TITLE":          "Test PS4 Game",
				"CATEGORY":       "gd",
				"APP_VER":        "01.00",
				"DISC_NUMBER":    uint32(1),
				"DISC_TOTAL":     uint32(2),
				"PARENTAL_LEVEL": uint32(7),
			},
			want: &SFOInfo{
				Platform:      core.PlatformPS4,
				DiscID:        "CUSA00001",
				Title:         "Test PS4 Game",
				Category:      "gd",
				AppVersion:    "01.00",
				DiscNumber:    1,
				DiscTotal:     2,
				ParentalLevel: 7,
			},
		},
		{
			name: "multi-disc game",
			entries: map[string]any{
				"DISC_ID":     "BLUS30002",
				"TITLE":       "Multi-Disc Game",
				"CATEGORY":    "DG",
				"DISC_NUMBER": uint32(2),
				"DISC_TOTAL":  uint32(3),
			},
			want: &SFOInfo{
				Platform:   core.PlatformPS3,
				DiscID:     "BLUS30002",
				Title:      "Multi-Disc Game",
				Category:   "DG",
				DiscNumber: 2,
				DiscTotal:  3,
			},
		},
		{
			name:     "missing DISC_ID and TITLE_ID",
			entries:  map[string]any{"TITLE": "No ID Game"},
			wantErr:  true,
			errMatch: "missing DISC_ID or TITLE_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := makeTestSFO(tt.entries)
			info, err := Parse(bytes.NewReader(data), int64(len(data)))

			if tt.wantErr {
				if err == nil {
					t.Error("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if info.Platform != tt.want.Platform {
				t.Errorf("Platform = %v, want %v", info.Platform, tt.want.Platform)
			}
			if info.DiscID != tt.want.DiscID {
				t.Errorf("DiscID = %q, want %q", info.DiscID, tt.want.DiscID)
			}
			if info.Title != tt.want.Title {
				t.Errorf("Title = %q, want %q", info.Title, tt.want.Title)
			}
			if info.Category != tt.want.Category {
				t.Errorf("Category = %q, want %q", info.Category, tt.want.Category)
			}
			if info.Version != tt.want.Version {
				t.Errorf("Version = %q, want %q", info.Version, tt.want.Version)
			}
			if info.AppVersion != tt.want.AppVersion {
				t.Errorf("AppVersion = %q, want %q", info.AppVersion, tt.want.AppVersion)
			}
			if info.DiscNumber != tt.want.DiscNumber {
				t.Errorf("DiscNumber = %d, want %d", info.DiscNumber, tt.want.DiscNumber)
			}
			if info.DiscTotal != tt.want.DiscTotal {
				t.Errorf("DiscTotal = %d, want %d", info.DiscTotal, tt.want.DiscTotal)
			}
			if info.ParentalLevel != tt.want.ParentalLevel {
				t.Errorf("ParentalLevel = %d, want %d", info.ParentalLevel, tt.want.ParentalLevel)
			}
			if info.SystemVersion != tt.want.SystemVersion {
				t.Errorf("SystemVersion = %q, want %q", info.SystemVersion, tt.want.SystemVersion)
			}
			if info.Region != tt.want.Region {
				t.Errorf("Region = %d, want %d", info.Region, tt.want.Region)
			}
		})
	}
}

func TestParse_Errors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "file too small",
			data: []byte{0x00, 'P', 'S', 'F'},
		},
		{
			name: "invalid magic",
			data: append([]byte("XXXX"), make([]byte, 16)...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(bytes.NewReader(tt.data), int64(len(tt.data)))
			if err == nil {
				t.Error("Parse() expected error, got nil")
			}
		})
	}
}

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		discID string
		want   core.Platform
	}{
		// PSP prefixes (UMD)
		{"ULUS10041", core.PlatformPSP},
		{"UCUS98632", core.PlatformPSP},
		{"ULES00125", core.PlatformPSP},
		{"UCES00001", core.PlatformPSP},
		{"ULJS00001", core.PlatformPSP},
		{"UCJS10001", core.PlatformPSP},
		{"ULAS42001", core.PlatformPSP},
		{"UCAS40001", core.PlatformPSP},
		{"ULKS46001", core.PlatformPSP},
		{"UCKS46001", core.PlatformPSP},
		// PSP prefixes (PSN digital)
		{"NPUG80001", core.PlatformPSP},
		{"NPUH10000", core.PlatformPSP},
		{"NPEG00001", core.PlatformPSP},
		{"NPEH00001", core.PlatformPSP},
		{"NPJG00001", core.PlatformPSP},
		{"NPJH50443", core.PlatformPSP}, // Final Fantasy Type-0
		{"NPJJ00001", core.PlatformPSP},
		{"NPAG00001", core.PlatformPSP},
		{"NPAH00001", core.PlatformPSP},
		{"NPHG00001", core.PlatformPSP},
		{"NPHH00001", core.PlatformPSP},

		// PS3 prefixes
		{"BLUS30001", core.PlatformPS3},
		{"BCUS98111", core.PlatformPS3},
		{"BLES00001", core.PlatformPS3},
		{"BCES00001", core.PlatformPS3},
		{"BLJM60001", core.PlatformPS3},
		{"BLJS10001", core.PlatformPS3},
		{"BCJS30001", core.PlatformPS3},
		{"BLAS50001", core.PlatformPS3},
		{"BCAS20001", core.PlatformPS3},
		{"BLKS20001", core.PlatformPS3},
		{"BCKS10001", core.PlatformPS3},
		{"NPUB30001", core.PlatformPS3},
		{"NPEB00001", core.PlatformPS3},

		// PS Vita prefixes
		{"PCSA00001", core.PlatformPSVita},
		{"PCSB00001", core.PlatformPSVita},
		{"PCSE00001", core.PlatformPSVita},
		{"PCSF00001", core.PlatformPSVita},
		{"PCSC00001", core.PlatformPSVita},
		{"PCSG00001", core.PlatformPSVita},
		{"VLJM35001", core.PlatformPSVita},
		{"PCSH00001", core.PlatformPSVita},
		{"PCSD00001", core.PlatformPSVita},

		// PS4 prefixes
		{"CUSA00001", core.PlatformPS4},
		{"PLAS00001", core.PlatformPS4},
		{"PCAS00001", core.PlatformPS4},

		// 2-char fallback matching
		{"ULXX99999", core.PlatformPSP},    // UL -> PSP
		{"UCXX99999", core.PlatformPSP},    // UC -> PSP
		{"BLXX99999", core.PlatformPS3},    // BL -> PS3
		{"BCXX99999", core.PlatformPS3},    // BC -> PS3
		{"PCXX99999", core.PlatformPSVita}, // PC -> Vita
		{"CUXX99999", core.PlatformPS4},    // CU -> PS4

		// Unknown/invalid
		{"XXXX00001", ""},
		{"AB", ""},
		{"A", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.discID, func(t *testing.T) {
			got := detectPlatform(tt.discID)
			if got != tt.want {
				t.Errorf("detectPlatform(%q) = %v, want %v", tt.discID, got, tt.want)
			}
		})
	}
}
