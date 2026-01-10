# Sega Saturn ROM Support Plan

## Overview

Sega Saturn games were distributed on CD-ROM using a proprietary format. The game metadata is stored in the IP.BIN (Initial Program) file in the first 16 sectors of the disc.

## File Extensions

- `.iso` / `.bin` - Raw disc image
- `.cue` + `.bin` - Cue sheet with binary data
- `.ccd` + `.img` - CloneCD format
- `.mdf` + `.mds` - Alcohol 120% format
- `.chd` - MAME compressed format

## IP.BIN Header Structure

The IP.BIN is located in the boot area (first 16 sectors = 32KB). The metadata header starts at offset 0x00:

| Offset | Size | Description                         |
| ------ | ---- | ----------------------------------- |
| 0x000  | 16   | Hardware ID ("SEGA SEGASATURN ")    |
| 0x010  | 16   | Maker ID (e.g., "SEGA ENTERPRISES") |
| 0x020  | 10   | Product Number                      |
| 0x02A  | 6    | Version                             |
| 0x030  | 8    | Release Date (YYYYMMDD)             |
| 0x038  | 8    | Device Info                         |
| 0x040  | 10   | Compatible Area Symbols             |
| 0x04A  | 6    | Reserved                            |
| 0x050  | 16   | Compatible Peripherals              |
| 0x060  | 112  | Game Title                          |
| 0x0D0  | 4    | Reserved                            |
| 0x0D4  | 4    | IP Size                             |
| 0x0D8  | 4    | Reserved                            |
| 0x0DC  | 4    | First Read File                     |
| 0x0E0  | 4    | First Read Size                     |
| 0x0E4  | 28   | Reserved                            |
| 0x100  | ...  | Security code and boot code         |

## Hardware ID

Must be exactly "SEGA SEGASATURN " (with trailing space) for valid Saturn discs.

## Area Symbols (Region Codes)

The Compatible Area Symbols field at offset 0x040 contains region flags:

| Symbol | Region           |
| ------ | ---------------- |
| J      | Japan            |
| T      | Asia (Taiwan)    |
| U      | USA              |
| B      | Brazil           |
| K      | Korea            |
| A      | PAL-A (Asia PAL) |
| E      | Europe           |
| L      | Latin America    |

Multiple regions can be specified (e.g., "JTUBKAEL" for all regions).

## Peripheral Codes

The Compatible Peripherals field indicates supported input devices:

- J = Joypad
- A = Analog controller
- K = Keyboard
- M = Mouse
- S = Steering wheel
- T = Multitap
- etc.

## Magic Detection

```go
var saturnMagic = []byte("SEGA SEGASATURN ")
var saturnMagicOffset = int64(0)

// For ISO images, the IP.BIN is at the start
// For CD audio tracks, need to handle multi-track discs
```

## Multi-Track Handling

Saturn games often have mixed-mode discs:

- Track 1: Data (IP.BIN + game data)
- Track 2+: Audio tracks

For .cue/.bin or .ccd/.img formats:

1. Parse the cue/ccd sheet
2. Locate the first data track
3. Read IP.BIN from data track offset

## CHD Format

For CHD files, use the existing CHD parsing to extract the first sectors, then parse IP.BIN.

## Extracted Metadata

- **Title**: 112-byte game title (may contain garbage after null)
- **Product Number**: 10-byte product code
- **Version**: 6-byte version string
- **Maker ID**: 16-byte publisher identifier
- **Release Date**: 8-byte YYYYMMDD format
- **Regions**: From area symbols
- **Peripherals**: Supported input devices

## Implementation Notes

1. **ISO vs multi-track** - Simple ISOs vs cue/bin require different handling
2. **Maker ID parsing** - May include "SEGA ENTERPRISES" or third-party names
3. **Title cleanup** - May need to trim at first null or garbage bytes
4. **CD-ROM sector format** - May be Mode 1 (2048 bytes) or Mode 2 (2352 bytes)

## Detection Strategy

1. For ISO/BIN files, check for "SEGA SEGASATURN " at offset 0
2. For Mode 2 raw sectors (2352 bytes), offset is different
3. Validate by checking hardware ID and maker ID format

## Test ROMs

- Saturn homebrew
- Public domain demos
- HomeBrew Saturn community releases

## Complexity: Medium

Straightforward header but multi-track disc handling adds complexity.
