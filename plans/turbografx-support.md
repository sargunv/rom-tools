# TurboGrafx-16 / PC Engine ROM Support Plan

## Overview

The TurboGrafx-16 (PC Engine in Japan) used HuCard cartridges and later CD-ROM. ROM files typically have no standardized header, making identification challenging.

## File Extensions

### HuCard (cartridge)

- `.pce` - PC Engine ROM
- `.sgx` - SuperGrafx ROM

### CD-ROM

- `.iso` / `.bin` - Disc image
- `.cue` + `.bin` - Cue sheet format
- `.chd` - MAME compressed format

## HuCard Format

**Important**: HuCard ROMs do NOT have a standard header. They are raw ROM dumps.

### Common ROM Sizes

- 256 KB (2 Mbit)
- 512 KB (4 Mbit)
- 768 KB (6 Mbit)
- 1 MB (8 Mbit)

### Memory Layout

- ROMs are mapped starting at address $0000 or $E000
- First bytes are typically executable code (6502 family)
- No embedded metadata (title, publisher, etc.)

## CD-ROM Format (PC Engine CD / TurboGrafx-CD)

CD games have more structure. The system disc contains:

### IPL (Initial Program Loader)

Located in the boot area of the disc:

| Offset | Size | Description                          |
| ------ | ---- | ------------------------------------ |
| 0x800  | 4    | "PC Engine CD-ROM SYSTEM" or similar |
| ...    | ...  | System-specific data                 |

### Track Layout

- Track 1: Usually audio (warning track)
- Track 2: Data track with IPL and game data
- Track 3+: May be audio or additional data

## SuperGrafx Detection

SuperGrafx games are rare (only 5 exclusive titles + 2 enhanced). Detection options:

1. Use database of known SuperGrafx game hashes
2. Check file extension (`.sgx`)
3. Community lists of SuperGrafx-only games

## Region Differences

TurboGrafx-16 and PC Engine had different:

1. Pin configurations (region lock via hardware)
2. Video output (NTSC-J vs NTSC-US)

The ROM content is often identical; region is typically determined by:

1. Filename/extension conventions
2. Database lookup by hash

## Magic Detection

No reliable magic bytes for HuCard ROMs. Detection relies on:

1. File extension (`.pce`, `.sgx`)
2. File size (common HuCard sizes)
3. Hash-based database lookup

For CD-ROM:

```go
// Check for PC Engine CD signature in boot sector
// Various strings exist: "PC Engine CD-ROM SYSTEM", etc.
```

## Extracted Metadata

Due to lack of headers, metadata extraction is limited:

### From ROM File

- **File size**: Only reliable intrinsic data
- **Hash**: For database lookup

### From Database (hash-based)

- Title
- Publisher
- Region
- Platform (PCE/TG16/SuperGrafx)
- Year

### From CD-ROM

- System type from boot sector
- Directory structure parsing for additional info

## Implementation Notes

1. **Hash-based identification** is the primary method
2. **No embedded metadata** in most ROMs
3. **CD-ROM support** requires ISO parsing
4. **SuperGrafx** is a separate platform designation

## Detection Strategy

1. Check extension (`.pce`, `.sgx`)
2. Validate file size against common HuCard sizes
3. Calculate hash for database lookup
4. For ISOs, check for PC Engine CD boot signature

## Database Requirement

Without headers, a hash-based database is essential:

```go
type PCEGameInfo struct {
    CRC32    string
    SHA1     string
    Title    string
    Platform string // "pce", "tg16", "sgx"
    Region   string // "jp", "us"
    Year     int
    Publisher string
}
```

Consider integrating with No-Intro or other DAT sources.

## Test ROMs

- PC Engine homebrew community exists
- Some public domain releases
- Test ROMs for basic validation

## Complexity: Low (ROM) / Medium (CD)

HuCard parsing is trivial (no header to parse), but meaningful identification requires external database. CD-ROM support adds ISO parsing complexity.

## Alternative Approach

Since metadata extraction is limited, consider:

1. Basic format detection only (PCE vs SGX vs CD)
2. Rely on filename and hash for identification
3. Integration with ScreenScraper for metadata lookup
