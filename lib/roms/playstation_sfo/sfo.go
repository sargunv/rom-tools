package playstation_sfo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/sargunv/rom-tools/lib/core"
)

// SFO (System File Object) binary format parser.
//
// SFO is a metadata format used by PlayStation platforms (PSP, PS3, PS Vita, PS4)
// to store game information like title, disc ID, and system requirements.
//
// Format structure:
//   - Header (20 bytes): magic, version, table offsets, entry count
//   - Index table: 16-byte entries describing each key-value pair
//   - Key table: null-terminated UTF-8 key names
//   - Data table: values (strings or integers)
//
// References:
//   - https://www.psdevwiki.com/psp/Param.sfo
//   - https://www.psdevwiki.com/ps3/PARAM.SFO

const (
	sfoMagic     = "\x00PSF"
	sfoHeaderMin = 20

	// Data format types
	formatUTF8Special = 0x0004 // UTF-8 not null-terminated
	formatUTF8        = 0x0204 // UTF-8 null-terminated
	formatInt32       = 0x0404 // 32-bit unsigned integer
)

// sfoData represents parsed SFO key-value data.
// String values are returned as string, integers as uint32.
type sfoData map[string]any

// SFOInfo contains metadata extracted from an SFO file with platform detection.
type SFOInfo struct {
	// Platform is PSP, PS3, Vita, or PS4, determined from DISC_ID prefix.
	Platform core.Platform `json:",omitempty"`
	// DiscID is the game identifier (e.g., "ULUS10041", "BLUS30001").
	DiscID string `json:",omitempty"`
	// Title is the game title from the SFO.
	Title string `json:",omitempty"`
	// Category is the content category (e.g., "UG" for UMD game).
	Category string `json:",omitempty"`
	// Version is the disc media version (DISC_VERSION).
	Version string `json:",omitempty"`
	// AppVersion is the application/patch version (APP_VER).
	AppVersion string `json:",omitempty"`
	// DiscNumber is the disc number for multi-disc games (DISC_NUMBER, 1-indexed).
	DiscNumber int `json:",omitempty"`
	// DiscTotal is the total number of discs for multi-disc games (DISC_TOTAL).
	DiscTotal int `json:",omitempty"`
	// ParentalLevel is the content rating level (PARENTAL_LEVEL).
	ParentalLevel int `json:",omitempty"`
	// SystemVersion is the required system version (PSP_SYSTEM_VER or PS3_SYSTEM_VER).
	SystemVersion string `json:",omitempty"`
	// Region is the geographic region code (REGION).
	Region int `json:",omitempty"`
}

// Parse reads an SFO file and returns high-level game information.
func Parse(r io.ReaderAt, size int64) (*SFOInfo, error) {
	data, err := parsesfoData(r, size)
	if err != nil {
		return nil, err
	}

	// Try DISC_ID first, fall back to TITLE_ID (PS3 compatibility)
	discID := getString(data, "DISC_ID")
	if discID == "" {
		discID = getString(data, "TITLE_ID")
	}
	if discID == "" {
		return nil, fmt.Errorf("not a valid SFO: missing DISC_ID or TITLE_ID")
	}

	platform := detectPlatform(discID)

	// Try PSP_SYSTEM_VER first, then PS3_SYSTEM_VER
	systemVer := getString(data, "PSP_SYSTEM_VER")
	if systemVer == "" {
		systemVer = getString(data, "PS3_SYSTEM_VER")
	}

	return &SFOInfo{
		Platform:      platform,
		DiscID:        discID,
		Title:         getString(data, "TITLE"),
		Category:      getString(data, "CATEGORY"),
		Version:       getString(data, "DISC_VERSION"),
		AppVersion:    getString(data, "APP_VER"),
		DiscNumber:    getInt(data, "DISC_NUMBER"),
		DiscTotal:     getInt(data, "DISC_TOTAL"),
		ParentalLevel: getInt(data, "PARENTAL_LEVEL"),
		SystemVersion: systemVer,
		Region:        getInt(data, "REGION"),
	}, nil
}

