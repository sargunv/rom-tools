package md

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
)

// Mega Drive (Genesis) ROM format parsing.
//
// Mega Drive ROM header specification:
// https://plutiedev.com/rom-header
//
// Header layout (starting at $100):
//
//	Offset   Size  Description
//	$100     16    System Type (e.g., "SEGA MEGA DRIVE" or "SEGA GENESIS")
//	$110     16    Copyright/Release Date (e.g., "(C)SEGA YYYY.MM")
//	$120     48    Domestic Title (Japanese)
//	$150     48    Overseas Title (International)
//	$180     14    Serial Number/Product Code
//	$18E     2     Checksum (big-endian)
//	$190     16    Device Support (I/O info)
//	$1A0     8     ROM Address Range
//	$1A8     8     RAM Address Range
//	$1B0     12    Extra Memory (SRAM info)
//	$1BC     12    Modem Support
//	$1C8     40    Reserved
//	$1F0     16    Region Support (first 3 chars typically significant)

// Format represents the source ROM format.
type Format int

const (
	FormatMD  Format = iota // Native Mega Drive/Genesis format
	FormatSMD               // Super Magic Drive interleaved format
)

// Region represents a Mega Drive region code as a bitfield.
// Multiple regions can be combined with bitwise OR.
type Region uint8

const (
	RegionDomestic60Hz Region = 1 << 0 // 0x01 - bit 0: Japan (NTSC)
	RegionOverseas60Hz Region = 1 << 2 // 0x04 - bit 2: Americas (NTSC)
	RegionOverseas50Hz Region = 1 << 3 // 0x08 - bit 3: Europe (PAL)
)

// Device represents a supported input device.
type Device string

const (
	DeviceJoypad3Button Device = "J" // 3-button joypad
	DeviceJoypad6Button Device = "6" // 6-button joypad
	DeviceKeyboard      Device = "K" // Keyboard
	DeviceMouse         Device = "M" // Mouse
	DeviceTrackball     Device = "T" // Trackball
	DeviceLightgun      Device = "L" // Menacer/Justifier
	DevicePaddle        Device = "P" // Paddle
	DeviceActivator     Device = "A" // Activator
	DeviceTeamPlayer    Device = "4" // Team Player
	DeviceMasterSystem  Device = "0" // Master System joypad
)

const (
	mdHeaderStart        = 0x100
	mdHeaderSize         = 0x100 // 256 bytes ($100-$1FF)
	mdSystemTypeOffset   = 0x100
	mdSystemTypeLen      = 16
	mdCopyrightOffset    = 0x110
	mdCopyrightLen       = 16
	mdDomesticTitleOff   = 0x120
	mdDomesticTitleLen   = 48
	mdOverseasTitleOff   = 0x150
	mdOverseasTitleLen   = 48
	mdSerialNumberOffset = 0x180
	mdSerialNumberLen    = 14
	mdChecksumOffset     = 0x18E
	mdDeviceSupportOff   = 0x190
	mdDeviceSupportLen   = 16
	mdROMStartOffset     = 0x1A0
	mdROMEndOffset       = 0x1A4
	mdRAMStartOffset     = 0x1A8
	mdRAMEndOffset       = 0x1AC
	mdSRAMInfoOffset     = 0x1B0
	mdSRAMInfoLen        = 12
	mdModemOffset        = 0x1BC
	mdModemLen           = 12
	mdRegionOffset       = 0x1F0
	mdRegionLen          = 16

	// 32X-specific constants
	// The 32X MARS header at offset 0x3C0 identifies 32X ROMs.
	// It typically starts with "MARS" (e.g., "MARS CHECK MODE").
	md32XHeaderOffset = 0x3C0
	md32XMagicLen     = 4
	md32XMagic        = "MARS"

	// Minimum size needed for full parsing including 32X detection
	mdMinParseSize = md32XHeaderOffset + md32XMagicLen // 0x3C4
)

