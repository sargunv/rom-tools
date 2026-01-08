## screenscraper download game

Download game media

### Synopsis

Download game media (box art, logo, screenshot, etc.)

```
screenscraper download game [flags]
```

### Examples

```
  # Download game box art
  screenscraper download game --system=1 --game-id=3 --media="box-2D(us)" --output=box.png
  
  # Download game wheel logo
  screenscraper download game -s 1 -g 3 -m "wheel-hd(eu)" -o logo.png
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

* [screenscraper download](screenscraper_download.md)	 - Download media files

