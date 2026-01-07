# screenscraper-go

A Go client library and CLI for Screenscraper, a community platform for retro video game data and media.

## Installation

Install the CLI:

    go install sargunv/screenscraper-go/cmd/screenscraper@latest

Or use the library in your Go project:

    go get sargunv/screenscraper-go

## CLI Usage

See the [CLI documentation](docs/cli/screenscraper.md) for complete usage information.

Quick start:

- [Search for games](docs/cli/screenscraper_search.md)
- [Get game information](docs/cli/screenscraper_detail_game.md)
- [Download media files](docs/cli/screenscraper_download.md)
- [List metadata and reference data](docs/cli/screenscraper_list.md)

## Library Usage

    import "sargunv/screenscraper-go/client"

    c := client.NewClient(devID, devPassword, "my-app/1.0", ssID, ssPassword)
    game, err := client.GetGame(client.GetGameParams{GameID: "12345"})

## API Endpoint Implementation Status

### Core Information

- [x] `ssinfraInfos.php` - Infrastructure/server information
- [x] `ssuserInfos.php` - User information and quotas

### Metadata Lists

- [x] `regionsListe.php` - List of regions
- [x] `languesListe.php` - List of languages
- [x] `genresListe.php` - List of genres
- [x] `mediasSystemeListe.php` - List of system media types
- [x] `mediasJeuListe.php` - List of game media types
- [x] `systemesListe.php` - List of systems/consoles
- [x] `userlevelsListe.php` - List of user levels
- [x] `nbJoueursListe.php` - List of player counts
- [x] `supportTypesListe.php` - List of support types
- [x] `romTypesListe.php` - List of ROM types
- [x] `famillesListe.php` - List of families
- [x] `classificationsListe.php` - List of classifications
- [x] `infosJeuListe.php` - List of game info types
- [x] `infosRomListe.php` - List of ROM info types

### Game Data

- [x] `jeuRecherche.php` - Search for games by name
- [x] `jeuInfos.php` - Get detailed game information

### Media Downloads

- [x] `mediaJeu.php` - Download game media
- [x] `mediaSysteme.php` - Download system media
- [ ] ~~`mediaVideoSysteme.php` - Download system video media~~ (`mediaSysteme.php` seems to work for videos as well)
- [ ] ~~`mediaVideoJeu.php` - Download game video media~~ (`mediaJeu.php` seems to work for videos as well)
- [ ] ~~`mediaManuelJeu.php` - Download game manuals (PDF)~~ (`mediaJeu.php` seems to work for manuals as well)
- [x] `mediaGroup.php` - Download group media (genres, modes, etc.)
- [x] `mediaCompagnie.php` - Download company media

### Community Features

- [ ] `botNote.php` - Submit game ratings
- [ ] `botProposition.php` - Submit info/media proposals

## Credentials

You need a Screenscraper developer account. Register at https://www.screenscraper.fr.

User credentials are optional but provide higher rate limits.
