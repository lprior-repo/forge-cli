package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
)

// GoBuild is a pure function that builds Go Lambda functions
func GoBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
	// For Go Lambda functions, the binary must be named "bootstrap"
	outputPath := cfg.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(cfg.SourceDir, "bootstrap")
	}

	// Build function returns either error or artifact
	artifact, err := func() (Artifact, error) {
		// Ensure output directory exists
		//nolint:gosec // G301: Lambda build directory permissions are intentionally permissive
		if err := os.MkdirAll(filepath.Dir(outputPath), 0754); err != nil {
			return Artifact{}, fmt.Errorf("failed to create output directory: %w", err)
		}

		// Build command
		cmd := exec.CommandContext(ctx, "go", "build",
			"-tags", "lambda.norpc",
			"-ldflags", "-s -w",
			"-o", outputPath,
			"./"+cfg.Handler,
		)

		// Set environment for Lambda (Linux AMD64)
		cmd.Env = append(os.Environ(),
			"GOOS=linux",
			"GOARCH=amd64",
			"CGO_ENABLED=0",
		)

		// Add any additional env vars from config
		for k, v := range cfg.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}

		// Set working directory
		cmd.Dir = cfg.SourceDir

		// Run build
		output, err := cmd.CombinedOutput()
		if err != nil {
			return Artifact{}, fmt.Errorf("go build failed: %w\nOutput: %s", err, string(output))
		}

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
