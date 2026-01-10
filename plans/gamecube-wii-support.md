# Nintendo GameCube / Wii ROM Support Plan

## Overview

GameCube and Wii share similar disc formats. Both use proprietary optical discs with a consistent header structure. Wii also supports the WBFS (Wii Backup File System) format.

## File Extensions

### GameCube

- `.iso` / `.gcm` - GameCube disc image
- `.gcz` - Dolphin compressed format
- `.rvz` - Modern compressed format (Dolphin)
- `.nkit.iso` - NKit format

### Wii

- `.iso` - Wii disc image
- `.wbfs` - Wii Backup File System
- `.wia` - Wii compressed format
- `.rvz` - Modern compressed format (Dolphin)

## GameCube/Wii Disc Header (boot.bin)

The disc header is at offset 0x00 and is 0x440 bytes:

| Offset | Size | Description                        |
| ------ | ---- | ---------------------------------- |
| 0x000  | 1    | Disc ID (system)                   |
| 0x001  | 2    | Game Code                          |
| 0x003  | 1    | Region Code                        |
| 0x004  | 2    | Maker Code                         |
| 0x006  | 1    | Disc Number                        |
| 0x007  | 1    | Disc Version                       |
| 0x008  | 1    | Audio Streaming                    |
| 0x009  | 1    | Stream Buffer Size                 |
| 0x00A  | 14   | Reserved                           |
| 0x018  | 4    | Wii Magic (0x5D1C9EA3 for Wii)     |
| 0x01C  | 4    | GC Magic (0xC2339F3D for GameCube) |
| 0x020  | 64   | Game Title (null-terminated)       |
| 0x060  | ...  | Additional data                    |

## Disc ID (System Byte at 0x000)

| Value | System              |
| ----- | ------------------- |
| G     | GameCube game       |
| D     | GameCube demo disc  |
| P     | GC promotional disc |
| U     | GC utility disc     |
| R     | Wii game            |
| S     | Wii game (alt)      |
| 0-4   | Wii Channel/System  |

## Game ID (ID6)

The first 6 bytes form the ID6:

- Byte 0: Disc ID (system type)
- Bytes 1-2: Game Code (unique identifier)
- Byte 3: Region Code
- Bytes 4-5: Maker Code

Example: "GALE01" = GameCube (G) + "ALE" game code + USA (E) + Nintendo (01)

## Region Codes (Byte 3)

| Code | Region       |
| ---- | ------------ |
| E    | USA          |
| P    | Europe (PAL) |
| J    | Japan        |
| K    | Korea        |
| W    | Taiwan       |

## Maker Codes (Bytes 4-5)

Common codes:

- 01 = Nintendo
- 08 = Capcom
- 41 = Ubisoft
- 4Q = Disney
- 52 = Activision
- 69 = EA
- 78 = THQ
- 8P = Sega

## Magic Numbers

```go
// GameCube magic at offset 0x1C
var gcMagic = []byte{0xC2, 0x33, 0x9F, 0x3D}
var gcMagicOffset = int64(0x1C)

// Wii magic at offset 0x18
var wiiMagic = []byte{0x5D, 0x1C, 0x9E, 0xA3}
var wiiMagicOffset = int64(0x18)
```

## WBFS Format

WBFS is a container format for Wii disc images.

### WBFS Header

| Offset | Size | Description       |
| ------ | ---- | ----------------- |
| 0x000  | 4    | Magic "WBFS"      |
| 0x004  | 4    | Number of sectors |
| 0x008  | 1    | HD sector shift   |
| 0x009  | 1    | WBFS sector shift |
| 0x00A  | 2    | Reserved          |
| 0x00C  | ...  | Disc table        |

### WBFS Disc Entry

The actual disc header (ID6 and game title) is found at offset 0x200 from the start of the WBFS file.

## RVZ/WIA Compressed Formats

These are modern compressed formats. The header contains:

- Magic identifier ("WIA\x01" or "RVZ\x01")
- Compression parameters
- Embedded disc header data

For basic identification, the disc header is typically preserved early in the file.

## Extracted Metadata

- **Game ID (ID6)**: First 6 bytes
- **Title**: 64-byte null-terminated string at 0x20
- **Region**: From byte 3 of ID6
- **Maker Code**: Bytes 4-5
- **Disc Number**: For multi-disc games
- **Version**: Disc version
- **Platform**: GameCube or Wii based on magic/disc ID

## Implementation Notes

1. **Dual platform support** - Same header format for GC and Wii
2. **WBFS requires special handling** - Disc header at offset 0x200
3. **Compressed formats** may need partial decompression
4. **Magic bytes** distinguish GC from Wii

## Detection Strategy

1. Check for GC magic at 0x1C or Wii magic at 0x18
2. For WBFS, check "WBFS" at 0x00, then read header at 0x200
3. Parse ID6 and title from appropriate offset
4. Distinguish platform by disc ID byte or magic

## Test ROMs

- GameCube/Wii homebrew
- Dolphin test cases
- Public domain demos

## Complexity: Medium

Multiple formats (ISO, WBFS, RVZ) but consistent header structure once located.
