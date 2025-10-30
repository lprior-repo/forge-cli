package discovery

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateStubZip(t *testing.T) {
	t.Run("creates valid empty zip file", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "test.zip")

		err := CreateStubZip(outputPath)
		require.NoError(t, err)

		// Verify file exists
		info, err := os.Stat(outputPath)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0), "stub zip should not be empty")

		// Verify it's a valid zip file
		reader, err := zip.OpenReader(outputPath)
		require.NoError(t, err)
		defer reader.Close()

		// Should have no files but be valid
		assert.Equal(t, 0, len(reader.File))
		assert.Equal(t, "Forge stub - will be replaced by actual build", reader.Comment)
	})

	t.Run("creates parent directories if they don't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "nested", "dir", "test.zip")

		err := CreateStubZip(outputPath)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(outputPath)
		require.NoError(t, err)
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "test.zip")

		// Create initial file
		require.NoError(t, os.WriteFile(outputPath, []byte("old content"), 0644))

		// Create stub (should overwrite)
		err := CreateStubZip(outputPath)
		require.NoError(t, err)

		// Verify it's now a valid zip
		reader, err := zip.OpenReader(outputPath)
		require.NoError(t, err)
		defer reader.Close()
	})

	t.Run("fails when directory creation fails", func(t *testing.T) {
		// Use a path that cannot be created (parent is a file, not a directory)
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0644))

		// Try to create a zip inside the file (should fail)
		outputPath := filepath.Join(filePath, "test.zip")
		err := CreateStubZip(outputPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})

	t.Run("fails when file creation fails on read-only parent", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping test when running as root (permissions don't apply)")
		}

		tmpDir := t.TempDir()
		parentDir := filepath.Join(tmpDir, "readonly")
		require.NoError(t, os.Mkdir(parentDir, 0755))

		// Make directory read-only
		require.NoError(t, os.Chmod(parentDir, 0444))
		defer os.Chmod(parentDir, 0755) // Cleanup

		outputPath := filepath.Join(parentDir, "test.zip")
		err := CreateStubZip(outputPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create stub zip")
	})
}

func TestCreateStubZips(t *testing.T) {
	t.Run("creates stubs for all functions", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, ".forge/build")

		functions := []Function{
			{Name: "api", Runtime: "provided.al2023"},
			{Name: "worker", Runtime: "nodejs20.x"},
			{Name: "processor", Runtime: "python3.13"},
		}

		count, err := CreateStubZips(functions, buildDir)
		require.NoError(t, err)
		assert.Equal(t, 3, count)

		// Verify all stubs exist
		for _, fn := range functions {
			stubPath := filepath.Join(buildDir, fn.Name+".zip")
			info, err := os.Stat(stubPath)
			require.NoError(t, err)
			assert.Greater(t, info.Size(), int64(0))
		}
	})

	t.Run("skips existing files", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, ".forge/build")
		require.NoError(t, os.MkdirAll(buildDir, 0755))

		// Create one existing file
		existingPath := filepath.Join(buildDir, "api.zip")
		require.NoError(t, os.WriteFile(existingPath, []byte("existing"), 0644))

		functions := []Function{
			{Name: "api", Runtime: "provided.al2023"},
			{Name: "worker", Runtime: "nodejs20.x"},
		}

		count, err := CreateStubZips(functions, buildDir)
		require.NoError(t, err)
		assert.Equal(t, 1, count) // Only created stub for worker

		// Verify existing file wasn't overwritten
		content, err := os.ReadFile(existingPath)
		require.NoError(t, err)
		assert.Equal(t, "existing", string(content))

		// Verify new stub was created
		workerPath := filepath.Join(buildDir, "worker.zip")
		_, err = os.Stat(workerPath)
		require.NoError(t, err)
	})

	t.Run("creates build directory if it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, ".forge/build")

		functions := []Function{
			{Name: "api", Runtime: "provided.al2023"},
		}

		count, err := CreateStubZips(functions, buildDir)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify directory was created
		info, err := os.Stat(buildDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("handles empty function list", func(t *testing.T) {
		tmpDir := t.TempDir()
		buildDir := filepath.Join(tmpDir, ".forge/build")

		count, err := CreateStubZips([]Function{}, buildDir)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("fails when build directory creation fails", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping test when running as root (permissions don't apply)")
		}

		tmpDir := t.TempDir()
		// Create a file where we need a directory
		buildDirPath := filepath.Join(tmpDir, "build")
		require.NoError(t, os.WriteFile(buildDirPath, []byte("not a dir"), 0644))

		functions := []Function{
			{Name: "api", Runtime: "provided.al2023"},
		}

		count, err := CreateStubZips(functions, buildDirPath)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to create build directory")
	})
}
