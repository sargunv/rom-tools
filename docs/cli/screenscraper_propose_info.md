## screenscraper propose info

Submit a text info proposal

### Synopsis

Submit a text info proposal for a game or ROM.

Game info types (--game-id): name, editeur, developpeur, players, score, rating, genres, datessortie, rotation, resolution, modes, familles, numero, styles, themes, description

ROM info types (--rom-id): developpeur, editeur, datessortie, players, regions, langues, clonetype, hacktype, friendly, serial, description

```
screenscraper propose info [flags]
```

### Examples

```
  # Add a game name for US region
  screenscraper propose info --game-id=123 --type=name --text="Super Mario Bros." --region=us

  # Add a synopsis in English
  screenscraper propose info --game-id=123 --type=description --text="A classic platformer..." --language=en

  # Add a ROM serial number
  screenscraper propose info --rom-id=456 --type=serial --text="SLUS-01234"
```

### Options

```
      --game-id string    Game ID to submit info for
  -h, --help              help for info
  -l, --language string   Language short name (required for description)
  -r, --region string     Region short name (required for name, datessortie)
      --rom-id string     ROM ID to submit info for
  -s, --source string     Source URL or info (optional)
      --text string       The text content
  -t, --type string       Info type (e.g. name, editeur, description)
  -v, --version string    Version (optional)
```

### Options inherited from parent commands

```
      --json            Output results as JSON
      --locale string   Override locale for output (e.g., en, fr, de)
```

### SEE ALSO

- [screenscraper propose](screenscraper_propose.md) - Submit proposals to ScreenScraper
