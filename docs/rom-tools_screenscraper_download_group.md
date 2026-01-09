## rom-tools screenscraper download group

Download group media

### Synopsis

Download group media (genres, modes, families, themes, styles)

```
rom-tools screenscraper download group [flags]
```

### Examples

```
  # Download genre logo
  rom-tools screenscraper download group --group-id=1 --media="logo-monochrome" --output=genre.png
```

### Options

```
      --format string       Output format: png or jpg
  -g, --group-id string     Group ID (required)
  -h, --help                help for group
      --max-height string   Maximum height in pixels
      --max-width string    Maximum width in pixels
  -m, --media string        Media identifier (required, e.g. 'logo-monochrome')
  -o, --output string       Output file path (default: stdout)
```

### Options inherited from parent commands

```
      --json            Output results as JSON
      --locale string   Override locale for output (e.g., en, fr, de)
```

### SEE ALSO

- [rom-tools screenscraper download](rom-tools_screenscraper_download.md) - Download media files
