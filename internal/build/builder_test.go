package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateChecksum(t *testing.T) {
	t.Run("calculates checksum for valid file", func(t *testing.T) {
		// Create temp file with known content
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		checksum, err := calculateChecksum(testFile)
		require.NoError(t, err)
		assert.NotEmpty(t, checksum)
		assert.Len(t, checksum, 64) // SHA256 produces 64 hex characters
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		checksum, err := calculateChecksum("/nonexistent/file.txt")
		assert.Error(t, err)
		assert.Empty(t, checksum)
	})

	t.Run("returns error for directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		checksum, err := calculateChecksum(tmpDir)
		assert.Error(t, err)
		assert.Empty(t, checksum)
	})
}

func TestGetFileSize(t *testing.T) {
	t.Run("returns size for valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := []byte("test content")
		err := os.WriteFile(testFile, content, 0644)
		require.NoError(t, err)

		size, err := getFileSize(testFile)
		require.NoError(t, err)
		assert.Equal(t, int64(len(content)), size)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		size, err := getFileSize("/nonexistent/file.txt")
		assert.Error(t, err)
		assert.Equal(t, int64(0), size)
	})

	t.Run("returns size for directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		size, err := getFileSize(tmpDir)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, size, int64(0))
	})

	t.Run("returns size for empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		emptyFile := filepath.Join(tmpDir, "empty.txt")
		err := os.WriteFile(emptyFile, []byte{}, 0644)
		require.NoError(t, err)

		size, err := getFileSize(emptyFile)
		require.NoError(t, err)
		assert.Equal(t, int64(0), size)
	})
}
