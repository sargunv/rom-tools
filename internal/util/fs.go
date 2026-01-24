package util

import (
	"io"
)

// FileEntry represents a file within a container.
type FileEntry struct {
	Name  string // Relative path within container
	Size  int64  // Uncompressed size
	CRC32 uint32 // Pre-computed CRC32 (0 if not available, e.g., for folders)
}

// FileContainer represents a container format (ZIP, folder, etc.) that can enumerate
// and provide access to its contents.
type FileContainer interface {
	// Entries returns all files in the container.
	Entries() []FileEntry

	// OpenFile opens a file for sequential reading.
	OpenFile(name string) (io.ReadCloser, error)

	// OpenFileAt opens a file with random access support (for format detection/identification).
	// Returns the reader and the file size.
	OpenFileAt(name string) (RandomAccessReader, int64, error)

	// Close releases resources associated with the container.
	Close() error
}

// RandomAccessReader combines io.ReaderAt and io.Closer.
// This is needed for format detection and identification which require random access.
type RandomAccessReader interface {
	io.ReaderAt
	io.Closer
}
