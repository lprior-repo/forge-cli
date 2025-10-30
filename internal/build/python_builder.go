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

// PythonBuildSpec represents the pure specification for a Python build
// PURE: No side effects, deterministic output from inputs
type PythonBuildSpec struct {
	OutputPath       string
	SourceDir        string
	RequirementsPath string
	HasRequirements  bool
	Env              []string
	DependencyCmd    []string // Command to install dependencies
	UsesUV           bool     // Whether uv is available
}

// GeneratePythonBuildSpec creates a build specification from config
// PURE: Calculation - same inputs always produce same outputs
func GeneratePythonBuildSpec(cfg Config, hasUV bool, hasRequirements bool) PythonBuildSpec {
	outputPath := cfg.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(cfg.SourceDir, "lambda.zip")
	}

	requirementsPath := filepath.Join(cfg.SourceDir, "requirements.txt")

	var depCmd []string
	if hasRequirements {
		if hasUV {
			// Use uv pip install (much faster)
			depCmd = []string{
				"uv", "pip", "install",
				"-r", requirementsPath,
				"--target", "{tempDir}", // Placeholder for temp dir
				"--python-platform", "linux",
				"--python-version", "3.11",
			}
		} else {
			// Fallback to pip
			depCmd = []string{
				"pip", "install",
				"-r", requirementsPath,
				"-t", "{tempDir}", // Placeholder for temp dir
				"--upgrade",
			}
		}
	}

	return PythonBuildSpec{
		OutputPath:       outputPath,
		SourceDir:        cfg.SourceDir,
		RequirementsPath: requirementsPath,
		HasRequirements:  hasRequirements,
		Env:              envSlice(cfg.Env),
		DependencyCmd:    depCmd,
		UsesUV:           hasUV,
	}
}

// ExecutePythonBuildSpec executes a Python build specification
// ACTION: Performs I/O operations (file system, process execution)
func ExecutePythonBuildSpec(ctx context.Context, spec PythonBuildSpec) E.Either[error, Artifact] {
	artifact, err := func() (Artifact, error) {
		// I/O: Create temporary directory for dependencies
		tempDir, err := os.MkdirTemp("", "forge-python-*")
		if err != nil {
			return Artifact{}, fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer func() {
			_ = os.RemoveAll(tempDir) // Best effort cleanup
		}()

		// I/O: Install dependencies if requirements.txt exists
		if spec.HasRequirements {
			// Replace placeholder with actual temp dir
			cmd := make([]string, len(spec.DependencyCmd))
			for i, arg := range spec.DependencyCmd {
				cmd[i] = strings.ReplaceAll(arg, "{tempDir}", tempDir)
			}

			if err := executeCommand(ctx, cmd, spec.Env, spec.SourceDir); err != nil {
				return Artifact{}, err
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

		// I/O: Add dependencies from temp directory
		if spec.HasRequirements {
			if err := addDirToZip(zipWriter, tempDir, ""); err != nil {
				return Artifact{}, fmt.Errorf("failed to add dependencies: %w", err)
			}
		}

		// I/O: Add source code
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

// PythonBuild composes pure specification generation with impure execution
// COMPOSITION: Pure core + Imperative shell
func PythonBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
	// I/O: Check for uv availability (this is I/O but minimal)
	_, hasUV := exec.LookPath("uv")

	// I/O: Check for requirements.txt
	requirementsPath := filepath.Join(cfg.SourceDir, "requirements.txt")
	_, err := os.Stat(requirementsPath)
	hasRequirements := err == nil

	// PURE: Generate build specification
	spec := GeneratePythonBuildSpec(cfg, hasUV == nil, hasRequirements)

	// ACTION: Execute build
	return ExecutePythonBuildSpec(ctx, spec)
}

// envSlice converts the env map to a slice
// PURE: Calculation
func envSlice(envMap map[string]string) []string {
	var env []string
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

// addDirToZip recursively adds directory contents to zip
// ACTION: Performs I/O (file system reads, zip writes)
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
// PURE: Calculation - deterministic based on filename
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
