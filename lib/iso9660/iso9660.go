// Package iso9660 provides support for reading ISO 9660 filesystem images.
//
// This package handles ISO 9660 filesystem parsing, supporting both
// cooked (.iso) and raw (.bin) CD images by detecting the sector format
// (MODE1/2048, MODE1/2352, MODE2/2352).
//
// The API mirrors archive/zip: use NewReader to open an ISO, then access
// files via OpenFile or read raw sectors via ReadAt.
//
// ISO 9660 layout (relevant parts):
//   - Sectors 0-15: System area (platform-specific, e.g., Saturn/Dreamcast headers)
//   - Sector 16 (offset 0x8000): Primary Volume Descriptor
//   - PVD offset 156: Root directory record (34 bytes)
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

// Reader provides access to an ISO 9660 filesystem image.
// It implements io.ReaderAt for raw sector access.
type Reader struct {
	r             io.ReaderAt
	size          int64
	rootExtentLoc uint32
	rootExtentLen uint32
}

// NewReader opens an ISO 9660 image and validates the primary volume descriptor.
// Automatically detects the sector format (cooked or raw).
func NewReader(r io.ReaderAt, size int64) (*Reader, error) {
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

		return &Reader{
			r:             reader,
			size:          logicalSize,
			rootExtentLoc: rootExtentLoc,
			rootExtentLen: rootExtentLen,
		}, nil
	}

	return nil, fmt.Errorf("not a valid ISO 9660: no CD001 magic found")
}

// ReadAt implements io.ReaderAt, reading from the logical (2048-byte sector) view.
// This allows direct access to any part of the ISO, including the system area
// at offset 0 (used for Saturn/Dreamcast identification).
func (r *Reader) ReadAt(p []byte, off int64) (n int, err error) {
	return r.r.ReadAt(p, off)
}

// Size returns the logical size of the ISO image in bytes.
func (r *Reader) Size() int64 {
	return r.size
}

// OpenFile opens a file by path (case-insensitive) and returns a reader for its contents.
// Supports subdirectory paths like "PSP_GAME/PARAM.SFO".
// Handles ISO 9660 version suffixes (e.g., ";1").
func (r *Reader) OpenFile(path string) (io.ReaderAt, int64, error) {
	// Split path into components
	parts := strings.Split(path, "/")

	// Start from root directory
	dirExtentLoc := r.rootExtentLoc
	dirExtentLen := r.rootExtentLen

	// Traverse directories
	for i, part := range parts {
		isLast := i == len(parts)-1

		extentLoc, extentLen, isDir, err := r.findEntry(dirExtentLoc, dirExtentLen, part)
		if err != nil {
			return nil, 0, fmt.Errorf("path component %q not found: %w", part, err)
		}

		if isLast {
			// Final component - return a reader for the file
			if isDir {
				return nil, 0, fmt.Errorf("%q is a directory, not a file", part)
			}
			fileOffset := int64(extentLoc) * sectorSize2048
			fileSize := int64(extentLen)
			return io.NewSectionReader(r.r, fileOffset, fileSize), fileSize, nil
		}

		// Intermediate component - must be a directory
		if !isDir {
			return nil, 0, fmt.Errorf("%q is not a directory", part)
		}
		dirExtentLoc = extentLoc
		dirExtentLen = extentLen
	}

	return nil, 0, fmt.Errorf("empty path")
}

// findEntry searches a directory for an entry by name.
// Returns the entry's extent location, size, whether it's a directory, and any error.
func (r *Reader) findEntry(dirExtentLoc, dirExtentLen uint32, name string) (uint32, uint32, bool, error) {
	// Read directory
	dirData := make([]byte, dirExtentLen)
	if _, err := r.r.ReadAt(dirData, int64(dirExtentLoc)*sectorSize2048); err != nil {
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
