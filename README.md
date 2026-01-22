# rom-tools

A CLI to scrape ROM metadata and media from Screenscraper, a community platform for retro video data.

## CLI

Install the CLI:

    go install github.com/sargunv/rom-tools/cmd/rom-tools

See the [CLI documentation](./docs/rom-tools.md) for complete usage information.

## Libraries

- API Clients
  - [./lib/screenscraper](./lib/screenscraper): OpenAPI spec and generated client for the ScreenScraper API.
- Utilities
  - [./lib/identify](./lib/identify/): Utility to identify the title, serial, and other info of a ROM.
  - [./lib/esde](./lib/esde): Implementation of the ESDE gamelist.xml format.
  - [./lib/datfile](./lib/datfile): Implementation of the Logiqx DAT XML format with No-Intro extensions.
- ROM format implementations:
  - [./lib/format/chd](./lib/format/chd): Implementation of the CHD (Compressed Hunks of Data) disc image format.
  - [./lib/format/dreamcast](./lib/format/dreamcast): Sega Dreamcast disc identification from IP.BIN headers.
  - [./lib/format/folder](./lib/format/folder): Directory-based ROM container.
  - [./lib/format/gamecube](./lib/format/gamecube): GameCube and Wii disc header parsing, including RVZ support.
  - [./lib/format/gb](./lib/format/gb): Game Boy and Game Boy Color ROM header parsing.
  - [./lib/format/gba](./lib/format/gba): Game Boy Advance ROM header parsing.
  - [./lib/format/iso9660](./lib/format/iso9660): ISO 9660 filesystem image parsing for optical disk platforms.
  - [./lib/format/megadrive](./lib/format/megadrive): Sega Mega Drive (Genesis) ROM header parsing, including SMD format.
  - [./lib/format/n64](./lib/format/n64): Nintendo 64 ROM parsing with support for Z64, V64, and N64 byte orders.
  - [./lib/format/nds](./lib/format/nds): Nintendo DS ROM header parsing.
  - [./lib/format/nes](./lib/format/nes): NES ROM parsing for iNES and NES 2.0 formats.
  - [./lib/format/playstation_cnf](./lib/format/playstation_cnf): PlayStation 1/2 SYSTEM.CNF parsing for disc identification.
  - [./lib/format/playstation_sfo](./lib/format/playstation_sfo): PlayStation SFO metadata format for PSP, PS3, PS Vita, and PS4.
  - [./lib/format/saturn](./lib/format/saturn): Sega Saturn disc identification from system area headers.
  - [./lib/format/sms](./lib/format/sms): Sega Master System and Game Gear ROM header parsing.
  - [./lib/format/snes](./lib/format/snes): Super Nintendo ROM header parsing with LoROM/HiROM detection.
  - [./lib/format/xbox](./lib/format/xbox): Original Xbox XBE executable and XISO disc image parsing.
  - [./lib/format/zip](./lib/format/zip): ZIP archive ROM container.

## Test Data

ROM files in `**/testdata/` are sourced from:

- [XboxDev/cromwell](https://github.com/XboxDev/cromwell) (LGPL-2.1)
- [Zophar's Domain PD ROMs](https://www.zophar.net/pdroms/) (public domain)

These files are used as sample data for automated tests.
