package romident

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sargunv/rom-tools/lib/romident/container"
	"github.com/sargunv/rom-tools/lib/romident/container/folder"
	"github.com/sargunv/rom-tools/lib/romident/container/zip"
	"github.com/sargunv/rom-tools/lib/romident/disc/chd"
	"github.com/sargunv/rom-tools/lib/romident/game/gb"
	"github.com/sargunv/rom-tools/lib/romident/game/gba"
	"github.com/sargunv/rom-tools/lib/romident/game/md"
	"github.com/sargunv/rom-tools/lib/romident/game/n64"
	"github.com/sargunv/rom-tools/lib/romident/game/nds"
	"github.com/sargunv/rom-tools/lib/romident/game/nes"
	"github.com/sargunv/rom-tools/lib/romident/game/snes"
	"github.com/sargunv/rom-tools/lib/romident/game/xbox"
)

// IdentifyROM identifies a ROM file, ZIP archive, or folder.
// Returns a ROM struct with file information and hashes.
func IdentifyROM(path string, opts Options) (*ROM, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return identifyFolder(absPath, opts)
	}

	return identifyFile(absPath, info.Size(), opts)
}

// identifyContainer processes any Container (folder, ZIP, etc.) and identifies all files within.
func identifyContainer(c container.Container, containerType ROMType, containerPath string, opts Options) (*ROM, error) {
	entries := c.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("container is empty")
	}

	files := make(Files)
	detector := NewDetector()
	var romIdent *GameIdent

	for _, entry := range entries {
		// Open file for identification
		reader, err := c.OpenFileAt(entry.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", entry.Name, err)
		}

		romFile, fileIdent, err := identifySingleReader(reader, entry.Name, detector, opts)
		if err != nil {
			reader.Close()
			return nil, fmt.Errorf("failed to identify %s: %w", entry.Name, err)
		}

		// Add pre-computed CRC32 from container metadata if available and not already calculated
		if entry.CRC32 != 0 {
			hasCRC32 := false
			for _, h := range romFile.Hashes {
				if h.Algorithm == HashCRC32 {
					hasCRC32 = true
					break
				}
			}
			if !hasCRC32 {
				romFile.Hashes = append(romFile.Hashes, NewHash(HashCRC32, fmt.Sprintf("%08x", entry.CRC32), HashSourceZIPMetadata))
			}
		}

		// Collect identification (error if multiple identifications found)
		if fileIdent != nil {
			if romIdent != nil {
				reader.Close()
				return nil, fmt.Errorf("container has multiple game identifications: %s and %s", romIdent.TitleID, fileIdent.TitleID)
			}
			romIdent = fileIdent
		}

		reader.Close()
		files[entry.Name] = *romFile
	}

	return &ROM{
		Path:  containerPath,
		Type:  containerType,
		Files: files,
		Ident: romIdent,
	}, nil
}

func identifyFolder(path string, opts Options) (*ROM, error) {
	c, err := folder.NewFolderContainer(path)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return identifyContainer(c, ROMTypeFolder, path, opts)
}

func identifyFile(path string, size int64, opts Options) (*ROM, error) {
	// Open file for format detection
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Detect format
	detector := NewDetector()
	detectedFormat, err := detector.Detect(f, size, filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("failed to detect format: %w", err)
	}

	// Handle ZIP specially
	if detectedFormat == FormatZIP {
		return identifyZIP(path, opts)
	}

	// Single file
	romFile, ident, err := identifySingleFile(path, size, detector, opts)
	if err != nil {
		return nil, err
	}

	files := Files{
		filepath.Base(path): *romFile,
	}

	return &ROM{
		Path:  path,
		Type:  ROMTypeFile,
		Files: files,
		Ident: ident,
	}, nil
}

