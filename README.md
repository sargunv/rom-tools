# screenscraper-go

A Go client library and CLI for Screenscraper, a community platform for retro video game data and media.

## Installation

Install the CLI:

    go install github.com/sargunv/screenscraper-go/cmd/screenscraper@latest

Or use the library in your Go project:

    go get github.com/sargunv/screenscraper-go

## CLI Usage

See the [CLI documentation](docs/cli/screenscraper.md) for complete usage information.

Quick start:

- [Search for games](docs/cli/screenscraper_search.md)
- [Get game information](docs/cli/screenscraper_detail_game.md)
- [Download media files](docs/cli/screenscraper_download.md)
- [List metadata and reference data](docs/cli/screenscraper_list.md)

## Library Usage

    import "github.com/sargunv/screenscraper-go/client"

    c := client.NewClient(devID, devPassword, "my-app/1.0", ssID, ssPassword)
    game, err := client.GetGame(client.GetGameParams{GameID: "12345"})

## API Client Status

- [x] Get infrastructure / user status
- [x] List metadata entries (regions, genres, systems, etc.)
- [x] Search for games and roms
- [x] Get detailed game information
- [x] Download game, system, group, and company media
- [x] Submit game ratings
- [x] Submit info / media proposals

## Credentials

You need a Screenscraper developer account. Register at https://www.screenscraper.fr.

User credentials are optional but provide higher rate limits.