// MDInfo contains metadata extracted from a Mega Drive/Genesis ROM file.
type MDInfo struct {
	// SourceFormat indicates whether the ROM was in MD or SMD format.
	SourceFormat Format `json:"source_format"`
	// SystemType identifies the console (e.g., "SEGA MEGA DRIVE", "SEGA GENESIS").
	SystemType string `json:"system_type,omitempty"`
	// Copyright contains copyright and release date info.
	Copyright string `json:"copyright,omitempty"`
	// DomesticTitle is the Japanese title.
	DomesticTitle string `json:"domestic_title,omitempty"`
	// OverseasTitle is the international title.
	OverseasTitle string `json:"overseas_title,omitempty"`
	// SerialNumber is the product code (e.g., "GM XXXXXXXX-XX").
	SerialNumber string `json:"serial_number,omitempty"`
	// Checksum is the ROM checksum (big-endian).
	Checksum uint16 `json:"checksum"`
	// Devices contains supported input devices.
	Devices []Device `json:"devices,omitempty"`
	// Region is a bitfield of supported regions.
	// Note: Some games have incorrect region codes in their headers.
	// For example, Shadow Squadron (USA, Europe) has only "E" in its header.
	// See: https://misterfpga.org/viewtopic.php?t=4569&start=30
	Region Region `json:"region"`
	// ROMStart is the start address of ROM (typically 0x000000).
	ROMStart uint32 `json:"rom_start"`
	// ROMEnd is the end address of ROM.
	ROMEnd uint32 `json:"rom_end"`
	// RAMStart is the start address of RAM (typically 0xFF0000).
	RAMStart uint32 `json:"ram_start"`
	// RAMEnd is the end address of RAM (typically 0xFFFFFF).
	RAMEnd uint32 `json:"ram_end"`
	// SRAMInfo contains backup memory information (if present).
	SRAMInfo string `json:"sram_info,omitempty"`
	// ModemInfo contains modem/network support information (rarely used).
	ModemInfo string `json:"modem_info,omitempty"`
	// Is32X indicates whether this ROM is for the Sega 32X add-on.
	// Detected by presence of "MARS" at offset 0x3C0.
	Is32X bool `json:"is_32x,omitempty"`
}

// GamePlatform implements identify.GameInfo.
func (i *MDInfo) GamePlatform() core.Platform {
	if i.Is32X {
		return core.Platform32X
	}
	return core.PlatformMD
}

// GameTitle implements identify.GameInfo. Returns overseas title if available, otherwise domestic.
func (i *MDInfo) GameTitle() string {
	if i.OverseasTitle != "" {
		return i.OverseasTitle
	}
	return i.DomesticTitle
}

// GameSerial implements identify.GameInfo.
func (i *MDInfo) GameSerial() string { return i.SerialNumber }

// Parse extracts game information from a Mega Drive ROM file.
// It automatically detects and handles both native MD and SMD formats.
func Parse(r io.ReaderAt, size int64) (*MDInfo, error) {
	if isSMDROM(r, size) {
		return parseSMD(r, size)
	}
	return parseMD(r, size)
}

// parseMD extracts game information from a native Mega Drive ROM.
func parseMD(r io.ReaderAt, size int64) (*MDInfo, error) {
	if size < mdHeaderStart+mdHeaderSize {
		return nil, fmt.Errorf("file too small for Mega Drive header: %d bytes", size)
	}

	// Read enough for header + 32X detection (0x3C4 bytes)
	// If file is smaller, read what we can (32X detection will be skipped)
	readSize := min(size, int64(mdMinParseSize))

	data := make([]byte, readSize)
	if _, err := r.ReadAt(data, 0); err != nil {
		return nil, fmt.Errorf("failed to read Mega Drive ROM: %w", err)
	}

	info, err := parseMDBytes(data)
	if err != nil {
		return nil, err
	}
	info.SourceFormat = FormatMD
	return info, nil
}

