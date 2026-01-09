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
	ISO9660 Format = "iso9660"
	ZIP     Format = "zip"
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
)

// Detector can detect the format of a file.
type Detector struct{}

// NewDetector creates a new format detector.
func NewDetector() *Detector {
	return &Detector{}
}

// DetectByExtension returns a likely format based on file extension.
// Returns Unknown if the extension is not recognized.
func (d *Detector) DetectByExtension(filename string) Format {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".chd":
		return CHD
	case ".zip":
		return ZIP
	case ".iso":
		// Could be XISO or ISO9660, need magic detection
		return Unknown
	case ".xiso":
		return XISO
	default:
		return Unknown
	}
}

// ReaderAtSeeker combines io.ReaderAt and io.Seeker for random access reading.
type ReaderAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

// DetectByMagic detects the format by reading magic bytes.
// Requires a seekable reader to check various offsets.
func (d *Detector) DetectByMagic(r ReaderAtSeeker, size int64) (Format, error) {
	// Check for ZIP
	if size >= zipOffset+int64(len(zipMagic)) {
		buf := make([]byte, len(zipMagic))
		if _, err := r.ReadAt(buf, zipOffset); err == nil {
			if bytesEqual(buf, zipMagic) {
				return ZIP, nil
			}
		}
	}

	// Check for CHD
	if size >= chdOffset+int64(len(chdMagic)) {
		buf := make([]byte, len(chdMagic))
		if _, err := r.ReadAt(buf, chdOffset); err == nil {
			if bytesEqual(buf, chdMagic) {
				return CHD, nil
			}
		}
	}

	// Check for Xbox XISO
	if size >= xisoOffset+int64(len(xisoMagic)) {
		buf := make([]byte, len(xisoMagic))
		if _, err := r.ReadAt(buf, xisoOffset); err == nil {
			if bytesEqual(buf, xisoMagic) {
				return XISO, nil
			}
		}
	}

	// Check for ISO9660
	if size >= iso9660Offset+int64(len(iso9660Magic)) {
		buf := make([]byte, len(iso9660Magic))
		if _, err := r.ReadAt(buf, iso9660Offset); err == nil {
			if bytesEqual(buf, iso9660Magic) {
				return ISO9660, nil
			}
		}
	}

	return Unknown, nil
}

// Detect combines extension and magic detection.
// Uses extension as a hint, then verifies with magic bytes when possible.
func (d *Detector) Detect(r ReaderAtSeeker, size int64, filename string) (Format, error) {
	// First try magic detection
	format, err := d.DetectByMagic(r, size)
	if err != nil {
		return Unknown, err
	}

	if format != Unknown {
		return format, nil
	}

	// Fall back to extension
	return d.DetectByExtension(filename), nil
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
