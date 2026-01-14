package romident

import (
	"path/filepath"
	"strings"

	"github.com/sargunv/rom-tools/lib/romident/core"
	"github.com/sargunv/rom-tools/lib/romident/gb"
	"github.com/sargunv/rom-tools/lib/romident/gba"
	"github.com/sargunv/rom-tools/lib/romident/gcm"
	"github.com/sargunv/rom-tools/lib/romident/iso9660"
	"github.com/sargunv/rom-tools/lib/romident/md"
	"github.com/sargunv/rom-tools/lib/romident/n64"
	"github.com/sargunv/rom-tools/lib/romident/nds"
	"github.com/sargunv/rom-tools/lib/romident/nes"
	"github.com/sargunv/rom-tools/lib/romident/rvz"
	"github.com/sargunv/rom-tools/lib/romident/smd"
	"github.com/sargunv/rom-tools/lib/romident/snes"
	v64 "github.com/sargunv/rom-tools/lib/romident/v64"
	"github.com/sargunv/rom-tools/lib/romident/xbe"
	"github.com/sargunv/rom-tools/lib/romident/xiso"
	"github.com/sargunv/rom-tools/lib/romident/z64"
)

// FormatEntry associates a format with its extensions and identification function.
type FormatEntry struct {
	Format     Format
	Extensions []string
	Identify   core.IdentifyFunc
}

// registry contains all registered format entries.
// Entries are ordered by specificity - more specific extensions first.
// For ambiguous extensions like .iso, multiple formats are registered and
// the detection logic tries each candidate in order.
var registry = []FormatEntry{
	{FormatGBA, []string{".gba"}, gba.Identify},
	{FormatNDS, []string{".nds", ".dsi", ".ids"}, nds.Identify},
	{FormatNES, []string{".nes"}, nes.Identify},
	{FormatSNES, []string{".sfc", ".smc"}, snes.Identify},
	{FormatGB, []string{".gb", ".gbc"}, gb.Identify},
	{FormatZ64, []string{".z64"}, z64.Identify},
	{FormatV64, []string{".v64"}, v64.Identify},
	{FormatN64, []string{".n64"}, n64.Identify},
	{FormatMD, []string{".md", ".gen"}, md.Identify},
	{FormatSMD, []string{".smd"}, smd.Identify},
	{FormatXISO, []string{".xiso", ".iso"}, xiso.Identify},
	{FormatXBE, []string{".xbe"}, xbe.Identify},
	{FormatGCM, []string{".gcm", ".iso"}, gcm.Identify},
	{FormatRVZ, []string{".rvz", ".wia"}, rvz.Identify},
	{FormatCHD, []string{".chd"}, nil},
	{FormatZIP, []string{".zip"}, nil},
	{FormatISO9660, []string{".iso", ".bin"}, iso9660.Identify},
}

// FormatsByExtension returns all format entries that match the given filename extension.
// Returns entries in order of preference (more specific formats first).
func FormatsByExtension(filename string) []FormatEntry {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil
	}

	// Find all matching entries
	var entries []FormatEntry
	for _, entry := range registry {
		for _, entryExt := range entry.Extensions {
			if entryExt == ext {
				entries = append(entries, entry)
				break
			}
		}
	}
	return entries
}
