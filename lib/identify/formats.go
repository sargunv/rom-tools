package identify

import (
	"bytes"
	"io"

	"github.com/sargunv/rom-tools/lib/chd"
	"github.com/sargunv/rom-tools/lib/iso9660"
	"github.com/sargunv/rom-tools/lib/roms/dreamcast"
	"github.com/sargunv/rom-tools/lib/roms/megadrive"
	"github.com/sargunv/rom-tools/lib/roms/playstation_cnf"
	"github.com/sargunv/rom-tools/lib/roms/playstation_sfo"
	"github.com/sargunv/rom-tools/lib/roms/saturn"
)

func identifyCHD(r io.ReaderAt, size int64) (GameInfo, error) {
	reader, err := chd.NewReader(r, size)
	if err != nil {
		return nil, err
	}

	// Find first non-audio track
	for _, track := range reader.Tracks {
		if track.Type != "AUDIO" {
			return identifyISO9660(track.Open(), track.Size())
		}
	}

	// No tracks or all audio - try raw CHD access
	return identifyISO9660(reader, reader.Size())
}

func identifyISO9660(r io.ReaderAt, size int64) (GameInfo, error) {
	reader, err := iso9660.NewReader(r, size)
	if err != nil {
		return nil, err
	}

	// Try to read system area (sector 0) for Sega CD/Saturn/Dreamcast identification
	systemArea := make([]byte, 2048)
	if _, err := reader.ReadAt(systemArea, 0); err == nil {
		if info, err := megadrive.ParseSegaCD(bytes.NewReader(systemArea), int64(len(systemArea))); err == nil {
			return info, nil
		}
		if info, err := saturn.ParseSaturn(bytes.NewReader(systemArea), int64(len(systemArea))); err == nil {
			return info, nil
		}
		if info, err := dreamcast.ParseDreamcast(bytes.NewReader(systemArea), int64(len(systemArea))); err == nil {
			return info, nil
		}
	}

	// Try to read SYSTEM.CNF (PS1/PS2 discs)
	if fileReader, fileSize, err := reader.OpenFile("SYSTEM.CNF"); err == nil {
		data := make([]byte, fileSize)
		if _, err := fileReader.ReadAt(data, 0); err == nil {
			if info, err := playstation_cnf.Parse(bytes.NewReader(data), fileSize); err == nil {
				return info, nil
			}
		}
	}

	// Try to read PSP_GAME/PARAM.SFO (PSP/PS3/Vita/PS4 discs)
	if fileReader, fileSize, err := reader.OpenFile("PSP_GAME/PARAM.SFO"); err == nil {
		data := make([]byte, fileSize)
		if _, err := fileReader.ReadAt(data, 0); err == nil {
			if info, err := playstation_sfo.Parse(bytes.NewReader(data), fileSize); err == nil {
				return info, nil
			}
		}
	}

	// No platform identified - valid ISO but unknown content
	return nil, nil
}
