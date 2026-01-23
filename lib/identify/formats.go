package identify

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/lib/chd"
	"github.com/sargunv/rom-tools/lib/iso9660"
	"github.com/sargunv/rom-tools/lib/roms/dreamcast"
	"github.com/sargunv/rom-tools/lib/roms/gamecube"
	"github.com/sargunv/rom-tools/lib/roms/gb"
	"github.com/sargunv/rom-tools/lib/roms/gba"
	"github.com/sargunv/rom-tools/lib/roms/megadrive"
	"github.com/sargunv/rom-tools/lib/roms/n64"
	"github.com/sargunv/rom-tools/lib/roms/nds"
	"github.com/sargunv/rom-tools/lib/roms/nes"
	"github.com/sargunv/rom-tools/lib/roms/playstation_cnf"
	"github.com/sargunv/rom-tools/lib/roms/playstation_sfo"
	"github.com/sargunv/rom-tools/lib/roms/saturn"
	"github.com/sargunv/rom-tools/lib/roms/sms"
	"github.com/sargunv/rom-tools/lib/roms/snes"
	"github.com/sargunv/rom-tools/lib/roms/xbox"
)

// Simple parsers: GBA, GB, NES, NDS, SNES, SMS

func identifyGBA(r io.ReaderAt, size int64) (GameInfo, error) {
	return gba.ParseGBA(r, size)
}

func identifyGB(r io.ReaderAt, size int64) (GameInfo, error) {
	return gb.ParseGB(r, size)
}

func identifyNES(r io.ReaderAt, size int64) (GameInfo, error) {
	return nes.ParseNES(r, size)
}

func identifyNDS(r io.ReaderAt, size int64) (GameInfo, error) {
	return nds.ParseNDS(r, size)
}

func identifySNES(r io.ReaderAt, size int64) (GameInfo, error) {
	return snes.ParseSNES(r, size)
}

func identifySMS(r io.ReaderAt, size int64) (GameInfo, error) {
	return sms.ParseSMS(r, size)
}

// MD, GCM, XBE

func identifyMD(r io.ReaderAt, size int64) (GameInfo, error) {
	info, err := megadrive.Parse(r, size)
	if err != nil {
		return nil, err
	}
	if info.SourceFormat != megadrive.FormatMD {
		return nil, fmt.Errorf("format mismatch: expected MD, got SMD")
	}
	return info, nil
}

func identifyGCM(r io.ReaderAt, size int64) (GameInfo, error) {
	return gamecube.ParseGCM(r, size)
}

func identifyXBE(r io.ReaderAt, size int64) (GameInfo, error) {
	return xbox.ParseXBE(r, size)
}

// N64: unified ParseN64 handles all byte orderings

func identifyZ64(r io.ReaderAt, size int64) (GameInfo, error) {
	info, err := n64.ParseN64(r, size)
	if err != nil {
		return nil, err
	}
	if info.ByteOrder != n64.N64BigEndian {
		return nil, fmt.Errorf("byte order mismatch: expected z64, got %s", info.ByteOrder)
	}
	return info, nil
}

func identifyV64(r io.ReaderAt, size int64) (GameInfo, error) {
	info, err := n64.ParseN64(r, size)
	if err != nil {
		return nil, err
	}
	if info.ByteOrder != n64.N64ByteSwapped {
		return nil, fmt.Errorf("byte order mismatch: expected v64, got %s", info.ByteOrder)
	}
	return info, nil
}

func identifyN64(r io.ReaderAt, size int64) (GameInfo, error) {
	info, err := n64.ParseN64(r, size)
	if err != nil {
		return nil, err
	}
	if info.ByteOrder != n64.N64LittleEndian {
		return nil, fmt.Errorf("byte order mismatch: expected n64, got %s", info.ByteOrder)
	}
	return info, nil
}

// Delegation: SMD->MD, RVZ->GCM, XISO->XBE, CHD->ISO9660

func identifySMD(r io.ReaderAt, size int64) (GameInfo, error) {
	info, err := megadrive.Parse(r, size)
	if err != nil {
		return nil, err
	}
	if info.SourceFormat != megadrive.FormatSMD {
		return nil, fmt.Errorf("format mismatch: expected SMD, got MD")
	}
	return info, nil
}

func identifyRVZ(r io.ReaderAt, size int64) (GameInfo, error) {
	return gamecube.ParseRVZ(r, size)
}

func identifyXISO(r io.ReaderAt, size int64) (GameInfo, error) {
	return xbox.ParseXISO(r, size)
}

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

// Dispatcher: ISO9660 (tries Saturn, Dreamcast, PlayStation, PSP)

func identifyISO9660(r io.ReaderAt, size int64) (GameInfo, error) {
	reader, err := iso9660.NewReader(r, size)
	if err != nil {
		return nil, err
	}

	// Try to read system area (sector 0) for Saturn/Dreamcast identification
	systemArea := make([]byte, 2048)
	if _, err := reader.ReadAt(systemArea, 0); err == nil {
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
