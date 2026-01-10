# SNES/Super Famicom ROM Support Plan

## Overview

Super Nintendo ROMs have an internal header that provides rich metadata. The header location varies based on memory mapping mode (LoROM vs HiROM).

## File Extensions

- `.sfc` - Super Famicom (headerless, preferred)
- `.smc` - Super MagiCom (may have 512-byte copier header)
- `.fig` - Pro Fighter format
- `.swc` - Super Wild Card format

## Header Locations

The SNES internal header is 48 bytes (extended to 64 bytes with maker/game code):

| Mode    | Header Offset (no copier header) | With 512-byte copier header |
| ------- | -------------------------------- | --------------------------- |
| LoROM   | 0x7FB0 (extended) / 0x7FC0       | 0x81B0 / 0x81C0             |
| HiROM   | 0xFFB0 (extended) / 0xFFC0       | 0x101B0 / 0x101C0           |
| ExHiROM | 0x40FFB0 / 0x40FFC0              | + 0x200                     |

## Header Format (64-byte extended at 0x7FB0/0xFFB0)

| Offset    | Size | Description                                       |
| --------- | ---- | ------------------------------------------------- |
| 0x00      | 2    | Maker Code (ASCII)                                |
| 0x04      | 4    | Game Code (ASCII)                                 |
| 0x08      | 7    | Reserved                                          |
| 0x0F      | 1    | Expansion RAM size                                |
| 0x10      | 1    | Special version                                   |
| 0x11      | 1    | Cartridge type sub-number                         |
| 0x12-0x14 | 3    | Reserved                                          |
| 0x15      | 21   | Game Title (ASCII/JIS, space-padded)              |
| 0x2A      | 1    | Map Mode                                          |
| 0x2B      | 1    | Cartridge Type (ROM, ROM+RAM, ROM+RAM+SRAM, etc.) |
| 0x2C      | 1    | ROM Size (2^n KB)                                 |
| 0x2D      | 1    | RAM Size (2^n KB)                                 |
| 0x2E      | 1    | Destination Code (region)                         |
| 0x2F      | 1    | Old Maker Code (0x33 = use new maker code)        |
| 0x30      | 1    | ROM Version                                       |
| 0x31      | 2    | Checksum Complement                               |
| 0x33      | 2    | Checksum                                          |

## Map Mode Byte (offset 0x2A)

- Bits 0-3: Memory speed (0=SlowROM, 1=FastROM)
- Bit 4: 0=LoROM, 1=HiROM
- Bit 5: ExHiROM flag

Common values:

- 0x20: LoROM
- 0x21: HiROM
- 0x23: SA-1
- 0x25: ExHiROM
- 0x30: LoROM + FastROM
- 0x31: HiROM + FastROM
- 0x35: ExHiROM + FastROM

## Destination Codes (Region)

| Code | Region            |
| ---- | ----------------- |
| 0x00 | Japan             |
| 0x01 | USA               |
| 0x02 | Europe (PAL)      |
| 0x03 | Sweden (PAL)      |
| 0x04 | Finland (PAL)     |
| 0x05 | Denmark (PAL)     |
| 0x06 | France (PAL)      |
| 0x07 | Netherlands (PAL) |
| 0x08 | Spain (PAL)       |
| 0x09 | Germany (PAL)     |
| 0x0A | Italy (PAL)       |
| 0x0B | China             |
| 0x0C | Indonesia         |
| 0x0D | Korea             |
| 0x0F | Canada            |
| 0x10 | Brazil            |
| 0x11 | Australia (PAL)   |

## Header Detection Strategy

1. Check for 512-byte copier header (if file size % 0x8000 == 0x200)
2. Read potential header at both LoROM and HiROM offsets
3. Validate using checksum: `checksum + checksum_complement == 0xFFFF`
4. Validate map mode byte matches expected location
5. Use the header with better validation score

## Magic Detection

No fixed magic bytes. Detection relies on:

1. File extension
2. Checksum validation
3. Map mode byte consistency

## Extracted Metadata

- **Title**: 21-byte game title
- **Game Code**: 4-byte unique identifier
- **Maker Code**: 2-byte publisher code
- **Version**: ROM version number
- **Region**: Destination code
- **Checksum**: For validation

## Implementation Notes

1. **Copier header detection** is important - many dumps have 512-byte prefix
2. **Header location detection** requires trying multiple offsets
3. **Checksum validation** helps verify correct header location
4. **JIS encoding** may be used for Japanese titles

## Test ROMs

- SNES homebrew from pdroms.de
- Homebrew competitions
- SNES test ROMs

## Complexity: Medium-High

Multiple header locations and formats require careful detection logic.
