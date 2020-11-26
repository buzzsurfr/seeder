package local

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// File is a local file seed
type File struct {
	Path      string
	Name      string
	file      *os.File
	w         io.WriteCloser
	isWritten bool
}

// NewFile creates a new local file
func NewFile(path, name string) *File {
	lf := File{
		Path: path,
		Name: name,
	}
	lf.initialize()

	return &lf
}

// Write is a wrapper for an io.Writer
func (f *File) Write(p []byte) (int, error) {
	if f.isWritten {
		f.initialize()
	}
	n, err := f.w.Write(p)

	// Trap io.EOF and reset reader (so that the reader is always ready)
	if err == io.EOF {
		f.isWritten = true
	}
	return n, err
}

// Close is a wrapper for an io.Closer
func (f *File) Close() error {
	f.isWritten = true
	return f.w.Close()
}

func (f *File) initialize() {
	// Ensure path exists
	info, statErr := os.Stat(f.Path)
	if statErr != nil || !info.IsDir() {
		os.MkdirAll(f.Path, 0755)
	}

	// Create File
	lf, openErr := os.OpenFile(filepath.Join(f.Path, f.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if openErr != nil {
		fmt.Println("Error creating or opening file")
	}

	f.w = lf
	f.isWritten = false
}
