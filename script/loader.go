package script

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// ZipLoader handles extraction of zip files to temporary directory
type ZipLoader struct {
	tempDir string
}

// LoadZipFromFS loads a zip file from the provided filesystem with progress logging
func LoadZipFromFS(assets fs.FS, zipPath string, logFunc ...func(string)) (*ZipLoader, error) {
	var logger func(string)
	if len(logFunc) > 0 && logFunc[0] != nil {
		logger = logFunc[0]
	} else {
		logger = func(msg string) {} // no-op if not provided
	}

	logger("Reading zip file...")
	zipFile, err := fs.ReadFile(assets, zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read zip file: %w", err)
	}

	logger(fmt.Sprintf("Zip file size: %.2f MB", float64(len(zipFile))/(1024*1024)))
	reader, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	logger(fmt.Sprintf("Found %d files to extract", len(reader.File)))

	// Create temporary directory
	tempDir := filepath.Join(os.TempDir(), "GoFyneInstaller")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	logger(fmt.Sprintf("Extracting to: %s", tempDir))

	// Extract all files
	extractedCount := 0
	for _, file := range reader.File {
		path := filepath.Join(tempDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
		src, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file in zip: %w", err)
		}

		dst, err := os.Create(path)
		if err != nil {
			src.Close()
			return nil, fmt.Errorf("failed to create file: %w", err)
		}

		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			return nil, fmt.Errorf("failed to extract file: %w", err)
		}

		src.Close()
		dst.Close()

		extractedCount++
		if extractedCount%10 == 0 || extractedCount == len(reader.File) {
			logger(fmt.Sprintf("Extracted %d/%d files", extractedCount, len(reader.File)))
		}
	}

	logger("Extraction complete")
	return &ZipLoader{tempDir: tempDir}, nil
}

// GetFS returns a filesystem interface for the extracted contents
func (zl *ZipLoader) GetFS() fs.FS {
	return os.DirFS(zl.tempDir)
}

// GetFileContent retrieves content of a file in the extracted archive
func (zl *ZipLoader) GetFileContent(filename string) ([]byte, error) {
	if zl.tempDir == "" {
		return nil, errors.New("loader not initialized")
	}
	return os.ReadFile(filepath.Join(zl.tempDir, filename))
}

// CleanupTemp removes the temporary directory
func (zl *ZipLoader) CleanupTemp() error {
	if zl.tempDir != "" {
		return os.RemoveAll(zl.tempDir)
	}
	return nil
}
