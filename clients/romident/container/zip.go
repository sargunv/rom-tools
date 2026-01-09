// Package container provides handlers for container formats (ZIP, folders).
package container

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// ZIPHandler handles ZIP archive files.
type ZIPHandler struct{}

// NewZIPHandler creates a new ZIP handler.
func NewZIPHandler() *ZIPHandler {
	return &ZIPHandler{}
}

// ZIPFileInfo contains information about a file within a ZIP archive.
type ZIPFileInfo struct {
	Name  string
	Size  int64
	CRC32 uint32
}

// ZIPArchive represents an open ZIP archive.
type ZIPArchive struct {
	reader *zip.ReadCloser
	Files  []ZIPFileInfo
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
// Returns a ZIPEntryReader that implements io.ReaderAt by buffering decompressed data.
// This is useful for format detection and header parsing without decompressing the entire file.
func (z *ZIPArchive) OpenFileAt(name string) (*ZIPEntryReader, error) {
	for _, f := range z.reader.File {
		if f.Name == name {
			return newZIPEntryReader(f)
		}
	}
	return nil, fmt.Errorf("file not found in ZIP: %s", name)
}

// ZIPEntryReader provides io.ReaderAt access to a ZIP entry by buffering decompressed data.
// It decompresses data lazily, only reading as much as needed to satisfy ReadAt requests.
type ZIPEntryReader struct {
	file   *zip.File
	mu     sync.Mutex
	buffer []byte
	reader io.ReadCloser
	err    error // sticky error from decompression
	pos    int64 // current position for Seek/Read
}

func newZIPEntryReader(f *zip.File) (*ZIPEntryReader, error) {
	return &ZIPEntryReader{
		file:   f,
		buffer: make([]byte, 0, 64*1024), // pre-allocate 64KB, common for header reads
	}, nil
}

// Size returns the uncompressed size of the ZIP entry.
func (z *ZIPEntryReader) Size() int64 {
	return int64(z.file.UncompressedSize64)
}

// Seek implements io.Seeker by tracking a position for sequential reads.
func (z *ZIPEntryReader) Seek(offset int64, whence int) (int64, error) {
	z.mu.Lock()
	defer z.mu.Unlock()

	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = z.pos + offset
	case io.SeekEnd:
		newPos = int64(z.file.UncompressedSize64) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newPos < 0 {
		return 0, fmt.Errorf("negative position")
	}

	z.pos = newPos
	return z.pos, nil
}

// ReadAt implements io.ReaderAt by decompressing data on-demand.
// Data is buffered so subsequent reads don't re-decompress.
func (z *ZIPEntryReader) ReadAt(p []byte, off int64) (n int, err error) {
	z.mu.Lock()
	defer z.mu.Unlock()

	if z.err != nil {
		return 0, z.err
	}

	// Check if offset is beyond the file
	if off >= int64(z.file.UncompressedSize64) {
		return 0, io.EOF
	}

	// Ensure we've decompressed enough data
	needed := off + int64(len(p))
	if needed > int64(z.file.UncompressedSize64) {
		needed = int64(z.file.UncompressedSize64)
	}

	if int64(len(z.buffer)) < needed {
		if err := z.decompressTo(needed); err != nil {
			z.err = err
			return 0, err
		}
	}

	// Copy from buffer
	available := int64(len(z.buffer)) - off
	if available <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > available {
		p = p[:available]
	}
	copy(p, z.buffer[off:])
	return len(p), nil
}

// decompressTo ensures at least 'needed' bytes are decompressed into the buffer.
func (z *ZIPEntryReader) decompressTo(needed int64) error {
	// Open reader if not already open
	if z.reader == nil {
		r, err := z.file.Open()
		if err != nil {
			return fmt.Errorf("failed to open ZIP entry: %w", err)
		}
		z.reader = r
	}

	// Read until we have enough
	toRead := needed - int64(len(z.buffer))
	if toRead <= 0 {
		return nil
	}

	// Read in chunks for efficiency
	chunkSize := int64(64 * 1024) // 64KB chunks
	if toRead < chunkSize {
		chunkSize = toRead
	}

	buf := make([]byte, chunkSize)
	for int64(len(z.buffer)) < needed {
		n, err := z.reader.Read(buf)
		if n > 0 {
			z.buffer = append(z.buffer, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to decompress ZIP entry: %w", err)
		}
	}

	return nil
}

// Close releases resources associated with the reader.
func (z *ZIPEntryReader) Close() error {
	z.mu.Lock()
	defer z.mu.Unlock()

	if z.reader != nil {
		err := z.reader.Close()
		z.reader = nil
		return err
	}
	return nil
}

// Open opens a ZIP archive and returns metadata for all files.
func (h *ZIPHandler) Open(path string) (*ZIPArchive, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}

	var files []ZIPFileInfo
	for _, f := range r.File {
		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}

		files = append(files, ZIPFileInfo{
			Name:  f.Name,
			Size:  int64(f.UncompressedSize64),
			CRC32: f.CRC32,
		})
	}

	return &ZIPArchive{
		reader: r,
		Files:  files,
	}, nil
}

// FolderHandler handles directory-based ROMs.
type FolderHandler struct{}

// NewFolderHandler creates a new folder handler.
func NewFolderHandler() *FolderHandler {
	return &FolderHandler{}
}

// ListFiles lists all files in a folder.
func (h *FolderHandler) ListFiles(path string) ([]string, error) {
	var files []string

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(path, p)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list folder: %w", err)
	}

	return files, nil
}
