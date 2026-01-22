package identify

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/sargunv/rom-tools/lib/core"
	"github.com/sargunv/rom-tools/lib/format/chd"
	"github.com/sargunv/rom-tools/lib/format/dreamcast"
	"github.com/sargunv/rom-tools/lib/format/gamecube"
	"github.com/sargunv/rom-tools/lib/format/gb"
	"github.com/sargunv/rom-tools/lib/format/gba"
	"github.com/sargunv/rom-tools/lib/format/iso9660"
	"github.com/sargunv/rom-tools/lib/format/megadrive"
	"github.com/sargunv/rom-tools/lib/format/n64"
	"github.com/sargunv/rom-tools/lib/format/nds"
	"github.com/sargunv/rom-tools/lib/format/nes"
	"github.com/sargunv/rom-tools/lib/format/playstation_cnf"
	"github.com/sargunv/rom-tools/lib/format/playstation_sfo"
	"github.com/sargunv/rom-tools/lib/format/saturn"
	"github.com/sargunv/rom-tools/lib/format/sms"
	"github.com/sargunv/rom-tools/lib/format/snes"
	"github.com/sargunv/rom-tools/lib/format/xbox"
)

// Helper functions to convert format info to GameIdent

func xbeInfoToGameIdent(info *xbox.XBEInfo) *GameIdent {
	return &GameIdent{
		Platform: core.PlatformXbox,
		Title:    info.Title,
		Serial:   fmt.Sprintf("%s-%03d", info.PublisherCode, info.GameNumber),
		Extra:    info,
	}
}

func mdInfoToGameIdent(info *megadrive.MDInfo) *GameIdent {
	// Use overseas title if available, otherwise domestic title
	title := info.OverseasTitle
	if title == "" {
		title = info.DomesticTitle
	}
	return &GameIdent{
		Platform: core.PlatformMD,
		Title:    title,
		Serial:   info.SerialNumber,
		Extra:    info,
	}
}

func gcmInfoToGameIdent(info *gamecube.GCMInfo, extra any) *GameIdent {
	// Build the full game ID (DiscID + GameCode + RegionCode)
	serial := fmt.Sprintf("%c%s%c", info.DiscID, info.GameCode, info.RegionCode)

	// If no extra provided, use GCMInfo itself
	if extra == nil {
		extra = info
	}

	return &GameIdent{
		Platform: info.Platform,
		Title:    info.Title,
		Serial:   serial,
		Extra:    extra,
	}
}

func n64InfoToGameIdent(info *n64.N64Info) *GameIdent {
	return &GameIdent{
		Platform: core.PlatformN64,
		Title:    info.Title,
		Serial:   info.GameCode,
		Extra:    info,
	}
}

func saturnInfoToGameIdent(info *saturn.SaturnInfo) *GameIdent {
	return &GameIdent{
		Platform: core.PlatformSaturn,
		Title:    info.Title,
		Serial:   info.ProductNumber,
		Extra:    info,
	}
}

func dreamcastInfoToGameIdent(info *dreamcast.DreamcastInfo) *GameIdent {
	return &GameIdent{
		Platform: core.PlatformDreamcast,
		Title:    info.Title,
		Serial:   info.ProductNumber,
		Extra:    info,
	}
}

func cnfInfoToGameIdent(info *playstation_cnf.CNFInfo) *GameIdent {
	return &GameIdent{
		Platform: info.Platform,
		Serial:   info.DiscID(),
		Extra:    info,
	}
}

func sfoInfoToGameIdent(info *playstation_sfo.SFOInfo) *GameIdent {
	// Normalize disc ID: add hyphen after 4-char prefix if not present
	normalizedID := info.DiscID
	if !strings.Contains(normalizedID, "-") && len(normalizedID) > 4 {
		normalizedID = normalizedID[:4] + "-" + normalizedID[4:]
	}

	return &GameIdent{
		Platform: info.Platform,
		Title:    info.Title,
		Serial:   normalizedID,
		Extra:    info,
	}
}

// Simple parsers: GBA, GB, NES, NDS, SNES, SMS

func identifyGBA(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := gba.ParseGBA(r, size)
	if err != nil {
		return nil, err
	}
	return &GameIdent{
		Platform: core.PlatformGBA,
		Title:    info.Title,
		Serial:   info.GameCode,
		Extra:    info,
	}, nil
}

func identifyGB(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := gb.ParseGB(r, size)
	if err != nil {
		return nil, err
	}
	return &GameIdent{
		Platform: info.Platform,
		Title:    info.Title,
		Extra:    info,
	}, nil
}

func identifyNES(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := nes.ParseNES(r, size)
	if err != nil {
		return nil, err
	}
	return &GameIdent{
		Platform: core.PlatformNES,
		Extra:    info,
	}, nil
}

func identifyNDS(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := nds.ParseNDS(r, size)
	if err != nil {
		return nil, err
	}
	return &GameIdent{
		Platform: info.Platform,
		Title:    info.Title,
		Serial:   info.GameCode,
		Extra:    info,
	}, nil
}

