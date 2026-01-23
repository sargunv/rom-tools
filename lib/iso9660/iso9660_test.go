package iso9660

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

// mockReaderAt wraps a byte slice to implement io.ReaderAt
type mockReaderAt struct {
	data []byte
}

func (m *mockReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n = copy(p, m.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// createMinimalISO creates a minimal valid ISO 9660 image for testing.
// It includes the PVD at sector 16 with the CD001 magic.
func createMinimalISO() []byte {
	// Minimum size: 17 sectors (0-15 system area + 16 PVD) + 1 root dir sector
	data := make([]byte, 18*sectorSize2048)

	// Primary Volume Descriptor at sector 16
	pvdOffset := 16 * sectorSize2048
	data[pvdOffset+0] = 0x01          // Type: Primary Volume Descriptor
	copy(data[pvdOffset+1:], "CD001") // Standard Identifier
	data[pvdOffset+6] = 0x01          // Version

	// Root directory record at PVD offset 156
	rootRecordOffset := pvdOffset + pvdRootDirOffset
	data[rootRecordOffset+0] = 34                                                          // Directory record length
	data[rootRecordOffset+1] = 0                                                           // Extended attribute record length
	binary.LittleEndian.PutUint32(data[rootRecordOffset+dirEntryExtentLoc:], 17)           // Extent location (sector 17)
	binary.LittleEndian.PutUint32(data[rootRecordOffset+dirEntryDataLen:], sectorSize2048) // Data length

	// Root directory at sector 17
	rootDirOffset := 17 * sectorSize2048

	// First entry: current directory "."
	data[rootDirOffset+0] = 34 // Length
	binary.LittleEndian.PutUint32(data[rootDirOffset+dirEntryExtentLoc:], 17)
	binary.LittleEndian.PutUint32(data[rootDirOffset+dirEntryDataLen:], sectorSize2048)
	data[rootDirOffset+dirEntryFlags] = flagDirectory
	data[rootDirOffset+dirEntryNameLen] = 1
	data[rootDirOffset+dirEntryName] = 0x00 // "." = 0x00

	// Second entry: parent directory ".."
	data[rootDirOffset+34+0] = 34
	binary.LittleEndian.PutUint32(data[rootDirOffset+34+dirEntryExtentLoc:], 17)
	binary.LittleEndian.PutUint32(data[rootDirOffset+34+dirEntryDataLen:], sectorSize2048)
	data[rootDirOffset+34+dirEntryFlags] = flagDirectory
	data[rootDirOffset+34+dirEntryNameLen] = 1
	data[rootDirOffset+34+dirEntryName] = 0x01 // ".." = 0x01

	return data
}

// createISOWithFile creates an ISO with a test file at the root level.
func createISOWithFile(filename string, content []byte) []byte {
	// Size: system area + PVD + root dir + file data (rounded to sector)
	fileSectors := (len(content) + sectorSize2048 - 1) / sectorSize2048
	if fileSectors == 0 {
		fileSectors = 1
	}
	totalSectors := 18 + fileSectors // 16 system + 1 PVD + 1 root dir + file sectors
	data := make([]byte, totalSectors*sectorSize2048)

	// Primary Volume Descriptor at sector 16
	pvdOffset := 16 * sectorSize2048
	data[pvdOffset+0] = 0x01
	copy(data[pvdOffset+1:], "CD001")
	data[pvdOffset+6] = 0x01

	// Root directory record at PVD offset 156
	rootRecordOffset := pvdOffset + pvdRootDirOffset
	data[rootRecordOffset+0] = 34
	binary.LittleEndian.PutUint32(data[rootRecordOffset+dirEntryExtentLoc:], 17)
	binary.LittleEndian.PutUint32(data[rootRecordOffset+dirEntryDataLen:], sectorSize2048)

	// Root directory at sector 17
	rootDirOffset := 17 * sectorSize2048

	// "." entry
	data[rootDirOffset+0] = 34
	binary.LittleEndian.PutUint32(data[rootDirOffset+dirEntryExtentLoc:], 17)
	binary.LittleEndian.PutUint32(data[rootDirOffset+dirEntryDataLen:], sectorSize2048)
	data[rootDirOffset+dirEntryFlags] = flagDirectory
	data[rootDirOffset+dirEntryNameLen] = 1
	data[rootDirOffset+dirEntryName] = 0x00

	// ".." entry
	offset := 34
	data[rootDirOffset+offset+0] = 34
	binary.LittleEndian.PutUint32(data[rootDirOffset+offset+dirEntryExtentLoc:], 17)
	binary.LittleEndian.PutUint32(data[rootDirOffset+offset+dirEntryDataLen:], sectorSize2048)
	data[rootDirOffset+offset+dirEntryFlags] = flagDirectory
	data[rootDirOffset+offset+dirEntryNameLen] = 1
	data[rootDirOffset+offset+dirEntryName] = 0x01

	// File entry
	offset = 68
	filenameWithVersion := filename + ";1"
	entryLen := 33 + len(filenameWithVersion)
	if entryLen%2 == 1 {
		entryLen++ // Padding to even
	}
	data[rootDirOffset+offset+0] = byte(entryLen)
	binary.LittleEndian.PutUint32(data[rootDirOffset+offset+dirEntryExtentLoc:], 18) // File at sector 18
	binary.LittleEndian.PutUint32(data[rootDirOffset+offset+dirEntryDataLen:], uint32(len(content)))
	data[rootDirOffset+offset+dirEntryFlags] = 0 // Regular file
	data[rootDirOffset+offset+dirEntryNameLen] = byte(len(filenameWithVersion))
	copy(data[rootDirOffset+offset+dirEntryName:], filenameWithVersion)

	// File content at sector 18
	fileOffset := 18 * sectorSize2048
	copy(data[fileOffset:], content)

	return data
}

func TestNewReader_ValidISO(t *testing.T) {
	data := createMinimalISO()
	reader, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}
	if reader == nil {
		t.Fatal("NewReader returned nil reader")
	}
}

func TestNewReader_InvalidMagic(t *testing.T) {
	data := make([]byte, 18*sectorSize2048)
	// No CD001 magic
	_, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err == nil {
		t.Error("expected error for invalid magic, got nil")
	}
}

func TestNewReader_TooSmall(t *testing.T) {
	data := make([]byte, 1000)
	_, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err == nil {
		t.Error("expected error for too-small input, got nil")
	}
}

func TestReader_ReadAt(t *testing.T) {
	data := createMinimalISO()
	// Write some test data in system area
	testData := []byte("SYSTEM AREA TEST DATA")
	copy(data[0:], testData)

	reader, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}

	// Read back the system area
	buf := make([]byte, len(testData))
	n, err := reader.ReadAt(buf, 0)
	if err != nil {
		t.Fatalf("ReadAt failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("ReadAt read %d bytes, want %d", n, len(testData))
	}
	if !bytes.Equal(buf, testData) {
		t.Errorf("ReadAt data = %q, want %q", buf, testData)
	}
}

func TestReader_Size(t *testing.T) {
	data := createMinimalISO()
	reader, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}

	if reader.Size() != int64(len(data)) {
		t.Errorf("Size() = %d, want %d", reader.Size(), len(data))
	}
}

