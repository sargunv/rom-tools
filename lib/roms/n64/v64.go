package n64

// V64 (byte-swapped) N64 ROM format.
//
// V64 files have bytes swapped in pairs compared to native Z64 format.
// This package converts V64 to Z64 format and delegates to the z64 package.
