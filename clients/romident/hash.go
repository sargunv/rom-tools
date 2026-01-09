package romident

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
)

// CalculateHashes computes SHA1, MD5, and CRC32 hashes from a reader in a single pass.
func CalculateHashes(r io.Reader) ([]Hash, error) {
	sha1Hash := sha1.New()
	md5Hash := md5.New()
	crc32Hash := crc32.NewIEEE()

	// MultiWriter writes to all hashes simultaneously
	multiWriter := io.MultiWriter(sha1Hash, md5Hash, crc32Hash)

	if _, err := io.Copy(multiWriter, r); err != nil {
		return nil, fmt.Errorf("failed to read data for hashing: %w", err)
	}

	return []Hash{
		{Algorithm: HashSHA1, Value: hex.EncodeToString(sha1Hash.Sum(nil)), Source: "calculated"},
		{Algorithm: HashMD5, Value: hex.EncodeToString(md5Hash.Sum(nil)), Source: "calculated"},
		{Algorithm: HashCRC32, Value: fmt.Sprintf("%08x", crc32Hash.Sum32()), Source: "calculated"},
	}, nil
}

// CalculateSingleHash computes a single hash from a reader.
func CalculateSingleHash(r io.Reader, algorithm HashAlgorithm) (Hash, error) {
	var h hash.Hash
	switch algorithm {
	case HashSHA1:
		h = sha1.New()
	case HashMD5:
		h = md5.New()
	case HashCRC32:
		h = crc32.NewIEEE()
	default:
		return Hash{}, fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	if _, err := io.Copy(h, r); err != nil {
		return Hash{}, fmt.Errorf("failed to read data for hashing: %w", err)
	}

	var value string
	if algorithm == HashCRC32 {
		value = fmt.Sprintf("%08x", h.(hash.Hash32).Sum32())
	} else {
		value = hex.EncodeToString(h.Sum(nil))
	}

	return Hash{Algorithm: algorithm, Value: value, Source: "calculated"}, nil
}

// NewHash creates a Hash with the given parameters.
func NewHash(algorithm HashAlgorithm, value, source string) Hash {
	return Hash{
		Algorithm: algorithm,
		Value:     value,
		Source:    source,
	}
}
