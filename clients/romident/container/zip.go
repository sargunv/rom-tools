// Package container provides handlers for container formats (ZIP, folders).
package container

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ZIPHandler handles ZIP archive files.
type ZIPHandler struct{}

// NewZIPHandler creates a new ZIP handler.
func NewZIPHandler() *ZIPHandler {
	return &ZIPHandler{}
}

// ZIPFileInfo contains information about a file within a ZIP archive.
type ZIPFileInfo struct {
	Name  string
	Size  int64
	CRC32 uint32
}

// ZIPArchive represents an open ZIP archive.
type ZIPArchive struct {
	reader *zip.ReadCloser
	Files  []ZIPFileInfo
}

// Close closes the ZIP archive.
func (z *ZIPArchive) Close() error {
	return z.reader.Close()
}

// OpenFile opens a file within the ZIP archive for reading.
func (z *ZIPArchive) OpenFile(name string) (io.ReadCloser, error) {
	for _, f := range z.reader.File {
		if f.Name == name {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("file not found in ZIP: %s", name)
}

// Open opens a ZIP archive and returns metadata for all files.
func (h *ZIPHandler) Open(path string) (*ZIPArchive, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}

	var files []ZIPFileInfo
	for _, f := range r.File {
		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}

		files = append(files, ZIPFileInfo{
			Name:  f.Name,
			Size:  int64(f.UncompressedSize64),
			CRC32: f.CRC32,
		})
	}

	return &ZIPArchive{
		reader: r,
		Files:  files,
	}, nil
}

// FolderHandler handles directory-based ROMs.
type FolderHandler struct{}

// NewFolderHandler creates a new folder handler.
func NewFolderHandler() *FolderHandler {
	return &FolderHandler{}
}

// ListFiles lists all files in a folder.
func (h *FolderHandler) ListFiles(path string) ([]string, error) {
	var files []string

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(path, p)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list folder: %w", err)
	}

	return files, nil
}
