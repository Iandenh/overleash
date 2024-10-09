package storage

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
)

const (
	localAppData = "LocalAppData"
	xdgDataHome  = "XDG_DATA_HOME"
	dataDir      = "DATA_DIR"
)

type FileStore struct{}

func NewFileStore() *FileStore {
	return new(FileStore)
}

func (f *FileStore) dataDir() string {
	if dir := os.Getenv(dataDir); dir != "" {
		return filepath.Join(dir, "overleash")
	}
	if a := os.Getenv(xdgDataHome); a != "" {
		return filepath.Join(a, "overleash")
	}
	if b := os.Getenv(localAppData); runtime.GOOS == "windows" && b != "" {
		return filepath.Join(b, "Overleash")
	}

	c, _ := os.UserHomeDir()

	return filepath.Join(c, ".local", "share", "overleash")
}

func (f *FileStore) Read(filename string) ([]byte, error) {
	file, err := os.Open(filepath.Join(f.dataDir(), filename))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *FileStore) Write(filename string, data []byte) (writeErr error) {
	filename = filepath.Join(f.dataDir(), filename)

	if writeErr = os.MkdirAll(filepath.Dir(filename), 0771); writeErr != nil {
		return writeErr
	}

	var file *os.File
	if file, writeErr = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600); writeErr != nil {
		return
	}
	defer func() {
		if err := file.Close(); writeErr == nil && err != nil {
			writeErr = err
		}
	}()
	_, writeErr = file.Write(data)
	return writeErr
}
