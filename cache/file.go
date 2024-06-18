package cache

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

func DataDir() string {
	var path string
	if dir := os.Getenv(dataDir); dir != "" {
		path = filepath.Join(dir, "overleash")
	} else if a := os.Getenv(xdgDataHome); a != "" {
		path = filepath.Join(a, "overleash")
	} else if b := os.Getenv(localAppData); runtime.GOOS == "windows" && b != "" {
		path = filepath.Join(b, "Overleash")
	} else {
		c, _ := os.UserHomeDir()
		path = filepath.Join(c, ".local", "share", "overleash")
	}

	return path
}

func ReadFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func WriteFile(filename string, data []byte) (writeErr error) {
	if writeErr = os.MkdirAll(filepath.Dir(filename), 0771); writeErr != nil {
		return
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
	return
}
