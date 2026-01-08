## screenscraper download company

Download company media

### Synopsis

Download company media (publishers, developers)

```
screenscraper download company [flags]
```

### Examples

```
  # Download company logo
  screenscraper download company --company-id=3 --media="logo-monochrome" --output=company.png
```

### Options

```
  -c, --company-id string   Company ID (required)
      --format string       Output format: png or jpg
  -h, --help                help for company
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

* [screenscraper download](screenscraper_download.md)	 - Download media files

