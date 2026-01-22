package folder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sargunv/rom-tools/internal/util"
)

// FolderContainer implements Container for directory-based ROMs.
type FolderContainer struct {
	path    string
	entries []util.FileEntry
}

// NewFolderContainer creates a new folder container.
func NewFolderContainer(path string) (*FolderContainer, error) {
	var entries []util.FileEntry

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(path, p)
			if err != nil {
				return err
			}
			entries = append(entries, util.FileEntry{
				Name:  rel,
				Size:  info.Size(),
				CRC32: 0, // Folders don't have pre-computed CRC32
			})
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list folder: %w", err)
	}

	return &FolderContainer{
		path:    path,
		entries: entries,
	}, nil
}

// Entries returns all files in the folder.
func (f *FolderContainer) Entries() []util.FileEntry {
	return f.entries
}

// OpenFile opens a file within the folder for sequential reading.
func (f *FolderContainer) OpenFile(name string) (io.ReadCloser, error) {
	fullPath := filepath.Join(f.path, name)
	return os.Open(fullPath)
}

// OpenFileAt opens a file within the folder with random access support.
func (f *FolderContainer) OpenFileAt(name string) (util.RandomAccessReader, error) {
	fullPath := filepath.Join(f.path, name)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	return &fileReaderAt{
		file: file,
		size: info.Size(),
	}, nil
}

// Close releases resources (no-op for folders).
func (f *FolderContainer) Close() error {
	return nil
}

// fileReaderAt wraps *os.File to implement RandomAccessReader.
type fileReaderAt struct {
	file *os.File
	size int64
}

func (f *fileReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	return f.file.ReadAt(p, off)
}

func (f *fileReaderAt) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

func (f *fileReaderAt) Size() int64 {
	return f.size
}

func (f *fileReaderAt) Close() error {
	return f.file.Close()
}