func TestReader_OpenFile(t *testing.T) {
	content := []byte("Hello, ISO 9660!")
	data := createISOWithFile("TEST.TXT", content)

	reader, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}

	fileReader, size, err := reader.OpenFile("TEST.TXT")
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	if size != int64(len(content)) {
		t.Errorf("file size = %d, want %d", size, len(content))
	}

	buf := make([]byte, size)
	n, err := fileReader.ReadAt(buf, 0)
	if err != nil {
		t.Fatalf("file ReadAt failed: %v", err)
	}
	if n != len(content) {
		t.Errorf("read %d bytes, want %d", n, len(content))
	}
	if !bytes.Equal(buf, content) {
		t.Errorf("file content = %q, want %q", buf, content)
	}
}

func TestReader_OpenFile_CaseInsensitive(t *testing.T) {
	content := []byte("Test content")
	data := createISOWithFile("MYFILE.DAT", content)

	reader, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}

	// Try lowercase
	_, _, err = reader.OpenFile("myfile.dat")
	if err != nil {
		t.Errorf("OpenFile with lowercase failed: %v", err)
	}

	// Try mixed case
	_, _, err = reader.OpenFile("MyFile.Dat")
	if err != nil {
		t.Errorf("OpenFile with mixed case failed: %v", err)
	}
}

func TestReader_OpenFile_NotFound(t *testing.T) {
	data := createMinimalISO()
	reader, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}

	_, _, err = reader.OpenFile("NOTEXIST.TXT")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestReader_OpenFile_EmptyPath(t *testing.T) {
	data := createMinimalISO()
	reader, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}

	_, _, err = reader.OpenFile("")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

func TestNewReader_RawMODE1(t *testing.T) {
	// Create a raw MODE1/2352 ISO
	numSectors := 18
	data := make([]byte, numSectors*sectorSize2352)

	// For MODE1/2352, data starts at offset 16 within each sector
	pvdSector := 16
	pvdPhysicalOffset := pvdSector*sectorSize2352 + mode1SectorHeader

	data[pvdPhysicalOffset+0] = 0x01
	copy(data[pvdPhysicalOffset+1:], "CD001")
	data[pvdPhysicalOffset+6] = 0x01

	// Root directory record
	rootRecordOffset := pvdPhysicalOffset + pvdRootDirOffset
	data[rootRecordOffset+0] = 34
	binary.LittleEndian.PutUint32(data[rootRecordOffset+dirEntryExtentLoc:], 17)
	binary.LittleEndian.PutUint32(data[rootRecordOffset+dirEntryDataLen:], sectorSize2048)

	reader, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader failed for MODE1/2352: %v", err)
	}

	// Size should be translated to logical 2048-byte sectors
	expectedSize := int64(numSectors * sectorSize2048)
	if reader.Size() != expectedSize {
		t.Errorf("Size() = %d, want %d", reader.Size(), expectedSize)
	}
}

func TestNewReader_RawMODE2(t *testing.T) {
	// Create a raw MODE2/2352 ISO
	numSectors := 18
	data := make([]byte, numSectors*sectorSize2352)

	// For MODE2/2352, data starts at offset 24 within each sector
	pvdSector := 16
	pvdPhysicalOffset := pvdSector*sectorSize2352 + mode2SectorHeader

	data[pvdPhysicalOffset+0] = 0x01
	copy(data[pvdPhysicalOffset+1:], "CD001")
	data[pvdPhysicalOffset+6] = 0x01

	// Root directory record
	rootRecordOffset := pvdPhysicalOffset + pvdRootDirOffset
	data[rootRecordOffset+0] = 34
	binary.LittleEndian.PutUint32(data[rootRecordOffset+dirEntryExtentLoc:], 17)
	binary.LittleEndian.PutUint32(data[rootRecordOffset+dirEntryDataLen:], sectorSize2048)

	reader, err := NewReader(&mockReaderAt{data}, int64(len(data)))
	if err != nil {
		t.Fatalf("NewReader failed for MODE2/2352: %v", err)
	}

	expectedSize := int64(numSectors * sectorSize2048)
	if reader.Size() != expectedSize {
		t.Errorf("Size() = %d, want %d", reader.Size(), expectedSize)
	}
}
