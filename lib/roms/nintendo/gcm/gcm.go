package gcm

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
)

// GameCube/Wii disc header format parsing.
//
// Both GameCube and Wii use the same disc header format.
// Per WiiBrew wiki: "The first 0x400 bytes are like the GameCube disc header format."
// https://wiibrew.org/wiki/Wii_disc
//
// Disc header layout (relevant fields):
//
//	Offset  Size  Description
//	0x000   1     System code (G=GameCube, R/S=Wii, etc.)
//	0x001   2     Game code
//	0x003   1     Region code (E=USA, P=PAL, J=Japan, etc.)
//	0x004   2     Maker code
//	0x006   1     Disc number
//	0x007   1     Disc version
//	0x018   4     Wii magic word (0x5D1C9EA3 for Wii, 0x00000000 for GameCube)
//	0x01C   4     GameCube magic word (0xC2339F3D for GameCube, 0x00000000 for Wii)
//	0x020   64    Game title (ASCII, null-terminated)

// SystemCode represents the console/platform identifier (first byte of disc ID).
// Source: https://wiki.dolphin-emu.org/index.php?title=GameIDs
type SystemCode byte

const (
	// Physical disc platforms
	SystemCodeGameCube      SystemCode = 'G' // GameCube
	SystemCodeGameCubeRerel SystemCode = 'D' // GameCube rerelease (Master Quest, demo discs)
	SystemCodeGameCubePromo SystemCode = 'P' // Promotional GameCube (shared with TurboGrafx-16 VC)
	SystemCodeWii           SystemCode = 'R' // Older Wii releases
	SystemCodeWiiNew        SystemCode = 'S' // Newer Wii releases

	// Virtual Console platforms
	SystemCodeNES            SystemCode = 'F' // NES Virtual Console
	SystemCodeSNES           SystemCode = 'J' // Super Nintendo Virtual Console
	SystemCodeN64            SystemCode = 'N' // Nintendo 64 Virtual Console
	SystemCodeSMS            SystemCode = 'L' // Sega Master System Virtual Console
	SystemCodeGenesis        SystemCode = 'M' // Sega Genesis Virtual Console
	SystemCodeTurboGrafx16   SystemCode = 'P' // TurboGrafx-16 Virtual Console
	SystemCodeTurboGrafx16CD SystemCode = 'Q' // TurboGrafx-16 CD Virtual Console
	SystemCodeC64            SystemCode = 'C' // Commodore 64 Virtual Console
	SystemCodeArcade         SystemCode = 'E' // Arcade / Neo Geo Virtual Console
	SystemCodeMSX            SystemCode = 'X' // MSX Virtual Console (shared with WiiWare demos)

	// Wii digital
	SystemCodeWiiChannels  SystemCode = 'H' // Wii Channels
	SystemCodeWiiWare      SystemCode = 'W' // WiiWare
	SystemCodeWiiWareDemos SystemCode = 'X' // WiiWare Demos (shared with MSX)
)

// Region represents the target region (fourth byte of disc ID).
// Source: https://wiki.dolphin-emu.org/index.php?title=GameIDs
type Region byte

const (
	RegionJapan          Region = 'J' // Japan
	RegionNorthAmerica   Region = 'E' // USA / North America
	RegionEurope         Region = 'P' // Europe and other PAL regions
	RegionAustralia      Region = 'U' // Australia (also Europe alternate)
	RegionKorea          Region = 'K' // Korea
	RegionTaiwan         Region = 'W' // Taiwan / Hong Kong / Macau
	RegionGermany        Region = 'D' // Germany
	RegionFrance         Region = 'F' // France
	RegionSpain          Region = 'S' // Spain
	RegionItaly          Region = 'I' // Italy
	RegionNetherlands    Region = 'H' // Netherlands (also Europe alternate)
	RegionRussia         Region = 'R' // Russia
	RegionScandinavia    Region = 'V' // Scandinavia
	RegionSystemChannels Region = 'A' // System Wii Channels
	RegionJPImportPAL    Region = 'L' // Japanese import to PAL regions
	RegionUSImportPAL    Region = 'M' // American import to PAL regions
	RegionJPImportNTSC   Region = 'N' // Japanese import to NTSC regions
	RegionJPImportKorea  Region = 'Q' // Japanese VC import to Korea
	RegionUSImportKorea  Region = 'T' // American VC import to Korea
	RegionSpecialX       Region = 'X' // Europe/US special releases
	RegionSpecialY       Region = 'Y' // Europe/US special releases
	RegionSpecialZ       Region = 'Z' // Europe/US special releases
)

const (
	discHeaderSize = 0x60 // We only need first 96 bytes for identification

	systemCodeOffset  = 0x000
	gameCodeOffset    = 0x001
	gameCodeLen       = 2
	regionOffset      = 0x003
	makerCodeOffset   = 0x004
	makerCodeLen      = 2
	discNumberOffset  = 0x006
	discVersionOffset = 0x007
	wiiMagicOffset    = 0x018
	gcMagicOffset     = 0x01C
	titleOffset       = 0x020
	titleLen          = 64

	wiiMagicWord = 0x5D1C9EA3
	gcMagicWord  = 0xC2339F3D
)

