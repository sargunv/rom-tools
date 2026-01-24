// Package identify provides ROM identification and hashing utilities.
package identify

import "github.com/sargunv/rom-tools/lib/core"

// Re-export hash types from core for convenience.
type HashType = core.HashType

const (
	HashSHA1                = core.HashSHA1
	HashMD5                 = core.HashMD5
	HashCRC32               = core.HashCRC32
	HashZipCRC32            = core.HashZipCRC32
	HashCHDUncompressedSHA1 = core.HashCHDUncompressedSHA1
	HashCHDCompressedSHA1   = core.HashCHDCompressedSHA1
)

// Re-export Hashes from core for convenience.
type Hashes = core.Hashes

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

// Options controls ROM identification behavior.
type Options struct {
	// MaxHashSize is the maximum file size (in bytes) for which hashes will be calculated.
	// Files larger than this skip hash calculation entirely.
	// Use -1 for no limit (always calculate hashes).
	// Default is 64 MiB.
	MaxHashSize int64
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxHashSize: 64 * 1024 * 1024, // 64 MiB
	}
}