func parseMDBytes(data []byte) (*MDInfo, error) {
	// Extract system type and verify
	systemType := util.ExtractASCII(data[mdSystemTypeOffset : mdSystemTypeOffset+mdSystemTypeLen])
	if !strings.Contains(systemType, "SEGA") {
		return nil, fmt.Errorf("not a valid Mega Drive ROM: system type is %q", systemType)
	}

	// Extract all fields
	copyright := util.ExtractASCII(data[mdCopyrightOffset : mdCopyrightOffset+mdCopyrightLen])
	domesticTitle := util.ExtractShiftJIS(data[mdDomesticTitleOff : mdDomesticTitleOff+mdDomesticTitleLen])
	overseasTitle := util.ExtractASCII(data[mdOverseasTitleOff : mdOverseasTitleOff+mdOverseasTitleLen])
	serialNumber := util.ExtractASCII(data[mdSerialNumberOffset : mdSerialNumberOffset+mdSerialNumberLen])

	// Extract checksum (big-endian)
	checksum := binary.BigEndian.Uint16(data[mdChecksumOffset:])

	// Extract device support
	deviceData := data[mdDeviceSupportOff : mdDeviceSupportOff+mdDeviceSupportLen]
	devices := parseDevices(deviceData)

	// Extract region codes
	regionData := data[mdRegionOffset : mdRegionOffset+mdRegionLen]
	region := parseRegionCodes(regionData)

	// Extract ROM address range
	romStart := binary.BigEndian.Uint32(data[mdROMStartOffset:])
	romEnd := binary.BigEndian.Uint32(data[mdROMEndOffset:])

	// Extract RAM address range
	ramStart := binary.BigEndian.Uint32(data[mdRAMStartOffset:])
	ramEnd := binary.BigEndian.Uint32(data[mdRAMEndOffset:])

	// Extract SRAM info
	sramInfo := util.ExtractASCII(data[mdSRAMInfoOffset : mdSRAMInfoOffset+mdSRAMInfoLen])

	// Extract modem info
	modemInfo := util.ExtractASCII(data[mdModemOffset : mdModemOffset+mdModemLen])

	// Check for 32X by looking for "MARS" at offset 0x3C0
	// This is the start of the MARS header (e.g., "MARS CHECK MODE")
	is32X := false
	if len(data) >= md32XHeaderOffset+md32XMagicLen {
		marsData := string(data[md32XHeaderOffset : md32XHeaderOffset+md32XMagicLen])
		if marsData == md32XMagic {
			is32X = true
		}
	}

	return &MDInfo{
		SystemType:    systemType,
		Copyright:     copyright,
		DomesticTitle: domesticTitle,
		OverseasTitle: overseasTitle,
		SerialNumber:  serialNumber,
		Checksum:      checksum,
		Devices:       devices,
		Region:        region,
		ROMStart:      romStart,
		ROMEnd:        romEnd,
		RAMStart:      ramStart,
		RAMEnd:        ramEnd,
		SRAMInfo:      sramInfo,
		ModemInfo:     modemInfo,
		Is32X:         is32X,
	}, nil
}

// parseRegionCodes extracts region codes from the region field.
// Mega Drive uses two styles:
// - Old style: ASCII chars like J (Japan), U (USA), E (Europe)
// - New style: Single hex digit as bitfield (bit 0=Japan, bit 2=USA, bit 3=Europe)
// Both styles are normalized to a bitfield.
func parseRegionCodes(data []byte) Region {
	var chars []byte
	for _, b := range data {
		if b == 0x00 || b == ' ' {
			break
		}
		chars = append(chars, b)
	}

	if len(chars) == 0 {
		return 0
	}

	// Try old-style first: J, U, E characters
	var region Region
	for _, b := range chars {
		switch b {
		case 'J':
			region |= RegionDomestic60Hz
		case 'U':
			region |= RegionOverseas60Hz
		case 'E':
			region |= RegionOverseas50Hz
		}
	}
	if region != 0 {
		return region
	}

	// New-style: single hex digit is already a bitfield
	if len(chars) == 1 {
		b := chars[0]
		if b >= '0' && b <= '9' {
			return Region(b - '0')
		}
		if b >= 'A' && b <= 'F' {
			return Region(b - 'A' + 10)
		}
	}

	return 0
}

// parseDevices extracts device support codes from the device support field.
func parseDevices(data []byte) []Device {
	var devices []Device
	for _, b := range data {
		if b == 0x00 || b == ' ' {
			break
		}
		devices = append(devices, Device(b))
	}
	return devices
}
