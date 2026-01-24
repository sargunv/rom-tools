package identify

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/sargunv/rom-tools/lib/core"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/gb"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/gba"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/gcm"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/n3ds"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/n64"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/nds"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/nes"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/rvz"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/sfc"
	"github.com/sargunv/rom-tools/lib/roms/playstation/pkg"
	"github.com/sargunv/rom-tools/lib/roms/sega/md"
	"github.com/sargunv/rom-tools/lib/roms/sega/sms"
	"github.com/sargunv/rom-tools/lib/roms/xbox/xbe"
	"github.com/sargunv/rom-tools/lib/roms/xbox/xiso"
)

// identifyFunc attempts to identify content from a reader.
// Returns game info, optional embedded hashes (for formats like CHD), and error.
type identifyFunc func(r io.ReaderAt, size int64) (core.GameInfo, core.Hashes, error)

// wrapParser converts a typed parser function to the generic signature.
// This is needed because Go function types are invariant - a function returning
// *GBAInfo is not assignable to a function returning GameInfo even though
// *GBAInfo implements GameInfo.
func wrapParser[T core.GameInfo](fn func(io.ReaderAt, int64) (T, error)) identifyFunc {
	return func(r io.ReaderAt, size int64) (core.GameInfo, core.Hashes, error) {
		info, err := fn(r, size)
		return info, nil, err
	}
}

// registry maps file extensions to ordered list of parsers to try.
// Parsers are tried in order until one succeeds.
var registry = map[string][]identifyFunc{
	".gba":  {wrapParser(gba.ParseGBA)},
	".gb":   {wrapParser(gb.ParseGB)},
	".gbc":  {wrapParser(gb.ParseGB)},
	".nds":  {wrapParser(nds.ParseNDS)},
	".dsi":  {wrapParser(nds.ParseNDS)},
	".ids":  {wrapParser(nds.ParseNDS)},
	".3ds":  {wrapParser(n3ds.ParseN3DS)},
	".cci":  {wrapParser(n3ds.ParseN3DS)},
	".nes":  {wrapParser(nes.ParseNES)},
	".sfc":  {wrapParser(sfc.ParseSNES)},
	".smc":  {wrapParser(sfc.ParseSNES)},
	".z64":  {wrapParser(n64.ParseN64)},
	".v64":  {wrapParser(n64.ParseN64)},
	".n64":  {wrapParser(n64.ParseN64)},
	".md":   {wrapParser(md.Parse)},
	".gen":  {wrapParser(md.Parse)},
	".32x":  {wrapParser(md.Parse)},
	".smd":  {wrapParser(md.Parse)},
	".sms":  {wrapParser(sms.ParseSMS)},
	".gg":   {wrapParser(sms.ParseSMS)},
	".xbe":  {wrapParser(xbe.ParseXBE)},
	".pkg":  {wrapParser(pkg.ParsePKG)},
	".chd":  {identifyCHD},
	".rvz":  {wrapParser(rvz.ParseRVZ)},
	".wia":  {wrapParser(rvz.ParseRVZ)},
	".gcm":  {wrapParser(gcm.ParseGCM)},
	".xiso": {wrapParser(xiso.ParseXISO)},
	".iso":  {wrapParser(xiso.ParseXISO), wrapParser(gcm.ParseGCM), identifyISO9660},
	".bin":  {identifyISO9660, wrapParser(md.Parse)},
}

// identifyByExtension returns the list of parsers to try for a given filename.
func identifyByExtension(filename string) []identifyFunc {
	ext := strings.ToLower(filepath.Ext(filename))
	return registry[ext]
}
