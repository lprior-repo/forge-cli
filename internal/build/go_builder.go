package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
)

// GoBuildSpec represents the pure specification for a Go build
// PURE: No side effects, deterministic output from inputs
type GoBuildSpec struct {
	Command    []string
	Env        []string
	WorkDir    string
	OutputPath string
}

// GenerateGoBuildSpec creates a build specification from config
// PURE: Calculation - same inputs always produce same outputs
func GenerateGoBuildSpec(cfg Config) GoBuildSpec {
	outputPath := cfg.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(cfg.SourceDir, "bootstrap")
	}

	// Build command arguments
	command := []string{
		"go", "build",
		"-tags", "lambda.norpc",
		"-ldflags", "-s -w",
		"-o", outputPath,
		"./" + cfg.Handler,
	}

	// Set environment for Lambda (Linux AMD64)
	env := []string{
		"GOOS=linux",
		"GOARCH=amd64",
		"CGO_ENABLED=0",
	}

	// Add any additional env vars from config
	for k, v := range cfg.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return GoBuildSpec{
		Command:    command,
		Env:        env,
		WorkDir:    cfg.SourceDir,
		OutputPath: outputPath,
	}
}

// ExecuteGoBuildSpec executes a build specification
// ACTION: Performs I/O operations (file system, process execution)
func ExecuteGoBuildSpec(ctx context.Context, spec GoBuildSpec) E.Either[error, Artifact] {
	artifact, err := func() (Artifact, error) {
		// I/O: Ensure output directory exists
		//nolint:gosec // G301: Lambda build directory permissions are intentionally permissive
		if err := os.MkdirAll(filepath.Dir(spec.OutputPath), 0754); err != nil {
			return Artifact{}, fmt.Errorf("failed to create output directory: %w", err)
		}

		// I/O: Execute build command
		if err := executeCommand(ctx, spec.Command, spec.Env, spec.WorkDir); err != nil {
			return Artifact{}, err
		}

		// I/O: Calculate checksum
		checksum, err := calculateChecksum(spec.OutputPath)
		if err != nil {
			return Artifact{}, fmt.Errorf("failed to calculate checksum: %w", err)
		}

		// I/O: Get file size
		size, err := getFileSize(spec.OutputPath)
		if err != nil {
			return Artifact{}, fmt.Errorf("failed to get file size: %w", err)
		}

		return Artifact{
			Path:     spec.OutputPath,
			Checksum: checksum,
			Size:     size,
		}, nil
	}()

	if err != nil {
		return E.Left[Artifact](err)
	}
	return E.Right[error](artifact)
}

// GoBuild composes pure specification generation with impure execution
// COMPOSITION: Pure core + Imperative shell
func GoBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
	spec := GenerateGoBuildSpec(cfg)      // PURE: Calculation
	return ExecuteGoBuildSpec(ctx, spec) // ACTION: I/O
}
