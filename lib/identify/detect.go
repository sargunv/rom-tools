package identify

import (
	"bytes"
	"io"
)

// Magic bytes and offsets for container/disc formats (non-game formats).
// Game format verification is delegated to the format registry identifiers.
var (
	// CHD v5: magic at start
	chdMagic  = []byte("MComprHD")
	chdOffset = int64(0)

	// ISO9660: magic at offset 0x8001 (sector 16 + 1 byte for type code)
	iso9660Magic  = []byte("CD001")
	iso9660Offset = int64(0x8001)

	// ZIP: magic at start
	zipMagic  = []byte{0x50, 0x4B, 0x03, 0x04}
	zipOffset = int64(0)
)

// detector can detect the format of a file.
type detector struct{}

// newDetector creates a new format detector.
func newDetector() *detector {
	return &detector{}
}

// candidatesByExtension returns possible formats based on file extension.
// Returns nil for generic/unknown extensions (we don't do magic-only detection).
// This delegates to the format registry for most lookups.
func (d *detector) candidatesByExtension(filename string) []Format {
	entries := formatsByExtension(filename)
	if len(entries) == 0 {
		return nil
	}

	formats := make([]Format, len(entries))
	for i, entry := range entries {
		formats[i] = entry.Format
	}
	return formats
}

// verifyFormat checks if a file matches a specific format.
// For container/disc formats (CHD, ZIP, ISO9660), uses magic bytes.
// For game formats, uses the format registry's identifier.
func (d *detector) verifyFormat(r io.ReaderAt, size int64, format Format) bool {
	// Container/disc formats: use magic byte verification
	switch format {
	case FormatZIP:
		return checkMagic(r, size, zipOffset, zipMagic)
	case FormatCHD:
		return checkMagic(r, size, chdOffset, chdMagic)
	case FormatISO9660:
		return checkMagic(r, size, iso9660Offset, iso9660Magic)
	}

	// Game formats: find the identify function in the registry and try to identify
	for _, entry := range registry {
		if entry.Format == format && entry.Identify != nil {
			_, err := entry.Identify(r, size)
			return err == nil
		}
	}

	return false
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
	return bytes.Equal(buf, magic)
}

// detect identifies the format using extension to narrow candidates, then verifies.
// Returns Unknown for generic extensions (like .bin) or if verification fails.
// For game formats, this delegates to the format registry's identifiers.
func (d *detector) detect(r interface {
	io.ReaderAt
	io.Seeker
}, size int64, filename string) (Format, error) {
	candidates := d.candidatesByExtension(filename)
	if len(candidates) == 0 {
		// Generic or unknown extension: no identification
		return FormatUnknown, nil
	}

	// Try each candidate and return the first that verifies
	for _, candidate := range candidates {
		if d.verifyFormat(r, size, candidate) {
			return candidate, nil
		}
	}

	// Extension suggested format(s) but none verified
	return FormatUnknown, nil
}
