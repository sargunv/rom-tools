package codec

import (
	"bytes"
	"compress/flate"
	"io"
)

// Zlib decompresses raw zlib/flate compressed data.
func Zlib(data []byte, outputSize int) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()

	result := make([]byte, outputSize)
	n, err := io.ReadFull(r, result)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return result[:n], nil
}
