package romident

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/sargunv/rom-tools/lib/romident/game/gb"
	"github.com/sargunv/rom-tools/lib/romident/game/md"
	"github.com/sargunv/rom-tools/lib/romident/game/n64"
	"github.com/sargunv/rom-tools/lib/romident/game/nds"
	"github.com/sargunv/rom-tools/lib/romident/game/nes"
	"github.com/sargunv/rom-tools/lib/romident/game/snes"
)

// Magic bytes and offsets for various formats
var (
	// CHD v5: magic at start
	chdMagic  = []byte("MComprHD")
	chdOffset = int64(0)

	// Xbox XISO: magic at offset 0x10000
	xisoMagic  = []byte("MICROSOFT*XBOX*MEDIA")
	xisoOffset = int64(0x10000)

	// ISO9660: magic at offset 0x8001 (sector 16 + 1 byte for type code)
	iso9660Magic  = []byte("CD001")
	iso9660Offset = int64(0x8001)

	// ZIP: magic at start
	zipMagic  = []byte{0x50, 0x4B, 0x03, 0x04}
	zipOffset = int64(0)

	// XBE: magic at start
	xbeMagic  = []byte("XBEH")
	xbeOffset = int64(0)

	// GBA: fixed value 0x96 required at offset 0xB2
	gbaMagic  = []byte{0x96}
	gbaOffset = int64(0xB2)
)

// Detector can detect the format of a file.
type Detector struct{}

// NewDetector creates a new format detector.
func NewDetector() *Detector {
	return &Detector{}
}

// CandidatesByExtension returns possible formats based on file extension.
// Returns nil for generic/unknown extensions (we don't do magic-only detection).
func (d *Detector) CandidatesByExtension(filename string) []Format {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".chd":
		return []Format{FormatCHD}
	case ".zip":
		return []Format{FormatZIP}
	case ".iso":
		// Ambiguous: could be XISO or ISO9660
		return []Format{FormatXISO, FormatISO9660}
	case ".xiso":
		return []Format{FormatXISO}
	case ".xbe":
		return []Format{FormatXBE}
	case ".gba":
		return []Format{FormatGBA}
	case ".z64":
		return []Format{FormatZ64}
	case ".v64":
		return []Format{FormatV64}
	case ".n64":
		return []Format{FormatN64}
	case ".gb", ".gbc":
		return []Format{FormatGB}
	case ".md", ".gen":
		return []Format{FormatMD}
	case ".smd":
		return []Format{FormatSMD}
	case ".nds", ".dsi", ".ids":
		return []Format{FormatNDS}
	case ".nes":
		return []Format{FormatNES}
	case ".sfc", ".smc":
		return []Format{FormatSNES}
	default:
		// Generic extensions like .bin or no extension: no candidates
		return nil
	}
}

// ReaderAtSeeker combines io.ReaderAt and io.Seeker for random access reading.
type ReaderAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

// VerifyFormat checks if a file matches a specific format using magic bytes.
func (d *Detector) VerifyFormat(r io.ReaderAt, size int64, format Format) bool {
	switch format {
	case FormatZIP:
		return checkMagic(r, size, zipOffset, zipMagic)
	case FormatCHD:
		return checkMagic(r, size, chdOffset, chdMagic)
	case FormatXISO:
		return checkMagic(r, size, xisoOffset, xisoMagic)
	case FormatXBE:
		return checkMagic(r, size, xbeOffset, xbeMagic)
	case FormatISO9660:
		return checkMagic(r, size, iso9660Offset, iso9660Magic)
	case FormatGBA:
		return checkMagic(r, size, gbaOffset, gbaMagic)
	case FormatNES:
		return nes.IsNESROM(r, size)
	case FormatNDS:
		return nds.IsNDSROM(r, size)
	case FormatGB:
		return gb.IsGBROM(r, size)
	case FormatMD:
		return md.IsMDROM(r, size)
	case FormatSMD:
		return md.IsSMDROM(r, size)
	case FormatSNES:
		return snes.IsSNESROM(r, size)
	case FormatZ64:
		return checkN64Format(r, size) == FormatZ64
	case FormatV64:
		return checkN64Format(r, size) == FormatV64
	case FormatN64:
		return checkN64Format(r, size) == FormatN64
	default:
		return false
	}
}

// checkMagic is a helper to verify magic bytes at a specific offset.
func checkMagic(r io.ReaderAt, size int64, offset int64, magic []byte) bool {
	if size < offset+int64(len(magic)) {
		return false
	}
	buf := make([]byte, len(magic))
	if _, err := r.ReadAt(buf, offset); err != nil {
		return false
	}
	return bytesEqual(buf, magic)
}

// checkN64Format returns the N64 format variant if valid, Unknown otherwise.
func checkN64Format(r io.ReaderAt, size int64) Format {
	if size < 4 {
		return FormatUnknown
	}

	buf := make([]byte, 4)
	if _, err := r.ReadAt(buf, 0); err != nil {
		return FormatUnknown
	}

	return Format(n64.DetectN64ByteOrder(buf))
}

// Detect identifies the format using extension to narrow candidates, then magic to verify.
// Returns Unknown for generic extensions (like .bin) or if magic verification fails.
func (d *Detector) Detect(r ReaderAtSeeker, size int64, filename string) (Format, error) {
	candidates := d.CandidatesByExtension(filename)
	if len(candidates) == 0 {
		// Generic or unknown extension: no identification
		return FormatUnknown, nil
	}

	// Try each candidate and return the first that verifies
	for _, candidate := range candidates {
		if d.VerifyFormat(r, size, candidate) {
			return candidate, nil
		}
	}

	// Extension suggested format(s) but none verified
	return FormatUnknown, nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
