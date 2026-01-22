// Package zip provides utilities for reading ZIP archives with random access support.
//
// The main functionality is NewEntryReader, which wraps a zip.File to provide
// io.ReaderAt and io.Seeker interfaces over the decompressed content.
package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"sync"
)

// EntryReader provides random access to decompressed ZIP entry content.
// It decompresses data lazily, only reading as much as needed to satisfy ReadAt requests.
// Data is buffered so subsequent reads don't re-decompress.
type EntryReader struct {
	file   *zip.File
	mu     sync.Mutex
	buffer []byte
	reader io.ReadCloser
	err    error // sticky error from decompression
	pos    int64 // current position for Seek/Read
}

// NewEntryReader creates a new EntryReader for random access to a ZIP entry.
func NewEntryReader(f *zip.File) *EntryReader {
	return &EntryReader{
		file:   f,
		buffer: make([]byte, 0, 64*1024), // pre-allocate 64KB, common for header reads
	}
}

// Size returns the uncompressed size of the ZIP entry.
func (r *EntryReader) Size() int64 {
	return int64(r.file.UncompressedSize64)
}

// Seek implements io.Seeker by tracking a position for sequential reads.
func (r *EntryReader) Seek(offset int64, whence int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = r.pos + offset
	case io.SeekEnd:
		newPos = int64(r.file.UncompressedSize64) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newPos < 0 {
		return 0, fmt.Errorf("negative position")
	}

	r.pos = newPos
	return r.pos, nil
}

// ReadAt implements io.ReaderAt by decompressing data on-demand.
// Data is buffered so subsequent reads don't re-decompress.
func (r *EntryReader) ReadAt(p []byte, off int64) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.err != nil {
		return 0, r.err
	}

	// Check if offset is beyond the file
	if off >= int64(r.file.UncompressedSize64) {
		return 0, io.EOF
	}

	// Ensure we've decompressed enough data
	needed := off + int64(len(p))
	if needed > int64(r.file.UncompressedSize64) {
		needed = int64(r.file.UncompressedSize64)
	}

	if int64(len(r.buffer)) < needed {
		if err := r.decompressTo(needed); err != nil {
			r.err = err
			return 0, err
		}
	}

	// Copy from buffer
	available := int64(len(r.buffer)) - off
	if available <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > available {
		p = p[:available]
	}
	copy(p, r.buffer[off:])
	return len(p), nil
}

// decompressTo ensures at least 'needed' bytes are decompressed into the buffer.
func (r *EntryReader) decompressTo(needed int64) error {
	// Open reader if not already open
	if r.reader == nil {
		rd, err := r.file.Open()
		if err != nil {
			return fmt.Errorf("failed to open ZIP entry: %w", err)
		}
		r.reader = rd
	}

	// Read until we have enough
	toRead := needed - int64(len(r.buffer))
	if toRead <= 0 {
		return nil
	}

	// Read in chunks for efficiency
	chunkSize := int64(64 * 1024) // 64KB chunks
	if toRead < chunkSize {
		chunkSize = toRead
	}

	buf := make([]byte, chunkSize)
	for int64(len(r.buffer)) < needed {
		n, err := r.reader.Read(buf)
		if n > 0 {
			r.buffer = append(r.buffer, buf[:n]...)
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
func (r *EntryReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.reader != nil {
		err := r.reader.Close()
		r.reader = nil
		return err
	}
	return nil
}

// Read implements io.Reader using the current position from Seek.
func (r *EntryReader) Read(p []byte) (n int, err error) {
	n, err = r.ReadAt(p, r.pos)
	r.mu.Lock()
	r.pos += int64(n)
	r.mu.Unlock()
	return n, err
}
