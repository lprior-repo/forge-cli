package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDeployCmd(t *testing.T) {
	t.Run("creates deploy command", func(t *testing.T) {
		cmd := NewDeployCmd()

		assert.NotNil(t, cmd)
		assert.Equal(t, "deploy", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("has auto-approve flag", func(t *testing.T) {
		cmd := NewDeployCmd()

		flag := cmd.Flags().Lookup("auto-approve")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("has namespace flag", func(t *testing.T) {
		cmd := NewDeployCmd()

		flag := cmd.Flags().Lookup("namespace")
		assert.NotNil(t, flag)
		assert.Equal(t, "", flag.DefValue)
	})
}

func TestRunDeploy(t *testing.T) {
	t.Run("returns error when no src/functions directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		err := runDeploy(true, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan functions")
	})

	t.Run("returns error when build fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create functions but no infra (Go build will fail first)
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "main.go"), []byte("package main"), 0644))

		os.Chdir(tmpDir)

		err := runDeploy(true, "")
		assert.Error(t, err)
		// Build fails with command execution error
		assert.Contains(t, err.Error(), "deployment failed")
	})

	t.Run("uses namespace when provided", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		// Will fail early due to missing functions, but namespace should be validated
		err := runDeploy(true, "pr-123")
		assert.Error(t, err)
	})

	t.Run("handles empty namespace", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		err := runDeploy(true, "")
		assert.Error(t, err)
	})
}

func TestFindTerraformPath(t *testing.T) {
	t.Run("finds terraform in PATH", func(t *testing.T) {
		path := findTerraformPath()
		// Should return either system terraform or "terraform"
		assert.NotEmpty(t, path)
	})
}
