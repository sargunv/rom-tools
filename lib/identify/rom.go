package identify

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sargunv/rom-tools/internal/container/folder"
	"github.com/sargunv/rom-tools/internal/container/zip"
	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/chd"
)

// Identify identifies a ROM file, ZIP archive, or folder.
// Returns a Result with identified items and their hashes.
func Identify(path string, opts Options) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return identifyFolder(absPath, opts)
	}

	return identifyFile(absPath, info.Size(), opts)
}

// identifyFile handles a single file (may be a container like ZIP).
func identifyFile(path string, size int64, opts Options) (*Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Check if it's a ZIP file (special handling required)
	if isZIP(f, size) {
		return identifyZIP(path, opts)
	}

	// Single file - identify it
	item, err := identifyReader(f, size, filepath.Base(path), opts)
	if err != nil {
		return nil, err
	}

	return &Result{
		Path:  path,
		Items: []Item{*item},
	}, nil
}

// identifyFolder handles a directory containing ROM files.
func identifyFolder(path string, opts Options) (*Result, error) {
	c, err := folder.NewFolderContainer(path)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	entries := c.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("folder is empty")
	}

	items := make([]Item, 0, len(entries))

	for _, entry := range entries {
		reader, size, err := c.OpenFileAt(entry.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", entry.Name, err)
		}

		item, err := identifyReader(reader, size, entry.Name, opts)
		reader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to identify %s: %w", entry.Name, err)
		}

		items = append(items, *item)
	}

	return &Result{
		Path:  path,
		Items: items,
	}, nil
}

// identifyZIP handles a ZIP archive.
func identifyZIP(path string, opts Options) (*Result, error) {
	handler := zip.NewZIPHandler()

	archive, err := handler.Open(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	entries := archive.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("ZIP archive is empty")
	}

	items := make([]Item, 0, len(entries))

	for _, entry := range entries {
		if opts.HashMode == HashModeSlow {
			// Slow mode: decompress and fully identify
			reader, size, err := archive.OpenFileAt(entry.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to open %s: %w", entry.Name, err)
			}

			item, err := identifyReader(reader, size, entry.Name, opts)
			reader.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to identify %s: %w", entry.Name, err)
			}

			items = append(items, *item)
		} else {
			// Fast/default mode: use ZIP metadata only (no decompression)
			candidates := candidatesByExtension(entry.Name)
			detectedFormat := FormatUnknown
			if len(candidates) == 1 {
				detectedFormat = candidates[0]
			}

			hashes := make(Hashes)
			if entry.CRC32 != 0 {
				hashes[HashZipCRC32] = fmt.Sprintf("%08x", entry.CRC32)
			}

			items = append(items, Item{
				Name:   entry.Name,
				Size:   entry.Size,
				Format: detectedFormat,
				Hashes: hashes,
				Game:   nil, // No identification in fast mode
			})
		}
	}

	return &Result{
		Path:  path,
		Items: items,
	}, nil
}

// identifyReader identifies a single file from a reader.
// Returns an Item with format, hashes, and game info.
// Accepts either *os.File or util.RandomAccessReader.
func identifyReader(r util.RandomAccessReader, size int64, name string, opts Options) (*Item, error) {
	// Try to identify format and game in one pass
	format, game := identifyGame(r, size, name)

	item := &Item{
		Name:   name,
		Size:   size,
		Format: format,
		Game:   game,
	}

	// Handle hashes based on format
	if format == FormatCHD {
		// CHD: extract hashes from header (fast, no decompression needed)
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to seek: %w", err)
		}
		chdReader, err := chd.NewReader(r, size)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CHD header: %w", err)
		}
		header := chdReader.Header()
		item.Hashes = Hashes{
			HashCHDUncompressedSHA1: header.RawSHA1,
			HashCHDCompressedSHA1:   header.SHA1,
		}
		return item, nil
	}

	// Fast mode: skip hashes for large files
	if opts.HashMode == HashModeFast && size >= fastModeSmallFileThreshold {
		return item, nil
	}

	// Calculate hashes - wrap reader to provide io.Reader interface
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek: %w", err)
	}

	hashes, err := calculateHashes(&readerAtWrapper{r: r})
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hashes: %w", err)
	}

	item.Hashes = hashes
	return item, nil
}

// readerAtWrapper wraps a ReaderAt+Seeker to implement io.Reader.
type readerAtWrapper struct {
	r   io.ReaderAt
	pos int64
}

func (w *readerAtWrapper) Read(p []byte) (n int, err error) {
	n, err = w.r.ReadAt(p, w.pos)
	w.pos += int64(n)
	return n, err
}

// identifyGame tries to identify the format and game from a reader.
// Returns the detected format and game info (nil if not identifiable).
func identifyGame(r util.RandomAccessReader, size int64, name string) (Format, GameInfo) {
	// Get candidate formats by extension
	entries := formatsByExtension(name)
	if len(entries) == 0 {
		return FormatUnknown, nil
	}

	// Try each candidate's identifier
	for _, entry := range entries {
		// Reset reader position
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			continue
		}

		// If no identify function, just verify format using magic bytes
		if entry.Identify == nil {
			if verifyFormat(r, size, entry.Format) {
				return entry.Format, nil
			}
			continue
		}

		// Try to identify using the entry's function
		game, err := entry.Identify(r, size)
		if err == nil {
			return entry.Format, game
		}
		// If identification fails, try next candidate
	}

	return FormatUnknown, nil
}

// isZIP checks if a file is a ZIP archive by checking magic bytes.
func isZIP(r io.ReaderAt, size int64) bool {
	return checkMagic(r, size, zipOffset, zipMagic)
}
