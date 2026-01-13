package iso9660

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/sargunv/rom-tools/lib/romident/cnf"
	"github.com/sargunv/rom-tools/lib/romident/core"
)

// ISO 9660 disc identification with platform dispatch.
//
// This package handles ISO 9660 filesystem parsing and dispatches to
// platform-specific handlers based on the disc contents:
//   - SYSTEM.CNF with BOOT2 → PS2
//   - SYSTEM.CNF with BOOT → PS1
//
// ISO 9660 layout (relevant parts):
//   - Sector 16 (offset 0x8000): Primary Volume Descriptor
//   - PVD offset 156: Root directory record (34 bytes)
//   - Root directory record contains extent location and size

const (
	sectorSize        = 2048
	pvdOffset         = 16 * sectorSize // Sector 16
	pvdMagicOffset    = 1
	pvdRootDirOffset  = 156
	dirEntryExtentLoc = 2  // Offset within directory entry
	dirEntryDataLen   = 10 // Offset within directory entry
	dirEntryNameLen   = 32 // Offset within directory entry
	dirEntryName      = 33 // Offset within directory entry
)

// Identify parses an ISO 9660 image and attempts to identify the platform.
// Returns nil (not an error) if the ISO is valid but the platform is unknown.
func Identify(r io.ReaderAt, size int64) (*core.GameIdent, error) {
	img, err := openImage(r, size)
	if err != nil {
		return nil, err
	}

	// Try to read SYSTEM.CNF (PS1/PS2 discs)
	if data, err := img.readFile("SYSTEM.CNF"); err == nil {
		if ident := cnf.IdentifyFromSystemCNF(data); ident != nil {
			return ident, nil
		}
	}

	// No platform identified - valid ISO but unknown content
	return nil, nil
}

// image represents a minimal ISO 9660 image reader.
type image struct {
	r             io.ReaderAt
	rootExtentLoc uint32
	rootExtentLen uint32
}

// openImage opens an ISO 9660 image and validates the primary volume descriptor.
func openImage(r io.ReaderAt, size int64) (*image, error) {
	if size < pvdOffset+sectorSize {
		return nil, fmt.Errorf("file too small for ISO 9660")
	}

	// Read Primary Volume Descriptor
	pvd := make([]byte, sectorSize)
	if _, err := r.ReadAt(pvd, pvdOffset); err != nil {
		return nil, fmt.Errorf("failed to read PVD: %w", err)
	}

	// Validate magic "CD001"
	if string(pvd[pvdMagicOffset:pvdMagicOffset+5]) != "CD001" {
		return nil, fmt.Errorf("not a valid ISO 9660: missing CD001 magic")
	}

	// Extract root directory record info
	rootRecord := pvd[pvdRootDirOffset:]
	rootExtentLoc := binary.LittleEndian.Uint32(rootRecord[dirEntryExtentLoc:])
	rootExtentLen := binary.LittleEndian.Uint32(rootRecord[dirEntryDataLen:])

	return &image{
		r:             r,
		rootExtentLoc: rootExtentLoc,
		rootExtentLen: rootExtentLen,
	}, nil
}

// readFile reads a file from the root directory by name (case-insensitive).
// Handles ISO 9660 version suffixes (e.g., ";1").
func (img *image) readFile(name string) ([]byte, error) {
	// Read root directory
	rootDir := make([]byte, img.rootExtentLen)
	if _, err := img.r.ReadAt(rootDir, int64(img.rootExtentLoc)*sectorSize); err != nil {
		return nil, fmt.Errorf("failed to read root directory: %w", err)
	}

	// Search for the file
	name = strings.ToUpper(name)
	offset := 0
	for offset < len(rootDir) {
		entryLen := int(rootDir[offset])
		if entryLen == 0 {
			// End of directory entries in this sector, try next sector
			nextSector := ((offset / sectorSize) + 1) * sectorSize
			if nextSector >= len(rootDir) {
				break
			}
			offset = nextSector
			continue
		}

		if offset+dirEntryName >= len(rootDir) {
			break
		}

		nameLen := int(rootDir[offset+dirEntryNameLen])
		if offset+dirEntryName+nameLen > len(rootDir) {
			break
		}

		entryName := strings.ToUpper(string(rootDir[offset+dirEntryName : offset+dirEntryName+nameLen]))

		// Strip version suffix (";1")
		if idx := strings.Index(entryName, ";"); idx != -1 {
			entryName = entryName[:idx]
		}

		if entryName == name {
			// Found it - read the file
			fileExtentLoc := binary.LittleEndian.Uint32(rootDir[offset+dirEntryExtentLoc:])
			fileSize := binary.LittleEndian.Uint32(rootDir[offset+dirEntryDataLen:])

			data := make([]byte, fileSize)
			if _, err := img.r.ReadAt(data, int64(fileExtentLoc)*sectorSize); err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			return data, nil
		}

		offset += entryLen
	}

	return nil, fmt.Errorf("file not found: %s", name)
}
