// Package container provides handlers for container formats (ZIP, folders).
package container

import (
	"io"
)

// ReaderAtSeekCloser combines io.ReaderAt, io.Seeker, and io.Closer with Size().
// This is needed for format detection and identification which require random access.
type ReaderAtSeekCloser interface {
	io.ReaderAt
	io.Seeker
	io.Closer
	Size() int64
}

// FileEntry represents a file within a container.
type FileEntry struct {
	Name  string // Relative path within container
	Size  int64  // Uncompressed size
	CRC32 uint32 // Pre-computed CRC32 (0 if not available, e.g., for folders)
}

// Container represents a container format (ZIP, folder, etc.) that can enumerate
// and provide access to its contents.
type Container interface {
	// Entries returns all files in the container.
	Entries() []FileEntry

	// OpenFile opens a file for sequential reading.
	OpenFile(name string) (io.ReadCloser, error)

	// OpenFileAt opens a file with random access support (for format detection/identification).
	// Only available in slow mode for containers that require decompression (e.g., ZIP).
	OpenFileAt(name string) (ReaderAtSeekCloser, error)

	// Close releases resources associated with the container.
	Close() error
}
