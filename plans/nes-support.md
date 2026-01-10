# NES/Famicom ROM Support Plan

## Overview

The Nintendo Entertainment System (NES) and Famicom use the iNES file format for ROM distribution. This is a well-documented format with a 16-byte header.

## File Extensions

- `.nes` - iNES format (most common)
- `.unf` - UNIF format (less common)
- `.fds` - Famicom Disk System

## Header Format (iNES)

The iNES header is 16 bytes at offset 0x00:

| Offset    | Size | Description                                             |
| --------- | ---- | ------------------------------------------------------- |
| 0x00      | 4    | Magic: "NES" + 0x1A                                     |
| 0x04      | 1    | PRG-ROM size (16 KB units)                              |
| 0x05      | 1    | CHR-ROM size (8 KB units)                               |
| 0x06      | 1    | Flags 6: Mapper low nibble, mirroring, battery, trainer |
| 0x07      | 1    | Flags 7: Mapper high nibble, console type               |
| 0x08      | 1    | PRG-RAM size (8 KB units, 0 = 8KB for compatibility)    |
| 0x09      | 1    | Flags 9: TV system (0=NTSC, 1=PAL)                      |
| 0x0A-0x0F | 6    | Reserved (should be 0)                                  |

### Flags 6 (Byte 6)

- Bit 0: Mirroring (0=horizontal, 1=vertical)
- Bit 1: Battery-backed RAM
- Bit 2: 512-byte trainer present
- Bit 3: Four-screen VRAM
- Bits 4-7: Lower 4 bits of mapper number

### Flags 7 (Byte 7)

- Bits 0-1: Console type (0=NES, 1=Vs.System, 2=PlayChoice-10)
- Bits 2-3: NES 2.0 identifier (if == 2, this is NES 2.0 format)
- Bits 4-7: Upper 4 bits of mapper number

## Magic Detection

```go
var nesMagic = []byte{0x4E, 0x45, 0x53, 0x1A} // "NES" + 0x1A
var nesOffset = int64(0)
```

## Extracted Metadata

- **PRG-ROM Size**: Byte 4 × 16 KB
- **CHR-ROM Size**: Byte 5 × 8 KB
- **Mapper Number**: (Flags7 & 0xF0) | (Flags6 >> 4)
- **TV System**: NTSC or PAL from Flags 9
- **Battery**: Has save capability from Flags 6 bit 1

## Region Detection

NES ROMs don't have explicit region codes in the header. Region is determined by:

1. Flags 9 bit 0: 0 = NTSC, 1 = PAL
2. Often embedded in filename (USA, EUR, JPN)

## Implementation Notes

1. **No game title in header** - iNES format doesn't include game title
2. **Mapper identification** is useful metadata for emulation
3. **NES 2.0** is an extended format (check byte 7 bits 2-3 == 2)
4. Consider supporting UNIF format as secondary detection

## Test ROMs

Look for public domain NES homebrew:

- Homebrew games on itch.io
- NESdev competition entries
- Test ROMs from nesdev.org

## Complexity: Low-Medium

Simple header format, but no embedded title makes identification limited.
