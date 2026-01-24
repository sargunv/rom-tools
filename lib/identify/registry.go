package identify

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/sargunv/rom-tools/lib/core"
	"github.com/sargunv/rom-tools/lib/roms/gamecube"
	"github.com/sargunv/rom-tools/lib/roms/gb"
	"github.com/sargunv/rom-tools/lib/roms/gba"
	"github.com/sargunv/rom-tools/lib/roms/megadrive"
	"github.com/sargunv/rom-tools/lib/roms/n3ds"
	"github.com/sargunv/rom-tools/lib/roms/n64"
	"github.com/sargunv/rom-tools/lib/roms/nds"
	"github.com/sargunv/rom-tools/lib/roms/nes"
	"github.com/sargunv/rom-tools/lib/roms/playstation/psnpkg"
	"github.com/sargunv/rom-tools/lib/roms/sms"
	"github.com/sargunv/rom-tools/lib/roms/snes"
	"github.com/sargunv/rom-tools/lib/roms/xbox"
)

// IdentifyFunc attempts to identify a ROM from a reader.
// Returns game info, optional embedded hashes (for formats like CHD), and error.
type IdentifyFunc func(r io.ReaderAt, size int64) (core.GameInfo, core.Hashes, error)

// wrapParser converts a typed parser function to the generic signature.
// This is needed because Go function types are invariant - a function returning
// *GBAInfo is not assignable to a function returning GameInfo even though
// *GBAInfo implements GameInfo.
func wrapParser[T core.GameInfo](fn func(io.ReaderAt, int64) (T, error)) IdentifyFunc {
	return func(r io.ReaderAt, size int64) (core.GameInfo, core.Hashes, error) {
		info, err := fn(r, size)
		return info, nil, err
	}
}

// registry maps file extensions to ordered list of parsers to try.
// Parsers are tried in order until one succeeds.
var registry = map[string][]IdentifyFunc{
	".gba":  {wrapParser(gba.ParseGBA)},
	".gb":   {wrapParser(gb.ParseGB)},
	".gbc":  {wrapParser(gb.ParseGB)},
	".nds":  {wrapParser(nds.ParseNDS)},
	".dsi":  {wrapParser(nds.ParseNDS)},
	".ids":  {wrapParser(nds.ParseNDS)},
	".3ds":  {wrapParser(n3ds.ParseN3DS)},
	".cci":  {wrapParser(n3ds.ParseN3DS)},
	".nes":  {wrapParser(nes.ParseNES)},
	".sfc":  {wrapParser(snes.ParseSNES)},
	".smc":  {wrapParser(snes.ParseSNES)},
	".z64":  {wrapParser(n64.ParseN64)},
	".v64":  {wrapParser(n64.ParseN64)},
	".n64":  {wrapParser(n64.ParseN64)},
	".md":   {wrapParser(megadrive.Parse)},
	".gen":  {wrapParser(megadrive.Parse)},
	".32x":  {wrapParser(megadrive.Parse)},
	".smd":  {wrapParser(megadrive.Parse)},
	".sms":  {wrapParser(sms.ParseSMS)},
	".gg":   {wrapParser(sms.ParseSMS)},
	".xbe":  {wrapParser(xbox.ParseXBE)},
	".pkg":  {wrapParser(psnpkg.ParsePKG)},
	".chd":  {identifyCHD},
	".rvz":  {wrapParser(gamecube.ParseRVZ)},
	".wia":  {wrapParser(gamecube.ParseRVZ)},
	".gcm":  {wrapParser(gamecube.ParseGCM)},
	".xiso": {wrapParser(xbox.ParseXISO)},
	".iso":  {wrapParser(xbox.ParseXISO), wrapParser(gamecube.ParseGCM), identifyISO9660},
	".bin":  {identifyISO9660},
}

// identifyByExtension returns the list of parsers to try for a given filename.
func identifyByExtension(filename string) []IdentifyFunc {
	ext := strings.ToLower(filepath.Ext(filename))
	return registry[ext]
}
