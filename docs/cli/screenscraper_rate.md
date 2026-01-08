## screenscraper rate

Submit a rating for a game

### Synopsis

Submit a rating (1-20) for a game on ScreenScraper.

This command requires user credentials (SCREENSCRAPER_ID and SCREENSCRAPER_PASSWORD).
Your rating will be associated with your ScreenScraper account.

```
screenscraper rate [flags]
```

### Examples

```
  # Rate a game 18 out of 20
  screenscraper rate --id=3 --rating=18
```

### Options

```
  -h, --help         help for rate
      --id string    Game ID to rate (required)
      --rating int   Rating from 1 to 20 (required)
```

### Options inherited from parent commands

```
      --json            Output results as JSON
      --locale string   Override locale for output (e.g., en, fr, de)
```

### SEE ALSO

* [screenscraper](screenscraper.md)	 - Screenscraper CLI client

