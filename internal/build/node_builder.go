package build

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
)

type (
	// NodeEnv represents the detected Node.js build environment.
	// PURE: Data structure with no behavior.
	NodeEnv struct {
		HasPackageJSON bool
		HasTypeScript  bool
	}

	// NodeBuildSpec represents the pure specification for a Node.js build.
	// PURE: No side effects, deterministic output from inputs.
	NodeBuildSpec struct {
		OutputPath     string
		SourceDir      string
		HasPackageJSON bool
		HasTypeScript  bool
		Env            []string
		InstallCmd     []string // npm install command
		BuildCmd       []string // npm run build command (for TypeScript)
	}
)

// GenerateNodeBuildSpec creates a build specification from config.
// PURE: Calculation - same inputs always produce same outputs.
func GenerateNodeBuildSpec(cfg Config, hasPackageJSON, hasTypeScript bool) NodeBuildSpec {
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

// ExecuteNodeBuildSpec executes a build specification using functional composition.
// ACTION: Performs I/O operations (file system, process execution).
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
		zipWriter := zip.NewWriter(zipFile)
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

// DetectNodeEnv detects the Node.js build environment capabilities.
// ACTION: Performs I/O operations (os.Stat).
func DetectNodeEnv(sourceDir string) E.Either[error, NodeEnv] {
	// I/O: Check for package.json
	packageJSONPath := filepath.Join(sourceDir, "package.json")
	_, err1 := os.Stat(packageJSONPath)
	hasPackageJSON := err1 == nil

	// I/O: Check for tsconfig.json (TypeScript)
	tsconfigPath := filepath.Join(sourceDir, "tsconfig.json")
	_, err2 := os.Stat(tsconfigPath)
	hasTypeScript := err2 == nil

	return E.Right[error](NodeEnv{
		HasPackageJSON: hasPackageJSON,
		HasTypeScript:  hasTypeScript,
	})
}

// NodeBuild composes pure specification generation with impure execution.
// COMPOSITION: Pure core + Imperative shell using monadic composition.
func NodeBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
	// ACTION: Detect environment
	// PURE: Generate specification
	// ACTION: Execute build
	return E.Chain(func(env NodeEnv) E.Either[error, Artifact] {
		spec := GenerateNodeBuildSpec(cfg, env.HasPackageJSON, env.HasTypeScript)
		return ExecuteNodeBuildSpec(ctx, spec)
	})(DetectNodeEnv(cfg.SourceDir))
}
