package bin

import "io"

// CD sector formats
const (
	SectorSize2048 = 2048 // Standard ISO9660 sector
	SectorSize2352 = 2352 // Raw CD sector (MODE1/MODE2)

	// For MODE2/2352, user data starts at offset 24 within each sector:
	// 12 bytes sync + 4 bytes header + 8 bytes subheader = 24 bytes before data
	RawSectorHeader = 24
)

// SectorReader wraps an io.ReaderAt to translate logical sector reads
// (2048 bytes/sector as expected by ISO9660) to physical sector reads
// in a raw BIN file (which may use 2352 bytes/sector).
type SectorReader struct {
	r              io.ReaderAt
	physicalSector int64 // bytes per physical sector (2352 or 2048)
	dataOffset     int64 // offset to data within sector (16 for raw, 0 for cooked)
	size           int64 // logical size (in 2048-byte terms)
}

// NewSectorReader creates a sector-translating reader.
// physicalSector: actual sector size in the BIN file (2352 or 2048)
// dataOffset: offset to user data within each sector (16 for MODE2/2352, 0 for 2048)
// physicalSize: total size of the underlying file
func NewSectorReader(r io.ReaderAt, physicalSector, dataOffset, physicalSize int64) *SectorReader {
	// Calculate logical size
	numSectors := physicalSize / physicalSector
	logicalSize := numSectors * SectorSize2048

	return &SectorReader{
		r:              r,
		physicalSector: physicalSector,
		dataOffset:     dataOffset,
		size:           logicalSize,
	}
}

// ReadAt implements io.ReaderAt, translating logical offsets to physical.
func (s *SectorReader) ReadAt(p []byte, off int64) (int, error) {
	if off >= s.size {
		return 0, io.EOF
	}

	n := 0
	for n < len(p) && off+int64(n) < s.size {
		logicalOffset := off + int64(n)

		// Which logical sector?
		logicalSector := logicalOffset / SectorSize2048
		offsetInSector := logicalOffset % SectorSize2048

		// Calculate physical offset
		physicalOffset := logicalSector*s.physicalSector + s.dataOffset + offsetInSector

		// How many bytes can we read from this sector?
		bytesInSector := int64(SectorSize2048) - offsetInSector
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
func (s *SectorReader) Size() int64 {
	return s.size
}
