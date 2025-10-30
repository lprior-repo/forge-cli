// Package build provides serverless function build functionality with multi-runtime support.
// It follows functional programming principles using Either monads for error handling
// and composable build decorators for cross-cutting concerns like caching and logging.
package build

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Config holds build configuration (immutable).
type (
	// Config struct stores all build parameters.
	Config struct {
		SourceDir  string            // Source code directory
		OutputPath string            // Output file path
		Handler    string            // Handler path/name
		Runtime    string            // Runtime (go1.x, python3.11, etc.)
		Env        map[string]string // Environment variables for build
	}

	// Artifact represents a built artifact with metadata.
	Artifact struct {
		Path     string
		Checksum string
		Size     int64
	}
)

// calculateChecksum computes SHA256 checksum of a file.
func calculateChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		//nolint:errcheck // Defer close errors are not critical after successful read
		_ = f.Close()
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// getFileSize returns the size of a file.
func getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// executeCommand executes a command with given environment and working directory.
// ACTION: Performs I/O (process execution).
func executeCommand(ctx context.Context, command, env []string, workDir string) error {
	if len(command) == 0 {
		return errors.New("empty command")
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
