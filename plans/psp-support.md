# PlayStation Portable (PSP) ROM Support Plan

## Overview

PSP games are distributed as UMD (Universal Media Disc) images or in compressed formats. Game metadata is stored in multiple locations within the disc structure.

## File Extensions

- `.iso` - Uncompressed UMD image
- `.cso` - Compressed ISO (CISO format)
- `.pbp` - PlayStation Portable Bundle (EBOOT format)
- `.dax` - DAX compressed format

## ISO/UMD Structure

PSP UMD images are ISO 9660 filesystems with a specific directory structure:

```
/
├── PSP_GAME/
│   ├── PARAM.SFO      <- Game metadata
│   ├── ICON0.PNG      <- Game icon
│   ├── PIC0.PNG       <- Background image
│   ├── PIC1.PNG       <- Background image (alt)
│   ├── SYSDIR/
│   │   └── EBOOT.BIN  <- Main executable
│   └── USRDIR/        <- Game data
└── UMD_DATA.BIN       <- UMD metadata
```

## UMD_DATA.BIN Format

Located at root of the disc, contains basic identification:

| Offset | Size | Description                 |
| ------ | ---- | --------------------------- | --- |
| 0x00   | 10   | Game ID (e.g., "UCUS98645") |
| 0x0A   | 1    | Pipe separator "            | "   |
| 0x0B   | var  | Additional info             |

## PARAM.SFO Format

The PARAM.SFO file is a structured binary format containing key-value pairs:

### Header (20 bytes)

| Offset | Size | Description                  |
| ------ | ---- | ---------------------------- |
| 0x00   | 4    | Magic: 0x00505346 ("\0PSF")  |
| 0x04   | 4    | Version (usually 0x00000101) |
| 0x08   | 4    | Key table offset             |
| 0x0C   | 4    | Data table offset            |
| 0x10   | 4    | Number of entries            |

### Index Entry (16 bytes each)

| Offset | Size | Description                             |
| ------ | ---- | --------------------------------------- |
| 0x00   | 2    | Key offset (relative to key table)      |
| 0x02   | 1    | Data alignment                          |
| 0x03   | 1    | Data type (0=binary, 2=string, 4=int32) |
| 0x04   | 4    | Data size used                          |
| 0x08   | 4    | Data size total                         |
| 0x0C   | 4    | Data offset (relative to data table)    |

### Important Keys

| Key            | Description                              |
| -------------- | ---------------------------------------- |
| DISC_ID        | Game ID (e.g., "UCUS98645")              |
| DISC_VERSION   | Version string                           |
| TITLE          | Game title                               |
| TITLE_xx       | Localized title (e.g., TITLE_JP)         |
| CATEGORY       | Type (UG=UMD Game, MG=Memory Stick Game) |
| REGION         | Region code                              |
| PARENTAL_LEVEL | Age rating                               |

## Game ID Format

PSP Game IDs follow this pattern: `XXXX99999`

| Prefix | Region             |
| ------ | ------------------ |
| UCUS   | USA                |
| ULUS   | USA (alternate)    |
| UCES   | Europe             |
| ULES   | Europe (alternate) |
| UCJS   | Japan              |
| ULJS   | Japan (alternate)  |
| UCAS   | Asia               |
| ULAS   | Asia (alternate)   |
| NPUH   | USA (PSN)          |
| NPEH   | Europe (PSN)       |
| NPJH   | Japan (PSN)        |

## CSO (Compressed ISO) Format

### Header (24 bytes)

| Offset | Size | Description               |
| ------ | ---- | ------------------------- |
| 0x00   | 4    | Magic "CISO"              |
| 0x04   | 4    | Header size (usually 24)  |
| 0x08   | 8    | Uncompressed size         |
| 0x10   | 4    | Block size (usually 2048) |
| 0x14   | 1    | Version                   |
| 0x15   | 1    | Alignment                 |
| 0x16   | 2    | Reserved                  |

After header: block index table pointing to compressed blocks.

## Magic Detection

```go
// For ISO: Check for PSP_GAME directory or UMD_DATA.BIN
// For CSO:
var csoMagic = []byte("CISO")
var csoOffset = int64(0)

// For PARAM.SFO:
var sfoMagic = []byte{0x00, 0x50, 0x53, 0x46} // "\0PSF"
```

## Extracted Metadata

- **Title**: From PARAM.SFO TITLE key
- **Game ID**: From DISC_ID or UMD_DATA.BIN
- **Version**: From DISC_VERSION
- **Region**: Derived from game ID prefix
- **Category**: Game type (UMD, Memory Stick, etc.)

## Implementation Notes

1. **ISO 9660 parsing** required to read PARAM.SFO
2. **CSO decompression** needed for compressed images
3. **PARAM.SFO parser** is reusable for PS Vita
4. **Multiple metadata sources** - UMD_DATA.BIN and PARAM.SFO

## Implementation Strategy

1. Detect format (ISO vs CSO)
2. For CSO, decompress header area or index
3. Parse ISO 9660 to locate PSP_GAME/PARAM.SFO
4. Parse PARAM.SFO for metadata
5. Fallback to UMD_DATA.BIN if PARAM.SFO unavailable

## Test ROMs

- PSP homebrew from psp-hacks
- Homebrew games
- Scene demos

## Complexity: Medium-High

Requires ISO 9660 parsing and PARAM.SFO parsing. CSO support adds compression handling.
