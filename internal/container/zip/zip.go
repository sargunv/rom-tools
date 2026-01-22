// Package zip provides ZIP archive handling for ROM identification.
// It wraps the standard library zip package for random access.
package zip

import (
	"archive/zip"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
)

// ZIPHandler handles ZIP archive files.
type ZIPHandler struct{}

// NewZIPHandler creates a new ZIP handler.
func NewZIPHandler() *ZIPHandler {
	return &ZIPHandler{}
}

// ZIPArchive represents an open ZIP archive and implements Container.
type ZIPArchive struct {
	reader  *zip.ReadCloser
	entries []util.FileEntry
}

// Entries returns all files in the ZIP archive.
func (z *ZIPArchive) Entries() []util.FileEntry {
	return z.entries
}

// Close closes the ZIP archive.
func (z *ZIPArchive) Close() error {
	return z.reader.Close()
}

// OpenFile opens a file within the ZIP archive for reading.
func (z *ZIPArchive) OpenFile(name string) (io.ReadCloser, error) {
	for _, f := range z.reader.File {
		if f.Name == name {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("file not found in ZIP: %s", name)
}

// ZIPEntryReader wraps EntryReader to implement util.RandomAccessReader.
type ZIPEntryReader struct {
	*EntryReader
}

// OpenFileAt opens a file within the ZIP archive with random access support.
// Returns a RandomAccessReader that implements io.ReaderAt by buffering decompressed data.
// This is useful for format detection and header parsing without decompressing the entire file.
func (z *ZIPArchive) OpenFileAt(name string) (util.RandomAccessReader, error) {
	for _, f := range z.reader.File {
		if f.Name == name {
			return &ZIPEntryReader{NewEntryReader(f)}, nil
		}
	}
	return nil, fmt.Errorf("file not found in ZIP: %s", name)
}

// Open opens a ZIP archive and returns metadata for all files.
func (h *ZIPHandler) Open(path string) (*ZIPArchive, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}

	var entries []util.FileEntry
	for _, f := range r.File {
		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}

		entries = append(entries, util.FileEntry{
			Name:  f.Name,
			Size:  int64(f.UncompressedSize64),
			CRC32: f.CRC32, // Pre-computed CRC32 from ZIP metadata
		})
	}

	return &ZIPArchive{
		reader:  r,
		entries: entries,
	}, nil
}
