package playstation_sfo

import (
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

// SFOData represents parsed SFO key-value data.
// String values are returned as string, integers as uint32.
type SFOData map[string]any

// SFOInfo contains metadata extracted from an SFO file with platform detection.
type SFOInfo struct {
	// Platform is PSP, PS3, Vita, or PS4, determined from DISC_ID prefix.
	Platform core.Platform
	// DiscID is the game identifier (e.g., "ULUS10041", "BLUS30001").
	DiscID string
	// Title is the game title from the SFO.
	Title string
	// Category is the content category (e.g., "UG" for UMD game).
	Category string
}

// ParseSFO reads an SFO file and returns high-level game information.
func ParseSFO(r io.ReaderAt, size int64) (*SFOInfo, error) {
	data, err := parseSFOData(r, size)
	if err != nil {
		return nil, err
	}

	discID := getString(data, "DISC_ID")
	if discID == "" {
		return nil, fmt.Errorf("not a valid SFO: missing DISC_ID")
	}

	platform := detectPlatform(discID)

	return &SFOInfo{
		Platform: platform,
		DiscID:   discID,
		Title:    getString(data, "TITLE"),
		Category: getString(data, "CATEGORY"),
	}, nil
}

// detectPlatform determines the PlayStation platform from the DISC_ID prefix.
//
// Platform prefixes:
//   - PSP: ULUS, UCUS, ULES, UCES, ULJS, UCJS, ULAS, ULKS, NPxx (digital)
//   - PS3: BLUS, BLES, BLJM, BCUS, BCES, NPUB, NPEB, NPJB, etc.
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
		"ULAS",                                 // Asia
		"ULKS",                                 // Korea
		"NPUG", "NPEG", "NPJG", "NPAG", "NPHG": // PSN digital
		return core.PlatformPSP
	}

	// PS3 prefixes
	switch prefix {
	case "BLUS", "BCUS", // US
		"BLES", "BCES", // EU
		"BLJM", "BCJS", // JP
		"BLAS", "BCAS", // Asia
		"NPUB", "NPEB", "NPJB", "NPAB", "NPHB": // PSN digital
		return core.PlatformPS3
	}

	// Vita prefixes
	switch prefix {
	case "PCSA", // US
		"PCSB",         // EU
		"PCSE",         // EU (alternate)
		"PCSG", "PCSH", // JP/Asia
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

// parseSFOData reads an SFO file and returns raw key-value pairs.
func parseSFOData(r io.ReaderAt, size int64) (SFOData, error) {
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

	result := make(SFOData)

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
			// Trim null terminator if present
			strData := data[dataStart : dataStart+dataLen]
			for len(strData) > 0 && strData[len(strData)-1] == 0 {
				strData = strData[:len(strData)-1]
			}
			result[key] = string(strData)
		default:
			// Unknown format, store as raw bytes
			result[key] = data[dataStart : dataStart+dataLen]
		}
	}

	return result, nil
}

// getString returns a string value from parsed SFO data.
func getString(sfo SFOData, key string) string {
	if v, ok := sfo[key].(string); ok {
		return v
	}
	return ""
}
