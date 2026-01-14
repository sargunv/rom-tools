package psp

import (
	"strings"

	"github.com/sargunv/rom-tools/lib/romident/core"
	"github.com/sargunv/rom-tools/lib/romident/sfo"
)

// PSP PARAM.SFO identification.
//
// PSP games store metadata in PSP_GAME/PARAM.SFO using the SFO binary format.
// Key fields:
//   - DISC_ID: Game identifier (e.g., "ULUS10041")
//   - TITLE: Game title
//   - CATEGORY: Content type ("UG" = UMD game)
//
// Region prefixes:
//   - ULUS/UCUS/NPUG = US
//   - ULES/UCES/NPEG = EU
//   - ULJS/UCJS/NPJG = JP
//   - ULAS/NPAG = Asia
//   - ULKS/NPHG = Korea
//   - NPxH/NPxZ = Digital variants
//
// References:
//   - https://www.psdevwiki.com/psp/UMD
//   - https://www.psdevwiki.com/psp/Param.sfo

// Info contains metadata extracted from a PSP disc.
type Info struct {
	DiscID   string // Original DISC_ID from SFO
	Title    string // Game title
	Category string // Content category
}

// IdentifyFromSFO parses PSP PARAM.SFO content.
// Returns nil if not a valid PSP PARAM.SFO.
func IdentifyFromSFO(data []byte) *core.GameIdent {
	parsed, err := sfo.Parse(data)
	if err != nil {
		return nil
	}

	discID := sfo.GetString(parsed, "DISC_ID")
	if discID == "" {
		return nil
	}

	info := &Info{
		DiscID:   discID,
		Title:    sfo.GetString(parsed, "TITLE"),
		Category: sfo.GetString(parsed, "CATEGORY"),
	}

	// Normalize disc ID: add hyphen after 4-char prefix if not present
	normalizedID := normalizeDiscID(discID)

	return &core.GameIdent{
		Platform: core.PlatformPSP,
		TitleID:  normalizedID,
		Title:    info.Title,
		Extra:    info,
	}
}

// normalizeDiscID converts "ULUS10041" to "ULUS-10041".
func normalizeDiscID(discID string) string {
	// If already has hyphen, return as-is
	if strings.Contains(discID, "-") {
		return discID
	}
	// Insert hyphen after 4-character prefix
	if len(discID) > 4 {
		return discID[:4] + "-" + discID[4:]
	}
	return discID
}
