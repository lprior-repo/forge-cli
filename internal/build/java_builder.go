package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
)

// JavaBuildSpec represents the pure specification for a Java build
// PURE: No side effects, deterministic output from inputs
type JavaBuildSpec struct {
	OutputPath string
	SourceDir  string
	TargetDir  string
	Env        []string
	BuildCmd   []string // Maven build command
}

// GenerateJavaBuildSpec creates a build specification from config
// PURE: Calculation - same inputs always produce same outputs
func GenerateJavaBuildSpec(cfg Config) JavaBuildSpec {
	outputPath := cfg.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(cfg.SourceDir, "target", "lambda.jar")
	}

	targetDir := filepath.Join(cfg.SourceDir, "target")

	// Maven clean package command
	buildCmd := []string{"mvn", "clean", "package", "-DskipTests"}

	return JavaBuildSpec{
		OutputPath: outputPath,
		SourceDir:  cfg.SourceDir,
		TargetDir:  targetDir,
		Env:        envSlice(cfg.Env),
		BuildCmd:   buildCmd,
	}
}

// ExecuteJavaBuildSpec executes a Java build specification
// ACTION: Performs I/O operations (file system, process execution)
func ExecuteJavaBuildSpec(ctx context.Context, spec JavaBuildSpec) E.Either[error, Artifact] {
	artifact, err := func() (Artifact, error) {
		// I/O: Check for pom.xml
		pomPath := filepath.Join(spec.SourceDir, "pom.xml")
		if _, err := os.Stat(pomPath); os.IsNotExist(err) {
			return Artifact{}, fmt.Errorf("pom.xml not found in %s", spec.SourceDir)
		}

		// I/O: Clean and package with Maven
		if err := executeCommand(ctx, spec.BuildCmd, spec.Env, spec.SourceDir); err != nil {
			return Artifact{}, err
		}

		// I/O: Find the jar file in target directory
		jarPath, err := findJar(spec.TargetDir)
		if err != nil {
			return Artifact{}, fmt.Errorf("failed to find jar: %w", err)
		}

		// I/O: Copy to output path if different
		if jarPath != spec.OutputPath {
			if err := os.MkdirAll(filepath.Dir(spec.OutputPath), 0755); err != nil {
				return Artifact{}, fmt.Errorf("failed to create output directory: %w", err)
			}

			input, err := os.ReadFile(jarPath)
			if err != nil {
				return Artifact{}, fmt.Errorf("failed to read jar: %w", err)
			}

			//nolint:gosec // G306: JAR file permissions are standard
			if err := os.WriteFile(spec.OutputPath, input, 0644); err != nil {
				return Artifact{}, fmt.Errorf("failed to write jar: %w", err)
			}
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

// JavaBuild composes pure specification generation with impure execution
// COMPOSITION: Pure core + Imperative shell
func JavaBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
	// PURE: Generate build specification
	spec := GenerateJavaBuildSpec(cfg)

	// ACTION: Execute build
	return ExecuteJavaBuildSpec(ctx, spec)
}

// findJar finds the first JAR file in the target directory (excluding sources and javadoc jars)
// ACTION: Performs I/O (directory reading)
func findJar(targetDir string) (string, error) {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to read target directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip sources, javadoc, and original jars
		// Check if filename ends with .jar and doesn't have excluded suffixes
		if filepath.Ext(name) == ".jar" &&
			!hasSuffixBeforeExtension(name, "-sources.jar") &&
			!hasSuffixBeforeExtension(name, "-javadoc.jar") &&
			!hasSuffixBeforeExtension(name, "-original.jar") {
			return filepath.Join(targetDir, name), nil
		}
	}

	return "", fmt.Errorf("no jar file found in %s", targetDir)
}

// hasSuffixBeforeExtension checks if a filename has a specific suffix (e.g., "myapp-sources.jar")
// PURE: Calculation - deterministic string matching
func hasSuffixBeforeExtension(filename, suffix string) bool {
	// Use standard library - safe and tested
	return len(filename) >= len(suffix) && filename[len(filename)-len(suffix):] == suffix
}
