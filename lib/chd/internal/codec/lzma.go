package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/ulikunitz/xz/lzma"
)

// LZMA decompresses raw LZMA data (no header) as used by CHD.
// CHD stores raw LZMA compressed data without the standard 13-byte header.
func LZMA(data []byte, outputSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("LZMA data empty")
	}

	// CHD LZMA uses default properties: lc=3, lp=0, pb=2
	// Properties byte = (pb * 5 + lp) * 9 + lc = (2*5 + 0)*9 + 3 = 93 = 0x5D
	const propsByte = 0x5D

	// Dictionary size: use output size or 64KB, whichever is larger.
	// LZMA decoders accept dictionaries larger than needed.
	dictSize := uint32(65536)
	if outputSize > 65536 {
		dictSize = uint32(outputSize)
	}

	// Build standard LZMA header:
	// [0]   Properties byte
	// [1-4] Dictionary size (little-endian)
	// [5-12] Uncompressed size (little-endian)
	header := make([]byte, 13)
	header[0] = propsByte
	binary.LittleEndian.PutUint32(header[1:5], dictSize)
	binary.LittleEndian.PutUint64(header[5:13], uint64(outputSize))

	r, err := lzma.NewReader(bytes.NewReader(append(header, data...)))
	if err != nil {
		return nil, err
	}

	result := make([]byte, outputSize)
	n, err := io.ReadFull(r, result)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return result[:n], nil
}
