# Sega Dreamcast ROM Support Plan

## Overview

Sega Dreamcast used GD-ROM (Gigabyte Disc), a proprietary high-density format. Images are commonly distributed in GDI or CDI formats, with CHD becoming more popular for archival.

## File Extensions

- `.gdi` + track files - GD-ROM Image format
- `.cdi` - DiscJuggler format (may be truncated/modified)
- `.chd` - MAME compressed format
- `.iso` - Sometimes used for single-track dumps

## GDI Format

GDI is a text descriptor file that references track files:

```
3
1 0 4 2352 track01.bin 0
2 600 0 2352 track02.raw 0
3 45000 4 2352 track03.bin 0
```

Format:

- Line 1: Number of tracks
- Lines 2+: Track definitions

Track definition fields:

1. Track number
2. Starting LBA (Logical Block Address)
3. Track type (0=audio, 4=data)
4. Sector size (2048 or 2352)
5. Filename
6. Offset (usually 0)

## IP.BIN Header (Dreamcast)

Similar to Saturn, the metadata is in IP.BIN at the start of the last data track:

| Offset | Size | Description                          |
| ------ | ---- | ------------------------------------ |
| 0x000  | 16   | Hardware ID ("SEGA SEGAKATANA ")     |
| 0x010  | 16   | Maker ID                             |
| 0x020  | 10   | Product Number                       |
| 0x02A  | 6    | Version                              |
| 0x030  | 8    | Release Date (YYYYMMDD)              |
| 0x038  | 8    | Device Info ("GD-ROM " or "CD-ROM ") |
| 0x040  | 8    | Area Symbols (regions)               |
| 0x048  | 8    | Peripherals                          |
| 0x050  | 10   | Product Number (duplicate)           |
| 0x05A  | 6    | Version (duplicate)                  |
| 0x060  | 16   | Boot Filename                        |
| 0x070  | 16   | Publisher Name                       |
| 0x080  | 128  | Game Title                           |

## Hardware ID

Valid Dreamcast discs have "SEGA SEGAKATANA " (the codename was "Katana").

## Area Symbols (Region Codes)

| Symbol | Region |
| ------ | ------ |
| J      | Japan  |
| U      | USA    |
| E      | Europe |

Example: "JUE " = all regions

## Peripheral Codes

| Code | Device                              |
| ---- | ----------------------------------- |
| 0    | Uses Windows CE                     |
| 1    | VGA supported                       |
| 4    | Start + A + B + Triggers controller |
| 5    | Analog controller supported         |
| 6    | Analog controller required          |
| E    | Jump Pack supported                 |
| F    | Jump Pack required                  |
| etc. |

## GDI Parsing Strategy

1. Read the GDI text file
2. Parse track definitions
3. Find the high-density data track (usually track 3, starting around LBA 45000)
4. Read IP.BIN from the start of that track

## CDI Format

CDI is a proprietary DiscJuggler format. Structure:

1. Track data (concatenated)
2. CDI header at end of file

CDI parsing is more complex - consider using a library or only supporting GDI/CHD.

## CHD Format

For CHD files:

1. Use CHD metadata to identify track layout
2. Locate the high-density data area
3. Extract and parse IP.BIN

## Magic Detection

```go
var dreamcastMagic = []byte("SEGA SEGAKATANA ")
var dreamcastMagicOffset = int64(0)

// For GDI: Parse the .gdi file, locate IP.BIN in main data track
// For CHD: Use CHD track info
```

## Extracted Metadata

- **Title**: 128-byte game title
- **Product Number**: 10-byte product code
- **Version**: 6-byte version string
- **Maker ID**: Publisher name
- **Release Date**: YYYYMMDD format
- **Regions**: From area symbols
- **Device Info**: GD-ROM or CD-ROM indicator
- **Boot File**: Main executable name

## Implementation Notes

1. **GDI is preferred** - Most accurate format
2. **CDI may be truncated** - Some CDI dumps remove high-density area
3. **Track layout** - Need to find correct data track
4. **IP.BIN location** varies between GD-ROM and CD-R burns

## Detection Strategy

1. For GDI: Parse .gdi, find main data track, read IP.BIN
2. For CHD: Parse CHD metadata, locate data track
3. For ISO: Check for "SEGA SEGAKATANA " at offset 0
4. Validate hardware ID

## Test ROMs

- Dreamcast homebrew (plenty available)
- Public domain demos
- Homebrew competitions

## Complexity: High

Complex due to multi-format support (GDI, CDI, CHD) and track layout handling.
