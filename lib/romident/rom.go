package romident

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sargunv/rom-tools/lib/romident/chd"
	"github.com/sargunv/rom-tools/lib/romident/core"
	"github.com/sargunv/rom-tools/lib/romident/folder"
	"github.com/sargunv/rom-tools/lib/romident/zip"
)

// IdentifyROM identifies a ROM file, ZIP archive, or folder.
// Returns a ROM struct with file information and hashes.
func IdentifyROM(path string, opts Options) (*ROM, error) {
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

// identifyContainer processes any Container (folder, ZIP, etc.) and identifies all files within.
func identifyContainer(c core.Container, containerType ROMType, containerPath string, opts Options) (*ROM, error) {
	entries := c.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("container is empty")
	}

	files := make(Files)
	detector := NewDetector()
	var romIdent *GameIdent

	for _, entry := range entries {
		// Open file for identification
		reader, err := c.OpenFileAt(entry.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", entry.Name, err)
		}

		romFile, fileIdent, err := identifySingleReader(reader, entry.Name, detector, opts)
		if err != nil {
			reader.Close()
			return nil, fmt.Errorf("failed to identify %s: %w", entry.Name, err)
		}

		// Add pre-computed CRC32 from container metadata if available and not already calculated
		if entry.CRC32 != 0 {
			hasCRC32 := false
			for _, h := range romFile.Hashes {
				if h.Algorithm == HashCRC32 {
					hasCRC32 = true
					break
				}
			}
			if !hasCRC32 {
				romFile.Hashes = append(romFile.Hashes, NewHash(HashCRC32, fmt.Sprintf("%08x", entry.CRC32), HashSourceZIPMetadata))
			}
		}

		// Collect identification (error if multiple conflicting identifications found)
		if fileIdent != nil {
			if romIdent != nil && !romIdent.Equal(fileIdent) {
				reader.Close()
				return nil, fmt.Errorf("container has multiple game identifications: %s and %s", romIdent.TitleID, fileIdent.TitleID)
			}
			romIdent = fileIdent
		}

		reader.Close()
		files[entry.Name] = *romFile
	}

	return &ROM{
		Path:  containerPath,
		Type:  containerType,
		Files: files,
		Ident: romIdent,
	}, nil
}

func identifyFolder(path string, opts Options) (*ROM, error) {
	c, err := folder.NewFolderContainer(path)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return identifyContainer(c, ROMTypeFolder, path, opts)
}

func identifyFile(path string, size int64, opts Options) (*ROM, error) {
	// Open file for format detection
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Detect format
	detector := NewDetector()
	detectedFormat, err := detector.Detect(f, size, filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("failed to detect format: %w", err)
	}

	// Handle ZIP specially
	if detectedFormat == FormatZIP {
		return identifyZIP(path, opts)
	}

	// Single file
	romFile, ident, err := identifySingleFile(path, size, detector, opts)
	if err != nil {
		return nil, err
	}

	files := Files{
		filepath.Base(path): *romFile,
	}

	return &ROM{
		Path:  path,
		Type:  ROMTypeFile,
		Files: files,
		Ident: ident,
	}, nil
}

func identifyZIP(path string, opts Options) (*ROM, error) {
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

	// Slow mode: use full container introspection (decompresses files)
	if opts.HashMode == HashModeSlow {
		return identifyContainer(archive, ROMTypeZIP, path, opts)
	}

	// Fast/default mode: use ZIP metadata only (no decompression)
	files := make(Files)
	detector := NewDetector()

	for _, entry := range entries {
		// Use extension-based format detection (no decompression)
		// In fast mode, we can't verify with magic, so only trust unambiguous extensions
		candidates := detector.CandidatesByExtension(entry.Name)
		detectedFormat := FormatUnknown
		if len(candidates) == 1 {
			detectedFormat = candidates[0]
		}

		hashes := []Hash{}
		if entry.CRC32 != 0 {
			hashes = []Hash{
				NewHash(HashCRC32, fmt.Sprintf("%08x", entry.CRC32), HashSourceZIPMetadata),
			}
		}

		files[entry.Name] = ROMFile{
			Size:   entry.Size,
			Format: detectedFormat,
			Hashes: hashes,
		}
	}

	return &ROM{
		Path:  path,
		Type:  ROMTypeZIP,
		Files: files,
		Ident: nil, // No identification in fast/default mode
	}, nil
}

