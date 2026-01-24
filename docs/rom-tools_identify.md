## rom-tools identify

Identify ROM files and extract metadata

### Synopsis

Extract hashes and game identification data from ROM files.

Supports:

- Platform specific ROMs: identifies game information from the ROM header. Supported formats:
  - Famicom (NES): .nes
  - Super Famicom (SNES): .sfc, .smc
  - Nintendo 64: .z64, .v64, .n64
  - Nintendo GameCube / Wii: .gcm, .iso, .rvz, .wia
  - Nintendo Game Boy / Color: .gb, .gbc
  - Nintendo Game Boy Advance: .gba
  - Nintendo DS: .nds, .dsi, .ids
  - Nintendo 3DS: .3ds, .cci
  - Sega Master System / Game Gear: .sms, .gg
  - Sega Mega Drive (Genesis): .md, .gen, .smd, .32x
  - Sega CD: .bin, .chd
  - Sega Saturn: .bin, .chd
  - Sega Dreamcast: .bin, .chd
  - Sony PlayStation 1: .bin, .chd
  - Sony PlayStation 2: .iso, .bin, .chd
  - Sony PlayStation 3: .pkg
  - Sony PlayStation Portable: .iso, .chd
  - Sony PlayStation Vita: .pkg
  - Microsoft Xbox: .iso, .chd, .xbe
- .chd discs: extracts SHA1 hashes from header (no decompression needed)
- .zip archives: extracts CRC32 hashes from metadata (no decompression needed)
- All files: calculates SHA1, MD5, CRC32 for uncompressed files under --max-hash-size
- All folders: identifies files within

```
rom-tools identify <file>... [flags]
```

### Options

```
  -h, --help                help for identify
  -j, --json                Output results as JSON Lines (one JSON object per line)
      --max-hash-size int   Max file size in bytes for hash calculation (-1 = no limit) (default 67108864)
```

### SEE ALSO

- [rom-tools](rom-tools.md) - ROM management and metadata tools