// detectPlatform determines the PlayStation platform from the DISC_ID prefix.
//
// Platform prefixes:
//   - PSP: ULUS, UCUS, ULES, UCES, ULJS, UCJS, ULAS, ULKS, NPxG/NPxH (digital)
//   - PS3: BLUS, BLES, BLJM, BCUS, BCES, NPxB (digital), etc.
//   - Vita: PCSA, PCSB, PCSE, PCSH, PCSG, PCSD, etc.
//   - PS4: CUSA, PLAS, etc.
func detectPlatform(discID string) core.Platform {
	if len(discID) < 4 {
		return ""
	}

	prefix := strings.ToUpper(discID[:4])

	// PSP prefixes (UMD and digital)
	switch prefix {
	case "ULUS", "UCUS", // US
		"ULES", "UCES", // EU
		"ULJS", "UCJS", // JP
		"ULAS", "UCAS", // Asia
		"ULKS", "UCKS", // Korea
		"NPUG", "NPUH", // PSN digital US
		"NPEG", "NPEH", // PSN digital EU
		"NPJG", "NPJH", "NPJJ", // PSN digital JP
		"NPAG", "NPAH", // PSN digital Asia
		"NPHG", "NPHH": // PSN digital HK
		return core.PlatformPSP
	}

	// PS3 prefixes
	switch prefix {
	case "BLUS", "BCUS", // US
		"BLES", "BCES", // EU
		"BLJM", "BLJS", "BCJS", // JP
		"BLAS", "BCAS", // Asia
		"BLKS", "BCKS", // Korea
		"NPUB", "NPEB", "NPJB", "NPAB", "NPHB": // PSN digital
		return core.PlatformPS3
	}

	// Vita prefixes
	switch prefix {
	case "PCSA", "PCSE", // US
		"PCSB", "PCSF", // EU
		"PCSC", "PCSG", "VLJM", // JP
		"PCSH", // Asia
		"PCSD": // Demo
		return core.PlatformPSVita
	}

	// PS4 prefixes
	switch prefix {
	case "CUSA", // Universal
		"PLAS", "PCAS": // Asia
		return core.PlatformPS4
	}

	// Check 2-char prefix for broader matching
	if len(discID) >= 2 {
		prefix2 := strings.ToUpper(discID[:2])
		switch prefix2 {
		case "UL", "UC": // PSP UMD
			return core.PlatformPSP
		case "BL", "BC": // PS3 Blu-ray
			return core.PlatformPS3
		case "PC": // Vita
			return core.PlatformPSVita
		case "CU": // PS4
			return core.PlatformPS4
		}
	}

	return ""
}

// parsesfoData reads an SFO file and returns raw key-value pairs.
func parsesfoData(r io.ReaderAt, size int64) (sfoData, error) {
	if size < sfoHeaderMin {
		return nil, fmt.Errorf("file too small for SFO header: need %d bytes, got %d", sfoHeaderMin, size)
	}

	data := make([]byte, size)
	if _, err := r.ReadAt(data, 0); err != nil {
		return nil, fmt.Errorf("failed to read SFO: %w", err)
	}

	// Validate magic
	if string(data[0:4]) != sfoMagic {
		return nil, fmt.Errorf("invalid SFO magic: %x", data[0:4])
	}

	// Read header
	keyTableOffset := binary.LittleEndian.Uint32(data[8:12])
	dataTableOffset := binary.LittleEndian.Uint32(data[12:16])
	numEntries := binary.LittleEndian.Uint32(data[16:20])

	// Validate offsets
	if keyTableOffset > uint32(len(data)) || dataTableOffset > uint32(len(data)) {
		return nil, fmt.Errorf("SFO table offsets out of bounds")
	}

	result := make(sfoData)

	// Parse index entries (16 bytes each, starting at offset 20)
	indexOffset := uint32(20)
	for i := uint32(0); i < numEntries; i++ {
		entryOffset := indexOffset + i*16
		if entryOffset+16 > uint32(len(data)) {
			return nil, fmt.Errorf("SFO index entry %d out of bounds", i)
		}

		keyOffset := binary.LittleEndian.Uint16(data[entryOffset:])
		dataFormat := binary.LittleEndian.Uint16(data[entryOffset+2:])
		dataLen := binary.LittleEndian.Uint32(data[entryOffset+4:])
		dataOffset := binary.LittleEndian.Uint32(data[entryOffset+12:])

		// Read key name (null-terminated string)
		keyStart := keyTableOffset + uint32(keyOffset)
		if keyStart >= uint32(len(data)) {
			return nil, fmt.Errorf("SFO key %d offset out of bounds", i)
		}
		keyEnd := keyStart
		for keyEnd < uint32(len(data)) && data[keyEnd] != 0 {
			keyEnd++
		}
		if keyEnd >= uint32(len(data)) {
			return nil, fmt.Errorf("SFO key %d has no null terminator", i)
		}
		key := string(data[keyStart:keyEnd])

		// Read data value
		dataStart := dataTableOffset + dataOffset
		if dataStart+dataLen > uint32(len(data)) {
			return nil, fmt.Errorf("SFO data for key %q out of bounds", key)
		}

		switch dataFormat {
		case formatInt32:
			if dataLen >= 4 {
				result[key] = binary.LittleEndian.Uint32(data[dataStart:])
			}
		case formatUTF8, formatUTF8Special:
			strData := data[dataStart : dataStart+dataLen]
			// Truncate at first null byte (everything after is garbage)
			if idx := bytes.IndexByte(strData, 0); idx >= 0 {
				strData = strData[:idx]
			}
			// Trim whitespace
			result[key] = strings.TrimSpace(string(strData))
		default:
			// Unknown format, store as raw bytes
			result[key] = data[dataStart : dataStart+dataLen]
		}
	}

	return result, nil
}

// getString returns a string value from parsed SFO data.
func getString(sfo sfoData, key string) string {
	if v, ok := sfo[key].(string); ok {
		return v
	}
	return ""
}

// getInt returns an integer value from parsed SFO data.
func getInt(sfo sfoData, key string) int {
	if v, ok := sfo[key].(uint32); ok {
		return int(v)
	}
	return 0
}
