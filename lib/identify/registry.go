package identify

import (
	"io"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sargunv/rom-tools/lib/roms/gamecube"
	"github.com/sargunv/rom-tools/lib/roms/gb"
	"github.com/sargunv/rom-tools/lib/roms/gba"
	"github.com/sargunv/rom-tools/lib/roms/megadrive"
	"github.com/sargunv/rom-tools/lib/roms/n64"
	"github.com/sargunv/rom-tools/lib/roms/nds"
	"github.com/sargunv/rom-tools/lib/roms/nes"
	"github.com/sargunv/rom-tools/lib/roms/playstation/psnpkg"
	"github.com/sargunv/rom-tools/lib/roms/sms"
	"github.com/sargunv/rom-tools/lib/roms/snes"
	"github.com/sargunv/rom-tools/lib/roms/xbox"
)

// formatEntry associates a format with its extensions and identification function.
// TODO: Format should be returned by parsers via GameInfo (like Platform) rather than
// being defined here. This would allow removing the Format field from the registry.
type formatEntry struct {
	Format     Format
	Extensions []string
	Identify   func(r io.ReaderAt, size int64) (GameInfo, error)
}

// wrapParser converts a typed parser function to the generic GameInfo signature.
// This is needed because Go function types are invariant - a function returning
// *GBAInfo is not assignable to a function returning GameInfo even though
// *GBAInfo implements GameInfo.
func wrapParser[T GameInfo](fn func(io.ReaderAt, int64) (T, error)) func(io.ReaderAt, int64) (GameInfo, error) {
	return func(r io.ReaderAt, size int64) (GameInfo, error) {
		return fn(r, size)
	}
}

// registry contains all registered format entries.
// Entries are ordered by specificity - more specific extensions first.
// For ambiguous extensions like .iso, multiple formats are registered and
// the detection logic tries each candidate in order.
var registry = []formatEntry{
	{FormatGBA, []string{".gba"}, wrapParser(gba.ParseGBA)},
	{FormatNDS, []string{".nds", ".dsi", ".ids"}, wrapParser(nds.ParseNDS)},
	{FormatNES, []string{".nes"}, wrapParser(nes.ParseNES)},
	{FormatSNES, []string{".sfc", ".smc"}, wrapParser(snes.ParseSNES)},
	{FormatGB, []string{".gb", ".gbc"}, wrapParser(gb.ParseGB)},
	{FormatZ64, []string{".z64"}, wrapParser(n64.ParseN64)},
	{FormatV64, []string{".v64"}, wrapParser(n64.ParseN64)},
	{FormatN64, []string{".n64"}, wrapParser(n64.ParseN64)},
	{FormatMD, []string{".32x", ".md", ".gen"}, wrapParser(megadrive.Parse)},
	{FormatSMD, []string{".smd"}, wrapParser(megadrive.Parse)},
	{FormatSMS, []string{".sms", ".gg"}, wrapParser(sms.ParseSMS)},
	{FormatXISO, []string{".xiso", ".iso"}, wrapParser(xbox.ParseXISO)},
	{FormatXBE, []string{".xbe"}, wrapParser(xbox.ParseXBE)},
	{FormatGCM, []string{".gcm", ".iso"}, wrapParser(gamecube.ParseGCM)},
	{FormatRVZ, []string{".rvz", ".wia"}, wrapParser(gamecube.ParseRVZ)},
	{FormatCHD, []string{".chd"}, identifyCHD},
	{FormatZIP, []string{".zip"}, nil},
	{FormatISO9660, []string{".iso", ".bin"}, identifyISO9660},
	{FormatPKG, []string{".pkg"}, wrapParser(psnpkg.ParsePKG)},
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
