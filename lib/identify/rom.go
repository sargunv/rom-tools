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
		item, err := identifyContainerEntry(c, entry, opts)
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

// identifyContainerEntry identifies a single entry within a container.
// If the entry has pre-computed hashes, those are used (never calculated).
func identifyContainerEntry(c util.FileContainer, entry util.FileEntry, opts Options) (*Item, error) {
	item := &Item{
		Name: entry.Name,
		Size: entry.Size,
	}

	// Use pre-computed hashes from container metadata if available
	if entry.Hashes != nil {
		item.Hashes = entry.Hashes
	}

	// Open and identify the file
	reader, size, err := c.OpenFileAt(entry.Name)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Identify the game
	item.Game = identifyGame(reader, size, entry.Name)

	// If no pre-computed hashes, calculate them (respecting MaxHashSize)
	if entry.Hashes == nil {
		// Handle CHD: extract hashes from the parsed info
		if chdInfo, ok := item.Game.(*chd.Info); ok {
			item.Hashes = Hashes{
				HashCHDUncompressedSHA1: chdInfo.RawSHA1,
				HashCHDCompressedSHA1:   chdInfo.SHA1,
			}
		} else if opts.MaxHashSize < 0 || size <= opts.MaxHashSize {
			// Calculate hashes if within size limit
			hashes, err := calculateHashes(reader, size)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate hashes: %w", err)
			}
			item.Hashes = hashes
		}
	}

	return item, nil
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

	// Skip hashes for files exceeding MaxHashSize (-1 = no limit)
	if opts.MaxHashSize >= 0 && size > opts.MaxHashSize {
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
