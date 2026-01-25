// Package psnpkg provides PlayStation PKG (package) file format parsing.
//
// PKG is the package format used for digital distribution on PS3, PSP, PS Vita, and PSM.
// This parser extracts identification metadata from PKG headers and embedded PARAM.SFO.
//
// Reference: https://www.psdevwiki.com/ps3/PKG_files
//
// PKG header layout (0x80 bytes, big-endian):
//
//	Offset  Size  Description
//	0x00    4     Magic ("\x7FPKG")
//	0x04    2     Revision (0x8000=finalized, 0x0000=debug)
//	0x06    2     Type (0x0001=PS3, 0x0002=PSP/Vita)
//	0x08    4     Metadata offset (absolute)
//	0x0C    4     Metadata count
//	0x10    4     Header size
//	0x14    4     Item count
//	0x18    8     Total size
//	0x20    8     Data offset
//	0x28    8     Data size
//	0x30    48    Content ID (null-terminated)
//	0x60    16    Digest
//	0x70    16    Data RIV (AES-CTR IV)
//
// Metadata entries (after header):
//
//	Each entry: ID (4 bytes BE) + Size (4 bytes BE) + Data (Size bytes)
//	Key IDs: 0x02 = Content Type, 0x0E = SFO Info
package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/core"
	"github.com/sargunv/rom-tools/lib/roms/playstation/sfo"
)

const (
	pkgMagic         = "\x7FPKG"
	pkgExtMagic      = "\x7Fext"
	pkgHeaderSize    = 0x80
	pkgPS3DigestSize = 0x40

	// Main header offsets
	pkgRevisionOffset    = 0x04
	pkgTypeOffset        = 0x06
	pkgMetadataOffOffset = 0x08
	pkgMetadataCountOff  = 0x0C
	pkgItemCountOffset   = 0x14
	pkgTotalSizeOffset   = 0x18
	pkgContentIDOffset   = 0x30
	pkgContentIDLen      = 48

	// Extended header (for PSP/Vita, located after main header + PS3 digest)
	pkgExtHeaderOffset = pkgHeaderSize + pkgPS3DigestSize // 0xC0
	pkgExtKeyIDOffset  = 0x24                             // within extended header

	// Metadata IDs
	metadataContentType = 0x02
	metadataSFOInfo     = 0x0E
)

// Type indicates the target platform category.
type Type uint16

// Type values per PSDevWiki.
const (
	TypePS3     Type = 0x0001 // PS3 packages
	TypePSPVita Type = 0x0002 // PSP and PS Vita packages
)

// ContentType indicates the type of content in the package.
type ContentType uint32

// ContentType values per PSDevWiki.
const (
	ContentTypeGameData    ContentType = 0x04 // PS3 GameData
	ContentTypeGameExec    ContentType = 0x05 // PS3 GameExec
	ContentTypePS1Emu      ContentType = 0x06 // PS1 emulator (PS3/PSP)
	ContentTypePSP         ContentType = 0x07 // PSP game
	ContentTypeTheme       ContentType = 0x09 // PS3 Theme
	ContentTypeWidget      ContentType = 0x0A // PS3 Widget
	ContentTypeLicense     ContentType = 0x0B // PS3 License
	ContentTypeVSHModule   ContentType = 0x0C // PS3 VSH Module
	ContentTypePSNAvatar   ContentType = 0x0D // PS3 PSN Avatar
	ContentTypePSPGo       ContentType = 0x0E // PSPgo
	ContentTypeMinis       ContentType = 0x0F // PS Minis
	ContentTypeNeoGeo      ContentType = 0x10 // NEOGEO
	ContentTypePSVitaGame  ContentType = 0x15 // PS Vita game (PSP2GD)
	ContentTypePSVitaDLC   ContentType = 0x16 // PS Vita DLC (PSP2AC)
	ContentTypePSVitaLA    ContentType = 0x17 // PS Vita LiveArea (PSP2LA)
	ContentTypePSM         ContentType = 0x18 // PlayStation Mobile
	ContentTypePSMUnity    ContentType = 0x1D // PSM for Unity
	ContentTypePSVitaTheme ContentType = 0x1F // PS Vita Theme
)

