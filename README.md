# rom-tools

A suite of utilities for working with video game ROMs for classic platforms.

The project is a work in progress and portions are AI-generated. See the maturity legend to understand the quality of the components listed in this README:

- 🔴 Vibe coded AI slop with little or no human review; very rough
- 🟡 Reviewed or written by a human; experimental but usable
- 🟢 Tested on my real game library; no longer slop

## CLI

Install the CLI:

    go install github.com/sargunv/rom-tools/cmd/rom-tools

- 🔴 `rom-tools screenscraper`: CLI client for the ScreenScraper API.
- 🔴 `rom-tools identify`: Hash roms and parse their metadata.
- 🔴 `rom-tools scrape`: Scrape metadata for frontends from a list of roms.

See the [CLI documentation](./docs/rom-tools.md) for complete usage information.

## Packages

### Metadata sources

- 🟡 [./lib/screenscraper](./lib/screenscraper): OpenAPI spec and generated client for the ScreenScraper API.
- Hasheous: TODO
- Launchbox: TODO

### Metadata destinations

- 🟡 [./lib/esde](./lib/esde): Implementation of the ES-DE gamelist.xml format.
- MuOS: TODO
- MinUI/NextUI: TODO

### General utilities

- 🔴 [./lib/identify](./lib/identify/): Utility to identify the title, serial, and other info of a ROM.
- 🟢 [./lib/datfile](./lib/datfile): Implementation of the Logiqx DAT XML format with No-Intro extensions.
- 🟡 [./lib/chd](./lib/chd): Implementation of the CHD (Compressed Hunks of Data) disc image format.
- 🟡 [./lib/iso9660](./lib/iso9660): ISO 9660 filesystem image parsing for optical disk platforms.

### Nintendo formats

- 🟢 [./lib/roms/nes](./lib/roms/nes): NES ROM parsing for iNES and NES 2.0 formats.
- 🟢 [./lib/roms/snes](./lib/roms/snes): Super Nintendo ROM header parsing with LoROM/HiROM detection.
- 🟢 [./lib/roms/n64](./lib/roms/n64): Nintendo 64 ROM parsing with support for Z64, V64, and N64 byte orders.
- 🟡 [./lib/roms/gamecube](./lib/roms/gamecube): GameCube and Wii disc header parsing, including RVZ support.
- 🟢 [./lib/roms/gb](./lib/roms/gb): Game Boy and Game Boy Color ROM header parsing.
- 🟢 [./lib/roms/gba](./lib/roms/gba): Game Boy Advance ROM header parsing.
- 🔴 [./lib/roms/nds](./lib/roms/nds): Nintendo DS ROM header parsing.
- 3DS: [TODO](https://github.com/sargunv/rom-tools/issues/18)
- Wii U: [TODO](https://github.com/sargunv/rom-tools/issues/25)

### Sega formats

- 🟢 [./lib/roms/sms](./lib/roms/sms): Sega Master System and Game Gear ROM header parsing.
- 🟢 [./lib/roms/megadrive](./lib/roms/megadrive): Sega Mega Drive (Genesis) ROM header parsing, including SMD format, 32X detection, and Sega CD disc identification.
- 🟢 [./lib/roms/saturn](./lib/roms/saturn): Sega Saturn disc identification from system area headers.
- 🟢 [./lib/roms/dreamcast](./lib/roms/dreamcast): Sega Dreamcast disc identification from IP.BIN headers.

### Sony formats

- 🟢 [./lib/roms/playstation_cnf](./lib/roms/playstation_cnf): PlayStation 1/2 SYSTEM.CNF parsing for disc identification.
- 🟢 [./lib/roms/playstation_sfo](./lib/roms/playstation_sfo): PlayStation SFO metadata format for PSP, PS3, PS Vita, and PS4.
- PS3: [TODO](https://github.com/sargunv/rom-tools/issues/27)
- PS Vita: [TODO](https://github.com/sargunv/rom-tools/issues/28)

### Xbox formats

- 🟢 [./lib/roms/xbox](./lib/roms/xbox): Original Xbox XBE executable and XISO disc image parsing.
- Xbox 360: [TODO](https://github.com/sargunv/rom-tools/issues/26)

### Other formats

- Neo Geo: [TODO](https://github.com/sargunv/rom-tools/issues/19)
- Atari 7800: [TODO](https://github.com/sargunv/rom-tools/issues/20)
- Atari Lynx: [TODO](https://github.com/sargunv/rom-tools/issues/21)
- Wonderswan and Color: [TODO](https://github.com/sargunv/rom-tools/issues/22)

## Test Data

ROM files in `**/testdata/` are sourced from:

- [XboxDev/cromwell](https://github.com/XboxDev/cromwell) ([LGPL-2.1](https://github.com/XboxDev/cromwell/blob/86f547387184d1001d377cce97d3756c8acf91cc/COPYING))
- [Zophar's Domain PD ROMs](https://www.zophar.net/pdroms/) (public domain)

These files are used as sample data for automated tests.
