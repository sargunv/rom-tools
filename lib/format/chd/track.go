package chd

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/sargunv/rom-tools/internal/util"
)

// rawSectorSize is the size of a raw CD sector (2352 bytes).
const rawSectorSize = 2352

// Track represents a single track in the CHD (like zip.File).
type Track struct {
	Number int    // Track number (1-based)
	Frames int    // Number of frames in the track
	Pregap int    // Pregap frames
	Type   string // Raw type string: "AUDIO", "MODE1_RAW", "MODE2_RAW", etc.

	// unexported
	reader     *Reader
	startFrame int64
}

// Open returns a reader for this track's raw sector data (2352 bytes/sector).
func (t *Track) Open() io.ReaderAt {
	return &trackReader{
		reader:     t.reader,
		track:      t,
		numSectors: int64(t.Frames),
	}
}

// Size returns the track size in bytes (Frames * 2352).
func (t *Track) Size() int64 {
	return int64(t.Frames) * rawSectorSize
}

// trackReader provides access to a track's raw sector data within a CHD file.
type trackReader struct {
	reader     *Reader
	track      *Track
	numSectors int64
}

// ReadAt implements io.ReaderAt for raw track data.
func (tr *trackReader) ReadAt(p []byte, off int64) (int, error) {
	n := 0
	for n < len(p) {
		logicalOffset := off + int64(n)

		// Which sector?
		sector := logicalOffset / rawSectorSize
		offsetInSector := int(logicalOffset % rawSectorSize)

		if sector >= tr.numSectors {
			if n > 0 {
				return n, nil
			}
			return 0, io.EOF
		}

		// Calculate actual sector number in the CHD (skip pregap)
		actualSector := uint64(tr.track.startFrame + int64(tr.track.Pregap) + sector)

		// Read the physical sector from CHD
		sectorData, err := tr.reader.readSector(actualSector)
		if err != nil {
			if n > 0 {
				return n, nil
			}
			return 0, err
		}

		// Return raw sector data (first 2352 bytes, strip subcode if present)
		endOffset := min(rawSectorSize, len(sectorData))
		bytesToCopy := min(endOffset-offsetInSector, len(p)-n)
		if bytesToCopy > 0 {
			copy(p[n:n+bytesToCopy], sectorData[offsetInSector:offsetInSector+bytesToCopy])
		}
		n += bytesToCopy
	}

	return n, nil
}

// MetadataTag represents a 4-character CHD metadata tag.
type MetadataTag string

// Known metadata tag types from MAME chd.h.
const (
	TagHardDisk      MetadataTag = "GDDD" // Hard disk geometry
	TagHardDiskIdent MetadataTag = "IDNT" // Hard disk identify information
	TagHardDiskKey   MetadataTag = "KEY " // Hard disk key information
	TagPCMCIACIS     MetadataTag = "CIS " // PCMCIA CIS information
	TagCDROMOld      MetadataTag = "CHCD" // Legacy CD-ROM metadata
	TagCDROM         MetadataTag = "CHTR" // CD-ROM track metadata
	TagCDROM2        MetadataTag = "CHT2" // CD-ROM track metadata v2
	TagGDROMOld      MetadataTag = "CHGT" // Legacy GD-ROM metadata
	TagGDROM         MetadataTag = "CHGD" // GD-ROM track metadata
	TagDVD           MetadataTag = "DVD " // DVD metadata
	TagAV            MetadataTag = "AVAV" // A/V metadata
	TagAVLaserdisc   MetadataTag = "AVLD" // A/V laserdisc frame metadata
)

// parseTrackMetadata reads metadata and extracts track information.
func parseTrackMetadata(r io.ReaderAt, header *Header, reader *Reader) ([]*Track, error) {
	metaOffset := binary.BigEndian.Uint64(make([]byte, 8))

	// Read metadata offset from header (bytes 48-55)
	buf := make([]byte, 8)
	if _, err := r.ReadAt(buf, 48); err != nil {
		return nil, fmt.Errorf("read metadata offset: %w", err)
	}
	metaOffset = binary.BigEndian.Uint64(buf)

	if metaOffset == 0 {
		return nil, nil // No metadata
	}

	var tracks []*Track
	offset := metaOffset

	for offset != 0 {
		// Read metadata entry header (16 bytes):
		//   [0-3]   uint32 tag (big-endian, ASCII)
		//   [4-7]   uint32 length + flags (24-bit length, 8-bit flags)
		//   [8-15]  uint64 next offset
		entryHeader := make([]byte, 16)
		if _, err := r.ReadAt(entryHeader, int64(offset)); err != nil {
			return nil, fmt.Errorf("read metadata header at offset %d: %w", offset, err)
		}

		tag := MetadataTag(util.ExtractASCII(entryHeader[0:4]))
		lengthFlags := binary.BigEndian.Uint32(entryHeader[4:8])
		length := lengthFlags & 0x00FFFFFF // Lower 24 bits
		nextOffset := binary.BigEndian.Uint64(entryHeader[8:16])

		// Read payload
		data := make([]byte, length)
		if length > 0 {
			if _, err := r.ReadAt(data, int64(offset)+16); err != nil {
				return nil, fmt.Errorf("read metadata payload at offset %d: %w", offset+16, err)
			}
		}

		// Parse track metadata (CHTR, CHT2, CHGD all use same format)
		if tag == TagCDROM || tag == TagCDROM2 || tag == TagGDROM {
			if track, err := parseTrackMetadataEntry(data); err == nil {
				track.reader = reader
				tracks = append(tracks, track)
			}
		}

		offset = nextOffset
	}

	// Calculate start frames for each track
	var currentFrame int64
	for _, track := range tracks {
		track.startFrame = currentFrame
		currentFrame += int64(track.Pregap + track.Frames)
	}

	return tracks, nil
}

// parseTrackMetadataEntry parses track metadata from CHTR, CHT2, or CHGD format.
// All formats use space-separated KEY:VALUE pairs with at least TRACK, TYPE, FRAMES.
func parseTrackMetadataEntry(data []byte) (*Track, error) {
	str := strings.TrimRight(string(data), "\x00")
	fields := parseMetadataFields(str)

	track := &Track{}

	if v, ok := fields["TRACK"]; ok {
		track.Number, _ = strconv.Atoi(v)
	}
	if v, ok := fields["TYPE"]; ok {
		track.Type = v
	}
	if v, ok := fields["FRAMES"]; ok {
		track.Frames, _ = strconv.Atoi(v)
	}
	if v, ok := fields["PREGAP"]; ok {
		track.Pregap, _ = strconv.Atoi(v)
	}

	if track.Number == 0 {
		return nil, fmt.Errorf("invalid track metadata")
	}
	return track, nil
}

// parseMetadataFields parses space-separated KEY:VALUE pairs.
func parseMetadataFields(s string) map[string]string {
	fields := make(map[string]string)
	for _, part := range strings.Fields(s) {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			fields[kv[0]] = kv[1]
		}
	}
	return fields
}
