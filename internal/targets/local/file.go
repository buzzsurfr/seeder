package local

import (
	"io"
	"os"
	"path/filepath"
)

type File struct {
	Path string
	Name string
	w    io.WriteCloser
}

func NewFile(path, name string) *File {
	// Ensure path exists
	info, statErr := os.Stat(path)
	if statErr != nil || !info.IsDir() {
		os.MkdirAll(path, 0755)
	}

	// Create File
	f, createErr := os.Create(filepath.Join(path, name))
	if createErr != nil {
		return &File{}
	}

	return &File{
		Path: path,
		Name: name,
		w: f
	}
}
