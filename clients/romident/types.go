// Package romident provides ROM identification and hashing utilities.
package romident

// ROMType indicates how the ROM is packaged.
type ROMType string

const (
	ROMTypeFile   ROMType = "file"
	ROMTypeZIP    ROMType = "zip"
	ROMTypeFolder ROMType = "folder"
)

// Format indicates the detected file format.
type Format string

const (
	FormatUnknown Format = "unknown"
	FormatCHD     Format = "chd"
	FormatXISO    Format = "xiso"
	FormatISO9660 Format = "iso9660"
	FormatZIP     Format = "zip"
	// Future: FormatGBA, FormatSFC, etc.
)

// HashAlgorithm identifies a hash algorithm.
type HashAlgorithm string

const (
	HashSHA1  HashAlgorithm = "sha1"
	HashMD5   HashAlgorithm = "md5"
	HashCRC32 HashAlgorithm = "crc32"
)

// Hash represents a computed or extracted hash value.
type Hash struct {
	Algorithm HashAlgorithm `json:"algorithm"`
	Value     string        `json:"value"`  // hex-encoded
	Source    string        `json:"source"` // "calculated", "zip-metadata", "chd-raw", "chd-compressed"
}

// ROMFile represents a single file within a ROM.
type ROMFile struct {
	Size      int64  `json:"size"`
	Format    Format `json:"format"`
	Hashes    []Hash `json:"hashes"`
	IsPrimary bool   `json:"is_primary,omitempty"` // true if used for identification
}

// Files is a map of file path to file info.
type Files map[string]ROMFile

// GameIdent represents platform-specific identification data.
type GameIdent struct {
	Platform string            `json:"platform"`
	TitleID  string            `json:"title_id,omitempty"`
	Title    string            `json:"title,omitempty"`
	Region   string            `json:"region,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

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