func identifyZIP(path string, opts Options) (*ROM, error) {
	handler := zip.NewZIPHandler()

	archive, err := handler.Open(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	entries := archive.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("ZIP archive is empty")
	}

	// Slow mode: use full container introspection (decompresses files)
	if opts.HashMode == HashModeSlow {
		return identifyContainer(archive, ROMTypeZIP, path, opts)
	}

	// Fast/default mode: use ZIP metadata only (no decompression)
	files := make(Files)
	detector := NewDetector()

	for _, entry := range entries {
		// Use extension-based format detection (no decompression)
		// In fast mode, we can't verify with magic, so only trust unambiguous extensions
		candidates := detector.CandidatesByExtension(entry.Name)
		detectedFormat := FormatUnknown
		if len(candidates) == 1 {
			detectedFormat = candidates[0]
		}

		hashes := []Hash{}
		if entry.CRC32 != 0 {
			hashes = []Hash{
				NewHash(HashCRC32, fmt.Sprintf("%08x", entry.CRC32), HashSourceZIPMetadata),
			}
		}

		files[entry.Name] = ROMFile{
			Size:   entry.Size,
			Format: detectedFormat,
			Hashes: hashes,
		}
	}

	return &ROM{
		Path:  path,
		Type:  ROMTypeZIP,
		Files: files,
		Ident: nil, // No identification in fast/default mode
	}, nil
}

// identifySingleReader identifies a file from a ReaderAtSeekCloser (works for any container).
// Returns the ROMFile, game identification (if any), and an error.
func identifySingleReader(r container.ReaderAtSeekCloser, name string, detector *Detector, opts Options) (*ROMFile, *GameIdent, error) {
	size := r.Size()

	// Detect format
	detectedFormat, err := detector.Detect(r, size, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to detect format: %w", err)
	}

	romFile := &ROMFile{
		Size:   size,
		Format: detectedFormat,
	}

	var ident *GameIdent

	// For CHD, always extract hashes from header (fast, no decompression)
	if detectedFormat == FormatCHD {
		chdInfo, err := chd.ParseCHDHeader(r)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse CHD header: %w", err)
		}
		romFile.Hashes = []Hash{
			NewHash(HashSHA1, chdInfo.RawSHA1, HashSourceCHDRaw),
			NewHash(HashSHA1, chdInfo.SHA1, HashSourceCHDCompressed),
		}
		return romFile, ident, nil
	}

	// Extract identification for identifiable formats
	if detectedFormat == FormatXISO {
		// Reset reader position for XISO parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyXISO(r, size)
		}
	}
	if detectedFormat == FormatXBE {
		// Reset reader position for XBE parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyXBE(r, size)
		}
	}
	if detectedFormat == FormatGBA {
		// Reset reader position for GBA parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyGBA(r, size)
		}
	}
	if detectedFormat == FormatZ64 || detectedFormat == FormatV64 || detectedFormat == FormatN64 {
		// Reset reader position for N64 parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyN64(r, size)
		}
	}
	if detectedFormat == FormatGB {
		// Reset reader position for GB parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyGB(r, size)
		}
	}
	if detectedFormat == FormatMD {
		// Reset reader position for MD parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyMD(r, size)
		}
	}
	if detectedFormat == FormatSMD {
		// Reset reader position for SMD parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifySMD(r, size)
		}
	}
	if detectedFormat == FormatNDS {
		// Reset reader position for NDS parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyNDS(r, size)
		}
	}
	if detectedFormat == FormatNES {
		// Reset reader position for NES parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyNES(r, size)
		}
	}
	if detectedFormat == FormatSNES {
		// Reset reader position for SNES parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifySNES(r, size)
		}
	}

	// Fast mode: skip calculating hashes for large files, but allow small files
	if opts.HashMode == HashModeFast && size >= FastModeSmallFileThreshold {
		return romFile, ident, nil
	}

	// For other formats, calculate hashes
	// Reset reader position
	if _, err := r.Seek(0, 0); err != nil {
		return nil, nil, fmt.Errorf("failed to seek: %w", err)
	}

	// Wrap ReaderAtSeekCloser as io.Reader for CalculateHashes
	reader := &readerAtWrapper{r: r}
	hashes, err := CalculateHashes(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate hashes: %w", err)
	}

	romFile.Hashes = hashes

	return romFile, ident, nil
}

