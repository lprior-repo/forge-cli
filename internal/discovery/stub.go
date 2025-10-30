package discovery

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

// This allows Terraform to initialize even before actual builds are complete.
func CreateStubZip(outputPath string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create zip file
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create stub zip: %w", err)
	}
	defer func() {
		if err := zipFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close stub zip %s: %v\n", outputPath, err)
		}
	}()

	// Create empty zip archive with a placeholder comment
	zipWriter := zip.NewWriter(zipFile)
	if err := zipWriter.SetComment("Forge stub - will be replaced by actual build"); err != nil {
		return fmt.Errorf("failed to set zip comment: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close stub zip: %w", err)
	}

	return nil
}

// Returns the number of stubs created.
func CreateStubZips(functions []Function, buildDir string) (int, error) {
	if err := os.MkdirAll(buildDir, 0o750); err != nil {
		return 0, fmt.Errorf("failed to create build directory: %w", err)
	}

	count := 0
	for _, fn := range functions {
		outputPath := filepath.Join(buildDir, fn.Name+".zip")

		// Only create stub if file doesn't exist
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			if err := CreateStubZip(outputPath); err != nil {
				return count, fmt.Errorf("failed to create stub for %s: %w", fn.Name, err)
			}
			count++
		}
	}

	return count, nil
}
