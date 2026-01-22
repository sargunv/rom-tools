package n64

// N64 (word-swapped/little-endian) ROM format.
//
// N64 files have 32-bit words reversed compared to native Z64 format.
// This package converts N64 to Z64 format and delegates to the z64 package.
