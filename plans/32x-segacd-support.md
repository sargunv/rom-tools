# Sega 32X / Sega CD ROM Support Plan

## Overview

The Sega 32X was an add-on for the Mega Drive/Genesis, while Sega CD was a CD-ROM attachment. Both share characteristics with the base Mega Drive format.

---

## Sega 32X

### File Extensions

- `.32x` - 32X ROM
- `.bin` - Generic binary (needs detection)

### Header Format

The 32X uses the same header structure as Mega Drive at offset $100-$1FF, but with different system identification:

| Offset | Size | Description                          |
| ------ | ---- | ------------------------------------ |
| $100   | 16   | System Type ("SEGA 32X" or variants) |
| $110   | 16   | Copyright/Release Date               |
| $120   | 48   | Domestic Title                       |
| $150   | 48   | Overseas Title                       |
| $180   | 14   | Serial Number                        |
| $18E   | 2    | Checksum                             |
| $190   | 16   | Device Support                       |
| $1A0   | 8    | ROM Address Range                    |
| $1A8   | 8    | RAM Address Range                    |
| $1B0   | 12   | Extra Memory                         |
| $1BC   | 12   | Modem Support                        |
| $1C8   | 40   | Reserved                             |
| $1F0   | 16   | Region Support                       |

### 32X-Specific Header (at $3C0)

Additional user header for 32X programs:

| Offset | Size | Description         |
| ------ | ---- | ------------------- |
| $3C0   | 4    | Module Name         |
| $3C4   | 4    | Version             |
| $3C8   | 4    | Source Address      |
| $3CC   | 4    | Destination Address |
| $3D0   | 4    | Size                |
| $3D4   | 4    | Master SH2 Start    |
| $3D8   | 4    | Slave SH2 Start     |
| $3DC   | 4    | Master SH2 VBR      |
| $3E0   | 4    | Slave SH2 VBR       |

### System Type Variants

- "SEGA 32X "
- "SEGA MARS " (codename)
- "SEGA MEGA DRIVE" + 32X indicator

### Magic Detection

```go
// Check for "SEGA 32X" at offset 0x100
var sega32xMagic = []byte("SEGA 32X")
var sega32xOffset = int64(0x100)

// Or check for "32X" substring in system type
```

---

## Sega CD / Mega CD

### File Extensions

- `.iso` / `.bin` - Disc image
- `.cue` + `.bin` - Cue sheet format
- `.chd` - MAME compressed format

### IP.BIN Structure

The Sega CD uses IP.BIN similar to Saturn, located in the boot sector:

| Offset | Size | Description                                 |
| ------ | ---- | ------------------------------------------- |
| 0x000  | 16   | System ID ("SEGADISCSYSTEM ")               |
| 0x010  | 11   | Volume Name                                 |
| 0x01B  | 5    | Volume Version (e.g., "$03.00")             |
| 0x020  | 2    | Volume Type (0x0001 = CD-ROM)               |
| 0x022  | 11   | System Name                                 |
| 0x02D  | 2    | System Version                              |
| 0x02F  | 8    | Build Date (mmddyyyy)                       |
| 0x037  | ...  | More metadata                               |
| 0x100  | 16   | Hardware ID ("SEGA MEGA CD " or "SEGA CD ") |
| 0x110  | 16   | Maker ID                                    |
| 0x120  | 48   | Domestic Title                              |
| 0x150  | 48   | Overseas Title                              |
| 0x180  | 14   | Product Number                              |
| 0x18E  | 2    | Version                                     |
| 0x190  | 16   | Device Support                              |
| 0x1F0  | 16   | Region Support                              |

### System ID Variants

- "SEGADISCSYSTEM " - Standard
- "SEGA MEGA CD " - Japanese
- "SEGA CD " - Western

### Region Support

Same as Mega Drive:

- J = Japan
- U = USA
- E = Europe

### Magic Detection

```go
// Primary magic
var segaCDMagic = []byte("SEGADISCSYSTEM")
var segaCDOffset = int64(0)

// Alternative check at 0x100 for hardware ID
var megaCDMagic = []byte("SEGA MEGA CD")
var megaCDMagicOffset = int64(0x100)
```

---

## Extracted Metadata

### 32X

- Reuse Mega Drive extraction
- Distinguish by system type containing "32X"

### Sega CD

- **Title**: Domestic and Overseas titles
- **Product Number**: 14 bytes
- **Version**: 2 bytes
- **Maker ID**: Publisher
- **Regions**: J/U/E codes
- **Volume Name**: Disc name

## Implementation Notes

### 32X

1. Same header location as Mega Drive ($100)
2. Check system type for "32X" or "MARS"
3. Reuse MD parsing, change platform designation

### Sega CD

1. Multiple possible header locations
2. May need to handle multi-track discs
3. CHD support useful for compressed images

## Detection Strategy

### 32X

1. Check offset $100 for "SEGA 32X" or "SEGA MARS"
2. Fall back to MD detection, check system type string
3. Use `.32x` extension as hint

### Sega CD

1. Check offset 0 for "SEGADISCSYSTEM"
2. Check offset $100 for "SEGA MEGA CD" or "SEGA CD"
3. For ISO/BIN, may need to check multiple sectors

## Complexity

- **32X**: Low (extends Mega Drive support)
- **Sega CD**: Medium (disc image handling)

## Test ROMs

- 32X homebrew is rare but exists
- Sega CD homebrew available
- Public domain demos
