package iso9660

import "io"

// CD sector formats
const (
	sectorSize2048 = 2048 // Standard ISO9660 sector (cooked)
	sectorSize2352 = 2352 // Raw CD sector (MODE1/MODE2)

	// For MODE1/2352, user data starts at offset 16 within each sector:
	// 12 bytes sync + 4 bytes header = 16 bytes before data
	mode1SectorHeader = 16

	// For MODE2/2352, user data starts at offset 24 within each sector:
	// 12 bytes sync + 4 bytes header + 8 bytes subheader = 24 bytes before data
	mode2SectorHeader = 24
)

// sectorFormat describes the physical layout of a CD image.
type sectorFormat struct {
	sectorSize int64  // bytes per physical sector
	dataOffset int64  // offset to user data within sector
	pvdOffset  int64  // physical offset to ISO9660 PVD magic
	name       string // format name for debugging
}

// Probe locations for ISO9660 PVD (logical sector 16).
// We probe multiple sector formats to detect cooked ISOs and raw BIN images.
var sectorFormats = []sectorFormat{
	// Cooked ISO (2048 bytes/sector, no headers)
	{sectorSize2048, 0, 16 * sectorSize2048, "MODE1/2048"},
	// Raw MODE1 (2352 bytes/sector, 16-byte header) - used by Saturn, some PS1
	{sectorSize2352, mode1SectorHeader, 16*sectorSize2352 + mode1SectorHeader, "MODE1/2352"},
	// Raw MODE2 (2352 bytes/sector, 24-byte header) - used by PS1/PS2
	{sectorSize2352, mode2SectorHeader, 16*sectorSize2352 + mode2SectorHeader, "MODE2/2352"},
}

// sectorReader wraps an io.ReaderAt to translate logical sector reads
// (2048 bytes/sector as expected by ISO9660) to physical sector reads
// in a raw BIN file (which may use 2352 bytes/sector).
type sectorReader struct {
	r              io.ReaderAt
	physicalSector int64 // bytes per physical sector (2352 or 2048)
	dataOffset     int64 // offset to data within sector (16/24 for raw, 0 for cooked)
	size           int64 // logical size (in 2048-byte terms)
}

// newSectorReader creates a sector-translating reader.
func newSectorReader(r io.ReaderAt, format sectorFormat, physicalSize int64) *sectorReader {
	// Calculate logical size
	numSectors := physicalSize / format.sectorSize
	logicalSize := numSectors * sectorSize2048

	return &sectorReader{
		r:              r,
		physicalSector: format.sectorSize,
		dataOffset:     format.dataOffset,
		size:           logicalSize,
	}
}

// ReadAt implements io.ReaderAt, translating logical offsets to physical.
func (s *sectorReader) ReadAt(p []byte, off int64) (int, error) {
	if off >= s.size {
		return 0, io.EOF
	}

	n := 0
	for n < len(p) && off+int64(n) < s.size {
		logicalOffset := off + int64(n)

		// Which logical sector?
		logicalSector := logicalOffset / sectorSize2048
		offsetInSector := logicalOffset % sectorSize2048

		// Calculate physical offset
		physicalOffset := logicalSector*s.physicalSector + s.dataOffset + offsetInSector

		// How many bytes can we read from this sector?
		bytesInSector := int64(sectorSize2048) - offsetInSector
		bytesToRead := min(int64(len(p)-n), bytesInSector, s.size-logicalOffset)

		// Read from physical location
		bytesRead, err := s.r.ReadAt(p[n:n+int(bytesToRead)], physicalOffset)
		n += bytesRead
		if err != nil {
			return n, err
		}
	}

	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// Size returns the logical size (as if it were a standard 2048-byte sector ISO).
func (s *sectorReader) Size() int64 {
	return s.size
}
