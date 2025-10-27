package build

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	E "github.com/IBM/fp-go/either"
)

// PythonBuild is a pure function that builds Python Lambda functions
func PythonBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
	artifact, err := func() (Artifact, error) {
		outputPath := cfg.OutputPath
		if outputPath == "" {
			outputPath = filepath.Join(cfg.SourceDir, "lambda.zip")
		}

		// Create temporary directory for dependencies
		tempDir, err := os.MkdirTemp("", "forge-python-*")
		if err != nil {
			return Artifact{}, fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(tempDir)

		// Check for requirements.txt
		requirementsPath := filepath.Join(cfg.SourceDir, "requirements.txt")
		if _, err := os.Stat(requirementsPath); err == nil {
			// Install dependencies
			cmd := exec.CommandContext(ctx, "pip", "install",
				"-r", requirementsPath,
				"-t", tempDir,
				"--upgrade",
			)
			cmd.Env = append(os.Environ(), envSlice(cfg.Env)...)
			cmd.Dir = cfg.SourceDir

			output, err := cmd.CombinedOutput()
			if err != nil {
				return Artifact{}, fmt.Errorf("pip install failed: %w\nOutput: %s", err, string(output))
			}
		}

		// Create zip file
		zipFile, err := os.Create(outputPath)
		if err != nil {
			return Artifact{}, fmt.Errorf("failed to create zip: %w", err)
		}
		defer zipFile.Close()

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		// Add dependencies from temp directory
		if err := addDirToZip(zipWriter, tempDir, ""); err != nil {
			return Artifact{}, fmt.Errorf("failed to add dependencies: %w", err)
		}

		// Add source code
		if err := addDirToZip(zipWriter, cfg.SourceDir, ""); err != nil {
			return Artifact{}, fmt.Errorf("failed to add source code: %w", err)
		}

		// Close zip before calculating checksum
		zipWriter.Close()
		zipFile.Close()

		// Calculate checksum
		checksum, err := calculateChecksum(outputPath)
		if err != nil {
			return Artifact{}, fmt.Errorf("failed to calculate checksum: %w", err)
		}

		// Get file size
		size, err := getFileSize(outputPath)
		if err != nil {
			return Artifact{}, fmt.Errorf("failed to get file size: %w", err)
		}

		return Artifact{
			Path:     outputPath,
			Checksum: checksum,
			Size:     size,
		}, nil
	}()

	if err != nil {
		return E.Left[Artifact](err)
	}
	return E.Right[error](artifact)
}

// envSlice converts the env map to a slice
func envSlice(envMap map[string]string) []string {
	var env []string
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

// addDirToZip recursively adds directory contents to zip
func addDirToZip(zipWriter *zip.Writer, baseDir, prefix string) error {
	return filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip certain files
		if shouldSkipFile(info.Name()) {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}

		// Create zip entry
		zipPath := filepath.Join(prefix, relPath)
		zipPath = filepath.ToSlash(zipPath) // Use forward slashes in zip

		w, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}

		// Copy file contents
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		return err
	})
}

// shouldSkipFile determines if a file should be excluded from the zip
func shouldSkipFile(name string) bool {
	skipSuffixes := []string{
		".pyc",
		".pyo",
		".pyd",
		"__pycache__",
		".git",
		".DS_Store",
	}

	for _, suffix := range skipSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}

	return false
}
