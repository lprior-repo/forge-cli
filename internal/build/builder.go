package build

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// Config holds build configuration (immutable)
type Config struct {
	SourceDir  string            // Source code directory
	OutputPath string            // Output file path
	Handler    string            // Handler path/name
	Runtime    string            // Runtime (go1.x, python3.11, etc.)
	Env        map[string]string // Environment variables for build
}

// Artifact represents a built artifact (immutable)
type Artifact struct {
	Path     string
	Checksum string
	Size     int64
}

// calculateChecksum computes SHA256 checksum of a file
func calculateChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// getFileSize returns the size of a file
func getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