// Info contains metadata extracted from a GameCube/Wii disc header.
type Info struct {
	// SystemCode is the console/platform identifier (G=GameCube, R/S=Wii, etc.).
	SystemCode SystemCode `json:"system_code"`
	// GameCode is the 2-character unique game identifier.
	GameCode string `json:"game_code,omitempty"`
	// Region is the region code (E=USA, P=PAL, J=Japan, etc.).
	Region Region `json:"region"`
	// MakerCode is the 2-character publisher identifier.
	MakerCode string `json:"maker_code,omitempty"`
	// DiscNumber is the disc number for multi-disc games.
	DiscNumber int `json:"disc_number"`
	// Version is the disc version/revision.
	Version int `json:"version"`
	// Title is the game title.
	Title string `json:"title,omitempty"`
	// platform is the target platform (GameCube or Wii) (internal, used by GamePlatform).
	platform core.Platform
}

// GamePlatform implements core.GameInfo.
func (i *Info) GamePlatform() core.Platform { return i.platform }

// GameTitle implements core.GameInfo.
func (i *Info) GameTitle() string { return i.Title }

// GameSerial implements core.GameInfo. Returns the full game ID (SystemCode + GameCode + Region).
func (i *Info) GameSerial() string {
	return fmt.Sprintf("%c%s%c", i.SystemCode, i.GameCode, i.Region)
}

// GameRegions implements core.GameInfo.
func (i *Info) GameRegions() []core.Region {
	switch i.Region {
	case RegionJapan:
		return []core.Region{core.RegionJapan}
	case RegionNorthAmerica:
		return []core.Region{core.RegionUSA}
	case RegionEurope:
		return []core.Region{core.RegionEurope}
	case RegionAustralia:
		return []core.Region{core.RegionAustralia}
	case RegionKorea:
		return []core.Region{core.RegionKorea}
	case RegionTaiwan:
		return []core.Region{core.RegionTaiwan}
	case RegionGermany:
		return []core.Region{core.RegionGermany}
	case RegionFrance:
		return []core.Region{core.RegionFrance}
	case RegionSpain:
		return []core.Region{core.RegionSpain}
	case RegionItaly:
		return []core.Region{core.RegionItaly}
	case RegionNetherlands:
		return []core.Region{core.RegionNetherlands}
	case RegionRussia:
		return []core.Region{core.RegionRussia}
	case RegionScandinavia:
		return []core.Region{core.RegionDenmark, core.RegionNorway, core.RegionSweden}
	case RegionSystemChannels:
		return []core.Region{core.RegionWorld}
	case RegionJPImportPAL, RegionJPImportNTSC, RegionJPImportKorea:
		return []core.Region{core.RegionJapan}
	case RegionUSImportPAL, RegionUSImportKorea:
		return []core.Region{core.RegionUSA}
	case RegionSpecialX, RegionSpecialY, RegionSpecialZ:
		return []core.Region{core.RegionEurope, core.RegionUSA}
	default:
		return []core.Region{}
	}
}

// Parse parses a GameCube/Wii disc header from a reader.
func Parse(r io.ReaderAt, size int64) (*Info, error) {
	if size < discHeaderSize {
		return nil, fmt.Errorf("file too small for disc header: need %d bytes, got %d", discHeaderSize, size)
	}

	header := make([]byte, discHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read disc header: %w", err)
	}

	return parseGCMBytes(header)
}

func parseGCMBytes(header []byte) (*Info, error) {
	// Check magic words to determine platform and validate
	wiiMagic := binary.BigEndian.Uint32(header[wiiMagicOffset:])
	gcMagic := binary.BigEndian.Uint32(header[gcMagicOffset:])

	isWii := wiiMagic == wiiMagicWord
	isGC := gcMagic == gcMagicWord

	if !isWii && !isGC {
		return nil, fmt.Errorf("not a valid GameCube/Wii disc: no magic word found (Wii: 0x%08X, GC: 0x%08X)",
			wiiMagic, gcMagic)
	}

	// Determine platform
	var platform core.Platform
	if isWii {
		platform = core.PlatformWii
	} else {
		platform = core.PlatformGC
	}

	// Extract fields
	systemCode := SystemCode(header[systemCodeOffset])
	gameCode := util.ExtractASCII(header[gameCodeOffset : gameCodeOffset+gameCodeLen])
	region := Region(header[regionOffset])
	makerCode := util.ExtractASCII(header[makerCodeOffset : makerCodeOffset+makerCodeLen])
	discNumber := int(header[discNumberOffset])
	version := int(header[discVersionOffset])
	title := util.ExtractASCII(header[titleOffset : titleOffset+titleLen])

	return &Info{
		SystemCode: systemCode,
		GameCode:   gameCode,
		Region:     region,
		MakerCode:  makerCode,
		DiscNumber: discNumber,
		Version:    version,
		Title:      title,
		platform:   platform,
	}, nil
}