func identifySingleFile(path string, size int64, detector *Detector, opts Options) (*ROMFile, *GameIdent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileReader := &containerFileReader{file: f, size: info.Size()}
	return identifySingleReader(fileReader, filepath.Base(path), detector, opts)
}

// containerFileReader wraps *os.File to implement ReaderAtSeekCloser.
type containerFileReader struct {
	file *os.File
	size int64
}

func (f *containerFileReader) ReadAt(p []byte, off int64) (n int, err error) {
	return f.file.ReadAt(p, off)
}

func (f *containerFileReader) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

func (f *containerFileReader) Size() int64 {
	return f.size
}

func (f *containerFileReader) Close() error {
	return f.file.Close()
}

// readerAtWrapper wraps ReaderAtSeekCloser to implement io.Reader.
type readerAtWrapper struct {
	r   container.ReaderAtSeekCloser
	pos int64
}

func (w *readerAtWrapper) Read(p []byte) (n int, err error) {
	n, err = w.r.ReadAt(p, w.pos)
	w.pos += int64(n)
	return n, err
}

// identifyXISO extracts game identification from an Xbox XISO.
// Returns nil if identification fails (non-fatal).
func identifyXISO(r io.ReaderAt, size int64) *GameIdent {
	info, err := xbox.ParseXISO(r, size)
	if err != nil {
		return nil
	}

	version := int(info.Version)
	discNumber := int(info.DiscNumber)

	return &GameIdent{
		Platform:   PlatformXbox,
		TitleID:    fmt.Sprintf("%s-%03d", info.PublisherCode, info.GameNumber),
		Title:      info.Title,
		Regions:    decodeXboxRegions(info.RegionFlags),
		MakerCode:  info.PublisherCode,
		Version:    &version,
		DiscNumber: &discNumber,
		Extra: map[string]string{
			"title_id_hex": info.TitleIDHex,
		},
	}
}

// identifyXBE extracts game identification from an Xbox XBE file.
// Returns nil if identification fails (non-fatal).
func identifyXBE(r io.ReaderAt, size int64) *GameIdent {
	info, err := xbox.ParseXBE(r, size)
	if err != nil {
		return nil
	}

	version := int(info.Version)
	discNumber := int(info.DiscNumber)

	return &GameIdent{
		Platform:   PlatformXbox,
		TitleID:    fmt.Sprintf("%s-%03d", info.PublisherCode, info.GameNumber),
		Title:      info.Title,
		Regions:    decodeXboxRegions(info.RegionFlags),
		MakerCode:  info.PublisherCode,
		Version:    &version,
		DiscNumber: &discNumber,
		Extra: map[string]string{
			"title_id_hex": info.TitleIDHex,
		},
	}
}

// identifyGBA extracts game identification from a GBA ROM file.
// Returns nil if identification fails (non-fatal).
func identifyGBA(r io.ReaderAt, size int64) *GameIdent {
	info, err := gba.ParseGBA(r, size)
	if err != nil {
		return nil
	}

	version := info.Version

	return &GameIdent{
		Platform:  PlatformGBA,
		TitleID:   info.GameCode,
		Title:     info.Title,
		Regions:   []Region{decodeGBARegion(info.RegionCode)},
		MakerCode: info.MakerCode,
		Version:   &version,
	}
}

// identifyN64 extracts game identification from an N64 ROM file.
// Returns nil if identification fails (non-fatal).
func identifyN64(r io.ReaderAt, size int64) *GameIdent {
	info, err := n64.ParseN64(r, size)
	if err != nil {
		return nil
	}

	version := info.Version

	return &GameIdent{
		Platform: PlatformN64,
		TitleID:  info.GameCode,
		Title:    info.Title,
		Regions:  []Region{decodeN64Region(info.RegionCode)},
		Version:  &version,
		Extra: map[string]string{
			"byte_order":    string(info.ByteOrder),
			"category_code": string(info.CategoryCode),
		},
	}
}

