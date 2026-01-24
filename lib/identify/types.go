// Package identify provides ROM identification and hashing utilities.
package identify

import "github.com/sargunv/rom-tools/lib/core"

// Item represents one identifiable unit (a file or entry within a container).
type Item struct {
	Name   string        `json:"name"`             // filename (basename for single files, relative path in containers)
	Size   int64         `json:"size"`             // file size in bytes
	Hashes core.Hashes   `json:"hashes,omitempty"` // hash values by type
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
	// This only applies when hashes are not already available from:
	//   - Container metadata (e.g., ZIP files provide zip-crc32)
	//   - Embedded format hashes (e.g., CHD files provide chd-*-sha1)
	// Files exceeding this limit will have no hashes unless provided by the above sources.
	// Use -1 for no limit (always calculate when needed).
	// Default is -1 (no limit).
	MaxHashSize int64
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxHashSize: -1, // no limit
	}
}
