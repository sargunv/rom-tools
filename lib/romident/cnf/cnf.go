package cnf

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/sargunv/rom-tools/lib/romident/core"
)

// PlayStation SYSTEM.CNF parsing for disc identification.
//
// PS1 and PS2 discs contain a SYSTEM.CNF file at the root with format:
//
// PS2:
//
//	BOOT2 = cdrom0:\SLUS_123.45;1
//	VER = 1.00
//	VMODE = NTSC
//
// PS1:
//
//	BOOT = cdrom:\SCUS_943.00;1
//	TCB = 4
//	EVENT = 16
//
// The disc ID (e.g., "SLUS-123.45") is the executable filename.
// The prefix encodes the region:
//   - SLUS/SCUS = US (NTSC-U)
//   - SLES/SCES = EU (PAL)
//   - SLPS/SCPS/SLPM = JP (NTSC-J)
//   - SLKA/SCKA = KR

// Info contains metadata extracted from a PlayStation disc.
type Info struct {
	DiscID    string // Original format from SYSTEM.CNF (e.g., "SCUS_943.00")
	Version   string // from VER line (PS2 only)
	VideoMode string // NTSC or PAL (PS2 only)
}

// IdentifyFromSystemCNF parses PlayStation SYSTEM.CNF content.
// Returns nil if not a valid PS1/PS2 SYSTEM.CNF.
func IdentifyFromSystemCNF(data []byte) *core.GameIdent {
	platform, info := parseSystemCNF(data)
	if info == nil {
		return nil
	}

	// Normalize disc ID: replace underscore with hyphen, remove periods
	normalizedID := strings.ReplaceAll(info.DiscID, "_", "-")
	normalizedID = strings.ReplaceAll(normalizedID, ".", "")

	return &core.GameIdent{
		Platform: platform,
		TitleID:  normalizedID,
		Regions:  []core.Region{decodeRegion(info.DiscID)},
		Extra:    info,
	}
}

// parseSystemCNF extracts PlayStation info from SYSTEM.CNF content.
// Returns the platform and info, or nil if neither BOOT nor BOOT2 found.
func parseSystemCNF(data []byte) (core.Platform, *Info) {
	info := &Info{}
	var platform core.Platform

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Split on '=' and trim spaces
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "BOOT2":
			// PS2 disc
			platform = core.PlatformPS2
			info.DiscID = extractDiscID(value)
		case "BOOT":
			// PS1 disc (only use if BOOT2 not found)
			if platform != core.PlatformPS2 {
				platform = core.PlatformPS1
				info.DiscID = extractDiscID(value)
			}
		case "VER":
			info.Version = value
		case "VMODE":
			info.VideoMode = value
		}
	}

	if info.DiscID == "" {
		return "", nil
	}

	return platform, info
}

// extractDiscID extracts the disc ID from a BOOT/BOOT2 path.
// Input: "cdrom0:\SLUS_123.45;1" or "cdrom:\SCUS_943.00;1"
// Output: "SLUS_123.45" or "SCUS_943.00"
func extractDiscID(bootPath string) string {
	// Find last backslash
	lastBackslash := strings.LastIndex(bootPath, "\\")
	if lastBackslash == -1 {
		// Try forward slash as fallback
		lastBackslash = strings.LastIndex(bootPath, "/")
	}

	var filename string
	if lastBackslash != -1 {
		filename = bootPath[lastBackslash+1:]
	} else {
		filename = bootPath
	}

	// Remove version suffix (;1)
	if semicolon := strings.Index(filename, ";"); semicolon != -1 {
		filename = filename[:semicolon]
	}

	return filename
}

// decodeRegion maps a PlayStation disc ID prefix to a region.
func decodeRegion(discID string) core.Region {
	if len(discID) < 4 {
		return core.RegionUnknown
	}

	prefix := strings.ToUpper(discID[:4])

	switch prefix {
	case "SLUS", "SCUS":
		return core.RegionUS
	case "SLES", "SCES":
		return core.RegionEU
	case "SLPS", "SCPS", "SLPM", "SCPM":
		return core.RegionJP
	case "SLKA", "SCKA":
		return core.RegionKR
	case "SLAJ": // Asia/Japan variant
		return core.RegionJP
	default:
		return core.RegionUnknown
	}
}