// identifyGB extracts game identification from a GB/GBC ROM file.
// Returns nil if identification fails (non-fatal).
func identifyGB(r io.ReaderAt, size int64) *GameIdent {
	info, err := gb.ParseGB(r, size)
	if err != nil {
		return nil
	}

	version := info.Version

	// Determine platform based on CGB flag
	var platform Platform
	if info.Platform == gb.GBPlatformGBC {
		platform = PlatformGBC
	} else {
		platform = PlatformGB
	}

	// Determine region from destination code
	var region Region
	if info.DestinationCode == 0x00 {
		region = RegionJP
	} else {
		region = RegionWorld // Non-Japanese = worldwide
	}

	extra := map[string]string{
		"licensee": info.LicenseeCode,
	}
	if info.ManufacturerCode != "" {
		extra["manufacturer"] = info.ManufacturerCode
	}
	if info.SGBFlag == gb.SGBFlagSupported {
		extra["sgb_support"] = "true"
	}
	extra["cartridge_type"] = fmt.Sprintf("%02X", info.CartridgeType)

	return &GameIdent{
		Platform:  platform,
		Title:     info.Title,
		Regions:   []Region{region},
		MakerCode: info.LicenseeCode,
		Version:   &version,
		Extra:     extra,
	}
}

// identifyMD extracts game identification from a Mega Drive/Genesis ROM file.
// Returns nil if identification fails (non-fatal).
func identifyMD(r io.ReaderAt, size int64) *GameIdent {
	info, err := md.ParseMD(r, size)
	if err != nil {
		return nil
	}

	// Use overseas title if available, otherwise domestic title
	title := info.OverseasTitle
	if title == "" {
		title = info.DomesticTitle
	}

	// Decode regions
	regions := decodeMDRegions(info.Regions)

	extra := map[string]string{
		"system_type": info.SystemType,
	}
	if info.Copyright != "" {
		extra["copyright"] = info.Copyright
	}
	if info.DomesticTitle != "" && info.DomesticTitle != info.OverseasTitle {
		extra["domestic_title"] = info.DomesticTitle
	}
	if info.DeviceSupport != "" {
		extra["device_support"] = info.DeviceSupport
	}
	extra["checksum"] = fmt.Sprintf("%04X", info.Checksum)

	return &GameIdent{
		Platform: PlatformMD,
		TitleID:  info.SerialNumber,
		Title:    title,
		Regions:  regions,
		Extra:    extra,
	}
}

// identifySNES extracts game identification from a SNES ROM file.
// Returns nil if identification fails (non-fatal).
func identifySNES(r io.ReaderAt, size int64) *GameIdent {
	info, err := snes.ParseSNES(r, size)
	if err != nil {
		return nil
	}

	version := info.Version

	extra := map[string]string{
		"map_mode": decodeSNESMapMode(info.MapMode),
	}
	if info.SRAMSize > 0 {
		extra["sram"] = formatROMSize(info.SRAMSize)
	}
	if info.HasCopierHeader {
		extra["copier_header"] = "true"
	}

	return &GameIdent{
		Platform: PlatformSNES,
		Title:    info.Title,
		Regions:  []Region{decodeSNESRegion(info.DestinationCode)},
		Version:  &version,
		Extra:    extra,
	}
}

// identifyNES extracts game identification from an NES ROM file.
// Returns nil if identification fails (non-fatal).
// Note: iNES format doesn't include game title, so identification is limited.
func identifyNES(r io.ReaderAt, size int64) *GameIdent {
	info, err := nes.ParseNES(r, size)
	if err != nil {
		return nil
	}

	// Determine region from TV system
	var region Region
	if info.TVSystem == nes.NESTVSystemPAL {
		region = RegionPAL
	} else {
		region = RegionNTSC
	}

	extra := map[string]string{
		"mapper":  fmt.Sprintf("%d", info.Mapper),
		"prg_rom": formatROMSize(info.PRGROMSize),
		"chr_rom": formatROMSize(info.CHRROMSize),
	}

	if info.HasBattery {
		extra["battery"] = "true"
	}
	if info.IsNES20 {
		extra["nes2.0"] = "true"
	}
	if info.Mirroring == nes.NESMirroringVertical {
		extra["mirroring"] = "vertical"
	} else {
		extra["mirroring"] = "horizontal"
	}

	return &GameIdent{
		Platform: PlatformNES,
		Regions:  []Region{region},
		Extra:    extra,
	}
}

