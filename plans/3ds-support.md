# Nintendo 3DS ROM Support Plan

## Overview

Nintendo 3DS games use several container formats (CCI, CIA, 3DS) that wrap NCCH (Nintendo Content Container Header) files. The format is more complex than previous Nintendo handhelds.

## File Extensions

- `.3ds` / `.cci` - CTR Cart Image (game cartridge dump)
- `.cia` - CTR Importable Archive (installable format)
- `.cxi` - CTR Executable Image
- `.cfa` - CTR File Archive

## CCI/3DS Format Structure

CCI files use NCSD (Nintendo Content Storage Device) container:

### NCSD Header (0x000 - 0x1FF)

| Offset      | Size | Description                                       |
| ----------- | ---- | ------------------------------------------------- |
| 0x000       | 256  | RSA-2048 signature                                |
| 0x100       | 4    | Magic "NCSD"                                      |
| 0x104       | 4    | Image size in media units                         |
| 0x108       | 8    | Media ID                                          |
| 0x110       | 8    | Partition FS type                                 |
| 0x118       | 8    | Partition crypt type                              |
| 0x120       | 64   | Partition offset/size table (8 entries × 8 bytes) |
| 0x160       | 32   | Extended header hash                              |
| 0x180       | 4    | Additional header size                            |
| 0x184       | 4    | Sector zero offset                                |
| 0x188       | 8    | Partition flags                                   |
| 0x190       | 64   | Partition ID table                                |
| 0x1D0       | 32   | Reserved                                          |
| 0x1F0       | 1    | Reserved                                          |
| 0x1F1       | 1    | Backup write wait time                            |
| 0x1F2       | 1    | Backup security type                              |
| 0x1F3-0x1FF | 13   | Reserved                                          |

### NCCH Header (at each partition)

| Offset | Size | Description                 |
| ------ | ---- | --------------------------- |
| 0x000  | 256  | RSA-2048 signature          |
| 0x100  | 4    | Magic "NCCH"                |
| 0x104  | 4    | Content size in media units |
| 0x108  | 8    | Partition ID                |
| 0x110  | 2    | Maker Code                  |
| 0x112  | 2    | Version                     |
| 0x114  | 4    | Hash of extended header     |
| 0x118  | 8    | Program ID                  |
| 0x120  | 16   | Reserved                    |
| 0x130  | 32   | Logo region hash            |
| 0x150  | 16   | Product Code (ASCII)        |
| 0x160  | 32   | Extended header hash        |
| 0x180  | 4    | Extended header size        |
| 0x184  | 4    | Reserved                    |
| 0x188  | 8    | Flags                       |
| 0x190  | 4    | Plain region offset         |
| 0x194  | 4    | Plain region size           |
| 0x198  | 4    | Logo region offset          |
| 0x19C  | 4    | Logo region size            |
| 0x1A0  | 4    | ExeFS offset                |
| 0x1A4  | 4    | ExeFS size                  |
| 0x1A8  | 4    | ExeFS hash region size      |
| 0x1AC  | 4    | Reserved                    |
| 0x1B0  | 4    | RomFS offset                |
| 0x1B4  | 4    | RomFS size                  |
| 0x1B8  | 4    | RomFS hash region size      |
| 0x1BC  | 4    | Reserved                    |
| 0x1C0  | 32   | ExeFS superblock hash       |
| 0x1E0  | 32   | RomFS superblock hash       |

## Product Code Format

The product code (at NCCH offset 0x150) follows this pattern:

- `CTR-P-XXXX` for game cards
- `CTR-N-XXXX` for downloadable titles

Where XXXX is:

- Char 1-2: Unique identifier
- Char 3: Region code
- Char 4: Language variant

## Region Codes

| Code | Region      |
| ---- | ----------- |
| J    | Japan       |
| E    | USA         |
| P    | Europe      |
| A    | All regions |
| K    | Korea       |
| T    | Taiwan      |
| C    | China       |

## Magic Detection

1. Check for "NCSD" at offset 0x100 (for CCI/3DS files)
2. Check for "NCCH" at partition start

```go
var ncsdMagic = []byte("NCSD")
var ncsdMagicOffset = int64(0x100)

var ncchMagic = []byte("NCCH")
```

## CIA Format

CIA files have a different structure:

1. CIA Header (aligns to 64 bytes)
2. Certificate chain
3. Ticket
4. Title Metadata (TMD)
5. Content (NCCH files)

The CIA header doesn't contain game metadata directly - it must be extracted from the embedded NCCH.

## Extracted Metadata

- **Product Code**: From NCCH header
- **Maker Code**: 2-byte publisher code
- **Title ID/Program ID**: 8-byte unique identifier
- **Version**: From NCCH header
- **Region**: From product code

## Implementation Notes

1. **Encryption** - Most 3DS ROMs are encrypted; parsing unencrypted headers is simpler
2. **Multiple partitions** - CCI can contain up to 8 NCCH partitions
3. **Primary content** is usually partition 0
4. **Title in SMDH** - Actual game title is in the icon (SMDH) file within ExeFS

## Complexity: High

Complex nested format with encryption. Consider supporting only unencrypted dumps initially.

## Test ROMs

- 3DS homebrew (devkitPro)
- Homebrew games
- Test ROMs for validation
