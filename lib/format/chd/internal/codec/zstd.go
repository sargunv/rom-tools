package codec

import (
	"fmt"

	"github.com/klauspost/compress/zstd"
)

var zstdDecoder *zstd.Decoder

func init() {
	var err error
	zstdDecoder, err = zstd.NewReader(nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create zstd decoder: %v", err))
	}
}

// Zstd decompresses Zstandard compressed data.
func Zstd(data []byte, outputSize int) ([]byte, error) {
	result, err := zstdDecoder.DecodeAll(data, make([]byte, 0, outputSize))
	if err != nil {
		return nil, err
	}
	return result, nil
}
