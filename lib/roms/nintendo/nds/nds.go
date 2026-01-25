package nds

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
)

// NDS (Nintendo DS) and DSi ROM format parsing.
//
// NDS/DSi cartridge header specification:
// https://problemkaputt.de/gbatek-ds-cartridge-header.htm
//
// Header layout (512 bytes):
//
//	Offset  Size  Description
//	0x000   12    Game Title (uppercase ASCII, null-padded)
//	0x00C   4     Game Code (category + unique code + destination)
//	0x010   2     Maker Code (e.g., "01" for Nintendo)
//	0x012   1     Unit Code (0x00=NDS, 0x02=NDS+DSi, 0x03=DSi)
//	0x013   1     Encryption Seed Select (0x00-0x07)
//	0x014   1     Device Capacity (ROM size = 128KB << n)
//	0x015   7     Reserved
//	0x01C   1     Reserved
//	0x01D   1     NDS Region (0x00=Normal, 0x40=Korea, 0x80=China)
//	0x01E   1     ROM Version
//	0x01F   1     Autostart (bit 2: skip "Press Button" after Health and Safety)
//	0x020   4     ARM9 ROM Offset
//	0x024   4     ARM9 Entry Address
//	0x028   4     ARM9 RAM Address
//	0x02C   4     ARM9 Size
//	0x030   4     ARM7 ROM Offset
//	0x034   4     ARM7 Entry Address
//	0x038   4     ARM7 RAM Address
//	0x03C   4     ARM7 Size
//	0x040-0x05F   File system info (FNT, FAT, overlays)
//	0x060-0x06F   Port settings and icon offset
//	0x06C   2     Secure Area Checksum (CRC-16)
//	0x06E   2     Secure Area Delay
//	0x070-0x0BF   Auto load, secure area, ROM size info
//	0x0C0   156   Nintendo Logo (compressed)
//	0x15C   2     Nintendo Logo Checksum (CRC-16)
//	0x15E   2     Header Checksum (CRC-16 of bytes 0x000-0x15D)
//	0x160-0x1FF   Debug info and reserved
//
// Game Code breakdown (4 bytes at 0x00C):
//   - Byte 0: Category - game type indicator (A/B/C=NDS, D=DSi-exclusive, K=DSiWare, V=DSi-enhanced)
//   - Bytes 1-2: Unique Code - 2-character game identifier
//   - Byte 3: Destination - target region (J=Japan, E=USA, P=Europe, etc.)

const (
	ndsHeaderSize           = 0x200 // 512 bytes
	ndsTitleOffset          = 0x000
	ndsTitleLen             = 12
	ndsGameCodeOffset       = 0x00C
	ndsGameCodeLen          = 4
	ndsMakerCodeOffset      = 0x010
	ndsMakerCodeLen         = 2
	ndsUnitCodeOffset       = 0x012
	ndsDeviceCapacityOffset = 0x014
	ndsRegionOffset         = 0x01D
	ndsVersionOffset        = 0x01E
	ndsHeaderChecksumOffset = 0x15E
)

// UnitCode indicates the target platform for the ROM.
type UnitCode byte

// UnitCode values per GBATEK.
const (
	UnitCodeNDS    UnitCode = 0x00 // NDS only
	UnitCodeNDSDSi UnitCode = 0x02 // NDS + DSi enhanced
	UnitCodeDSi    UnitCode = 0x03 // DSi only
)

// Region indicates the region lockout setting.
type Region byte

// Region values per GBATEK.
const (
	RegionNormal Region = 0x00 // Normal (worldwide/no region lock)
	RegionKorea  Region = 0x40 // Korea
	RegionChina  Region = 0x80 // China
)

// GameType represents the category from the first byte of the game code.
type GameType byte

