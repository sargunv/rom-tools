package util

import (
	"strings"
	"time"
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
