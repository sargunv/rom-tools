package identify

import (
	"bytes"
	"io"

	"github.com/sargunv/rom-tools/lib/chd"
	"github.com/sargunv/rom-tools/lib/core"
	"github.com/sargunv/rom-tools/lib/iso9660"
	"github.com/sargunv/rom-tools/lib/roms/playstation/cnf"
	"github.com/sargunv/rom-tools/lib/roms/playstation/sfo"
	"github.com/sargunv/rom-tools/lib/roms/sega/dreamcast"
	"github.com/sargunv/rom-tools/lib/roms/sega/md"
	"github.com/sargunv/rom-tools/lib/roms/sega/saturn"
)

func identifyCHD(r io.ReaderAt, size int64) (core.GameInfo, core.Hashes, error) {
	reader, err := chd.NewReader(r, size)
	if err != nil {
		return nil, nil, err
	}

	header := reader.Header()
	hashes := core.Hashes{
		core.HashCHDUncompressedSHA1: header.RawSHA1,
		core.HashCHDCompressedSHA1:   header.SHA1,
	}

	// Find first non-audio track and try to identify its content.
	// Errors are intentionally ignored: many disc formats (Sega CD, Saturn,
	// Dreamcast) use custom headers rather than ISO9660. Failure to parse
	// just means we return CHD hashes without game metadata, which is fine
	// since CHD hashes are the primary identifier for DAT matching.
	for _, track := range reader.Tracks {
		if track.Type != "AUDIO" {
			content, _, _ := identifyISO9660(track.Open(), track.Size())
			if content != nil {
				return content, hashes, nil
			}
			break
		}
	}

	// Try raw CHD access (for hard disk images, etc.)
	content, _, _ := identifyISO9660(reader, reader.Size())
	return content, hashes, nil
}

func identifyISO9660(r io.ReaderAt, size int64) (core.GameInfo, core.Hashes, error) {
	reader, err := iso9660.NewReader(r, size)
	if err != nil {
		return nil, nil, err
	}

	// Try to read system area (sector 0) for Sega CD/Saturn/Dreamcast identification
	systemArea := make([]byte, 2048)
	if _, err := reader.ReadAt(systemArea, 0); err == nil {
		if info, err := md.ParseSegaCD(bytes.NewReader(systemArea), int64(len(systemArea))); err == nil {
			return info, nil, nil
		}
		if info, err := saturn.ParseSaturn(bytes.NewReader(systemArea), int64(len(systemArea))); err == nil {
			return info, nil, nil
		}
		if info, err := dreamcast.ParseDreamcast(bytes.NewReader(systemArea), int64(len(systemArea))); err == nil {
			return info, nil, nil
		}
	}

	// Try to read SYSTEM.CNF (PS1/PS2 discs)
	if fileReader, fileSize, err := reader.OpenFile("SYSTEM.CNF"); err == nil {
		data := make([]byte, fileSize)
		if _, err := fileReader.ReadAt(data, 0); err == nil {
			if info, err := cnf.Parse(bytes.NewReader(data), fileSize); err == nil {
				return info, nil, nil
			}
		}
	}

	// Try to read PSP_GAME/PARAM.SFO (PSP/PS3/Vita/PS4 discs)
	if fileReader, fileSize, err := reader.OpenFile("PSP_GAME/PARAM.SFO"); err == nil {
		data := make([]byte, fileSize)
		if _, err := fileReader.ReadAt(data, 0); err == nil {
			if info, err := sfo.Parse(bytes.NewReader(data), fileSize); err == nil {
				return info, nil, nil
			}
		}
	}

	// Valid ISO9660 filesystem but no recognized game content.
	// This is expected for data discs, unsupported platforms, etc.
	// Returning nil allows the caller to try other parsers or fall back
	// to hash-only identification, which is sufficient for DAT matching.
	return nil, nil, nil
}
