# Atari Lynx ROM Support Plan

## Overview

The Atari Lynx was a handheld console with ROMs distributed in two formats: LNX (with 64-byte header) and LYX (raw, headerless).

## File Extensions

- `.lnx` - With 64-byte header (most common for emulators)
- `.lyx` - Raw/headerless format (used with hardware carts)
- `.o` - Object file format (rare)

## LNX Header Format (64 bytes at offset 0)

| Offset | Size | Description                                    |
| ------ | ---- | ---------------------------------------------- |
| 0x00   | 4    | Magic: "LYNX"                                  |
| 0x04   | 2    | Bank 0 page size (little-endian)               |
| 0x06   | 2    | Bank 1 page size (little-endian)               |
| 0x08   | 2    | Version (little-endian)                        |
| 0x0A   | 32   | Cartridge Name (null-terminated)               |
| 0x2A   | 16   | Manufacturer Name (null-terminated)            |
| 0x3A   | 1    | Rotation (0=none, 1=left, 2=right)             |
| 0x3B   | 1    | EEPROM flag (0=no EEPROM, non-zero=has EEPROM) |
| 0x3C   | 4    | Reserved (spare bytes)                         |

## LYX Format

LYX files have no header - they're raw ROM dumps. To identify:

1. Check file size (common sizes: 64KB, 128KB, 256KB, 512KB)
2. May need heuristics or database lookup

## Bank Size Values

The bank page size fields indicate how memory is organized:

- Typical values: 0x0100 (256 bytes), 0x0200 (512 bytes)
- A value of 0 means the bank is not used

## Rotation Values

| Value | Orientation       |
| ----- | ----------------- |
| 0     | None (horizontal) |
| 1     | Left (90° CCW)    |
| 2     | Right (90° CW)    |

This indicates how the Lynx should be held while playing.

## EEPROM Flag

Indicates if the cartridge has EEPROM for save data:

- 0x00 = No EEPROM
- Non-zero = Has EEPROM (value may indicate EEPROM size)

## Magic Detection

```go
var lynxMagic = []byte("LYNX")
var lynxMagicOffset = int64(0)
```

## Extracted Metadata

- **Title**: 32-byte cartridge name
- **Manufacturer**: 16-byte manufacturer name
- **Rotation**: Display orientation
- **EEPROM**: Save capability
- **Bank Sizes**: Memory configuration

## Implementation Notes

1. **LNX vs LYX detection**: Check for "LYNX" magic at offset 0
2. **No region info**: Lynx was region-free
3. **Rotation metadata** is useful for emulator configuration
4. **LYX fallback**: If no header, use extension and database

## Detection Strategy

1. Check for "LYNX" magic at offset 0
2. If found, parse LNX header
3. If not found and extension is `.lyx`, treat as headerless
4. Validate header fields (bank sizes, etc.)

## LYX to LNX Conversion

If needed, LYX files can be "upgraded" by prepending a header:

```go
func createLNXHeader(name string, bank0Size, bank1Size uint16) []byte {
    header := make([]byte, 64)
    copy(header[0:4], "LYNX")
    binary.LittleEndian.PutUint16(header[4:6], bank0Size)
    binary.LittleEndian.PutUint16(header[6:8], bank1Size)
    binary.LittleEndian.PutUint16(header[8:10], 1) // version
    copy(header[0x0A:0x2A], name)
    return header
}
```

## Test ROMs

- Atari Lynx homebrew community is active
- Homebrew games available on itch.io and dedicated sites
- Test ROMs for validation

## Complexity: Low

Simple 64-byte header with clear format. Main complexity is handling headerless LYX files.
