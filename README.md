# rom-tools

A CLI to scrape ROM metadata and media from Screenscraper, a community platform for retro video data.

## CLI

Install the CLI:

    go install github.com/sargunv/rom-tools/cmd/rom-tools

See the [CLI documentation](./docs/rom-tools.md) for complete usage information.

## Packages

Maturity legend:

- 🤖 Vibe coded with minimal human review, but it works on my machine.
- 🧹 AI generated, but reviewed for code style. May still have logic bugs of AI origin.
- 🏗️ Deeply reviewed for accuracy and style, or human-engineered from the start. Bugs here are of human origin only.

### API Clients

- 🏗️ [./lib/screenscraper](./lib/screenscraper): OpenAPI spec and generated client for the ScreenScraper API.

### Utilities

- 🤖 [./lib/identify](./lib/identify/): Utility to identify the title, serial, and other info of a ROM.
- 🏗️ [./lib/esde](./lib/esde): Implementation of the ES-DE gamelist.xml format.
- 🏗️ [./lib/datfile](./lib/datfile): Implementation of the Logiqx DAT XML format with No-Intro extensions.
- 🧹 [./lib/chd](./lib/chd): Implementation of the CHD (Compressed Hunks of Data) disc image format.

### ROM format implementations

- 🧹 [./lib/roms/dreamcast](./lib/roms/dreamcast): Sega Dreamcast disc identification from IP.BIN headers.
- 🏗️ [./lib/roms/gamecube](./lib/roms/gamecube): GameCube and Wii disc header parsing, including RVZ support.
- 🧹 [./lib/roms/gb](./lib/roms/gb): Game Boy and Game Boy Color ROM header parsing.
- 🧹 [./lib/roms/gba](./lib/roms/gba): Game Boy Advance ROM header parsing.
- 🤖 [./lib/roms/iso9660](./lib/roms/iso9660): ISO 9660 filesystem image parsing for optical disk platforms.
- 🧹 [./lib/roms/megadrive](./lib/roms/megadrive): Sega Mega Drive (Genesis) ROM header parsing, including SMD format.
- 🧹 [./lib/roms/n64](./lib/roms/n64): Nintendo 64 ROM parsing with support for Z64, V64, and N64 byte orders.
- 🧹 [./lib/roms/nds](./lib/roms/nds): Nintendo DS ROM header parsing.
- 🧹 [./lib/roms/nes](./lib/roms/nes): NES ROM parsing for iNES and NES 2.0 formats.
- 🧹 [./lib/roms/playstation_cnf](./lib/roms/playstation_cnf): PlayStation 1/2 SYSTEM.CNF parsing for disc identification.
- 🧹 [./lib/roms/playstation_sfo](./lib/roms/playstation_sfo): PlayStation SFO metadata format for PSP, PS3, PS Vita, and PS4.
- 🧹 [./lib/roms/saturn](./lib/roms/saturn): Sega Saturn disc identification from system area headers.
- 🧹 [./lib/roms/sms](./lib/roms/sms): Sega Master System and Game Gear ROM header parsing.
- 🧹 [./lib/roms/snes](./lib/roms/snes): Super Nintendo ROM header parsing with LoROM/HiROM detection.
- 🧹 [./lib/roms/xbox](./lib/roms/xbox): Original Xbox XBE executable and XISO disc image parsing.

## Test Data

ROM files in `**/testdata/` are sourced from:

- [XboxDev/cromwell](https://github.com/XboxDev/cromwell) (LGPL-2.1)
- [Zophar's Domain PD ROMs](https://www.zophar.net/pdroms/) (public domain)

These files are used as sample data for automated tests.
