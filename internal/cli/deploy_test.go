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
		// Build fails before we get to infra check
		assert.Contains(t, err.Error(), "failed to build")
	})
}
