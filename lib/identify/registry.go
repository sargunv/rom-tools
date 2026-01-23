package identify

import (
	"io"
	"path/filepath"
	"slices"
	"strings"
)

// formatEntry associates a format with its extensions and identification function.
type formatEntry struct {
	Format     Format
	Extensions []string
	Identify   func(r io.ReaderAt, size int64) (*GameIdent, error)
}

// registry contains all registered format entries.
// Entries are ordered by specificity - more specific extensions first.
// For ambiguous extensions like .iso, multiple formats are registered and
// the detection logic tries each candidate in order.
var registry = []formatEntry{
	{FormatGBA, []string{".gba"}, identifyGBA},
	{FormatNDS, []string{".nds", ".dsi", ".ids"}, identifyNDS},
	{FormatNES, []string{".nes"}, identifyNES},
	{FormatSNES, []string{".sfc", ".smc"}, identifySNES},
	{FormatGB, []string{".gb", ".gbc"}, identifyGB},
	{FormatZ64, []string{".z64"}, identifyZ64},
	{FormatV64, []string{".v64"}, identifyV64},
	{FormatN64, []string{".n64"}, identifyN64},
	{Format32X, []string{".32x"}, identify32X},
	{FormatMD, []string{".md", ".gen"}, identifyMD},
	{FormatSMD, []string{".smd"}, identifySMD},
	{FormatSMS, []string{".sms"}, identifySMS},
	{FormatGG, []string{".gg"}, identifySMS},
	{FormatXISO, []string{".xiso", ".iso"}, identifyXISO},
	{FormatXBE, []string{".xbe"}, identifyXBE},
	{FormatGCM, []string{".gcm", ".iso"}, identifyGCM},
	{FormatRVZ, []string{".rvz", ".wia"}, identifyRVZ},
	{FormatCHD, []string{".chd"}, identifyCHD},
	{FormatZIP, []string{".zip"}, nil},
	{FormatISO9660, []string{".iso", ".bin"}, identifyISO9660},
}

// formatsByExtension returns all format entries that match the given filename extension.
// Returns entries in order of preference (more specific formats first).
func formatsByExtension(filename string) []formatEntry {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil
	}

	// Find all matching entries
	var entries []formatEntry
	for _, entry := range registry {
		if slices.Contains(entry.Extensions, ext) {
			entries = append(entries, entry)
		}
	}
	return entries
}
