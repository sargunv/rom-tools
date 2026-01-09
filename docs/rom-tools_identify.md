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
  - CHD: by magic bytes
  - Xbox XISO: by magic bytes at offset 0x10000
  - ISO9660: by magic bytes
  - ZIP: by magic bytes

Hash modes:
  - Default: uses fast methods where available, calculates for loose files
  - --fast: skips hash calculation entirely
  - --slow: calculates full hashes even when fast methods are available

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

* [rom-tools](rom-tools.md)	 - ROM management and metadata tools

