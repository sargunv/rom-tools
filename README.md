# screenscraper-go

A Go client library and CLI for Screenscraper, a community platform for retro video game data and media.

## Installation

Install the CLI:

    go install sargunv/screenscraper-go/cmd/screenscraper@latest

Or use the library in your Go project:

    go get sargunv/screenscraper-go

## CLI Usage

```bash
screenscraper --help
```

```
A CLI client for the Screenscraper API to fetch game metadata and media.

Credentials are loaded from environment variables:
  SCREENSCRAPER_DEV_USER     - Developer username
  SCREENSCRAPER_DEV_PASSWORD - Developer password
  SCREENSCRAPER_ID           - User ID (optional)
  SCREENSCRAPER_PASSWORD     - User password (optional)

Usage:
  screenscraper [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  download    Download media files
  game        Get game information
  genres      Get list of genres
  help        Help about any command
  infra       Get infrastructure/server information
  languages   Get list of languages
  media-types Get list of media types
  regions     Get list of regions
  search      Search for games by name
  systems     Get list of systems/consoles
  user        Get user information and quotas

Flags:
      --dev-id string          Developer ID (or set SCREENSCRAPER_DEV_USER)
      --dev-password string    Developer password (or set SCREENSCRAPER_DEV_PASSWORD)
  -h, --help                   help for scraper
      --soft-name string       Software name identifier
      --user-id string         User ID (or set SCREENSCRAPER_ID)
      --user-password string   User password (or set SCREENSCRAPER_PASSWORD)

Use "scraper [command] --help" for more information about a command.
```

## Library Usage

    import "sargunv/screenscraper-go/client"

    c := client.NewClient(devID, devPassword, "my-app/1.0", ssID, ssPassword)
    game, err := client.GetGame(client.GetGameParams{GameID: "12345"})

## Credentials

You need a Screenscraper developer account. Register at https://www.screenscraper.fr.

User credentials are optional but provide higher rate limits.
