// Package format provides file format detection for ROM files.
package format

import (
	"io"
	"path/filepath"
	"strings"
)

// Format indicates the detected file format.
type Format string

const (
	Unknown Format = "unknown"
	CHD     Format = "chd"
	XISO    Format = "xiso"
	XBE     Format = "xbe"
	ISO9660 Format = "iso9660"
	ZIP     Format = "zip"
	GBA     Format = "gba"
	Z64     Format = "z64" // N64 big-endian (native)
	V64     Format = "v64" // N64 byte-swapped
	N64     Format = "n64" // N64 word-swapped (little-endian)
	GB      Format = "gb"
	MD      Format = "md"
	SMD     Format = "smd"
	NDS     Format = "nds"  // Nintendo DS
	NES     Format = "nes"  // Nintendo Entertainment System
	SNES    Format = "snes" // Super Nintendo Entertainment System
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
		return []Format{CHD}
	case ".zip":
		return []Format{ZIP}
	case ".iso":
		// Ambiguous: could be XISO or ISO9660
		return []Format{XISO, ISO9660}
	case ".xiso":
		return []Format{XISO}
	case ".xbe":
		return []Format{XBE}
	case ".gba":
		return []Format{GBA}
	case ".z64":
		return []Format{Z64}
	case ".v64":
		return []Format{V64}
	case ".n64":
		return []Format{N64}
	case ".gb", ".gbc":
		return []Format{GB}
	case ".md", ".gen":
		return []Format{MD}
	case ".smd":
		return []Format{SMD}
	case ".nds", ".dsi", ".ids":
		return []Format{NDS}
	case ".nes":
		return []Format{NES}
	case ".sfc", ".smc":
		return []Format{SNES}
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
	case ZIP:
		return checkMagic(r, size, zipOffset, zipMagic)
	case CHD:
		return checkMagic(r, size, chdOffset, chdMagic)
	case XISO:
		return checkMagic(r, size, xisoOffset, xisoMagic)
	case XBE:
		return checkMagic(r, size, xbeOffset, xbeMagic)
	case ISO9660:
		return checkMagic(r, size, iso9660Offset, iso9660Magic)
	case GBA:
		return checkMagic(r, size, gbaOffset, gbaMagic)
	case NES:
		return IsNESROM(r, size)
	case NDS:
		return IsNDSROM(r, size)
	case GB:
		return IsGBROM(r, size)
	case MD:
		return IsMDROM(r, size)
	case SMD:
		return IsSMDROM(r, size)
	case SNES:
		return IsSNESROM(r, size)
	case Z64:
		return checkN64Format(r, size) == Z64
	case V64:
		return checkN64Format(r, size) == V64
	case N64:
		return checkN64Format(r, size) == N64
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
		return Unknown
	}
	buf := make([]byte, 4)
	if _, err := r.ReadAt(buf, 0); err != nil {
		return Unknown
	}
	return DetectN64Format(buf)
}

// Detect identifies the format using extension to narrow candidates, then magic to verify.
// Returns Unknown for generic extensions (like .bin) or if magic verification fails.
func (d *Detector) Detect(r ReaderAtSeeker, size int64, filename string) (Format, error) {
	candidates := d.CandidatesByExtension(filename)
	if len(candidates) == 0 {
		// Generic or unknown extension: no identification
		return Unknown, nil
	}

	// Try each candidate and return the first that verifies
	for _, candidate := range candidates {
		if d.VerifyFormat(r, size, candidate) {
			return candidate, nil
		}
	}

	// Extension suggested format(s) but none verified
	return Unknown, nil
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
