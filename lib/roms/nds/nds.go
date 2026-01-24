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

// NDSUnitCode indicates the target platform for the ROM.
type NDSUnitCode byte

// NDSUnitCode values per GBATEK.
const (
	NDSUnitCodeNDS    NDSUnitCode = 0x00 // NDS only
	NDSUnitCodeNDSDSi NDSUnitCode = 0x02 // NDS + DSi enhanced
	NDSUnitCodeDSi    NDSUnitCode = 0x03 // DSi only
)

// NDSRegion indicates the region lockout setting.
type NDSRegion byte

// NDSRegion values per GBATEK.
const (
	NDSRegionNormal NDSRegion = 0x00 // Normal (worldwide/no region lock)
	NDSRegionKorea  NDSRegion = 0x40 // Korea
	NDSRegionChina  NDSRegion = 0x80 // China
)

// NDSGameType represents the category from the first byte of the game code.
type NDSGameType byte

// NDSGameType values per GBATEK.
const (
	NDSGameTypeNDS          NDSGameType = 'A' // NDS common (older titles)
	NDSGameTypeNDSB         NDSGameType = 'B' // NDS common (newer titles)
	NDSGameTypeNDSC         NDSGameType = 'C' // NDS common
	NDSGameTypeDSiExclusive NDSGameType = 'D' // DSi-exclusive
	NDSGameTypeDSiWareUtil  NDSGameType = 'H' // DSiWare (system utilities)
	NDSGameTypeNDSInfrared  NDSGameType = 'I' // NDS+DSi with infrared
	NDSGameTypeDSiWare      NDSGameType = 'K' // DSiWare (games)
	NDSGameTypeNDSDemo      NDSGameType = 'N' // NDS demos (Nintendo Channel)
	NDSGameTypeNDST         NDSGameType = 'T' // NDS common
	NDSGameTypeNDSSpecial   NDSGameType = 'U' // NDS with extra hardware
	NDSGameTypeDSiEnhanced  NDSGameType = 'V' // DSi-enhanced
	NDSGameTypeNDSY         NDSGameType = 'Y' // NDS common
)

// NDSDestination represents the target region from the fourth byte of the game code.
type NDSDestination byte

// NDSDestination values per GBATEK.
const (
	NDSDestinationAsia         NDSDestination = 'A'
	NDSDestinationChina        NDSDestination = 'C'
	NDSDestinationGermany      NDSDestination = 'D'
	NDSDestinationUSA          NDSDestination = 'E'
	NDSDestinationFrance       NDSDestination = 'F'
	NDSDestinationNetherlands  NDSDestination = 'H'
	NDSDestinationItaly        NDSDestination = 'I'
	NDSDestinationJapan        NDSDestination = 'J'
	NDSDestinationKorea        NDSDestination = 'K'
	NDSDestinationUSA2         NDSDestination = 'L'
	NDSDestinationSweden       NDSDestination = 'M'
	NDSDestinationNorway       NDSDestination = 'N'
	NDSDestinationIntl         NDSDestination = 'O'
	NDSDestinationEurope       NDSDestination = 'P'
	NDSDestinationDenmark      NDSDestination = 'Q'
	NDSDestinationRussia       NDSDestination = 'R'
	NDSDestinationSpain        NDSDestination = 'S'
	NDSDestinationUSAAustralia NDSDestination = 'T'
	NDSDestinationAustralia    NDSDestination = 'U'
	NDSDestinationEurAustralia NDSDestination = 'V'
	NDSDestinationEuropeW      NDSDestination = 'W'
	NDSDestinationEuropeX      NDSDestination = 'X'
	NDSDestinationEuropeY      NDSDestination = 'Y'
	NDSDestinationEuropeZ      NDSDestination = 'Z'
)

// NDSInfo contains metadata extracted from an NDS ROM file.
type NDSInfo struct {
	// Title is the game title (0x000, up to 12 uppercase ASCII characters).
	Title string `json:"title,omitempty"`
	// GameCode is the full 4-character game code (0x00C).
	GameCode string `json:"game_code,omitempty"`
	// GameType is the category from byte 0 of GameCode.
	GameType NDSGameType `json:"game_type"`
	// UniqueCode is the 2-character game identifier from bytes 1-2 of GameCode.
	UniqueCode string `json:"unique_code,omitempty"`
	// Destination is the target region from byte 3 of GameCode.
	Destination NDSDestination `json:"destination"`
	// MakerCode is the 2-character manufacturer code (0x010).
	MakerCode string `json:"maker_code,omitempty"`
	// UnitCode indicates the target platform (0x012).
	UnitCode NDSUnitCode `json:"unit_code"`
	// DeviceCapacity is the raw capacity value (0x014). ROM size = 128KB << DeviceCapacity.
	DeviceCapacity byte `json:"device_capacity"`
	// ROMSize is the calculated ROM size in bytes (128KB << DeviceCapacity).
	ROMSize int `json:"rom_size"`
	// NDSRegion is the region lockout setting (0x01D).
	NDSRegion NDSRegion `json:"nds_region"`
	// Version is the ROM version number (0x01E).
	Version int `json:"version"`
	// HeaderChecksum is the CRC-16 of header bytes 0x000-0x15D (0x15E).
	// TODO: validate this checksum
	HeaderChecksum uint16 `json:"header_checksum"`
	// platform is NDS or DSi based on unit code (internal, used by GamePlatform).
	platform core.Platform
}

// GamePlatform implements identify.GameInfo.
func (i *NDSInfo) GamePlatform() core.Platform { return i.platform }

// GameTitle implements identify.GameInfo.
func (i *NDSInfo) GameTitle() string { return i.Title }

// GameSerial implements identify.GameInfo.
func (i *NDSInfo) GameSerial() string { return i.GameCode }

// ParseNDS extracts game information from an NDS ROM file.
func ParseNDS(r io.ReaderAt, size int64) (*NDSInfo, error) {
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
	var gameType NDSGameType
	var uniqueCode string
	var destination NDSDestination
	if len(gameCode) >= 4 {
		gameType = NDSGameType(gameCode[0])
		uniqueCode = gameCode[1:3]
		destination = NDSDestination(gameCode[3])
	}

	// Extract maker code
	makerCode := util.ExtractASCII(header[ndsMakerCodeOffset : ndsMakerCodeOffset+ndsMakerCodeLen])

	// Extract unit code
	unitCode := NDSUnitCode(header[ndsUnitCodeOffset])

	// Determine platform based on unit code
	var platform core.Platform
	switch unitCode {
	case NDSUnitCodeDSi:
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
	ndsRegion := NDSRegion(header[ndsRegionOffset])

	// Extract software version
	version := int(header[ndsVersionOffset])

	// Extract header checksum (little-endian)
	headerChecksum := binary.LittleEndian.Uint16(header[ndsHeaderChecksumOffset:])

	return &NDSInfo{
		Title:          title,
		GameCode:       gameCode,
		GameType:       gameType,
		UniqueCode:     uniqueCode,
		Destination:    destination,
		MakerCode:      makerCode,
		UnitCode:       unitCode,
		DeviceCapacity: deviceCapacity,
		ROMSize:        romSize,
		NDSRegion:      ndsRegion,
		Version:        version,
		HeaderChecksum: headerChecksum,
		platform:       platform,
	}, nil
}
