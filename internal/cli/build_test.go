package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuildCmd(t *testing.T) {
	t.Run("creates build command", func(t *testing.T) {
		cmd := NewBuildCmd()

		assert.NotNil(t, cmd)
		assert.Equal(t, "build", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("has stub-only flag", func(t *testing.T) {
		cmd := NewBuildCmd()

		flag := cmd.Flags().Lookup("stub-only")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})
}

func TestRunBuild(t *testing.T) {
	t.Run("returns error when no src/functions directory", func(t *testing.T) {
		// Change to temp directory with no functions
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		err := runBuild(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan functions")
	})

	t.Run("succeeds with stub-only when functions exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create src/functions structure with a Go function
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "main.go"), []byte("package main"), 0644))

		os.Chdir(tmpDir)

		err := runBuild(true)
		assert.NoError(t, err)

		// Verify stub was created
		stubPath := filepath.Join(tmpDir, ".forge", "build", "api.zip")
		assert.FileExists(t, stubPath)
	})

	t.Run("returns nil when no functions found", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create functions dir with unsupported files (will be ignored)
		functionsDir := filepath.Join(tmpDir, "src", "functions")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))

		os.Chdir(tmpDir)

		// Should succeed with no functions message
		err := runBuild(false)
		assert.NoError(t, err)
	})
}
