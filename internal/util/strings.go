package util

import (
	"strings"
	"time"

	"golang.org/x/text/encoding/japanese"
)

// ExtractASCII extracts a null-terminated ASCII string from bytes.
func ExtractASCII(data []byte) string {
	// Find null terminator
	end := len(data)
	for i, b := range data {
		if b == 0 {
			end = i
			break
		}
	}
	return strings.TrimSpace(string(data[:end]))
}

// ParseYYYYMMDD parses a date string in YYYYMMDD format.
// Returns zero time if parsing fails.
func ParseYYYYMMDD(s string) time.Time {
	if len(s) != 8 {
		return time.Time{}
	}
	t, err := time.Parse("20060102", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ExtractShiftJIS extracts a null-terminated Shift-JIS string from bytes,
// decoding it to UTF-8. Used for Sega platforms (Mega Drive, Saturn, Dreamcast)
// where Japanese titles are encoded in Shift-JIS.
// Falls back to raw byte interpretation if decoding fails.
func ExtractShiftJIS(data []byte) string {
	// Find null terminator
	end := len(data)
	for i, b := range data {
		if b == 0 {
			end = i
			break
		}
	}

	// Decode Shift-JIS to UTF-8
	decoder := japanese.ShiftJIS.NewDecoder()
	decoded, err := decoder.Bytes(data[:end])
	if err != nil {
		// Fallback to raw byte interpretation
		return strings.TrimSpace(string(data[:end]))
	}
	return strings.TrimSpace(string(decoded))
}

// ExtractJISX0201 extracts a null-terminated JIS X 0201 string from bytes,
// decoding it to UTF-8. JIS X 0201 is ASCII (0x20-0x7F) plus half-width
// katakana (0xA1-0xDF). Used for SNES ROM headers.
// Stops at null bytes or invalid characters outside the JIS X 0201 range.
func ExtractJISX0201(data []byte) string {
	var result strings.Builder
	for _, b := range data {
		if b == 0 {
			break
		}
		// Printable ASCII range (0x20-0x7E)
		if b >= 0x20 && b <= 0x7E {
			result.WriteByte(b)
			continue
		}
		// Half-width katakana range (0xA1-0xDF) maps to U+FF61-U+FF9F
		if b >= 0xA1 && b <= 0xDF {
			// Unicode half-width katakana starts at U+FF61
			// 0xA1 -> U+FF61, 0xA2 -> U+FF62, etc.
			r := rune(0xFF61 + int(b) - 0xA1)
			result.WriteRune(r)
			continue
		}
		// Invalid byte - stop here
		break
	}
	return strings.TrimSpace(result.String())
}
