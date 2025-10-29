package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRootCmd tests root command creation
func TestNewRootCmd(t *testing.T) {
	t.Run("creates root command", func(t *testing.T) {
		cmd := NewRootCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "forge", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("has subcommands", func(t *testing.T) {
		cmd := NewRootCmd()
		assert.True(t, cmd.HasAvailableSubCommands())

		subcommands := cmd.Commands()
		commandNames := make([]string, len(subcommands))
		for i, sub := range subcommands {
			commandNames[i] = sub.Name()
		}

		assert.Contains(t, commandNames, "new")
		assert.Contains(t, commandNames, "deploy")
		assert.Contains(t, commandNames, "destroy")
		assert.Contains(t, commandNames, "version")
	})
}

// TestNewNewCmd tests new command creation
func TestNewNewCmd(t *testing.T) {
	t.Run("creates new command", func(t *testing.T) {
		cmd := NewNewCmd()
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Use, "new")
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("has required flags", func(t *testing.T) {
		cmd := NewNewCmd()

		// Check for stack name flag
		stackFlag := cmd.Flags().Lookup("stack")
		assert.NotNil(t, stackFlag, "Should have --stack flag")

		// Check for runtime flag
		runtimeFlag := cmd.Flags().Lookup("runtime")
		assert.NotNil(t, runtimeFlag, "Should have --runtime flag")
	})

	t.Run("creates new project with args", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := NewNewCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"test-project"})

		// Change to temp directory
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tmpDir)

		// Execute
		err := cmd.Execute()
		require.NoError(t, err)

		// Verify project directory was created
		projectDir := filepath.Join(tmpDir, "test-project")
		assert.DirExists(t, projectDir)

		// Verify forge.hcl was created
		forgeHCL := filepath.Join(projectDir, "forge.hcl")
		assert.FileExists(t, forgeHCL)
	})

	t.Run("creates new stack", func(t *testing.T) {
		tmpDir := t.TempDir()

		// First create a project
		projectDir := filepath.Join(tmpDir, "my-project")
		err := os.MkdirAll(projectDir, 0755)
		require.NoError(t, err)

		// Create forge.hcl
		forgeHCL := `project {
  name   = "my-project"
  region = "us-east-1"
}
`
		err = os.WriteFile(filepath.Join(projectDir, "forge.hcl"), []byte(forgeHCL), 0644)
		require.NoError(t, err)

		// Change to project directory
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(projectDir)

		cmd := NewNewCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		// Set flags
		cmd.Flags().Set("stack", "api")
		cmd.Flags().Set("runtime", "go1.x")

		// Execute
		err = cmd.Execute()
		require.NoError(t, err)

		// Verify stack directory was created
		stackDir := filepath.Join(projectDir, "api")
		assert.DirExists(t, stackDir)

		// Verify stack files
		assert.FileExists(t, filepath.Join(stackDir, "main.go"))
		assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
	})
}

// TestNewDeployCmd tests deploy command creation
func TestNewDeployCmd(t *testing.T) {
	t.Run("creates deploy command", func(t *testing.T) {
		cmd := NewDeployCmd()
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Use, "deploy")
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("has auto-approve flag", func(t *testing.T) {
		cmd := NewDeployCmd()

		autoApproveFlag := cmd.Flags().Lookup("auto-approve")
		assert.NotNil(t, autoApproveFlag, "Should have --auto-approve flag")
	})
}

// TestNewDestroyCmd tests destroy command creation
func TestNewDestroyCmd(t *testing.T) {
	t.Run("creates destroy command", func(t *testing.T) {
		cmd := NewDestroyCmd()
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Use, "destroy")
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("has auto-approve flag", func(t *testing.T) {
		cmd := NewDestroyCmd()

		autoApproveFlag := cmd.Flags().Lookup("auto-approve")
		assert.NotNil(t, autoApproveFlag, "Should have --auto-approve flag")
	})
}

// TestNewVersionCmd tests version command creation
func TestNewVersionCmd(t *testing.T) {
	t.Run("creates version command", func(t *testing.T) {
		cmd := NewVersionCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "version", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("executes without error", func(t *testing.T) {
		cmd := NewVersionCmd()

		output := &bytes.Buffer{}
		cmd.SetOut(output)
		cmd.SetErr(output)

		err := cmd.Execute()
		require.NoError(t, err)
	})
}

// TestExecute tests the Execute function
func TestExecute(t *testing.T) {
	t.Run("Execute creates and runs command", func(t *testing.T) {
		// This is a smoke test - Execute() should not panic
		// We can't easily test the full execution without mocking os.Exit
		cmd := NewRootCmd()
		assert.NotNil(t, cmd)
	})
}

// TestCreateProject tests project creation logic
func TestCreateProject(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	err := createProject("test-proj", "us-west-2")
	require.NoError(t, err)

	// Verify project structure
	projectDir := filepath.Join(tmpDir, "test-proj")
	assert.DirExists(t, projectDir)
	assert.FileExists(t, filepath.Join(projectDir, "forge.hcl"))
	assert.FileExists(t, filepath.Join(projectDir, ".gitignore"))
}

// TestCreateStack tests stack creation logic
func TestCreateStack(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal forge.hcl first
	forgeHCL := `project {
  name   = "test-project"
  region = "us-east-1"
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "forge.hcl"), []byte(forgeHCL), 0644)
	require.NoError(t, err)

	// Change to temp directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	err = createStack("my-stack", "python3.13", "Test stack")
	require.NoError(t, err)

	// Verify stack structure
	stackDir := filepath.Join(tmpDir, "my-stack")
	assert.DirExists(t, stackDir)
	assert.FileExists(t, filepath.Join(stackDir, "handler.py"))
	assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
}

// TestFindTerraformPath tests Terraform binary discovery
func TestFindTerraformPath(t *testing.T) {
	t.Run("returns terraform path", func(t *testing.T) {
		path := findTerraformPath()

		// Should return the terraform binary name
		assert.NotEmpty(t, path)
		assert.Equal(t, "terraform", path)
	})
}

// TestZipFile tests zip file creation
func TestZipFile(t *testing.T) {
	t.Run("creates zip file from source", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create source file
		srcFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(srcFile, []byte("test content"), 0644)
		require.NoError(t, err)

		// Create zip
		zipPath := filepath.Join(tmpDir, "test.zip")
		err = zipFile(srcFile, zipPath)
		require.NoError(t, err)

		// Verify zip was created
		assert.FileExists(t, zipPath)

		// Verify zip has content
		stat, err := os.Stat(zipPath)
		require.NoError(t, err)
		assert.Greater(t, stat.Size(), int64(0))
	})

	t.Run("fails with nonexistent source", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := zipFile("/nonexistent/file.txt", filepath.Join(tmpDir, "test.zip"))
		assert.Error(t, err)
	})
}