// GameType values per GBATEK.
const (
	GameTypeNDS          GameType = 'A' // NDS common (older titles)
	GameTypeNDSB         GameType = 'B' // NDS common (newer titles)
	GameTypeNDSC         GameType = 'C' // NDS common
	GameTypeDSiExclusive GameType = 'D' // DSi-exclusive
	GameTypeDSiWareUtil  GameType = 'H' // DSiWare (system utilities)
	GameTypeNDSInfrared  GameType = 'I' // NDS+DSi with infrared
	GameTypeDSiWare      GameType = 'K' // DSiWare (games)
	GameTypeNDSDemo      GameType = 'N' // NDS demos (Nintendo Channel)
	GameTypeNDST         GameType = 'T' // NDS common
	GameTypeNDSSpecial   GameType = 'U' // NDS with extra hardware
	GameTypeDSiEnhanced  GameType = 'V' // DSi-enhanced
	GameTypeNDSY         GameType = 'Y' // NDS common
)

// Destination represents the target region from the fourth byte of the game code.
type Destination byte

// Destination values per GBATEK.
const (
	DestinationAsia         Destination = 'A'
	DestinationChina        Destination = 'C'
	DestinationGermany      Destination = 'D'
	DestinationUSA          Destination = 'E'
	DestinationFrance       Destination = 'F'
	DestinationNetherlands  Destination = 'H'
	DestinationItaly        Destination = 'I'
	DestinationJapan        Destination = 'J'
	DestinationKorea        Destination = 'K'
	DestinationUSA2         Destination = 'L'
	DestinationSweden       Destination = 'M'
	DestinationNorway       Destination = 'N'
	DestinationIntl         Destination = 'O'
	DestinationEurope       Destination = 'P'
	DestinationDenmark      Destination = 'Q'
	DestinationRussia       Destination = 'R'
	DestinationSpain        Destination = 'S'
	DestinationUSAAustralia Destination = 'T'
	DestinationAustralia    Destination = 'U'
	DestinationEurAustralia Destination = 'V'
	DestinationEuropeW      Destination = 'W'
	DestinationEuropeX      Destination = 'X'
	DestinationEuropeY      Destination = 'Y'
	DestinationEuropeZ      Destination = 'Z'
)

// Info contains metadata extracted from an NDS ROM file.
type Info struct {
	// Title is the game title (0x000, up to 12 uppercase ASCII characters).
	Title string `json:"title,omitempty"`
	// GameCode is the full 4-character game code (0x00C).
	GameCode string `json:"game_code,omitempty"`
	// GameType is the category from byte 0 of GameCode.
	GameType GameType `json:"game_type"`
	// UniqueCode is the 2-character game identifier from bytes 1-2 of GameCode.
	UniqueCode string `json:"unique_code,omitempty"`
	// Destination is the target region from byte 3 of GameCode.
	Destination Destination `json:"destination"`
	// MakerCode is the 2-character manufacturer code (0x010).
	MakerCode string `json:"maker_code,omitempty"`
	// UnitCode indicates the target platform (0x012).
	UnitCode UnitCode `json:"unit_code"`
	// DeviceCapacity is the raw capacity value (0x014). ROM size = 128KB << DeviceCapacity.
	DeviceCapacity byte `json:"device_capacity"`
	// ROMSize is the calculated ROM size in bytes (128KB << DeviceCapacity).
	ROMSize int `json:"rom_size"`
	// Region is the region lockout setting (0x01D).
	Region Region `json:"region"`
	// Version is the ROM version number (0x01E).
	Version int `json:"version"`
	// HeaderChecksum is the CRC-16 of header bytes 0x000-0x15D (0x15E).
	// TODO: validate this checksum
	HeaderChecksum uint16 `json:"header_checksum"`
	// platform is NDS or DSi based on unit code (internal, used by GamePlatform).
	platform core.Platform
}

// GamePlatform implements core.GameInfo.
func (i *Info) GamePlatform() core.Platform { return i.platform }

// GameTitle implements core.GameInfo.
func (i *Info) GameTitle() string { return i.Title }

// GameSerial implements core.GameInfo.
func (i *Info) GameSerial() string { return i.GameCode }

