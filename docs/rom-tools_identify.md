## rom-tools identify

Identify ROM files and extract metadata

### Synopsis

Extract hashes and game identification data from ROM files.

Supports:

- Single files: calculates SHA1, MD5, CRC32
- ZIP archives: extracts CRC32 from metadata (fast, no decompression)
- CHD files: extracts SHA1 hashes from header (fast, no decompression)
- Folders: identifies all files within

Format detection:

- Loose files: by magic bytes (CHD, XISO, ISO9660, ZIP)
- ZIP contents: by extension (default), by magic bytes (--slow mode)
- Folders: by magic bytes for all files

Hash modes:

- Default: uses fast methods where available, calculates for loose files
- --fast: skips hash calculation for large loose files, but calculates for small loose files (<65MiB). ZIPs only use CRC32 from metadata (no decompression)
- --slow: calculates full hashes and enables format detection/identification for ZIP contents

```
rom-tools identify <file>... [flags]
```

### Options

```
      --fast   Skip hash calculation entirely
  -h, --help   help for identify
  -j, --json   Output results as JSON Lines (one JSON object per line)
      --slow   Calculate full hashes even for archives (requires decompression)
```

### SEE ALSO

- [rom-tools](rom-tools.md) - ROM management and metadata tools
