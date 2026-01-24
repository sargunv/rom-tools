package identify

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
)

// calculateHashes computes SHA1, MD5, and CRC32 hashes from a reader in a single pass.
func calculateHashes(r io.Reader) (Hashes, error) {
	sha1Hash := sha1.New()
	md5Hash := md5.New()
	crc32Hash := crc32.NewIEEE()

	// MultiWriter writes to all hashes simultaneously
	multiWriter := io.MultiWriter(sha1Hash, md5Hash, crc32Hash)

	if _, err := io.Copy(multiWriter, r); err != nil {
		return nil, fmt.Errorf("failed to read data for hashing: %w", err)
	}

	return Hashes{
		HashSHA1:  hex.EncodeToString(sha1Hash.Sum(nil)),
		HashMD5:   hex.EncodeToString(md5Hash.Sum(nil)),
		HashCRC32: fmt.Sprintf("%08x", crc32Hash.Sum32()),
	}, nil
}
