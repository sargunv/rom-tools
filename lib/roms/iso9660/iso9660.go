// Package iso9660 provides support for reading ISO 9660 filesystem images.
//
// This package handles ISO 9660 filesystem parsing, supporting both
// cooked (.iso) and raw (.bin) CD images by detecting the sector format
// (MODE1/2048, MODE1/2352, MODE2/2352).
//
// ISO 9660 layout (relevant parts):
//   - Sector 16 (offset 0x8000): Primary Volume Descriptor
//   - PVD offset 156: Root directory record (34 bytes)
//   - Root directory record contains extent location and size
package iso9660

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

const (
	pvdMagicOffset    = 1
	pvdRootDirOffset  = 156
	dirEntryExtentLoc = 2  // Offset within directory entry
	dirEntryDataLen   = 10 // Offset within directory entry
	dirEntryFlags     = 25 // Offset within directory entry (bit 1 = directory)
	dirEntryNameLen   = 32 // Offset within directory entry
	dirEntryName      = 33 // Offset within directory entry

	flagDirectory = 0x02 // Directory flag in file flags byte
)

// Image represents an ISO 9660 image.
type Image struct {
	r             io.ReaderAt
	size          int64
	rootExtentLoc uint32
	rootExtentLen uint32
}

// Open opens an ISO 9660 image and validates the primary volume descriptor.
// Automatically detects the sector format (cooked or raw).
func Open(r io.ReaderAt, size int64) (*Image, error) {
	// Try each sector format to find the ISO9660 PVD
	for _, format := range sectorFormats {
		// Check if file is large enough for this format
		magicOffset := format.pvdOffset + pvdMagicOffset
		if size < magicOffset+5 {
			continue
		}

		// Check for "CD001" magic
		magic := make([]byte, 5)
		if _, err := r.ReadAt(magic, magicOffset); err != nil {
			continue
		}
		if string(magic) != "CD001" {
			continue
		}

		// Found ISO9660! Create appropriate reader
		var reader io.ReaderAt = r
		var logicalSize int64 = size

		// For raw formats, wrap in a sector reader
		if format.sectorSize != sectorSize2048 {
			sr := newSectorReader(r, format, size)
			reader = sr
			logicalSize = sr.Size()
		}

		// Read Primary Volume Descriptor at logical sector 16
		pvdOffset := int64(16 * sectorSize2048)
		pvd := make([]byte, sectorSize2048)
		if _, err := reader.ReadAt(pvd, pvdOffset); err != nil {
			return nil, fmt.Errorf("failed to read PVD: %w", err)
		}

		// Extract root directory record info
		rootRecord := pvd[pvdRootDirOffset:]
		rootExtentLoc := binary.LittleEndian.Uint32(rootRecord[dirEntryExtentLoc:])
		rootExtentLen := binary.LittleEndian.Uint32(rootRecord[dirEntryDataLen:])

		return &Image{
			r:             reader,
			size:          logicalSize,
			rootExtentLoc: rootExtentLoc,
			rootExtentLen: rootExtentLen,
		}, nil
	}

	return nil, fmt.Errorf("not a valid ISO 9660: no CD001 magic found")
}

// ReadSystemArea reads the ISO 9660 system area (sectors 0-15).
// This area is reserved for system use and may contain platform-specific headers.
func (img *Image) ReadSystemArea() ([]byte, error) {
	// System area is sectors 0-15 (16 sectors * 2048 bytes = 32KB)
	// We only need the first sector for Saturn identification
	if img.size < sectorSize2048 {
		return nil, fmt.Errorf("file too small for system area")
	}

	data := make([]byte, sectorSize2048)
	if _, err := img.r.ReadAt(data, 0); err != nil {
		return nil, fmt.Errorf("failed to read system area: %w", err)
	}
	return data, nil
}

// ReadFile reads a file by path (case-insensitive).
// Supports subdirectory paths like "PSP_GAME/PARAM.SFO".
// Handles ISO 9660 version suffixes (e.g., ";1").
func (img *Image) ReadFile(path string) ([]byte, error) {
	// Split path into components
	parts := strings.Split(path, "/")

	// Start from root directory
	dirExtentLoc := img.rootExtentLoc
	dirExtentLen := img.rootExtentLen

	// Traverse directories
	for i, part := range parts {
		isLast := i == len(parts)-1

		extentLoc, extentLen, isDir, err := img.findEntry(dirExtentLoc, dirExtentLen, part)
		if err != nil {
			return nil, fmt.Errorf("path component %q not found: %w", part, err)
		}

		if isLast {
			// Final component - read the file
			if isDir {
				return nil, fmt.Errorf("%q is a directory, not a file", part)
			}
			data := make([]byte, extentLen)
			if _, err := img.r.ReadAt(data, int64(extentLoc)*sectorSize2048); err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			return data, nil
		}

		// Intermediate component - must be a directory
		if !isDir {
			return nil, fmt.Errorf("%q is not a directory", part)
		}
		dirExtentLoc = extentLoc
		dirExtentLen = extentLen
	}

	return nil, fmt.Errorf("empty path")
}

// findEntry searches a directory for an entry by name.
// Returns the entry's extent location, size, whether it's a directory, and any error.
func (img *Image) findEntry(dirExtentLoc, dirExtentLen uint32, name string) (uint32, uint32, bool, error) {
	// Read directory
	dirData := make([]byte, dirExtentLen)
	if _, err := img.r.ReadAt(dirData, int64(dirExtentLoc)*sectorSize2048); err != nil {
		return 0, 0, false, fmt.Errorf("failed to read directory: %w", err)
	}

	name = strings.ToUpper(name)
	offset := 0
	for offset < len(dirData) {
		entryLen := int(dirData[offset])
		if entryLen == 0 {
			// End of directory entries in this sector, try next sector
			nextSector := ((offset / sectorSize2048) + 1) * sectorSize2048
			if nextSector >= len(dirData) {
				break
			}
			offset = nextSector
			continue
		}

		if offset+dirEntryName >= len(dirData) {
			break
		}

		nameLen := int(dirData[offset+dirEntryNameLen])
		if offset+dirEntryName+nameLen > len(dirData) {
			break
		}

		entryName := strings.ToUpper(string(dirData[offset+dirEntryName : offset+dirEntryName+nameLen]))

		// Strip version suffix (";1")
		if idx := strings.Index(entryName, ";"); idx != -1 {
			entryName = entryName[:idx]
		}

		if entryName == name {
			extentLoc := binary.LittleEndian.Uint32(dirData[offset+dirEntryExtentLoc:])
			extentLen := binary.LittleEndian.Uint32(dirData[offset+dirEntryDataLen:])
			flags := dirData[offset+dirEntryFlags]
			isDir := (flags & flagDirectory) != 0
			return extentLoc, extentLen, isDir, nil
		}

		offset += entryLen
	}

	return 0, 0, false, fmt.Errorf("entry not found: %s", name)
}
