package util

import "strings"

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
