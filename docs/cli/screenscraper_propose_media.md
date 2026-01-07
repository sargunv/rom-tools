## screenscraper propose media

Submit a media proposal

### Synopsis

Submit a media proposal for a game or ROM.

Media types: sstitle, ss, fanart, video, overlay, steamgrid, wheel, wheel-hd, marquee, screenmarquee, box-2D, box-2D-side, box-2D-back, box-texture, manuel, flyer, maps, figurine, support-texture, box-scan, support-scan, bezel-4-3, bezel-4-3-v, bezel-4-3-cocktail, bezel-16-9, bezel-16-9-v, bezel-16-9-cocktail, wheel-tarcisios, videotable, videotable4k, themehs, themehb

You can provide the media either as a file (--file) or URL (--url). Use --file=- to read from stdin.

```
screenscraper propose media [flags]
```

### Examples

```
  # Upload box art from file
  screenscraper propose media --game-id=123 --type=box-2D --file=box_us.png --region=us

  # Submit media from URL
  screenscraper propose media --game-id=123 --type=wheel --url="https://example.com/logo.png" --region=eu

  # Upload screenshot from stdin
  cat screenshot.jpg | screenscraper propose media --game-id=123 --type=ss --file=- --region=us
```

### Options

```
  -f, --file string          File path to upload (use '-' for stdin)
      --game-id string       Game ID to submit media for
  -h, --help                 help for media
  -r, --region string        Region (required for ss, sstitle, wheel, box-*, bezel-*, etc.)
      --rom-id string        ROM ID to submit media for
  -s, --source string        Source URL or info (optional)
  -n, --support-num string   Support number 0-10 (required for box-*, flyer, support-*)
  -t, --type string          Media type (e.g. ss, box-2D, wheel)
  -u, --url string           URL of media to download
  -v, --version string       Version (for maps, box-scan, support-scan)
```

### Options inherited from parent commands

```
      --json            Output results as JSON
      --locale string   Override locale for output (e.g., en, fr, de)
```

### SEE ALSO

- [screenscraper propose](screenscraper_propose.md) - Submit proposals to ScreenScraper
