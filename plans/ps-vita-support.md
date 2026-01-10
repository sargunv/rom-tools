# PlayStation Vita ROM Support Plan

## Overview

PS Vita games use multiple formats: game cards (3DS-like chips), digital downloads, and backup formats. The PARAM.SFO file is the primary metadata source.

## File Extensions

- `.vpk` - Vita Package (homebrew/HENkaku format, essentially a zip)
- `.pkg` - PlayStation Store package (encrypted)
- `.psvita` - NoNpDrm decrypted format
- `.mai` - MaiDump format
- Folder dumps with `sce_sys/param.sfo`

## VPK Format

VPK is a ZIP archive containing:

```
/
├── sce_sys/
│   ├── param.sfo      <- Game metadata
│   ├── icon0.png      <- Game icon
│   ├── livearea/      <- LiveArea assets
│   └── ...
├── eboot.bin          <- Main executable
└── [game data]
```

## PKG Format (Official)

PKG files are encrypted Sony packages:

### PKG Header

| Offset | Size | Description                   |
| ------ | ---- | ----------------------------- |
| 0x00   | 4    | Magic: 0x7F434E54 ("\x7FCNT") |
| 0x04   | 2    | Revision                      |
| 0x06   | 2    | Type                          |
| 0x08   | 4    | Metadata offset               |
| 0x0C   | 4    | Metadata count                |
| 0x10   | 4    | Metadata size                 |
| 0x14   | 4    | Item count                    |
| 0x18   | 8    | Total size                    |
| 0x20   | 8    | Data offset                   |
| 0x28   | 8    | Data size                     |
| 0x30   | 36   | Content ID                    |
| ...    | ...  | More fields                   |

The Content ID at offset 0x30 follows the format: `XXYYYY-TITLEID_00-AAAAAAAAAAAAAAAA`

## PARAM.SFO Format

Same format as PSP (see PSP plan). Key fields:

| Key         | Description                    |
| ----------- | ------------------------------ |
| TITLE       | Game title                     |
| TITLE_ID    | Game ID (e.g., "PCSE00000")    |
| CONTENT_ID  | Full content identifier        |
| APP_VER     | Application version            |
| VERSION     | SFO version                    |
| CATEGORY    | Type (gd=game, gp=patch, etc.) |
| PUBTOOLINFO | Build info                     |
| ATTRIBUTE   | Game attributes                |

### Parsing PARAM.SFO

See PSP plan for SFO header/entry structure. Same format used.

## Title ID Format

| Prefix | Region/Type         |
| ------ | ------------------- |
| PCSA   | Asia                |
| PCSE   | USA                 |
| PCSF   | Europe              |
| PCSB   | Europe (alt)        |
| PCSC   | Japan               |
| PCSD   | Japan (alt)         |
| PCSG   | Japan (third-party) |
| PCSH   | Asia (third-party)  |

## Content ID Format

`{REGION}{PUBLISHER}-{TITLEID}_00-{UNIQUE}`

Example: `UP0001-PCSE00120_00-0000000000000001`

## NoNpDrm / Mai Format

These are decrypted dump formats that maintain the folder structure:

```
TITLEID/
├── sce_sys/
│   └── param.sfo
├── eboot.bin
└── ...
```

## Magic Detection

```go
// For VPK: It's a ZIP file
var vpkMagic = []byte{0x50, 0x4B, 0x03, 0x04} // PK\x03\x04
// Then check for sce_sys/param.sfo inside

// For PKG:
var pkgMagic = []byte{0x7F, 0x43, 0x4E, 0x54} // \x7FCNT

// For PARAM.SFO:
var sfoMagic = []byte{0x00, 0x50, 0x53, 0x46} // \0PSF
```

## Extracted Metadata

- **Title**: From PARAM.SFO TITLE key
- **Title ID**: From TITLE_ID key (e.g., "PCSE00120")
- **Content ID**: Full identifier
- **Version**: From APP_VER
- **Region**: From Title ID prefix
- **Category**: Game type

## Implementation Notes

1. **VPK handling** - Standard ZIP extraction
2. **PKG complexity** - Encrypted, consider supporting only VPK initially
3. **Folder dumps** - Just locate and parse sce_sys/param.sfo
4. **Reuse PSP SFO parser** - Same format

## Detection Strategy

### VPK

1. Check ZIP magic
2. Check for sce_sys/param.sfo in zip
3. Parse PARAM.SFO

### PKG

1. Check PKG magic at offset 0
2. Parse Content ID from header
3. Decrypt if keys available (complex)

### Folder

1. Check for sce_sys/param.sfo
2. Parse PARAM.SFO

## Test ROMs

- Vita homebrew community is active
- VPK homebrew readily available
- Test apps for validation

## Complexity: Low (VPK) / High (PKG)

VPK is straightforward (ZIP + SFO). PKG requires decryption keys and is more complex.

## Recommended Approach

1. Implement VPK support first (most homebrew uses this)
2. Support folder dumps (decrypted backups)
3. Consider PKG support later if needed (requires crypto)
