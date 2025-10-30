package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNewCmd(t *testing.T) {
	t.Run("creates new command", func(t *testing.T) {
		cmd := NewNewCmd()

		assert.NotNil(t, cmd)
		assert.Equal(t, "new [project-name]", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("has runtime flag", func(t *testing.T) {
		cmd := NewNewCmd()

		flag := cmd.Flags().Lookup("runtime")
		assert.NotNil(t, flag)
		assert.Equal(t, "go1.x", flag.DefValue)
	})

	t.Run("has auto-state flag", func(t *testing.T) {
		cmd := NewNewCmd()

		flag := cmd.Flags().Lookup("auto-state")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("has stack flag", func(t *testing.T) {
		cmd := NewNewCmd()

		flag := cmd.Flags().Lookup("stack")
		assert.NotNil(t, flag)
		assert.Equal(t, "", flag.DefValue)
	})

	t.Run("has description flag", func(t *testing.T) {
		cmd := NewNewCmd()

		flag := cmd.Flags().Lookup("description")
		assert.NotNil(t, flag)
		assert.Equal(t, "", flag.DefValue)
	})
}

func TestCreateProject(t *testing.T) {
	t.Run("creates project successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		err := createProject("test-project", "provided.al2023", false)
		require.NoError(t, err)

		// Verify project was created
		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
		assert.FileExists(t, filepath.Join(tmpDir, "test-project", "forge.hcl"))
	})
}

func TestCreateStack(t *testing.T) {
	t.Run("creates stack successfully in project", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		// Create forge.hcl to indicate we're in a project
		require.NoError(t, os.WriteFile("forge.hcl", []byte("service = \"test\""), 0o644))

		err := createStack("api", "provided.al2023", "API Lambda")
		require.NoError(t, err)

		// Verify stack was created
		stackDir := filepath.Join(tmpDir, "api")
		assert.DirExists(t, stackDir)
		assert.FileExists(t, filepath.Join(stackDir, "main.go"))
		assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
	})

	t.Run("creates Python stack", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		// Create forge.hcl to indicate we're in a project
		require.NoError(t, os.WriteFile("forge.hcl", []byte("service = \"test\""), 0o644))

		err := createStack("worker", "python3.13", "Worker Lambda")
		require.NoError(t, err)

		stackDir := filepath.Join(tmpDir, "worker")
		assert.DirExists(t, stackDir)
		assert.FileExists(t, filepath.Join(stackDir, "handler.py"))
		assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
	})

	t.Run("creates Node stack", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		// Create forge.hcl to indicate we're in a project
		require.NoError(t, os.WriteFile("forge.hcl", []byte("service = \"test\""), 0o644))

		err := createStack("frontend", "nodejs22.x", "Frontend Lambda")
		require.NoError(t, err)

		stackDir := filepath.Join(tmpDir, "frontend")
		assert.DirExists(t, stackDir)
		assert.FileExists(t, filepath.Join(stackDir, "index.js"))
		assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
	})

	t.Run("creates Java stack", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		// Create forge.hcl to indicate we're in a project
		require.NoError(t, os.WriteFile("forge.hcl", []byte("service = \"test\""), 0o644))

		err := createStack("service", "java21", "Service Lambda")
		require.NoError(t, err)

		stackDir := filepath.Join(tmpDir, "service")
		assert.DirExists(t, stackDir)
		assert.FileExists(t, filepath.Join(stackDir, "pom.xml"))
		assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
	})
}

// TestProvisionStateBackend is intentionally omitted as it's an integration test
// that requires actual Terraform and AWS resources. The underlying state package
// has 94% coverage with comprehensive unit tests.

func TestNewNewCmdExecution(t *testing.T) {
	t.Run("executes RunE with project name argument", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		cmd := NewNewCmd()
		cmd.SetArgs([]string{"test-project"})

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify project was created
		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
	})

	t.Run("executes RunE with --stack flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		// Create forge.hcl to indicate we're in a project
		require.NoError(t, os.WriteFile("forge.hcl", []byte("service = \"test\""), 0o644))

		cmd := NewNewCmd()
		cmd.SetArgs([]string{"--stack", "api", "--runtime", "go1.x", "--description", "API Lambda"})

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify stack was created
		assert.DirExists(t, filepath.Join(tmpDir, "api"))
	})

	t.Run("returns error when stack flag provided without value", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		cmd := NewNewCmd()
		cmd.SetArgs([]string{}) // No project name, no stack name

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--stack flag is required")
	})
}