// identifyNDS extracts game identification from a Nintendo DS ROM file.
// Returns nil if identification fails (non-fatal).
func identifyNDS(r io.ReaderAt, size int64) *GameIdent {
	info, err := nds.ParseNDS(r, size)
	if err != nil {
		return nil
	}

	version := info.Version

	// Determine platform based on unit code
	var platform Platform
	switch info.Platform {
	case nds.NDSPlatformDSi:
		platform = PlatformDSi
	default:
		platform = PlatformNDS
	}

	extra := map[string]string{}
	if info.Platform == nds.NDSPlatformDSiDual {
		extra["dsi_enhanced"] = "true"
	}

	return &GameIdent{
		Platform:  platform,
		TitleID:   info.GameCode,
		Title:     info.Title,
		Regions:   []Region{decodeNDSRegion(info.RegionCode)},
		MakerCode: info.MakerCode,
		Version:   &version,
		Extra:     extra,
	}
}

// identifySMD extracts game identification from an SMD (Super Magic Drive) ROM file.
// Returns nil if identification fails (non-fatal).
func identifySMD(r io.ReaderAt, size int64) *GameIdent {
	info, err := md.ParseSMD(r, size)
	if err != nil {
		return nil
	}

	// Use overseas title if available, otherwise domestic title
	title := info.OverseasTitle
	if title == "" {
		title = info.DomesticTitle
	}

	// Decode regions
	regions := decodeMDRegions(info.Regions)

	extra := map[string]string{
		"system_type": info.SystemType,
	}
	if info.Copyright != "" {
		extra["copyright"] = info.Copyright
	}
	if info.DomesticTitle != "" && info.DomesticTitle != info.OverseasTitle {
		extra["domestic_title"] = info.DomesticTitle
	}
	if info.DeviceSupport != "" {
		extra["device_support"] = info.DeviceSupport
	}
	extra["checksum"] = fmt.Sprintf("%04X", info.Checksum)

	return &GameIdent{
		Platform: PlatformMD,
		TitleID:  info.SerialNumber,
		Title:    title,
		Regions:  regions,
		Extra:    extra,
	}
}

// decodeMDRegions converts Mega Drive region codes to a slice of Region.
func decodeMDRegions(codes []byte) []Region {
	var regions []Region
	seen := make(map[Region]bool)

	for _, code := range codes {
		var region Region
		switch code {
		case 'J', '1':
			region = RegionJP
		case 'U', '4', '8':
			region = RegionUS
		case 'E':
			region = RegionEU
		case 'A':
			region = RegionWorld // Asia sometimes means world
		case 'B':
			region = RegionBR
		case 'K':
			region = RegionKR
		default:
			continue
		}

		if !seen[region] {
			seen[region] = true
			regions = append(regions, region)
		}
	}

	if len(regions) == 0 {
		regions = append(regions, RegionUnknown)
	}

	return regions
}

// decodeXboxRegions converts Xbox region flags to a slice of Region.
func decodeXboxRegions(flags uint32) []Region {
	var regions []Region
	if flags&uint32(xbox.XboxRegionNA) != 0 {
		regions = append(regions, RegionNA)
	}
	if flags&uint32(xbox.XboxRegionJapan) != 0 {
		regions = append(regions, RegionJP)
	}
	if flags&uint32(xbox.XboxRegionEUAU) != 0 {
		regions = append(regions, RegionEU, RegionAU)
	}
	if len(regions) == 0 {
		regions = append(regions, RegionUnknown)
	}
	return regions
}