// identifySingleReader identifies a file from a ReaderAtSeekCloser (works for any container).
// Returns the ROMFile, game identification (if any), and an error.
func identifySingleReader(r core.ReaderAtSeekCloser, name string, detector *Detector, opts Options) (*ROMFile, *GameIdent, error) {
	size := r.Size()

	// Detect format and identify game using registry
	detectedFormat, ident := identifyGameFromRegistry(r, size, name)

	romFile := &ROMFile{
		Size:   size,
		Format: detectedFormat,
	}

	// For CHD, always extract hashes from header (fast, no decompression)
	if detectedFormat == FormatCHD {
		chdInfo, err := chd.ParseCHDHeader(r)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse CHD header: %w", err)
		}
		romFile.Hashes = []Hash{
			NewHash(HashSHA1, chdInfo.RawSHA1, HashSourceCHDRaw),
			NewHash(HashSHA1, chdInfo.SHA1, HashSourceCHDCompressed),
		}
		return romFile, ident, nil
	}

	// Fast mode: skip calculating hashes for large files, but allow small files
	if opts.HashMode == HashModeFast && size >= FastModeSmallFileThreshold {
		return romFile, ident, nil
	}

	// For other formats, calculate hashes
	// Reset reader position
	if _, err := r.Seek(0, 0); err != nil {
		return nil, nil, fmt.Errorf("failed to seek: %w", err)
	}

	// Wrap ReaderAtSeekCloser as io.Reader for CalculateHashes
	reader := &readerAtWrapper{r: r}
	hashes, err := CalculateHashes(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate hashes: %w", err)
	}

	romFile.Hashes = hashes

	return romFile, ident, nil
}

// identifyGameFromRegistry uses the format registry to detect format and identify game.
// Returns the detected format and game identification (if any).
func identifyGameFromRegistry(r core.ReaderAtSeekCloser, size int64, name string) (Format, *GameIdent) {
	// Get candidate formats by extension
	entries := FormatsByExtension(name)
	if len(entries) == 0 {
		return FormatUnknown, nil
	}

	// Try each candidate's identifier
	for _, entry := range entries {
		// Reset reader position
		if _, err := r.Seek(0, 0); err != nil {
			continue
		}

		// If no identify function, just verify format using detector
		if entry.Identify == nil {
			detector := NewDetector()
			if detector.VerifyFormat(r, size, entry.Format) {
				return entry.Format, nil
			}
			continue
		}

		// Try to identify using the entry's function
		ident, err := entry.Identify(r, size)
		if err == nil {
			return entry.Format, ident
		}
		// If identification fails, try next candidate
	}

	return FormatUnknown, nil
}

func identifySingleFile(path string, size int64, detector *Detector, opts Options) (*ROMFile, *GameIdent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileReader := &containerFileReader{file: f, size: info.Size()}
	return identifySingleReader(fileReader, filepath.Base(path), detector, opts)
}

// containerFileReader wraps *os.File to implement ReaderAtSeekCloser.
type containerFileReader struct {
	file *os.File
	size int64
}

func (f *containerFileReader) ReadAt(p []byte, off int64) (n int, err error) {
	return f.file.ReadAt(p, off)
}

func (f *containerFileReader) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

func (f *containerFileReader) Size() int64 {
	return f.size
}

func (f *containerFileReader) Close() error {
	return f.file.Close()
}

// readerAtWrapper wraps ReaderAtSeekCloser to implement io.Reader.
type readerAtWrapper struct {
	r   core.ReaderAtSeekCloser
	pos int64
}

func (w *readerAtWrapper) Read(p []byte) (n int, err error) {
	n, err = w.r.ReadAt(p, w.pos)
	w.pos += int64(n)
	return n, err
}