func identifySNES(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := snes.ParseSNES(r, size)
	if err != nil {
		return nil, err
	}
	return &GameIdent{
		Platform: core.PlatformSNES,
		Title:    info.Title,
		Extra:    info,
	}, nil
}

func identifySMS(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := sms.ParseSMS(r, size)
	if err != nil {
		return nil, err
	}
	return &GameIdent{
		Platform: info.Platform,
		Serial:   info.ProductCode,
		Extra:    info,
	}, nil
}

// MD, GCM, XBE: use helper functions

func identifyMD(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := megadrive.ParseMD(r, size)
	if err != nil {
		return nil, err
	}
	return mdInfoToGameIdent(info), nil
}

func identifyGCM(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := gamecube.ParseGCM(r, size)
	if err != nil {
		return nil, err
	}
	return gcmInfoToGameIdent(info, nil), nil
}

func identifyXBE(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := xbox.ParseXBE(r, size)
	if err != nil {
		return nil, err
	}
	return xbeInfoToGameIdent(info), nil
}

// N64: unified ParseN64 handles all byte orderings

func identifyZ64(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := n64.ParseN64(r, size)
	if err != nil {
		return nil, err
	}
	if info.ByteOrder != n64.N64BigEndian {
		return nil, fmt.Errorf("byte order mismatch: expected z64, got %s", info.ByteOrder)
	}
	return n64InfoToGameIdent(info), nil
}

func identifyV64(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := n64.ParseN64(r, size)
	if err != nil {
		return nil, err
	}
	if info.ByteOrder != n64.N64ByteSwapped {
		return nil, fmt.Errorf("byte order mismatch: expected v64, got %s", info.ByteOrder)
	}
	return n64InfoToGameIdent(info), nil
}

func identifyN64(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := n64.ParseN64(r, size)
	if err != nil {
		return nil, err
	}
	if info.ByteOrder != n64.N64LittleEndian {
		return nil, fmt.Errorf("byte order mismatch: expected n64, got %s", info.ByteOrder)
	}
	return n64InfoToGameIdent(info), nil
}

// Delegation: SMD->MD, RVZ->GCM, XISO->XBE, CHD->ISO9660

func identifySMD(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := megadrive.ParseSMD(r, size)
	if err != nil {
		return nil, err
	}
	return mdInfoToGameIdent(info), nil
}

func identifyRVZ(r io.ReaderAt, size int64) (*GameIdent, error) {
	rvzInfo, err := gamecube.ParseRVZ(r, size)
	if err != nil {
		return nil, err
	}
	gcmInfo, err := gamecube.ParseGCM(bytes.NewReader(rvzInfo.DiscHeader), int64(len(rvzInfo.DiscHeader)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse disc header from RVZ: %w", err)
	}
	extra := &struct {
		GCMInfo *gamecube.GCMInfo
		RVZInfo *gamecube.RVZInfo
	}{
		GCMInfo: gcmInfo,
		RVZInfo: rvzInfo,
	}
	return gcmInfoToGameIdent(gcmInfo, extra), nil
}

func identifyXISO(r io.ReaderAt, size int64) (*GameIdent, error) {
	info, err := xbox.ParseXISO(r, size)
	if err != nil {
		return nil, err
	}
	return xbeInfoToGameIdent(info), nil
}

func identifyCHD(r io.ReaderAt, size int64) (*GameIdent, error) {
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

func identifyISO9660(r io.ReaderAt, size int64) (*GameIdent, error) {
	img, err := iso9660.Open(r, size)
	if err != nil {
		return nil, err
	}

	// Try to read system area (sectors 0-15) for Saturn/Dreamcast identification
	if data, err := img.ReadSystemArea(); err == nil {
		if info, err := saturn.ParseSaturn(bytes.NewReader(data), int64(len(data))); err == nil {
			return saturnInfoToGameIdent(info), nil
		}
		if info, err := dreamcast.ParseDreamcast(bytes.NewReader(data), int64(len(data))); err == nil {
			return dreamcastInfoToGameIdent(info), nil
		}
	}

	// Try to read SYSTEM.CNF (PS1/PS2 discs)
	if data, err := img.ReadFile("SYSTEM.CNF"); err == nil {
		if info, err := playstation_cnf.ParseCNF(bytes.NewReader(data), int64(len(data))); err == nil {
			return cnfInfoToGameIdent(info), nil
		}
	}

	// Try to read PSP_GAME/PARAM.SFO (PSP/PS3/Vita/PS4 discs)
	if data, err := img.ReadFile("PSP_GAME/PARAM.SFO"); err == nil {
		if info, err := playstation_sfo.ParseSFO(bytes.NewReader(data), int64(len(data))); err == nil {
			return sfoInfoToGameIdent(info), nil
		}
	}

	// No platform identified - valid ISO but unknown content
	return nil, nil
}
