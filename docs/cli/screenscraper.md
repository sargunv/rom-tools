## screenscraper

Screenscraper CLI client

### Synopsis

A CLI client for the Screenscraper API to fetch game metadata and media.

Credentials are loaded from environment variables:

- SCREENSCRAPER_DEV_USER - Developer username
- SCREENSCRAPER_DEV_PASSWORD - Developer password
- SCREENSCRAPER_ID - User ID (optional)
- SCREENSCRAPER_PASSWORD - User password (optional)

### Options

```
  -h, --help            help for screenscraper
      --json            Output results as JSON
      --locale string   Override locale for output (e.g., en, fr, de)
```

### SEE ALSO

- [screenscraper detail](screenscraper_detail.md) - Get detailed information about a specific item
- [screenscraper download](screenscraper_download.md) - Download media files
- [screenscraper list](screenscraper_list.md) - List metadata and reference data
- [screenscraper propose](screenscraper_propose.md) - Submit proposals to ScreenScraper
- [screenscraper rate](screenscraper_rate.md) - Submit a rating for a game
- [screenscraper search](screenscraper_search.md) - Search for games by name
- [screenscraper status](screenscraper_status.md) - Get status information