// GameRegions implements core.GameInfo.
func (i *Info) GameRegions() []core.Region {
	switch i.Destination {
	case DestinationJapan:
		return []core.Region{core.RegionJapan}
	case DestinationUSA, DestinationUSA2:
		return []core.Region{core.RegionUSA}
	case DestinationEurope, DestinationEuropeW, DestinationEuropeX, DestinationEuropeY, DestinationEuropeZ:
		return []core.Region{core.RegionEurope}
	case DestinationGermany:
		return []core.Region{core.RegionGermany}
	case DestinationFrance:
		return []core.Region{core.RegionFrance}
	case DestinationItaly:
		return []core.Region{core.RegionItaly}
	case DestinationSpain:
		return []core.Region{core.RegionSpain}
	case DestinationAustralia:
		return []core.Region{core.RegionAustralia}
	case DestinationChina:
		return []core.Region{core.RegionChina}
	case DestinationKorea:
		return []core.Region{core.RegionKorea}
	case DestinationNetherlands:
		return []core.Region{core.RegionNetherlands}
	case DestinationSweden:
		return []core.Region{core.RegionSweden}
	case DestinationNorway:
		return []core.Region{core.RegionNorway}
	case DestinationDenmark:
		return []core.Region{core.RegionDenmark}
	case DestinationRussia:
		return []core.Region{core.RegionRussia}
	case DestinationAsia:
		return []core.Region{core.RegionAsia}
	case DestinationIntl:
		return []core.Region{core.RegionWorld}
	case DestinationUSAAustralia:
		return []core.Region{core.RegionUSA, core.RegionAustralia}
	case DestinationEurAustralia:
		return []core.Region{core.RegionEurope, core.RegionAustralia}
	default:
		return []core.Region{}
	}
}

// Parse extracts game information from an NDS ROM file.
func Parse(r io.ReaderAt, size int64) (*Info, error) {
	if size < ndsHeaderSize {
		return nil, fmt.Errorf("file too small for NDS header: %d bytes", size)
	}

	header := make([]byte, ndsHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read NDS header: %w", err)
	}

	// Extract title (null-terminated ASCII)
	title := util.ExtractASCII(header[ndsTitleOffset : ndsTitleOffset+ndsTitleLen])

	// Extract game code
	gameCode := util.ExtractASCII(header[ndsGameCodeOffset : ndsGameCodeOffset+ndsGameCodeLen])

	// Parse game code components
	var gameType GameType
	var uniqueCode string
	var destination Destination
	if len(gameCode) >= 4 {
		gameType = GameType(gameCode[0])
		uniqueCode = gameCode[1:3]
		destination = Destination(gameCode[3])
	}

	// Extract maker code
	makerCode := util.ExtractASCII(header[ndsMakerCodeOffset : ndsMakerCodeOffset+ndsMakerCodeLen])

	// Extract unit code
	unitCode := UnitCode(header[ndsUnitCodeOffset])

	// Determine platform based on unit code
	var platform core.Platform
	switch unitCode {
	case UnitCodeDSi:
		platform = core.PlatformDSi
	default:
		platform = core.PlatformNDS
	}

	// Extract device capacity and calculate ROM size (max 512MB)
	deviceCapacity := header[ndsDeviceCapacityOffset]
	romSize := 0
	if deviceCapacity <= 12 {
		romSize = (128 * 1024) << deviceCapacity
	}

	// Extract NDS region
	region := Region(header[ndsRegionOffset])

	// Extract software version
	version := int(header[ndsVersionOffset])

	// Extract header checksum (little-endian)
	headerChecksum := binary.LittleEndian.Uint16(header[ndsHeaderChecksumOffset:])

	return &Info{
		Title:          title,
		GameCode:       gameCode,
		GameType:       gameType,
		UniqueCode:     uniqueCode,
		Destination:    destination,
		MakerCode:      makerCode,
		UnitCode:       unitCode,
		DeviceCapacity: deviceCapacity,
		ROMSize:        romSize,
		Region:         region,
		Version:        version,
		HeaderChecksum: headerChecksum,
		platform:       platform,
	}, nil
}
