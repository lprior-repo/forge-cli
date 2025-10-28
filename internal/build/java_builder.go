package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
)

// JavaBuild is a pure function that builds Java Lambda functions using Maven
func JavaBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
	artifact, err := func() (Artifact, error) {
		outputPath := cfg.OutputPath
		if outputPath == "" {
			outputPath = filepath.Join(cfg.SourceDir, "target", "lambda.jar")
		}

		// Check for pom.xml
		pomPath := filepath.Join(cfg.SourceDir, "pom.xml")
		if _, err := os.Stat(pomPath); os.IsNotExist(err) {
			return Artifact{}, fmt.Errorf("pom.xml not found in %s", cfg.SourceDir)
		}

		// Clean and package with Maven
		cmd := exec.CommandContext(ctx, "mvn", "clean", "package", "-DskipTests")
		cmd.Env = append(os.Environ(), envSlice(cfg.Env)...)
		cmd.Dir = cfg.SourceDir

		output, err := cmd.CombinedOutput()
		if err != nil {
			return Artifact{}, fmt.Errorf("mvn package failed: %w\nOutput: %s", err, string(output))
		}

		// Find the jar file in target directory
		targetDir := filepath.Join(cfg.SourceDir, "target")
		jarPath, err := findJar(targetDir)
		if err != nil {
			return Artifact{}, fmt.Errorf("failed to find jar: %w", err)
		}

		// Copy to output path if different
		if jarPath != outputPath {
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return Artifact{}, fmt.Errorf("failed to create output directory: %w", err)
			}

			input, err := os.ReadFile(jarPath)
			if err != nil {
				return Artifact{}, fmt.Errorf("failed to read jar: %w", err)
			}

			if err := os.WriteFile(outputPath, input, 0644); err != nil {
				return Artifact{}, fmt.Errorf("failed to write jar: %w", err)
			}
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

// findJar finds the first JAR file in the target directory (excluding sources and javadoc jars)
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
		if filepath.Ext(name) == ".jar" &&
			!contains(name, "-sources") &&
			!contains(name, "-javadoc") &&
			!contains(name, "-original") {
			return filepath.Join(targetDir, name), nil
		}
	}

	return "", fmt.Errorf("no jar file found in %s", targetDir)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && s[len(s)-len(substr)-len(".jar"):len(s)-len(".jar")] == substr)
}
