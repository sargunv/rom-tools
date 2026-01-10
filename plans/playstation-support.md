# PlayStation 1 (PSX) ROM Support Plan

## Overview

PlayStation 1 games were distributed on CD-ROM. Disc images contain game data with metadata in the SYSTEM.CNF file and potentially embedded in the executable.

## File Extensions

- `.bin` + `.cue` - Raw disc image with cue sheet
- `.iso` - ISO 9660 image
- `.img` + `.ccd` - CloneCD format
- `.mdf` + `.mds` - Alcohol 120% format
- `.pbp` - PSP EBOOT format (for PSP/PS3 compatibility)
- `.chd` - MAME compressed format
- `.ecm` - Error Code Modeler compressed

## Disc Structure

PS1 discs are standard CD-ROM with specific files:

```
/
├── SYSTEM.CNF       <- Boot configuration
├── SLUS_XXX.XX      <- Main executable (region-specific prefix)
└── [game files]
```

## SYSTEM.CNF Format

Plain text file with boot parameters:

```
BOOT = cdrom:\SLUS_012.34;1
TCB = 4
EVENT = 10
STACK = 801FFFF0
```

The `BOOT` line specifies the main executable, which contains the game ID.

## Game ID Format

The main executable filename IS the game ID:

| Prefix | Region                     |
| ------ | -------------------------- |
| SLUS   | USA (Sony)                 |
| SCUS   | USA (Sony - Greatest Hits) |
| SLPM   | Japan (third-party)        |
| SCPS   | Japan (Sony)               |
| SLES   | Europe (third-party)       |
| SCES   | Europe (Sony)              |
| SLKA   | Korea                      |

Format: `XXXX_XXX.XX` (e.g., "SLUS_012.34")

## Executable Header

PS1 executables (PS-X EXE) have a header:

| Offset | Size | Description                        |
| ------ | ---- | ---------------------------------- |
| 0x00   | 8    | Magic: "PS-X EXE"                  |
| 0x08   | 4    | Reserved                           |
| 0x0C   | 4    | Initial PC                         |
| 0x10   | 4    | Initial GP                         |
| 0x14   | 4    | Text section destination           |
| 0x18   | 4    | Text section size                  |
| 0x1C   | 8    | Reserved                           |
| 0x24   | 4    | BSS destination                    |
| 0x28   | 4    | BSS size                           |
| 0x2C   | 4    | Initial SP base                    |
| 0x30   | 4    | Initial SP offset                  |
| ...    | ...  | Reserved                           |
| 0x4C   | 1972 | ASCII region marker/licensing text |

The ASCII region at offset 0x4C often contains:

- "Licensed by Sony Computer Entertainment Inc."
- Region text ("for North America", "for Europe", etc.)

## PBP Format (PSP EBOOT)

See PSP support plan. PBP wraps PS1 disc images for PSP/PS3 playback.

### PBP Header

| Offset | Size | Description                 |
| ------ | ---- | --------------------------- |
| 0x00   | 4    | Magic: 0x00504250 ("\0PBP") |
| 0x04   | 4    | Version                     |
| 0x08   | 4    | Offset to PARAM.SFO         |
| 0x0C   | 4    | Offset to ICON0.PNG         |
| ...    | ...  | More offsets                |

PARAM.SFO contains DISC_ID with the game ID.

## Multi-Disc Games

PS1 had many multi-disc games. Each disc has:

- Same game ID prefix
- Different disc number in filename or content

## Magic Detection

```go
// For ISO/BIN: Look for SYSTEM.CNF in ISO filesystem
// Then parse the BOOT line

// For PS-X EXE:
var psxExeMagic = []byte("PS-X EXE")
var psxExeOffset = int64(0)

// For PBP:
var pbpMagic = []byte{0x00, 0x50, 0x42, 0x50}
var pbpOffset = int64(0)
```

## Extracted Metadata

- **Game ID**: From executable filename (SLUS_012.34)
- **Region**: From game ID prefix
- **Title**: From database lookup (not in disc)
- **Disc Number**: From multi-disc detection

## Implementation Notes

1. **ISO 9660 parsing** required to read SYSTEM.CNF
2. **Multi-track handling** for audio + data discs
3. **PBP support** leverages PSP's PARAM.SFO
4. **Title not embedded** - requires database lookup

## Detection Strategy

1. For BIN/ISO, parse ISO 9660 filesystem
2. Locate and read SYSTEM.CNF
3. Parse BOOT line for executable name
4. Extract game ID from executable filename
5. For PBP, parse header and PARAM.SFO

## Disc Image Handling

### CUE Sheet Parsing

```
FILE "game.bin" BINARY
  TRACK 01 MODE2/2352
    INDEX 01 00:00:00
  TRACK 02 AUDIO
    INDEX 01 03:00:00
```

Need to identify data track and offset.

### Mode 2 Sectors

PS1 often uses Mode 2 Form 1 (2352 bytes/sector with 2048 data). ISO offset calculations must account for this.

## Test ROMs

- PS1 homebrew (PS1 dev scene exists)
- Public domain demos
- Homebrew games

## Complexity: Medium-High

ISO parsing, multi-track handling, and lack of embedded title make this moderately complex.
