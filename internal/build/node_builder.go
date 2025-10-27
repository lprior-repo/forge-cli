package build

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
)

// NodeBuild is a pure function that builds Node.js Lambda functions
func NodeBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
	artifact, err := func() (Artifact, error) {
		outputPath := cfg.OutputPath
		if outputPath == "" {
			outputPath = filepath.Join(cfg.SourceDir, "lambda.zip")
		}

		// Check for package.json
		packageJsonPath := filepath.Join(cfg.SourceDir, "package.json")
		if _, err := os.Stat(packageJsonPath); err == nil {
			// Install dependencies
			cmd := exec.CommandContext(ctx, "npm", "install", "--production")
			cmd.Env = append(os.Environ(), envSlice(cfg.Env)...)
			cmd.Dir = cfg.SourceDir

			output, err := cmd.CombinedOutput()
			if err != nil {
				return Artifact{}, fmt.Errorf("npm install failed: %w\nOutput: %s", err, string(output))
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

		// Add source code and node_modules
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
