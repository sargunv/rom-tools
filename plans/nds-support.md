# Nintendo DS ROM Support Plan

## Overview

Nintendo DS ROMs have a well-structured 512-byte header at the beginning of the file with comprehensive metadata.

## File Extensions

- `.nds` - Nintendo DS ROM
- `.dsi` - DSi enhanced ROM
- `.ids` - DSi-only ROM
- `.srl` - SDK format

## Header Format (512 bytes at offset 0x00)

| Offset      | Size | Description                                     |
| ----------- | ---- | ----------------------------------------------- |
| 0x000       | 12   | Game Title (uppercase ASCII, null-padded)       |
| 0x00C       | 4    | Game Code (e.g., "AMFE")                        |
| 0x010       | 2    | Maker Code (e.g., "01" for Nintendo)            |
| 0x012       | 1    | Unit Code (0x00=DS, 0x02=DS+DSi, 0x03=DSi only) |
| 0x013       | 1    | Encryption seed select                          |
| 0x014       | 1    | Device capacity (ROM size = 2^(20+n) bytes)     |
| 0x015       | 7    | Reserved                                        |
| 0x01C       | 1    | DSi flags                                       |
| 0x01D       | 1    | NDS region                                      |
| 0x01E       | 1    | ROM Version                                     |
| 0x01F       | 1    | Autostart flag                                  |
| 0x020       | 4    | ARM9 ROM offset                                 |
| 0x024       | 4    | ARM9 entry address                              |
| 0x028       | 4    | ARM9 RAM address                                |
| 0x02C       | 4    | ARM9 size                                       |
| 0x030       | 4    | ARM7 ROM offset                                 |
| 0x034       | 4    | ARM7 entry address                              |
| 0x038       | 4    | ARM7 RAM address                                |
| 0x03C       | 4    | ARM7 size                                       |
| 0x040       | 4    | File Name Table offset                          |
| 0x044       | 4    | File Name Table size                            |
| 0x048       | 4    | File Allocation Table offset                    |
| 0x04C       | 4    | File Allocation Table size                      |
| 0x050       | 4    | ARM9 overlay offset                             |
| 0x054       | 4    | ARM9 overlay size                               |
| 0x058       | 4    | ARM7 overlay offset                             |
| 0x05C       | 4    | ARM7 overlay size                               |
| 0x060       | 4    | Port 0x40001A4 setting (normal)                 |
| 0x064       | 4    | Port 0x40001A4 setting (KEY1)                   |
| 0x068       | 4    | Icon/Title offset                               |
| 0x06C       | 2    | Secure Area checksum                            |
| 0x06E       | 2    | Secure Area delay                               |
| 0x070       | 4    | ARM9 auto-load list RAM address                 |
| 0x074       | 4    | ARM7 auto-load list RAM address                 |
| 0x078       | 8    | Secure Area disable                             |
| 0x080       | 4    | Total used ROM size                             |
| 0x084       | 4    | ROM header size                                 |
| 0x088-0x0BF | 56   | Reserved                                        |
| 0x0C0       | 156  | Nintendo Logo                                   |
| 0x15C       | 2    | Nintendo Logo checksum                          |
| 0x15E       | 2    | Header checksum                                 |
| 0x160       | 4    | Debug ROM offset                                |
| 0x164       | 4    | Debug size                                      |
| 0x168       | 4    | Debug RAM address                               |
| 0x16C       | 4    | Reserved                                        |
| 0x170-0x1FF | 144  | Reserved                                        |

## Game Code Format

The 4-character game code follows this pattern:

- Char 1: Category (A=Action, B=Sports, etc.)
- Char 2-3: Unique ID
- Char 4: Region (E=USA, P=Europe, J=Japan, K=Korea, etc.)

## Unit Code Values

| Value | Description                |
| ----- | -------------------------- |
| 0x00  | Nintendo DS                |
| 0x02  | Nintendo DS + DSi enhanced |
| 0x03  | Nintendo DSi only          |

## Region Codes (Character 4 of Game Code)

| Code | Region      |
| ---- | ----------- |
| J    | Japan       |
| E    | USA         |
| P    | Europe      |
| D    | Germany     |
| F    | France      |
| I    | Italy       |
| S    | Spain       |
| K    | Korea       |
| C    | China       |
| A    | All regions |
| U    | Australia   |

## Magic Detection

Check for Nintendo Logo at offset 0x0C0:

```go
// First few bytes of Nintendo Logo
var ndsLogoStart = []byte{0x24, 0xFF, 0xAE, 0x51, 0x69, 0x9A, 0xA2, 0x21}
```

Alternatively, validate header checksum at 0x15E.

## Extracted Metadata

- **Title**: 12-byte game title
- **Game Code**: 4-byte unique identifier
- **Maker Code**: 2-byte publisher code
- **Version**: ROM version number
- **Region**: From game code char 4
- **Platform**: DS, DS+DSi, or DSi-only from unit code

## Icon/Title Data

At the offset specified in header (0x068), there's icon and title data:

- Game icon (32x32, 4bpp)
- Title strings in multiple languages

## Implementation Notes

1. **Rich metadata** - Game title and code are always present
2. **Checksum validation** at 0x15E validates header integrity
3. **Unit code** distinguishes DS from DSi games
4. **Nintendo Logo** can be used for additional validation

## Test ROMs

- DS homebrew (devkitPro examples)
- Homebrew games from various sites
- NDS test ROMs

## Complexity: Low

Well-structured header with clear format. Straightforward to implement.
