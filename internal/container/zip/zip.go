// Package zip provides ZIP archive handling for ROM identification.
// It wraps the standard library zip package for random access.
package zip

import (
	"archive/zip"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
)

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

// OpenFileAt opens a file within the ZIP archive with random access support.
// Returns a RandomAccessReader that implements io.ReaderAt by buffering decompressed data.
// This is useful for format detection and header parsing without decompressing the entire file.
func (z *ZIPArchive) OpenFileAt(name string) (util.RandomAccessReader, int64, error) {
	for _, f := range z.reader.File {
		if f.Name == name {
			return NewEntryReader(f), int64(f.UncompressedSize64), nil
		}
	}
	return nil, 0, fmt.Errorf("file not found in ZIP: %s", name)
}

// Open opens a ZIP archive and returns metadata for all files.
func Open(path string) (*ZIPArchive, error) {
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
			Name: f.Name,
			Size: int64(f.UncompressedSize64),
			Hashes: core.Hashes{
				core.HashZipCRC32: fmt.Sprintf("%08x", f.CRC32),
			},
		})
	}

	return &ZIPArchive{
		reader:  r,
		entries: entries,
	}, nil
}
