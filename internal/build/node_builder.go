package build

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
)

// NodeBuildSpec represents the pure specification for a Node.js build
// PURE: No side effects, deterministic output from inputs
type NodeBuildSpec struct {
	OutputPath      string
	SourceDir       string
	HasPackageJSON  bool
	HasTypeScript   bool
	Env             []string
	InstallCmd      []string // npm install command
	BuildCmd        []string // npm run build command (for TypeScript)
}

// GenerateNodeBuildSpec creates a build specification from config
// PURE: Calculation - same inputs always produce same outputs
func GenerateNodeBuildSpec(cfg Config, hasPackageJSON bool, hasTypeScript bool) NodeBuildSpec {
	outputPath := cfg.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(cfg.SourceDir, "lambda.zip")
	}

	var installCmd []string
	var buildCmd []string

	if hasPackageJSON {
		// npm install command
		installCmd = []string{"npm", "install"}

		// TypeScript build command
		if hasTypeScript {
			buildCmd = []string{"npm", "run", "build"}
		}
	}

	return NodeBuildSpec{
		OutputPath:     outputPath,
		SourceDir:      cfg.SourceDir,
		HasPackageJSON: hasPackageJSON,
		HasTypeScript:  hasTypeScript,
		Env:            envSlice(cfg.Env),
		InstallCmd:     installCmd,
		BuildCmd:       buildCmd,
	}
}

// ExecuteNodeBuildSpec executes a Node.js build specification
// ACTION: Performs I/O operations (file system, process execution)
func ExecuteNodeBuildSpec(ctx context.Context, spec NodeBuildSpec) E.Either[error, Artifact] {
	artifact, err := func() (Artifact, error) {
		// I/O: Install dependencies if package.json exists
		if spec.HasPackageJSON {
			if err := executeCommand(ctx, spec.InstallCmd, spec.Env, spec.SourceDir); err != nil {
				return Artifact{}, err
			}

			// I/O: Build TypeScript if tsconfig.json exists
			if spec.HasTypeScript {
				if err := executeCommand(ctx, spec.BuildCmd, spec.Env, spec.SourceDir); err != nil {
					return Artifact{}, err
				}
			}
		}

		// I/O: Create zip file
		zipFile, err := os.Create(spec.OutputPath)
		if err != nil {
			return Artifact{}, fmt.Errorf("failed to create zip: %w", err)
		}
		defer func() {
			_ = zipFile.Close() // Best effort close in defer
		}()

		zipWriter := zip.NewWriter(zipFile)
		defer func() {
			_ = zipWriter.Close() // Best effort close in defer
		}()

		// I/O: Add source code and node_modules
		if err := addDirToZip(zipWriter, spec.SourceDir, ""); err != nil {
			return Artifact{}, fmt.Errorf("failed to add source code: %w", err)
		}

		// I/O: Close zip before calculating checksum
		if err := zipWriter.Close(); err != nil {
			return Artifact{}, fmt.Errorf("failed to close zip writer: %w", err)
		}
		if err := zipFile.Close(); err != nil {
			return Artifact{}, fmt.Errorf("failed to close zip file: %w", err)
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

// NodeBuild composes pure specification generation with impure execution
// COMPOSITION: Pure core + Imperative shell
func NodeBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
	// I/O: Check for package.json
	packageJSONPath := filepath.Join(cfg.SourceDir, "package.json")
	_, err1 := os.Stat(packageJSONPath)
	hasPackageJSON := err1 == nil

	// I/O: Check for tsconfig.json (TypeScript)
	tsconfigPath := filepath.Join(cfg.SourceDir, "tsconfig.json")
	_, err2 := os.Stat(tsconfigPath)
	hasTypeScript := err2 == nil

	// PURE: Generate build specification
	spec := GenerateNodeBuildSpec(cfg, hasPackageJSON, hasTypeScript)

	// ACTION: Execute build
	return ExecuteNodeBuildSpec(ctx, spec)
}
