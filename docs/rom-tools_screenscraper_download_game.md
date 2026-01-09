## rom-tools screenscraper download game

Download game media

### Synopsis

Download game media (box art, logo, screenshot, etc.)

```
rom-tools screenscraper download game [flags]
```

### Examples

```
  # Download game box art
  rom-tools screenscraper download game --system=1 --game-id=3 --media="box-2D(us)" --output=box.png

  # Download game wheel logo
  rom-tools screenscraper download game -s 1 -g 3 -m "wheel-hd(eu)" -o logo.png
```

### Options

```
      --format string       Output format: png or jpg
  -g, --game-id string      Game ID (required)
  -h, --help                help for game
      --max-height string   Maximum height in pixels
      --max-width string    Maximum width in pixels
  -m, --media string        Media identifier (required, e.g. 'box-2D(us)', 'wheel-hd(eu)')
  -o, --output string       Output file path (default: stdout)
  -s, --system string       System ID (required)
```

### Options inherited from parent commands

```
      --json            Output results as JSON
      --locale string   Override locale for output (e.g., en, fr, de)
```

### SEE ALSO

- [rom-tools screenscraper download](rom-tools_screenscraper_download.md) - Download media files
