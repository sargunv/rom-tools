package megadrive

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
)

// Sega CD / Mega CD disc identification from ISO 9660 system area.
//
// Sega CD discs store metadata in the ISO 9660 system area (sectors 0-15).
// The disc identifier and Genesis-style header are at the start of sector 0.
//
// Disc identifier (first 16 bytes at offset 0x00):
//   - "SEGADISCSYSTEM  " (bootable)
//   - "SEGABOOTDISC    " (bootable)
//   - "SEGADISC        " (non-bootable)
//   - "SEGADATADISC    " (non-bootable)
//
// The Genesis-style header at offset 0x100 uses the same format as cartridge ROMs.
//
// Documentation:
//   - https://www.retrodev.com/segacd.html
//   - https://www.plutiedev.com/rom-header

// DiscType indicates whether the disc is bootable.
type DiscType int

const (
	DiscTypeUnknown DiscType = iota
	DiscTypeBootable
	DiscTypeNonBootable
)

const (
	segaCDHeaderSize = 0x200 // Need 512 bytes (disc ID at 0x00 + header at 0x100-0x1FF)
	discIDOffset     = 0x00
	discIDLen        = 16
)

// Known disc identifiers
var (
	discIDBootable = []string{
		"SEGADISCSYSTEM  ",
		"SEGABOOTDISC    ",
	}
	discIDNonBootable = []string{
		"SEGADISC        ",
		"SEGADATADISC    ",
	}
)

// SegaCDInfo contains metadata extracted from a Sega CD disc header.
type SegaCDInfo struct {
	// DiscID is the 16-byte disc identifier.
	DiscID string `json:"disc_id,omitempty"`
	// DiscType indicates whether the disc is bootable.
	DiscType DiscType `json:"disc_type,omitempty"`
	// SystemType identifies the system (e.g., "SEGA MEGA DRIVE").
	SystemType string `json:"system_type,omitempty"`
	// Copyright contains copyright and release date info.
	Copyright string `json:"copyright,omitempty"`
	// DomesticTitle is the Japanese title.
	DomesticTitle string `json:"domestic_title,omitempty"`
	// OverseasTitle is the international title.
	OverseasTitle string `json:"overseas_title,omitempty"`
	// SerialNumber is the product code (e.g., "GM XXXXXXXX-XX").
	SerialNumber string `json:"serial_number,omitempty"`
	// Devices contains supported input devices.
	Devices []Device `json:"devices,omitempty"`
	// Region is a bitfield of supported regions.
	Region Region `json:"region,omitempty"`
}

// GamePlatform implements identify.GameInfo.
func (i *SegaCDInfo) GamePlatform() core.Platform { return core.PlatformSegaCD }

// GameTitle implements identify.GameInfo. Returns overseas title if available, otherwise domestic.
func (i *SegaCDInfo) GameTitle() string {
	if i.OverseasTitle != "" {
		return i.OverseasTitle
	}
	return i.DomesticTitle
}

// GameSerial implements identify.GameInfo.
func (i *SegaCDInfo) GameSerial() string { return i.SerialNumber }

// ParseSegaCD parses Sega CD metadata from a reader.
// The reader should contain the ISO 9660 system area data.
func ParseSegaCD(r io.ReaderAt, size int64) (*SegaCDInfo, error) {
	if size < segaCDHeaderSize {
		return nil, fmt.Errorf("data too small for Sega CD header: %d bytes", size)
	}

	data := make([]byte, segaCDHeaderSize)
	if _, err := r.ReadAt(data, 0); err != nil {
		return nil, fmt.Errorf("failed to read Sega CD header: %w", err)
	}

	return parseSegaCDBytes(data)
}

func parseSegaCDBytes(data []byte) (*SegaCDInfo, error) {
	// Validate disc identifier
	discID := string(data[discIDOffset : discIDOffset+discIDLen])
	discType := getDiscType(discID)
	if discType == DiscTypeUnknown {
		return nil, fmt.Errorf("not a valid Sega CD disc: invalid disc identifier %q", discID)
	}

	// Extract device support (reuse MD constants since header is at same offsets)
	deviceData := data[mdDeviceSupportOff : mdDeviceSupportOff+mdDeviceSupportLen]
	devices := parseDevices(deviceData)

	// Extract region codes
	regionData := data[mdRegionOffset : mdRegionOffset+mdRegionLen]
	region := parseRegionCodes(regionData)

	info := &SegaCDInfo{
		DiscID:        strings.TrimSpace(discID),
		DiscType:      discType,
		SystemType:    util.ExtractASCII(data[mdSystemTypeOffset : mdSystemTypeOffset+mdSystemTypeLen]),
		Copyright:     util.ExtractASCII(data[mdCopyrightOffset : mdCopyrightOffset+mdCopyrightLen]),
		DomesticTitle: util.ExtractShiftJIS(data[mdDomesticTitleOff : mdDomesticTitleOff+mdDomesticTitleLen]),
		OverseasTitle: util.ExtractASCII(data[mdOverseasTitleOff : mdOverseasTitleOff+mdOverseasTitleLen]),
		SerialNumber:  util.ExtractASCII(data[mdSerialNumberOffset : mdSerialNumberOffset+mdSerialNumberLen]),
		Devices:       devices,
		Region:        region,
	}

	return info, nil
}

// getDiscType returns the disc type based on the identifier.
func getDiscType(discID string) DiscType {
	if slices.Contains(discIDBootable, discID) {
		return DiscTypeBootable
	}
	if slices.Contains(discIDNonBootable, discID) {
		return DiscTypeNonBootable
	}
	return DiscTypeUnknown
}
