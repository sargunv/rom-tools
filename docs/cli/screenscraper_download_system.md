## screenscraper download system

Download system media

### Synopsis

Download system media (logo, wheel, photos, etc.)

```
screenscraper download system [flags]
```

### Examples

```
  # Download system wheel logo
  screenscraper download system --system=1 --media="wheel(wor)" --output=system.png
```

### Options

```
      --format string       Output format: png or jpg
  -h, --help                help for system
      --max-height string   Maximum height in pixels
      --max-width string    Maximum width in pixels
  -m, --media string        Media identifier (required)
  -o, --output string       Output file path (default: stdout)
  -s, --system string       System ID (required)
```

### Options inherited from parent commands

```
      --json            Output results as JSON
      --locale string   Override locale for output (e.g., en, fr, de)
```

### SEE ALSO

* [screenscraper download](screenscraper_download.md)	 - Download media files