// Info contains metadata extracted from a PlayStation PKG file.
// Info implements core.GameInfo.
type Info struct {
	// ContentID is the package content identifier (e.g., "UP0001-NPUA80472_00-LITTLEBIGPLAN001").
	ContentID string `json:"content_id,omitempty"`
	// Title is the game title from embedded PARAM.SFO.
	Title string `json:"title,omitempty"`
	// Platform is the detected platform (psp, psvita, psm, playstation3).
	Platform core.Platform `json:"platform"`
	// Type is the PKG type (0x0001=PS3, 0x0002=PSP/Vita).
	Type Type `json:"pkg_type"`
	// ContentType indicates the content category (determines platform).
	ContentType ContentType `json:"content_type,omitempty"`
	// TotalSize is the total package file size in bytes.
	TotalSize int64 `json:"total_size"`
	// ItemCount is the number of items in the package.
	ItemCount uint32 `json:"item_count"`
	// SFO contains the parsed PARAM.SFO data if available.
	SFO *sfo.Info `json:"sfo,omitempty"`
}

// GamePlatform implements core.GameInfo.
func (i *Info) GamePlatform() core.Platform { return i.Platform }

// GameTitle implements core.GameInfo.
func (i *Info) GameTitle() string {
	if i.SFO != nil && i.SFO.Title != "" {
		return i.SFO.Title
	}
	return i.Title
}

// GameSerial implements core.GameInfo. Returns the full Content ID.
func (i *Info) GameSerial() string { return i.ContentID }

// GameRegions implements core.GameInfo.
func (i *Info) GameRegions() []core.Region {
	// Delegate to SFO if available
	if i.SFO != nil {
		return i.SFO.GameRegions()
	}
	// Infer from ContentID prefix (format: "XX####-TITLEID_##-...")
	// The first 2 chars indicate region: UP=US, EP=EU, JP=JP, HP=Asia, KP=Korea
	if len(i.ContentID) >= 2 {
		prefix := i.ContentID[:2]
		switch prefix {
		case "UP":
			return []core.Region{core.RegionUSA}
		case "EP":
			return []core.Region{core.RegionEurope}
		case "JP":
			return []core.Region{core.RegionJapan}
		case "HP":
			return []core.Region{core.RegionAsia}
		case "KP":
			return []core.Region{core.RegionKorea}
		}
	}
	return []core.Region{}
}

// Parse extracts game information from a PlayStation PKG file.
func Parse(r io.ReaderAt, size int64) (*Info, error) {
	if size < pkgHeaderSize {
		return nil, fmt.Errorf("file too small for PKG header: need %d bytes, got %d", pkgHeaderSize, size)
	}

	header := make([]byte, pkgHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read PKG header: %w", err)
	}

	// Validate magic
	if string(header[0:4]) != pkgMagic {
		return nil, fmt.Errorf("invalid PKG magic: got %x, expected %x", header[0:4], []byte(pkgMagic))
	}

	// Parse header fields (big-endian)
	pkgType := Type(binary.BigEndian.Uint16(header[pkgTypeOffset:]))
	metadataOffset := binary.BigEndian.Uint32(header[pkgMetadataOffOffset:])
	metadataCount := binary.BigEndian.Uint32(header[pkgMetadataCountOff:])
	itemCount := binary.BigEndian.Uint32(header[pkgItemCountOffset:])
	totalSize := binary.BigEndian.Uint64(header[pkgTotalSizeOffset:])

	// Extract Content ID (null-terminated string)
	contentID := util.ExtractASCII(header[pkgContentIDOffset : pkgContentIDOffset+pkgContentIDLen])

	// Parse metadata to get content type and SFO info
	var contentType ContentType
	var sfoOffset, sfoSize uint32
	if metadataCount > 0 && metadataOffset > 0 {
		var err error
		contentType, sfoOffset, sfoSize, err = parseMetadata(r, metadataOffset, metadataCount, size)
		if err != nil {
			// Non-fatal: continue without metadata
			contentType = 0
			sfoOffset = 0
			sfoSize = 0
		}
	}

	// Read extended header key ID for PSP/Vita disambiguation
	var extKeyID uint32
	if pkgType == TypePSPVita && size >= pkgExtHeaderOffset+pkgExtKeyIDOffset+4 {
		extHeader := make([]byte, pkgExtKeyIDOffset+4)
		if _, err := r.ReadAt(extHeader, pkgExtHeaderOffset); err == nil {
			// Check for extended header magic
			if string(extHeader[0:4]) == pkgExtMagic {
				extKeyID = binary.BigEndian.Uint32(extHeader[pkgExtKeyIDOffset:])
			}
		}
	}

	// Detect platform
	platform := detectPlatform(pkgType, contentType, extKeyID)

	info := &Info{
		ContentID:   contentID,
		Platform:    platform,
		Type:        pkgType,
		ContentType: contentType,
		TotalSize:   int64(totalSize),
		ItemCount:   itemCount,
	}

	// Try to parse embedded PARAM.SFO
	if sfoOffset > 0 && sfoSize > 0 && int64(sfoOffset)+int64(sfoSize) <= size {
		sfoData := make([]byte, sfoSize)
		if _, err := r.ReadAt(sfoData, int64(sfoOffset)); err == nil {
			if sfoInfo, err := sfo.Parse(bytes.NewReader(sfoData), int64(sfoSize)); err == nil {
				info.SFO = sfoInfo
				if sfoInfo.Title != "" {
					info.Title = sfoInfo.Title
				}
			}
		}
	}

	return info, nil
}

