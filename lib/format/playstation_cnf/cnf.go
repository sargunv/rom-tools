package playstation_cnf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/sargunv/rom-tools/lib/core"
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

// VideoMode represents the video output mode for PS2 discs.
type VideoMode string

const (
	VideoModeNTSC VideoMode = "NTSC"
	VideoModePAL  VideoMode = "PAL"
)

// CNFInfo contains metadata extracted from a PlayStation SYSTEM.CNF file.
type CNFInfo struct {
	// Platform is PS1 or PS2, determined by the boot line type.
	Platform core.Platform `json:",omitempty"`
	// DiscID is the game identifier from the boot path (e.g., "SCUS_943.00").
	DiscID string `json:",omitempty"`
	// Version is the disc version from VER line (PS2 only).
	Version string `json:",omitempty"`
	// VideoMode is NTSC or PAL (PS2 only).
	VideoMode VideoMode `json:",omitempty"`
}

// ParseCNF parses PlayStation SYSTEM.CNF content from a reader.
func ParseCNF(r io.ReaderAt, size int64) (*CNFInfo, error) {
	data := make([]byte, size)
	if _, err := r.ReadAt(data, 0); err != nil {
		return nil, fmt.Errorf("failed to read SYSTEM.CNF: %w", err)
	}

	return parseCNFBytes(data)
}

func parseCNFBytes(data []byte) (*CNFInfo, error) {
	info := &CNFInfo{}

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
			info.Platform = core.PlatformPS2
			info.DiscID = extractDiscID(value)
		case "BOOT":
			// PS1 disc (only use if BOOT2 not found)
			if info.Platform != core.PlatformPS2 {
				info.Platform = core.PlatformPS1
				info.DiscID = extractDiscID(value)
			}
		case "VER":
			info.Version = value
		case "VMODE":
			info.VideoMode = VideoMode(value)
		}
	}

	if info.DiscID == "" {
		return nil, fmt.Errorf("not a valid PlayStation SYSTEM.CNF: no disc ID found")
	}

	return info, nil
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
