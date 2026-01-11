# ROM Format Support Plan

## Currently Implemented

| Platform             | Format   | Extensions             | Status  |
| -------------------- | -------- | ---------------------- | ------- |
| Game Boy Advance     | GBA      | `.gba`                 | ✅ Done |
| Game Boy / Color     | GB       | `.gb`, `.gbc`          | ✅ Done |
| Nintendo 64          | N64      | `.z64`, `.v64`, `.n64` | ✅ Done |
| Mega Drive / Genesis | MD       | `.md`, `.gen`, `.smd`  | ✅ Done |
| Xbox                 | XBE/XISO | `.xbe`, `.iso`         | ✅ Done |
| Famicom              | NES      | `.nes`                 | ✅ Done |
| Nintendo DS          | NDS      | `.nds`, `.dsi`, `.ids` | ✅ Done |
| Super Famicom        | SNES     | `.sfc`, `.smc`         | ✅ Done |

## Planned Support

### High Priority (Common Platforms)

| Platform | Complexity | Plan |
| --- | --- | --- |
| [PlayStation 1](playstation-support.md) | Medium-High | ISO + SYSTEM.CNF parsing |
| [PlayStation Portable](psp-support.md) | Medium-High | ISO + PARAM.SFO parsing |

### Medium Priority

| Platform | Complexity | Plan |
| --- | --- | --- |
| [Master System / Game Gear](sms-gg-support.md) | Low-Medium | Shared format, variable header location |
| [GameCube / Wii](gamecube-wii-support.md) | Medium | Multiple formats (ISO, WBFS) |
| [Sega 32X / Sega CD](32x-segacd-support.md) | Low-Medium | Extends Mega Drive format |
| [TurboGrafx-16 / PC Engine](turbografx-support.md) | Low (ROM) | No header, hash-based ID |

### Lower Priority (Complex or Niche)

| Platform | Complexity | Plan |
| --- | --- | --- |
| [Nintendo 3DS](3ds-support.md) | High | NCSD/NCCH nested format |
| [Sega Saturn](saturn-support.md) | Medium | IP.BIN header |
| [Sega Dreamcast](dreamcast-support.md) | High | GDI/CDI multi-track |
| [PS Vita](ps-vita-support.md) | Low-High | VPK easy, PKG encrypted |

## Complexity Ratings

- **Low**: Simple header at fixed offset, clear magic bytes
- **Low-Medium**: Simple header but variable location or optional
- **Medium**: Multiple formats or requires filesystem parsing
- **Medium-High**: Complex nested formats or encryption considerations
- **High**: Multi-file formats, complex containers, or encryption

## Implementation Priority Recommendations

### Phase 1: Quick Wins

1. **Master System/Game Gear** - Shares format, easy extension
2. **Atari Lynx** - Simple 64-byte header

### Phase 2: Major Platforms

1. **PSP** - Needs ISO + SFO parsing (reusable)
2. **GameCube/Wii** - Popular, straightforward once WBFS handled

### Phase 3: CD-Based Systems

1. **PlayStation 1** - Build on PSP's ISO parsing
2. **Sega Saturn** - Similar to existing Sega formats
3. **Sega CD** - Extends Mega Drive support

### Phase 4: Complex Formats

1. **Nintendo 3DS** - Complex but popular
2. **Dreamcast** - Multi-format (GDI, CDI, CHD)
3. **Neo Geo** - ROM set handling
4. **PS Vita** - VPK first, PKG later

## Common Patterns

### Reusable Components

1. **ISO 9660 Parser**: PSX, PSP, Saturn, Dreamcast, GameCube, Wii
2. **PARAM.SFO Parser**: PSP, PS Vita
3. **Sega Header Parser**: MD, 32X, SMS, GG, Saturn, CD
4. **Multi-track CD Handler**: PSX, Saturn, Dreamcast, Sega CD
5. **CHD Handler**: Already implemented, useful for all CD formats

### Common Challenges

1. **No embedded title**: NES, TG16, Atari 2600
2. **Variable header location**: SNES, SMS/GG
3. **Multi-file ROMs**: Neo Geo
4. **Encryption**: NDS, 3DS, PKG