// decodeGBARegion converts a GBA region code byte to a Region.
func decodeGBARegion(code byte) Region {
	switch code {
	case 'J':
		return RegionJP
	case 'E':
		return RegionUS
	case 'P':
		return RegionEU
	case 'F':
		return RegionFR
	case 'S':
		return RegionES
	case 'D':
		return RegionDE
	case 'I':
		return RegionIT
	default:
		return RegionUnknown
	}
}

// formatROMSize formats a ROM size in bytes to a human-readable string.
func formatROMSize(bytes int) string {
	if bytes == 0 {
		return "0"
	}
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%d MiB", bytes/(1024*1024))
	}
	if bytes >= 1024 {
		return fmt.Sprintf("%d KiB", bytes/1024)
	}
	return fmt.Sprintf("%d B", bytes)
}

// decodeSNESMapMode converts a SNES map mode byte to a human-readable string.
func decodeSNESMapMode(mode snes.SNESMapMode) string {
	switch mode {
	case snes.SNESMapModeLoROM:
		return "LoROM"
	case snes.SNESMapModeHiROM:
		return "HiROM"
	case snes.SNESMapModeLoROMSA1:
		return "LoROM+SA-1"
	case snes.SNESMapModeExLoROM:
		return "ExLoROM"
	case snes.SNESMapModeExHiROM:
		return "ExHiROM"
	case snes.SNESMapModeHiROMSPC, snes.SNESMapModeHiROMSPC2:
		return "HiROM+SPC7110"
	default:
		return fmt.Sprintf("0x%02X", mode)
	}
}

// decodeSNESRegion converts a SNES destination code to a Region.
func decodeSNESRegion(code byte) Region {
	switch code {
	case 0x00:
		return RegionJP
	case 0x01:
		return RegionNA
	case 0x02:
		return RegionEU
	case 0x03:
		return RegionSE
	case 0x04:
		return RegionFI
	case 0x05:
		return RegionDK
	case 0x06:
		return RegionFR
	case 0x07:
		return RegionNL
	case 0x08:
		return RegionES
	case 0x09:
		return RegionDE
	case 0x0A:
		return RegionIT
	case 0x0B:
		return RegionCN
	case 0x0D:
		return RegionKR
	case 0x0F:
		return RegionCA
	case 0x10:
		return RegionBR
	case 0x11:
		return RegionAU
	default:
		return RegionUnknown
	}
}

// decodeNDSRegion converts an NDS region code byte to a Region.
// The region is typically the 4th character of the game code.
func decodeNDSRegion(code byte) Region {
	switch code {
	case 'J':
		return RegionJP
	case 'E':
		return RegionUS
	case 'P':
		return RegionEU
	case 'D':
		return RegionDE
	case 'F':
		return RegionFR
	case 'I':
		return RegionIT
	case 'S':
		return RegionES
	case 'K':
		return RegionKR
	case 'C':
		return RegionCN
	case 'A':
		return RegionWorld
	case 'U':
		return RegionAU
	default:
		return RegionUnknown
	}
}

// decodeN64Region converts an N64 destination code byte to a Region.
// Based on the N64brew Wiki ROM Header specification.
func decodeN64Region(code byte) Region {
	switch code {
	case 'A':
		return RegionWorld
	case 'B':
		return RegionBR
	case 'C':
		return RegionCN
	case 'D':
		return RegionDE
	case 'E':
		return RegionUS
	case 'F':
		return RegionFR
	case 'G':
		return RegionNTSC
	case 'H':
		return RegionNL
	case 'I':
		return RegionIT
	case 'J':
		return RegionJP
	case 'K':
		return RegionKR
	case 'L':
		return RegionPAL
	case 'N':
		return RegionCA
	case 'P', 'X', 'Y', 'Z':
		return RegionEU
	case 'S':
		return RegionES
	case 'U':
		return RegionAU
	case 'W':
		return RegionNordic
	default:
		return RegionUnknown
	}
}
