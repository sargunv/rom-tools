package format

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

// CHD (Compressed Hunks of Data) is MAME's compressed disc image format.
// Format specification: https://github.com/mamedev/mame/blob/master/src/lib/util/chd.h
//
// V5 header layout (124 bytes):
//
//	Offset  Size  Description
//	0       8     Magic ("MComprHD")
//	8       4     Header length (big-endian)
//	12      4     Version (big-endian)
//	16      4     Compressors[0]
//	20      4     Compressors[1]
//	24      4     Compressors[2]
//	28      4     Compressors[3]
//	32      8     Logical bytes (big-endian)
//	40      8     Map offset (big-endian)
//	48      8     Metadata offset (big-endian)
//	56      4     Hunk bytes (big-endian)
//	60      4     Unit bytes (big-endian)
//	64      20    Raw SHA1 (of the raw, uncompressed data)
//	84      20    SHA1 (of the compressed data)
//	104     20    Parent SHA1 (all zeros if no parent)
const (
	chdHeaderSize       = 124
	chdRawSHA1Offset    = 64 // Raw SHA1 at offset 64
	chdSHA1Offset       = 84 // Compressed SHA1 at offset 84
	chdParentSHA1Offset = 104
	chdSHA1Size         = 20
)

// CHDInfo contains metadata extracted from a CHD file header.
type CHDInfo struct {
	Version    uint32
	RawSHA1    string // SHA1 of the raw (uncompressed) data
	SHA1       string // SHA1 of the compressed data
	ParentSHA1 string // SHA1 of parent CHD (for diffs), empty if standalone
}

// ParseCHDHeader reads and parses a CHD file header.
// Returns both the raw and compressed SHA1 hashes stored in the header.
func ParseCHDHeader(r io.ReaderAt) (*CHDInfo, error) {
	header := make([]byte, chdHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read CHD header: %w", err)
	}

	// Verify magic
	if string(header[0:8]) != "MComprHD" {
		return nil, fmt.Errorf("not a valid CHD file: invalid magic")
	}

	// Header length (big-endian uint32 at offset 8)
	headerLen := binary.BigEndian.Uint32(header[8:12])

	// Version (big-endian uint32 at offset 12)
	version := binary.BigEndian.Uint32(header[12:16])

	if version < 5 {
		return nil, fmt.Errorf("CHD version %d not supported (only v5+ supported)", version)
	}

	if headerLen < chdHeaderSize {
		return nil, fmt.Errorf("CHD header too small: %d bytes", headerLen)
	}

	// Raw SHA1 at offset 64 (20 bytes) - SHA1 of the uncompressed data
	rawSHA1 := hex.EncodeToString(header[chdRawSHA1Offset : chdRawSHA1Offset+chdSHA1Size])

	// SHA1 at offset 84 (20 bytes) - SHA1 of the compressed data
	sha1 := hex.EncodeToString(header[chdSHA1Offset : chdSHA1Offset+chdSHA1Size])

	// Parent SHA1 at offset 104 (20 bytes) - all zeros if no parent
	parentSHA1Bytes := header[chdParentSHA1Offset : chdParentSHA1Offset+chdSHA1Size]
	parentSHA1 := ""
	for _, b := range parentSHA1Bytes {
		if b != 0 {
			parentSHA1 = hex.EncodeToString(parentSHA1Bytes)
			break
		}
	}

	return &CHDInfo{
		Version:    version,
		RawSHA1:    rawSHA1,
		SHA1:       sha1,
		ParentSHA1: parentSHA1,
	}, nil
}
