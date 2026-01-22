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
	// BootPath is the raw boot path from BOOT/BOOT2 line (e.g., "cdrom0:\SLUS_123.45;1").
	BootPath string `json:",omitempty"`
	// DiscID is the game identifier extracted from the boot path (e.g., "SLUS_123.45").
	DiscID string `json:",omitempty"`
	// Version is the disc version from VER line (PS2 only).
	Version string `json:",omitempty"`
	// VideoMode is NTSC or PAL (PS2 only).
	VideoMode VideoMode `json:",omitempty"`
}

// Parse parses PlayStation SYSTEM.CNF content from a reader.
func Parse(r io.ReaderAt, size int64) (*CNFInfo, error) {
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
			info.BootPath = value
		case "BOOT":
			// PS1 disc (only use if BOOT2 not found)
			if info.Platform != core.PlatformPS2 {
				info.Platform = core.PlatformPS1
				info.BootPath = value
			}
		case "VER":
			info.Version = value
		case "VMODE":
			info.VideoMode = VideoMode(value)
		}
	}

	if info.BootPath == "" {
		return nil, fmt.Errorf("not a valid PlayStation SYSTEM.CNF: no boot path found")
	}

	info.DiscID = extractDiscID(info.BootPath)

	return info, nil
}

// extractDiscID extracts the disc ID from a BOOT/BOOT2 path.
// The disc ID is the portion after any of the characters `:`, `/`, or `\`
// and before the string `;1`.
//
// Examples:
//   - "cdrom0:\SLUS_123.45;1" → "SLUS_123.45"
//   - "cdrom:\SCUS_943.00;1" → "SCUS_943.00"
//   - "cdrom:SCUS_943.01" → "SCUS_943.01" (no backslash)
func extractDiscID(bootPath string) string {
	// Find last occurrence of any separator (`:`, `/`, or `\`)
	lastSep := -1
	for _, sep := range []string{":", "/", "\\"} {
		if idx := strings.LastIndex(bootPath, sep); idx > lastSep {
			lastSep = idx
		}
	}

	var filename string
	if lastSep != -1 {
		filename = bootPath[lastSep+1:]
	} else {
		filename = bootPath
	}

	// Remove ";1" suffix if present
	filename = strings.TrimSuffix(filename, ";1")

	return filename
}
