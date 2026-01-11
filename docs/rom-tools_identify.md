## rom-tools identify

Identify ROM files and extract metadata

### Synopsis

Extract hashes and game identification data from ROM files.

Supports:

- Platform specific ROMs: identifies game information from the ROM header, depending on the format. Supported ROM formats:
  - Nintendo Entertainment System: .nes
  - Super Nintendo Entertainment System: .sfc, .smc
  - Nintendo 64: .z64, .v64, .n64
  - Nintendo Game Boy / Color: .gb, .gbc
  - Nintendo Game Boy Advance: .gba
  - Nintendo DS: .nds, .dsi, .ids
  - Sega Mega Drive / Genesis: .md, .gen, .smd
  - Microsoft Xbox: .xiso, .xiso.iso, and .xbe
- .chd discs: extracts SHA1 hashes from header (fast, no decompression)
- .zip archives: extracts CRC32 from metadata (fast, no decompression). If in slow mode, also identifies files within the ZIP.
- all files: calculates SHA1, MD5, CRC32 (unless in fast mode).
- all folders: identifies files within.

```
rom-tools identify <file>... [flags]
```

### Options

```
      --fast   Skip hash calculation for large loose files, but calculates for small loose files (<65MiB).
  -h, --help   help for identify
  -j, --json   Output results as JSON Lines (one JSON object per line)
      --slow   Calculate full hashes and identify games inside archives (requires decompression).
```

### SEE ALSO

- [rom-tools](rom-tools.md) - ROM management and metadata tools