// parseMetadata reads PKG metadata entries to extract content type and SFO info.
func parseMetadata(r io.ReaderAt, offset uint32, count uint32, fileSize int64) (ContentType, uint32, uint32, error) {
	var contentType ContentType
	var sfoOffset, sfoSize uint32

	pos := int64(offset)
	for i := uint32(0); i < count; i++ {
		// Bounds check
		if pos+8 > fileSize {
			break
		}

		// Read entry header: ID (4 bytes) + Size (4 bytes)
		entryHeader := make([]byte, 8)
		if _, err := r.ReadAt(entryHeader, pos); err != nil {
			return contentType, sfoOffset, sfoSize, err
		}

		id := binary.BigEndian.Uint32(entryHeader[0:4])
		dataSize := binary.BigEndian.Uint32(entryHeader[4:8])

		// Bounds check for data
		if pos+8+int64(dataSize) > fileSize {
			break
		}

		switch id {
		case metadataContentType:
			if dataSize >= 4 {
				data := make([]byte, 4)
				if _, err := r.ReadAt(data, pos+8); err == nil {
					contentType = ContentType(binary.BigEndian.Uint32(data))
				}
			}
		case metadataSFOInfo:
			if dataSize >= 8 {
				data := make([]byte, 8)
				if _, err := r.ReadAt(data, pos+8); err == nil {
					sfoOffset = binary.BigEndian.Uint32(data[0:4])
					sfoSize = binary.BigEndian.Uint32(data[4:8])
				}
			}
		}

		pos += 8 + int64(dataSize)
	}

	return contentType, sfoOffset, sfoSize, nil
}

// detectPlatform determines the PlayStation platform from PKG metadata.
func detectPlatform(pkgType Type, contentType ContentType, extKeyID uint32) core.Platform {
	// Primary: content type
	switch contentType {
	case ContentTypePSVitaGame, ContentTypePSVitaDLC, ContentTypePSVitaLA, ContentTypePSVitaTheme:
		return core.PlatformPSVita
	case ContentTypePSM, ContentTypePSMUnity:
		return core.PlatformPSM
	case ContentTypePSP, ContentTypePSPGo, ContentTypeMinis, ContentTypeNeoGeo:
		return core.PlatformPSP
	case ContentTypeGameData, ContentTypeGameExec, ContentTypeTheme,
		ContentTypeWidget, ContentTypeLicense, ContentTypeVSHModule, ContentTypePSNAvatar:
		return core.PlatformPS3
	case ContentTypePS1Emu:
		// Could be PS3 or PSP hosting PS1
		if pkgType == TypePSPVita {
			return core.PlatformPSP
		}
		return core.PlatformPS3
	}

	// Secondary: pkg_type + extended header key_id
	if pkgType == TypePS3 {
		return core.PlatformPS3
	}

	// PKGTypePSPVita - check extended header key index
	keyIndex := extKeyID & 0xF
	switch keyIndex {
	case 2, 3: // Vita key indices
		return core.PlatformPSVita
	case 4: // PSM key index
		return core.PlatformPSM
	}

	return core.PlatformPSP // Default for type 0x0002
}
