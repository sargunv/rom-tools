// Package romident provides ROM identification and hashing utilities.
package romident

import "github.com/sargunv/rom-tools/lib/romident/core"

// ROMType indicates how the ROM is packaged.
type ROMType string

const (
	ROMTypeFile   ROMType = "file"
	ROMTypeZIP    ROMType = "zip"
	ROMTypeFolder ROMType = "folder"
)

// Format is an alias for core.Format.
type Format = core.Format

// Format constants re-exported from core package.
const (
	FormatUnknown = core.FormatUnknown
	FormatCHD     = core.FormatCHD
	FormatXISO    = core.FormatXISO
	FormatXBE     = core.FormatXBE
	FormatISO9660 = core.FormatISO9660
	FormatZIP     = core.FormatZIP
	FormatGBA     = core.FormatGBA
	FormatZ64     = core.FormatZ64
	FormatV64     = core.FormatV64
	FormatN64     = core.FormatN64
	FormatGB      = core.FormatGB
	FormatMD      = core.FormatMD
	FormatSMD     = core.FormatSMD
	FormatNDS     = core.FormatNDS
	FormatNES     = core.FormatNES
	FormatSNES    = core.FormatSNES
	FormatGCM     = core.FormatGCM
	FormatRVZ     = core.FormatRVZ
)

// HashAlgorithm identifies a hash algorithm.
type HashAlgorithm string

const (
	HashSHA1  HashAlgorithm = "sha1"
	HashMD5   HashAlgorithm = "md5"
	HashCRC32 HashAlgorithm = "crc32"
)

// HashSource indicates where a hash value came from.
type HashSource string

const (
	HashSourceCalculated    HashSource = "calculated"
	HashSourceZIPMetadata   HashSource = "zip-metadata"
	HashSourceCHDRaw        HashSource = "chd-raw"
	HashSourceCHDCompressed HashSource = "chd-compressed"
)

// Hash represents a computed or extracted hash value.
type Hash struct {
	Algorithm HashAlgorithm `json:"algorithm"`
	Value     string        `json:"value"` // hex-encoded
	Source    HashSource    `json:"source"`
}

// ROMFile represents a single file within a ROM.
type ROMFile struct {
	Size      int64  `json:"size"`
	Format    Format `json:"format"`
	Hashes    []Hash `json:"hashes"`
	IsPrimary bool   `json:"is_primary,omitempty"` // true if used for identification
}

type GameIdent = core.GameIdent
type Region = core.Region
type Platform = core.Platform

// Files is a map of file path to file info.
type Files map[string]ROMFile

// ROM represents a complete game unit (single file, zip, or folder).
type ROM struct {
	Path  string     `json:"path"`
	Type  ROMType    `json:"type"`
	Files Files      `json:"files"`
	Ident *GameIdent `json:"ident,omitempty"`
}

// HashMode controls how hashes are calculated.
type HashMode int

const (
	// HashModeDefault uses fast methods where available (CHD header, ZIP metadata),
	// calculates full hashes for loose files.
	HashModeDefault HashMode = iota

	// HashModeFast skips hash calculation for large files, but still calculates
	// hashes for small files (below FastModeSmallFileThreshold).
	HashModeFast

	// HashModeSlow calculates full hashes even when fast methods are available
	// (e.g., decompresses ZIP files to calculate SHA1/MD5).
	HashModeSlow
)

const (
	// FastModeSmallFileThreshold is the size threshold below which fast mode
	// will still calculate hashes. Files at or above this size skip hash calculation.
	// 65 MiB covers most cartridge ROMs (GBA, SNES, NES, etc.) but skips large disc images.
	FastModeSmallFileThreshold = 65 * 1024 * 1024 // 65 MiB
)

// Options controls ROM identification behavior.
type Options struct {
	HashMode HashMode
}
