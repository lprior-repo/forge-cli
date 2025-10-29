package discovery

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

// CreateStubZip creates an empty zip file as a placeholder
// This allows Terraform to initialize even before actual builds are complete
func CreateStubZip(outputPath string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create zip file
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create stub zip: %w", err)
	}
	defer zipFile.Close()

	// Create empty zip archive with a placeholder comment
	zipWriter := zip.NewWriter(zipFile)
	zipWriter.SetComment("Forge stub - will be replaced by actual build")

	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close stub zip: %w", err)
	}

	return nil
}

// CreateStubZips creates stub zip files for all discovered functions
// Returns the number of stubs created
func CreateStubZips(functions []Function, buildDir string) (int, error) {
	if err := os.MkdirAll(buildDir, 0755); err != nil {
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
