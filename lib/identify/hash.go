package identify

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"

	"github.com/sargunv/rom-tools/lib/core"
)

// calculateHashes computes SHA1, MD5, and CRC32 hashes from a ReaderAt in a single pass.
func calculateHashes(r io.ReaderAt, size int64) (core.Hashes, error) {
	sha1Hash := sha1.New()
	md5Hash := md5.New()
	crc32Hash := crc32.NewIEEE()

	// MultiWriter writes to all hashes simultaneously
	multiWriter := io.MultiWriter(sha1Hash, md5Hash, crc32Hash)

	// Use SectionReader to read from offset 0 to size
	sectionReader := io.NewSectionReader(r, 0, size)
	if _, err := io.Copy(multiWriter, sectionReader); err != nil {
		return nil, fmt.Errorf("failed to read data for hashing: %w", err)
	}

	return core.Hashes{
		core.HashSHA1:  hex.EncodeToString(sha1Hash.Sum(nil)),
		core.HashMD5:   hex.EncodeToString(md5Hash.Sum(nil)),
		core.HashCRC32: fmt.Sprintf("%08x", crc32Hash.Sum32()),
	}, nil
}
