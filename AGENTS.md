# AGENTS.md

## Project Overview

rom-tools is a Go CLI for scraping ROM metadata and media from Screenscraper, a community platform for retro video game data. It includes two main libraries:

- `lib/screenscraper`: API client for the Screenscraper API
- `lib/romident`: ROM identification engine supporting multiple platforms and formats

## Development Commands

**Important**: All commands must use `mise exec --` to ensure correct tool versions. Dev tools (Go, hk, pkl) are managed by mise.

Setup:

```bash
mise setup  # Install required tools and dependencies
```

Build and test:

```bash
mise build          # Build all packages: go build ./...
mise test           # Run tests (currently only lib/romident tests, screenscraper tests disabled)
mise exec -- go test ./lib/romident/...  # Run specific package tests
```

Run CLI:

```bash
mise rom-tools [args...]              # Run CLI via mise task
mise exec -- go run ./cmd/rom-tools   # Run CLI directly
```

Linting and formatting:

```bash
mise check  # Run linters (go fmt, gomod tidy, prettier) - runs hk check
mise fix    # Auto-fix formatting issues - runs hk fix with go fmt, gomod tidy, prettier
```

Generate documentation:

```bash
mise gen-docs  # Regenerate CLI docs in docs/ directory
```

## API Credentials

Screenscraper API requires credentials stored in `.env` (see `.env.example`):

- `SCREENSCRAPER_DEV_USER` and `SCREENSCRAPER_DEV_PASSWORD`: Developer credentials
- `SCREENSCRAPER_ID` and `SCREENSCRAPER_PASSWORD`: User credentials (optional)

## Architecture Overview

### lib/romident - ROM Identification Engine

The ROM identification system is modular and extensible:

**Core Components**:

- `rom.go`: Main entry point with `IdentifyROM()` function
- `detect.go`: Format detection via magic bytes and extensions
- `registry.go`: Central format registry mapping extensions to platforms
- `hash.go`: Multi-algorithm hash calculation (SHA1, MD5, CRC32) in single pass
- `types.go`: Core types (`ROM`, `ROMFile`, `GameIdent`, `Hash`)

**Format-Specific Handlers** (`lib/romident/[format]/`): Each format has its own package with an `Identify()` function that parses format-specific headers:

- Cartridge systems: `nes/`, `snes/`, `gba/`, `gb/`, `n64/`, `nds/`, `md/`, `smd/`
- Disc systems: `gcm/`, `xiso/`, `rvz/`, `ps2/`, `xbe/`
- Containers: `zip/`, `chd/`, `folder/`, `iso9660/`

**Identification Flow**:

1. Determine input type (file, ZIP, folder)
2. Select format candidates based on extension
3. Verify format via magic bytes or platform-specific identifiers
4. Extract platform-specific metadata from headers
5. Calculate hashes (supports fast/slow modes for containers)

**Hash Calculation Modes**:

- Default: Fast methods for containers (CHD headers, ZIP metadata), full hashing for loose files
- Fast (`--fast`): Skips large file hashing (>65MB)
- Slow (`--slow`): Decompresses containers to calculate full hashes

**Adding New Formats**:

1. Create new package under `lib/romident/[format]/`
2. Implement `Identify()` function matching the signature in `core/types.go`
3. Register format in `registry.go` with extensions and identification function

### lib/screenscraper - Screenscraper API Client

**Core Structure**:

- `client.go`: Main HTTP client with credential management and response validation
- `types.go`, `game.go`: Data models for API responses
- `download.go`: Media download endpoints (game, system, group, company media)
- `search.go`, `rating.go`, `proposal.go`: Search, rating, and submission operations
- `errors.go`: Custom error handling and API error parsing
- Enumeration files: `systems.go`, `genres.go`, `languages.go`, `regions.go`, etc.

**Client Design**:

- Stateless HTTP client with configurable timeouts
- URL building with proper parameter encoding
- Response validation via Header checking
- Error mapping from HTTP status codes and API response fields

### CLI Structure (cmd/rom-tools & internal/cli)

Built with Cobra framework:

**Top-Level Commands**:

- `identify`: Identify ROM files and extract metadata
  - Flags: `--json`, `--fast`, `--slow`
  - Uses `lib/romident` package
- `screenscraper`: Screenscraper API client with subcommands:
  - `detail game/system`: Get detailed metadata
  - `list [type]`: Reference data (systems, genres, languages, regions, media-types, etc.)
  - `download game/system/group/company`: Download media files
  - `status user/infra`: Check user quotas and server status
  - `search`: Search games by name
  - `propose info/media`: Submit metadata and media proposals
  - `rate`: Submit game ratings
- `scrape`: Batch ROM scraping (TODO: not implemented)

**Output Formatting** (`internal/format/`):

- `styles.go`: Terminal styling with lipgloss (colors, bold, tables)
- `render.go`: Complex rendering for Screenscraper data
- `locale.go`: Localization support with language selection
- `hyperlink.go`: Terminal hyperlink generation

### Shared Infrastructure

**internal/cli/screenscraper/shared/**: Global client and configuration state loaded from environment variables

**internal/util/**: Utility functions for string extraction from binary headers

## Data Flow

**ROM Identification**:

```
ROM file → IdentifyROM() → Detect format → Registry lookup
→ Format-specific Identify() → Extract metadata → Calculate hashes → Return ROM struct
```

**Screenscraper API**:

```
CLI command → shared.Client method → HTTP GET/POST to api.screenscraper.fr
→ Parse JSON → Format output → Display
```

## Key Design Patterns

1. **Format Registry Pattern**: Platform-specific handlers registered centrally for easy extensibility
2. **Separation of Concerns**: lib/ (business logic), internal/cli/ (UI), internal/format/ (presentation)
3. **Single Responsibility**: Each format handler only knows its own header format

## Tips

- When adding libraries, use `go get` to get the latest version
- When looking up how to use standard libraries or packages, use `go doc` to read the documentation
- Avoid many redundant tests. Err on the side of too few tests; we'll add more as needed.

## Test ROMs

The `testroms/` directory contains public domain test ROMs for various platforms used in tests. Do not commit copyrighted ROMs for testing. The user may place their legally dumped retail ROMs in the `tmp/` directory for local testing with real games. This directory is ignored by Git, so safe to place copyrighted ROMs here.
