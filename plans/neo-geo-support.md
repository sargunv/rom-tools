# Neo Geo ROM Support Plan

## Overview

The Neo Geo platform includes the arcade MVS (Multi Video System) and home AES (Advanced Entertainment System). ROMs consist of multiple files for different hardware components.

## File Extensions / ROM Sets

Neo Geo ROMs are typically distributed as ROM sets (zip files containing multiple files):

- `.zip` - Compressed ROM set (most common)
- Individual ROM files: `.p1`, `.p2`, `.s1`, `.c1`, `.c2`, `.v1`, `.v2`, `.m1`, etc.

## ROM Set Structure

A Neo Geo ROM set contains multiple files:

| File Type              | Content            | Processor |
| ---------------------- | ------------------ | --------- |
| P-ROM (.p1, .p2, .sp2) | Program code       | MC68000   |
| M-ROM (.m1)            | Sound program      | Z80       |
| V-ROM (.v1, .v2, ...)  | ADPCM samples      | YM2610    |
| C-ROM (.c1, .c2, ...)  | Sprite graphics    | GPU       |
| S-ROM (.s1)            | Fix layer graphics | GPU       |

## ROM Naming Convention

MAME naming convention: `gamename.zip` containing:

- `gamename/p1.rom` or `000-p1.p1`
- etc.

Older naming used game-specific prefixes: `201-p1.bin` (game number 201)

## P-ROM Header

The P-ROM (program ROM) contains a header at the start:

| Offset | Size | Description                    |
| ------ | ---- | ------------------------------ |
| 0x000  | 256  | Interrupt vectors (MC68000)    |
| 0x100  | 2    | Software DIPs settings         |
| 0x102  | 1    | Unknown                        |
| 0x103  | 1    | Unknown                        |
| 0x104  | 4    | Pointer to logo data           |
| 0x108  | 4    | Pointer to fix tile data       |
| 0x10C  | 4    | Pointer to sprite/palette data |
| 0x110  | 4    | Pointer to eye-catcher data    |
| 0x114  | 4    | Japanese game name             |
| 0x134  | 4    | English game name              |
| 0x154  | 2    | NGH number (game identifier)   |
| 0x156  | ...  | Additional data                |

## NGH Number

The NGH (Neo Geo Home) number at offset 0x154 is a unique 16-bit identifier:

- Used to identify games
- Determines save data slot
- Can be used for database lookup

## Game Name Pointers

At offsets 0x114 and 0x134, there are pointers to:

- Japanese game name string
- English game name string

These strings are typically at the end of the P-ROM vector table area.

## MVS vs AES

Differences are minimal at the ROM level:

- Same ROM data
- System mode detected via BIOS
- Some games have MVS-only or AES-only features

## Magic Detection

No simple magic bytes. Detection approaches:

1. **ROM set structure**: Identify Neo Geo by presence of P-ROM + M-ROM + V-ROM + C-ROM files
2. **P-ROM validation**: Check MC68000 vector table format
3. **Filename patterns**: "p1", "m1", "s1", "c1", "v1" in zip

## ROM Set Handling

For zip-based ROM sets:

1. List contents of zip
2. Identify P-ROM file (contains .p1 or p1 in name)
3. Read P-ROM header for game info
4. Extract NGH number and game names

## Extracted Metadata

- **NGH Number**: Primary game identifier
- **Japanese Title**: From P-ROM header pointer
- **English Title**: From P-ROM header pointer
- **ROM Set Name**: From filename (e.g., "mslug")

## Implementation Notes

1. **ROM set format** - Usually zipped collections
2. **Multiple files** per game - Need to identify main P-ROM
3. **MAME naming** is de facto standard
4. **No region distinction** - Same ROMs worldwide

## Detection Strategy

1. Check if zip contains Neo Geo ROM set structure
2. Locate P-ROM file (largest file or .p1 extension)
3. Read P-ROM header
4. Extract NGH number and game name pointers
5. Read game names from P-ROM

## Alternative: Database Approach

Given complexity, consider:

1. Identify as Neo Geo by ROM set structure
2. Use filename/romset name for database lookup
3. Retrieve full metadata from external source

## Test ROMs

- Neo Geo homebrew exists
- Public domain demos
- Fan-made games

## Complexity: Medium-High

Multiple files per game and pointer-based metadata add complexity. Database integration may be more practical than deep parsing.
