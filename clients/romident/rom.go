package romident

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sargunv/rom-tools/clients/romident/container"
	"github.com/sargunv/rom-tools/clients/romident/format"
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

func identifyFolder(path string, opts Options) (*ROM, error) {
	handler := container.NewFolderHandler()

	fileList, err := handler.ListFiles(path)
	if err != nil {
		return nil, err
	}

	if len(fileList) == 0 {
		return nil, fmt.Errorf("folder is empty")
	}

	files := make(Files)
	detector := format.NewDetector()

	for _, relPath := range fileList {
		fullPath := filepath.Join(path, relPath)

		info, err := os.Stat(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", relPath, err)
		}

		romFile, err := identifySingleFile(fullPath, info.Size(), detector, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to identify %s: %w", relPath, err)
		}

		files[relPath] = *romFile
	}

	return &ROM{
		Path:  path,
		Type:  ROMTypeFolder,
		Files: files,
		Ident: nil, // TODO: identification
	}, nil
}

func identifyFile(path string, size int64, opts Options) (*ROM, error) {
	// Open file for format detection
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Detect format
	detector := format.NewDetector()
	detectedFormat, err := detector.Detect(f, size, filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("failed to detect format: %w", err)
	}

	// Handle ZIP specially
	if detectedFormat == format.ZIP {
		return identifyZIP(path, opts)
	}

	// Single file
	romFile, err := identifySingleFile(path, size, detector, opts)
	if err != nil {
		return nil, err
	}

	files := Files{
		filepath.Base(path): *romFile,
	}

	// Try to get game identification
	var ident *GameIdent
	if detectedFormat == format.XISO {
		ident = identifyXISO(f, size)
	}

	return &ROM{
		Path:  path,
		Type:  ROMTypeFile,
		Files: files,
		Ident: ident,
	}, nil
}

func identifyZIP(path string, opts Options) (*ROM, error) {
	handler := container.NewZIPHandler()

	archive, err := handler.Open(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	if len(archive.Files) == 0 {
		return nil, fmt.Errorf("ZIP archive is empty")
	}

	files := make(Files)
	detector := format.NewDetector()

	for _, zipFile := range archive.Files {
		// Always extract CRC32 from ZIP metadata (fast, no decompression)
		hashes := []Hash{
			NewHash(HashCRC32, fmt.Sprintf("%08x", zipFile.CRC32), "zip-metadata"),
		}

		// Slow mode: also decompress and calculate full hashes
		if opts.HashMode == HashModeSlow {
			reader, err := archive.OpenFile(zipFile.Name)
			if err == nil {
				if calculated, err := CalculateHashes(reader); err == nil {
					// Replace CRC32 with calculated one (should match, but calculated is authoritative)
					hashes = calculated
				}
				reader.Close()
			}
		}

		files[zipFile.Name] = ROMFile{
			Size:   zipFile.Size,
			Format: formatToRomidentFormat(detector.DetectByExtension(zipFile.Name)),
			Hashes: hashes,
		}
	}

	return &ROM{
		Path:  path,
		Type:  ROMTypeZIP,
		Files: files,
		Ident: nil, // TODO: identification
	}, nil
}

func identifySingleFile(path string, size int64, detector *format.Detector, opts Options) (*ROMFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Detect format
	detectedFormat, err := detector.Detect(f, size, filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("failed to detect format: %w", err)
	}

	romFile := &ROMFile{
		Size:   size,
		Format: formatToRomidentFormat(detectedFormat),
	}

	// For CHD, always extract hashes from header (fast, no decompression)
	if detectedFormat == format.CHD {
		chdInfo, err := format.ParseCHDHeader(f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CHD header: %w", err)
		}
		romFile.Hashes = []Hash{
			NewHash(HashSHA1, chdInfo.RawSHA1, "chd-raw"),
			NewHash(HashSHA1, chdInfo.SHA1, "chd-compressed"),
		}
		return romFile, nil
	}

	// Fast mode: skip calculating hashes for loose files
	if opts.HashMode == HashModeFast {
		return romFile, nil
	}

	// For other formats, calculate hashes
	// Reset file position
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek: %w", err)
	}

	hashes, err := CalculateHashes(f)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hashes: %w", err)
	}

	romFile.Hashes = hashes

	return romFile, nil
}

// identifyXISO extracts game identification from an Xbox XISO.
// Returns nil if identification fails (non-fatal).
func identifyXISO(r io.ReaderAt, size int64) *GameIdent {
	info, err := format.ParseXISO(r, size)
	if err != nil {
		return nil
	}

	return &GameIdent{
		Platform: "xbox",
		TitleID:  fmt.Sprintf("%s-%03d", info.PublisherCode, info.GameNumber),
		Title:    info.Title,
		Region:   info.Region,
		Extra: map[string]string{
			"title_id_hex": info.TitleIDHex,
			"disc_number":  fmt.Sprintf("%d", info.DiscNumber),
			"version":      fmt.Sprintf("%d", info.Version),
		},
	}
}

// formatToRomidentFormat converts format.Format to romident.Format
func formatToRomidentFormat(f format.Format) Format {
	switch f {
	case format.CHD:
		return FormatCHD
	case format.XISO:
		return FormatXISO
	case format.ISO9660:
		return FormatISO9660
	case format.ZIP:
		return FormatZIP
	default:
		return FormatUnknown
	}
}
