package identify

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sargunv/rom-tools/internal/container/folder"
	"github.com/sargunv/rom-tools/internal/container/zip"
	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/chd"
	"github.com/sargunv/rom-tools/lib/core"
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
		container, err := folder.NewFolderContainer(absPath)
		if err != nil {
			return nil, err
		}
		defer container.Close()
		return identifyContainer(absPath, container, opts)
	}

	return identifyFile(absPath, info.Size(), opts)
}

// identifyFile handles a single file (may be a container like ZIP).
func identifyFile(path string, size int64, opts Options) (*Result, error) {
	ext := strings.ToLower(filepath.Ext(path))

	// ZIP files are containers - identify their contents
	if ext == ".zip" {
		container, err := zip.Open(path)
		if err != nil {
			return nil, err
		}
		defer container.Close()
		return identifyContainer(path, container, opts)
	}

	// Single file - open and identify it
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	item, err := identifyReader(f, size, filepath.Base(path), opts)
	if err != nil {
		return nil, err
	}

	return &Result{
		Path:  path,
		Items: []Item{*item},
	}, nil
}

// identifyContainer handles any container (ZIP, folder, etc.) using the FileContainer interface.
func identifyContainer(path string, c util.FileContainer, opts Options) (*Result, error) {
	entries := c.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("container is empty")
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

// identifyReader identifies a single file from a reader.
// Returns an Item with hashes and game info.
func identifyReader(r util.RandomAccessReader, size int64, name string, opts Options) (*Item, error) {
	// Try to identify game
	game := identifyGame(r, size, name)

	item := &Item{
		Name: name,
		Size: size,
		Game: game,
	}

	// Handle CHD: extract hashes from the parsed info
	if chdInfo, ok := game.(*chd.Info); ok {
		item.Hashes = Hashes{
			HashCHDUncompressedSHA1: chdInfo.RawSHA1,
			HashCHDCompressedSHA1:   chdInfo.SHA1,
		}
		return item, nil
	}

	// Fast mode: skip hashes for large files
	if opts.HashMode == HashModeFast && size >= fastModeSmallFileThreshold {
		return item, nil
	}

	// Calculate hashes
	hashes, err := calculateHashes(r, size)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hashes: %w", err)
	}

	item.Hashes = hashes
	return item, nil
}

// identifyGame tries to identify the game from a reader.
// Returns the game info (nil if not identifiable).
func identifyGame(r io.ReaderAt, size int64, name string) core.GameInfo {
	// Get candidate parsers by extension
	parsers := identifyByExtension(name)
	if len(parsers) == 0 {
		return nil
	}

	// Try each parser
	// TODO: log parser errors at debug level when logging is available
	for _, parser := range parsers {
		game, err := parser(r, size)
		if err == nil && game != nil {
			return game
		}
	}

	return nil
}
