package xbox

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// Xbox XISO (XDVDFS) format parsing.
//
// XDVDFS (Xbox DVD File System) specification:
// https://xboxdevwiki.net/Xbox_Game_Disc#Xbox_Game_Disc_filesystem_layout
//
// XISO layout:
//   - Offset 0x10000: Volume descriptor with "MICROSOFT*XBOX*MEDIA" magic
//   - Root directory entry follows, containing file entries in a binary tree
//   - default.xbe in root contains game metadata in its certificate

const (
	xisoVolumeDescOffset = 0x10000
	xisoMagicSize        = 20
	xisoRootDirOffset    = 0x14
	xisoRootDirSizeOff   = 0x18
)

// ParseXISO extracts game information from an Xbox XISO image.
func ParseXISO(r io.ReaderAt, size int64) (*XBEInfo, error) {
	// Read volume descriptor
	if size < xisoVolumeDescOffset+32 {
		return nil, fmt.Errorf("file too small for XISO header")
	}

	volDesc := make([]byte, 32)
	if _, err := r.ReadAt(volDesc, xisoVolumeDescOffset); err != nil {
		return nil, fmt.Errorf("failed to read XISO volume descriptor: %w", err)
	}

	// Verify magic
	if string(volDesc[:xisoMagicSize]) != "MICROSOFT*XBOX*MEDIA" {
		return nil, fmt.Errorf("not a valid XISO: invalid magic")
	}

	// Get root directory location
	rootDirSector := binary.LittleEndian.Uint32(volDesc[xisoRootDirOffset:])
	rootDirSize := binary.LittleEndian.Uint32(volDesc[xisoRootDirSizeOff:])

	// Find default.xbe in root directory
	xbeOffset, err := findDefaultXBE(r, int64(rootDirSector)*2048, int64(rootDirSize))
	if err != nil {
		return nil, fmt.Errorf("failed to find default.xbe: %w", err)
	}

	// Parse XBE using io.NewSectionReader to create a reader starting at the XBE offset
	xbeReader := io.NewSectionReader(r, xbeOffset, size-xbeOffset)
	return ParseXBE(xbeReader, size-xbeOffset)
}

// findDefaultXBE searches the XDVDFS directory tree for default.xbe.
// The directory uses a binary tree structure with left/right child offsets.
func findDefaultXBE(r io.ReaderAt, dirOffset, dirSize int64) (int64, error) {
	dirData := make([]byte, dirSize)
	if _, err := r.ReadAt(dirData, dirOffset); err != nil {
		return 0, fmt.Errorf("failed to read directory at offset %d (size %d): %w", dirOffset, dirSize, err)
	}

	return searchDirectory(dirData, "default.xbe")
}

// searchDirectory searches a directory's binary tree for a file.
// Directory entry format:
//
//	Offset  Size  Description
//	0       2     Left child offset (within directory, *4)
//	2       2     Right child offset (within directory, *4)
//	4       4     File start sector
//	8       4     File size
//	12      1     File attributes
//	13      1     Filename length
//	14      N     Filename (ASCII)
func searchDirectory(dirData []byte, target string) (int64, error) {
	target = strings.ToLower(target)
	return searchDirectoryAt(dirData, 0, target)
}

func searchDirectoryAt(dirData []byte, offset int, target string) (int64, error) {
	if offset+14 > len(dirData) {
		return 0, fmt.Errorf("directory entry offset out of bounds")
	}

	leftOffset := binary.LittleEndian.Uint16(dirData[offset:]) * 4
	rightOffset := binary.LittleEndian.Uint16(dirData[offset+2:]) * 4
	fileSector := binary.LittleEndian.Uint32(dirData[offset+4:])
	nameLen := int(dirData[offset+13])

	if nameLen == 0 || offset+14+nameLen > len(dirData) {
		return 0, fmt.Errorf("invalid directory entry")
	}

	name := strings.ToLower(string(dirData[offset+14 : offset+14+nameLen]))

	if name == target {
		return int64(fileSector) * 2048, nil
	}

	// Binary tree search
	if target < name && leftOffset != 0 {
		result, err := searchDirectoryAt(dirData, int(leftOffset), target)
		if err == nil {
			return result, nil
		}
	}
	if target > name && rightOffset != 0 {
		result, err := searchDirectoryAt(dirData, int(rightOffset), target)
		if err == nil {
			return result, nil
		}
	}

	return 0, fmt.Errorf("file not found")
}
