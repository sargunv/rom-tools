# Sega Master System / Game Gear ROM Support Plan

## Overview

The Sega Master System (SMS) and Game Gear (GG) share a similar ROM header format. The header is located near the end of the ROM rather than at the beginning.

## File Extensions

### Master System

- `.sms` - Master System ROM
- `.sg` - SG-1000 ROM

### Game Gear

- `.gg` - Game Gear ROM

## Header Location

The header location depends on ROM size:

| ROM Size | Header Offset |
| -------- | ------------- |
| 8 KB     | 0x1FF0        |
| 16 KB    | 0x3FF0        |
| 32 KB+   | 0x7FF0        |

Note: Some ROMs (especially unlicensed or Japanese) may not have a valid header.

## Header Format (16 bytes)

| Offset | Size | Description                                |
| ------ | ---- | ------------------------------------------ |
| 0x00   | 8    | Magic: "TMR SEGA"                          |
| 0x08   | 2    | Reserved (often 0x0000, 0xFFFF, or 0x2020) |
| 0x0A   | 2    | Checksum (little-endian)                   |
| 0x0C   | 3    | Product code (BCD format)                  |
| 0x0F   | 1    | Version and region/ROM size                |

### Byte 0x0F Structure

| Bits | Description   |
| ---- | ------------- |
| 7-4  | ROM size code |
| 3-0  | Region code   |

### ROM Size Codes (upper nibble)

| Code | Size   |
| ---- | ------ |
| 0xA  | 8 KB   |
| 0xB  | 16 KB  |
| 0xC  | 32 KB  |
| 0xD  | 48 KB  |
| 0xE  | 64 KB  |
| 0xF  | 128 KB |
| 0x0  | 256 KB |
| 0x1  | 512 KB |
| 0x2  | 1 MB   |

### Region Codes (lower nibble)

| Code | Region/System    |
| ---- | ---------------- |
| 0x3  | SMS Japan        |
| 0x4  | SMS Export       |
| 0x5  | GG Japan         |
| 0x6  | GG Export        |
| 0x7  | GG International |

## Magic Detection

```go
var smsMagic = []byte("TMR SEGA")
// Check at offsets 0x1FF0, 0x3FF0, or 0x7FF0
```

## Product Code

The 3-byte product code at offset 0x0C is in BCD (Binary Coded Decimal) format:

- Stored as: low byte, high byte, last nibble in version byte
- Example: bytes `90 51 4x` = product code 5190

To decode:

```go
func decodeProductCode(b []byte) string {
    // b[0] = low byte, b[1] = high byte, b[2] upper nibble = last digit
    low := b[0]
    high := b[1]
    last := (b[2] >> 4) & 0x0F
    // BCD decode each nibble
    return fmt.Sprintf("%d%d%d%d%d",
        high>>4, high&0x0F,
        low>>4, low&0x0F,
        last)
}
```

## Checksum

The checksum at offset 0x0A is calculated over ROM data (excluding the header area in some cases). The Master System BIOS uses this for validation, but Game Gear BIOS does not verify it.

## Platform Detection

Differentiate SMS from GG using:

1. File extension (`.sms` vs `.gg`)
2. Region code in header byte 0x0F

## Extracted Metadata

- **Product Code**: From BCD-encoded bytes
- **Version**: Lower 4 bits of the version/region byte
- **Region**: From region code (Japan vs Export vs International)
- **Platform**: SMS or GG based on region code
- **Checksum**: For validation

## Implementation Notes

1. **No game title** - The header doesn't contain a title
2. **Header may be absent** - Especially in unlicensed games
3. **Try multiple offsets** - Header location varies by ROM size
4. **Fallback to extension** - If no valid header, use file extension for platform

## Detection Strategy

1. Try to find "TMR SEGA" at standard offsets (0x7FF0, 0x3FF0, 0x1FF0)
2. If found, extract product code, region, version
3. If not found, fall back to extension-based detection

## Test ROMs

- Homebrew from SMS Power (smspower.org)
- Public domain demos
- Test ROMs

## Complexity: Low-Medium

Simple header format, but variable location and optional presence add complexity.
