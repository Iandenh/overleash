package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestFileStore_WriteRead writes data to a file and then reads it back,
// verifying that the contents match.
func TestFileStore_WriteRead(t *testing.T) {
	// Create a temporary directory and set DATA_DIR to point to it.
	tmpDir := t.TempDir()
	if err := os.Setenv(dataDir, tmpDir); err != nil {
		t.Fatalf("Failed to set env %s: %v", dataDir, err)
	}
	defer os.Unsetenv(dataDir)

	fs := NewFileStore()

	filename := "testfile.txt"
	content := []byte("Hello, FileStore!")

	// Write content to file.
	if err := fs.Write(filename, content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read the file back.
	readContent, err := fs.Read(filename)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Expected file content %q, got %q", content, readContent)
	}

	// Verify the file exists at the expected path.
	expectedPath := filepath.Join(tmpDir, "overleash", filename)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %q to exist, but it does not", expectedPath)
	}
}

// TestFileStore_Read_NonExistent verifies that reading a file that doesn't exist returns an error.
func TestFileStore_Read_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Setenv(dataDir, tmpDir); err != nil {
		t.Fatalf("Failed to set env %s: %v", dataDir, err)
	}
	defer os.Unsetenv(dataDir)

	fs := NewFileStore()

	_, err := fs.Read("nonexistent.txt")
	if err == nil {
		t.Error("Expected an error when reading a non-existent file, but got nil")
	}
}

// TestDataDirResolution verifies that the dataDir() method chooses the correct directory based on environment variables.
func TestDataDirResolution(t *testing.T) {
	// Test when DATA_DIR is set.
	tmpDir1 := t.TempDir()
	if err := os.Setenv(dataDir, tmpDir1); err != nil {
		t.Fatalf("Failed to set env %s: %v", dataDir, err)
	}
	defer os.Unsetenv(dataDir)

	fs := NewFileStore()
	dir := fs.dataDir()
	expected := filepath.Join(tmpDir1, "overleash")
	if dir != expected {
		t.Errorf("Expected dataDir %q, got %q", expected, dir)
	}

	// Test when DATA_DIR is not set but XDG_DATA_HOME is.
	os.Unsetenv(dataDir)
	tmpDir2 := t.TempDir()
	if err := os.Setenv(xdgDataHome, tmpDir2); err != nil {
		t.Fatalf("Failed to set env %s: %v", xdgDataHome, err)
	}
	defer os.Unsetenv(xdgDataHome)

	dir = fs.dataDir()
	expected = filepath.Join(tmpDir2, "overleash")
	if dir != expected {
		t.Errorf("Expected dataDir %q, got %q", expected, dir)
	}

	// On Windows, if XDG_DATA_HOME is not set, test using LocalAppData.
	if runtime.GOOS == "windows" {
		os.Unsetenv(xdgDataHome)
		tmpDir3 := t.TempDir()
		if err := os.Setenv(localAppData, tmpDir3); err != nil {
			t.Fatalf("Failed to set env %s: %v", localAppData, err)
		}
		defer os.Unsetenv(localAppData)

		dir = fs.dataDir()
		expected = filepath.Join(tmpDir3, "Overleash")
		if dir != expected {
			t.Errorf("Expected dataDir %q, got %q", expected, dir)
		}
	}

	// Test fallback: when none of DATA_DIR, XDG_DATA_HOME, or (on Windows) LocalAppData is set.
	os.Unsetenv(dataDir)
	os.Unsetenv(xdgDataHome)
	os.Unsetenv(localAppData)

	dir = fs.dataDir()
	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() failed: %v", err)
	}
	expected = filepath.Join(userHome, ".local", "share", "overleash")
	if dir != expected {
		t.Errorf("Expected fallback dataDir %q, got %q", expected, dir)
	}
}
