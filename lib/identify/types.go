// Package identify provides ROM identification and hashing utilities.
package identify

import "github.com/sargunv/rom-tools/lib/core"

// HashType identifies a specific hash (combines algorithm and source).
type HashType string

const (
	HashSHA1                HashType = "sha1"
	HashMD5                 HashType = "md5"
	HashCRC32               HashType = "crc32"
	HashZipCRC32            HashType = "zip-crc32"
	HashCHDUncompressedSHA1 HashType = "chd-uncompressed-sha1"
	HashCHDCompressedSHA1   HashType = "chd-compressed-sha1"
)

// Hashes maps hash type to hex-encoded value.
type Hashes map[HashType]string

// Item represents one identifiable unit (a file or entry within a container).
type Item struct {
	Name   string        `json:"name"`             // filename (basename for single files, relative path in containers)
	Size   int64         `json:"size"`             // file size in bytes
	Hashes Hashes        `json:"hashes,omitempty"` // hash values by type
	Game   core.GameInfo `json:"game,omitempty"`   // identified game info (platform-specific struct)
}

// Result is the result of identifying a path.
type Result struct {
	Path  string `json:"path"`  // absolute path that was identified
	Items []Item `json:"items"` // identified items (1 for single file, N for containers)
}

// HashMode controls how hashes are calculated.
type HashMode int

const (
	// HashModeDefault uses fast methods where available (CHD header, ZIP metadata),
	// calculates full hashes for loose files.
	HashModeDefault HashMode = iota

	// HashModeFast skips hash calculation for large files, but still calculates
	// hashes for small files (below fastModeSmallFileThreshold).
	HashModeFast

	// HashModeSlow calculates full hashes even when fast methods are available
	// (e.g., decompresses ZIP files to calculate SHA1/MD5).
	HashModeSlow
)

const (
	// fastModeSmallFileThreshold is the size threshold below which fast mode
	// will still calculate hashes. Files at or above this size skip hash calculation.
	// 65 MiB covers most cartridge ROMs (GBA, SNES, NES, etc.) but skips large disc images.
	fastModeSmallFileThreshold = 65 * 1024 * 1024 // 65 MiB
)

// Options controls ROM identification behavior.
type Options struct {
	HashMode HashMode
}
