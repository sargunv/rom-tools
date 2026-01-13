package bin

import (
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/lib/romident/core"
	"github.com/sargunv/rom-tools/lib/romident/iso9660"
)

// BIN file identification.
//
// .bin files are raw CD images with variable sector sizes:
//   - MODE1/2048: Standard 2048 bytes/sector (same as ISO9660)
//   - MODE2/2352: Raw 2352 bytes/sector (16 byte header + 2048 data + 276 ECC)
//
// To identify, we:
// 1. Detect sector size by probing for ISO9660 magic at different offsets
// 2. Create a sector-translating reader
// 3. Delegate to iso9660.Identify

// sectorConfig describes where to find ISO9660 PVD for a given sector format.
type sectorConfig struct {
	sectorSize int64  // bytes per physical sector
	dataOffset int64  // offset to user data within sector
	pvdOffset  int64  // offset to ISO9660 PVD magic
	name       string // format name for debugging
}

// Probe locations for ISO9660 PVD (sector 16)
var probeConfigs = []sectorConfig{
	{SectorSize2048, 0, 0x8000, "MODE1/2048"},               // Standard: sector 16 * 2048
	{SectorSize2352, RawSectorHeader, 0x9318, "MODE2/2352"}, // Raw: sector 16 * 2352 + 24 header
}

// Identify detects the sector format and identifies the disc.
func Identify(r io.ReaderAt, size int64) (*core.GameIdent, error) {
	// Probe for ISO9660 magic at known locations
	for _, cfg := range probeConfigs {
		// PVD magic is at offset 1 within sector 16, so check at pvdOffset + 1
		magicOffset := cfg.pvdOffset + 1
		if size < magicOffset+5 {
			continue
		}

		magic := make([]byte, 5)
		if _, err := r.ReadAt(magic, magicOffset); err != nil {
			continue
		}

		if string(magic) == "CD001" {
			// Found ISO9660! Create sector reader and delegate
			sr := NewSectorReader(r, cfg.sectorSize, cfg.dataOffset, size)
			return iso9660.Identify(sr, sr.Size())
		}
	}

	return nil, fmt.Errorf("not a valid BIN file: no ISO9660 filesystem found")
}
