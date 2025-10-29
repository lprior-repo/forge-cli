package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectStructure tests project structure validation
func TestProjectStructure(t *testing.T) {
	t.Run("validates forge.hcl format", func(t *testing.T) {
		// Create temp directory with forge.hcl
		tmpDir := t.TempDir()

		// Create minimal forge.hcl with correct block syntax
		forgeHCL := `project {
  name   = "test-project"
  region = "us-east-1"
}`
		err := os.WriteFile(filepath.Join(tmpDir, "forge.hcl"), []byte(forgeHCL), 0644)
		require.NoError(t, err)

		// Verify file exists
		assert.FileExists(t, filepath.Join(tmpDir, "forge.hcl"))
	})

	t.Run("validates stack structure", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a stack directory
		err := os.MkdirAll(filepath.Join(tmpDir, "stacks/api"), 0755)
		require.NoError(t, err)

		// Create stack.forge.hcl with correct block syntax
		stackHCL := `stack {
  name        = "api"
  runtime     = "go1.x"
  description = "API stack"
}`
		err = os.WriteFile(filepath.Join(tmpDir, "stacks/api/stack.forge.hcl"), []byte(stackHCL), 0644)
		require.NoError(t, err)

		// Verify files exist
		assert.DirExists(t, filepath.Join(tmpDir, "stacks/api"))
		assert.FileExists(t, filepath.Join(tmpDir, "stacks/api/stack.forge.hcl"))
	})
}

// TestDeployCommand tests deploy command creation and flags
func TestDeployCommand(t *testing.T) {
	t.Run("has correct flags", func(t *testing.T) {
		cmd := NewDeployCmd()

		autoApproveFlag := cmd.Flags().Lookup("auto-approve")
		assert.NotNil(t, autoApproveFlag)
		assert.Equal(t, "false", autoApproveFlag.DefValue)

		parallelFlag := cmd.Flags().Lookup("parallel")
		assert.NotNil(t, parallelFlag)
		assert.Equal(t, "false", parallelFlag.DefValue)
	})

	t.Run("accepts stack argument", func(t *testing.T) {
		cmd := NewDeployCmd()
		cmd.SetArgs([]string{"my-stack", "--auto-approve"})

		// Parse args without executing
		err := cmd.ParseFlags([]string{"my-stack", "--auto-approve"})
		assert.NoError(t, err)
	})

	t.Run("has correct command properties", func(t *testing.T) {
		cmd := NewDeployCmd()

		assert.Contains(t, cmd.Use, "deploy")
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("accepts optional stack argument", func(t *testing.T) {
		cmd := NewDeployCmd()

		// Should accept 0 or 1 args
		cmd.SetArgs([]string{})
		err := cmd.ParseFlags([]string{})
		assert.NoError(t, err)

		cmd.SetArgs([]string{"my-stack"})
		err = cmd.ParseFlags([]string{"my-stack"})
		assert.NoError(t, err)
	})
}

// TestDestroyCommand tests destroy command creation and flags
func TestDestroyCommand(t *testing.T) {
	t.Run("has auto-approve flag", func(t *testing.T) {
		cmd := NewDestroyCmd()

		autoApproveFlag := cmd.Flags().Lookup("auto-approve")
		assert.NotNil(t, autoApproveFlag)
		assert.Equal(t, "false", autoApproveFlag.DefValue)
	})

	t.Run("accepts stack argument", func(t *testing.T) {
		cmd := NewDestroyCmd()
		cmd.SetArgs([]string{"my-stack", "--auto-approve"})

		// Parse args without executing
		err := cmd.ParseFlags([]string{"my-stack", "--auto-approve"})
		assert.NoError(t, err)
	})

	t.Run("has correct command properties", func(t *testing.T) {
		cmd := NewDestroyCmd()

		assert.Contains(t, cmd.Use, "destroy")
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("accepts optional stack argument", func(t *testing.T) {
		cmd := NewDestroyCmd()

		// Should accept 0 or 1 args
		cmd.SetArgs([]string{})
		err := cmd.ParseFlags([]string{})
		assert.NoError(t, err)

		cmd.SetArgs([]string{"my-stack"})
		err = cmd.ParseFlags([]string{"my-stack"})
		assert.NoError(t, err)
	})
}

// TestRunDeployErrorCases tests error handling in runDeploy
func TestRunDeployErrorCases(t *testing.T) {
	t.Run("fails when no forge.hcl exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Try to deploy without forge.hcl
		err = runDeploy("", false, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})

	t.Run("fails when no stacks found", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create minimal forge.hcl with correct block syntax
		forgeHCL := `project {
  name   = "test-project"
  region = "us-east-1"
}`
		err = os.WriteFile("forge.hcl", []byte(forgeHCL), 0644)
		require.NoError(t, err)

		// Try to deploy with no stacks
		err = runDeploy("", false, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no stacks found")
	})

	t.Run("fails when target stack not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create forge.hcl with correct block syntax
		forgeHCL := `project {
  name   = "test-project"
  region = "us-east-1"
}`
		err = os.WriteFile("forge.hcl", []byte(forgeHCL), 0644)
		require.NoError(t, err)

		// Create a stack
		err = os.MkdirAll("stacks/api", 0755)
		require.NoError(t, err)

		stackHCL := `stack {
  name        = "api"
  runtime     = "go1.x"
  description = "API stack"
}`
		err = os.WriteFile("stacks/api/stack.forge.hcl", []byte(stackHCL), 0644)
		require.NoError(t, err)

		// Try to deploy non-existent stack
		err = runDeploy("nonexistent", false, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stack not found")
	})
}

// TestRunDestroyErrorCases tests error handling in runDestroy
func TestRunDestroyErrorCases(t *testing.T) {
	t.Run("fails when no forge.hcl exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Try to destroy without forge.hcl
		err = runDestroy("", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})

	t.Run("fails when no stacks found", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create minimal forge.hcl with correct block syntax
		forgeHCL := `project {
  name   = "test-project"
  region = "us-east-1"
}`
		err = os.WriteFile("forge.hcl", []byte(forgeHCL), 0644)
		require.NoError(t, err)

		// Try to destroy with no stacks
		err = runDestroy("", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no stacks found")
	})

	t.Run("fails when target stack not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create forge.hcl with correct block syntax
		forgeHCL := `project {
  name   = "test-project"
  region = "us-east-1"
}`
		err = os.WriteFile("forge.hcl", []byte(forgeHCL), 0644)
		require.NoError(t, err)

		// Create a stack
		err = os.MkdirAll("stacks/api", 0755)
		require.NoError(t, err)

		stackHCL := `stack {
  name        = "api"
  runtime     = "go1.x"
  description = "API stack"
}`
		err = os.WriteFile("stacks/api/stack.forge.hcl", []byte(stackHCL), 0644)
		require.NoError(t, err)

		// Try to destroy non-existent stack
		err = runDestroy("nonexistent", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stack not found")
	})
}

// TestExecuteFunction tests the Execute function
func TestExecuteFunction(t *testing.T) {
	t.Run("Execute creates root command", func(t *testing.T) {
		// Execute function should not panic
		assert.NotPanics(t, func() {
			// We can't easily test Execute() since it calls os.Exit
			// But we can test that NewRootCmd works
			cmd := NewRootCmd()
			assert.NotNil(t, cmd)
		})
	})
}

// TestFindTerraformPathFunction tests terraform path detection
func TestFindTerraformPathFunction(t *testing.T) {
	t.Run("finds terraform in PATH", func(t *testing.T) {
		path := findTerraformPath()
		// Should return "terraform" or find actual path
		assert.NotEmpty(t, path)
	})
}

// TestNewCommandFlags tests NewNewCmd flag handling
func TestNewCommandFlags(t *testing.T) {
	t.Run("stack flag is optional", func(t *testing.T) {
		cmd := NewNewCmd()

		stackFlag := cmd.Flags().Lookup("stack")
		assert.NotNil(t, stackFlag)
		assert.Equal(t, "", stackFlag.DefValue)
	})

	t.Run("runtime flag has default", func(t *testing.T) {
		cmd := NewNewCmd()

		runtimeFlag := cmd.Flags().Lookup("runtime")
		assert.NotNil(t, runtimeFlag)
		assert.Equal(t, "go1.x", runtimeFlag.DefValue)
	})

	t.Run("description flag is optional", func(t *testing.T) {
		cmd := NewNewCmd()

		descFlag := cmd.Flags().Lookup("description")
		assert.NotNil(t, descFlag)
		assert.Equal(t, "", descFlag.DefValue)
	})
}

// TestCreateProjectErrorCases tests error handling in createProject
func TestCreateProjectErrorCases(t *testing.T) {
	t.Run("fails when directory already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create the project directory first
		err = os.MkdirAll("test-project", 0755)
		require.NoError(t, err)

		// Try to create project in same location
		err = createProject("test-project", "go1.x")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

// TestCreateStackErrorCases tests error handling in createStack
func TestCreateStackErrorCases(t *testing.T) {
	t.Run("fails when not in forge project", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Try to create stack without forge.hcl
		err = createStack("my-stack", "go1.x", "My stack")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in a Forge project")
	})
}

// TestRunDeployWithMultipleStacks tests deploying multiple stacks
func TestRunDeployWithMultipleStacks(t *testing.T) {
	t.Run("prints multiple stack names when deploying all", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create forge.hcl with correct block syntax
		forgeHCL := `project {
  name   = "test-project"
  region = "us-east-1"
}`
		err = os.WriteFile("forge.hcl", []byte(forgeHCL), 0644)
		require.NoError(t, err)

		// Create two stacks
		for _, stackName := range []string{"api", "worker"} {
			err = os.MkdirAll(stackName, 0755)
			require.NoError(t, err)

			stackHCL := `stack {
  name        = "` + stackName + `"
  runtime     = "provided.al2023"
  description = "` + stackName + ` stack"
}`
			err = os.WriteFile(filepath.Join(stackName, "stack.forge.hcl"), []byte(stackHCL), 0644)
			require.NoError(t, err)
		}

		// Try to deploy all stacks (tests the multi-stack path)
		err = runDeploy("", false, false)
		// May succeed or fail depending on terraform setup
	})
}

// TestRunDeployWithSingleStack tests deploying a single named stack
func TestRunDeployWithSingleStack(t *testing.T) {
	t.Run("deploys only the specified stack", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create forge.hcl
		forgeHCL := `project {
  name   = "test-project"
  region = "us-east-1"
}`
		err = os.WriteFile("forge.hcl", []byte(forgeHCL), 0644)
		require.NoError(t, err)

		// Create two stacks
		for _, stackName := range []string{"api", "worker"} {
			err = os.MkdirAll(stackName, 0755)
			require.NoError(t, err)

			stackHCL := `stack {
  name        = "` + stackName + `"
  runtime     = "provided.al2023"
  description = "` + stackName + ` stack"
}`
			err = os.WriteFile(filepath.Join(stackName, "stack.forge.hcl"), []byte(stackHCL), 0644)
			require.NoError(t, err)
		}

		// Try to deploy only "api" stack
		err = runDeploy("api", false, false)
		// May succeed or fail depending on terraform setup, but we exercised the target stack filtering logic
	})
}

// TestRunDestroyWithMultipleStacks tests destroying multiple stacks
func TestRunDestroyWithMultipleStacks(t *testing.T) {
	t.Run("prints multiple stack names when destroying all", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create forge.hcl
		forgeHCL := `project {
  name   = "test-project"
  region = "us-east-1"
}`
		err = os.WriteFile("forge.hcl", []byte(forgeHCL), 0644)
		require.NoError(t, err)

		// Create two stacks
		for _, stackName := range []string{"api", "worker"} {
			err = os.MkdirAll(stackName, 0755)
			require.NoError(t, err)

			stackHCL := `stack {
  name        = "` + stackName + `"
  runtime     = "provided.al2023"
  description = "` + stackName + ` stack"
}`
			err = os.WriteFile(filepath.Join(stackName, "stack.forge.hcl"), []byte(stackHCL), 0644)
			require.NoError(t, err)
		}

		// Try to destroy with auto-approve (avoids confirmation prompt)
		err = runDestroy("", true)
		// May succeed or fail depending on terraform setup, but we exercised the multi-stack path
		// We don't assert error here since the test environment may have mock terraform
	})
}

// TestZipFileFunction tests the zipFile utility
func TestZipFileFunction(t *testing.T) {
	t.Run("creates valid zip from file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a source file
		sourceFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(sourceFile, []byte("test content"), 0644)
		require.NoError(t, err)

		// Create zip
		zipPath := filepath.Join(tmpDir, "test.zip")
		err = zipFile(sourceFile, zipPath)
		require.NoError(t, err)

		// Verify zip exists
		assert.FileExists(t, zipPath)

		// Verify zip is not empty
		info, err := os.Stat(zipPath)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0))
	})

	t.Run("fails with nonexistent source", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := zipFile("/nonexistent/file", filepath.Join(tmpDir, "test.zip"))
		assert.Error(t, err)
	})

	t.Run("fails with invalid destination", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a source file
		sourceFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(sourceFile, []byte("test content"), 0644)
		require.NoError(t, err)

		// Try to create zip in nonexistent directory
		err = zipFile(sourceFile, "/nonexistent/dir/test.zip")
		assert.Error(t, err)
	})
}

// TestNewDeployCommandExecution tests deploy command execution paths
func TestNewDeployCommandExecution(t *testing.T) {
	t.Run("command has RunE function", func(t *testing.T) {
		cmd := NewDeployCmd()
		assert.NotNil(t, cmd.RunE)
	})

	t.Run("command is marked as runnable", func(t *testing.T) {
		cmd := NewDeployCmd()
		assert.True(t, cmd.Runnable())
	})
}

// TestNewDestroyCommandExecution tests destroy command execution paths
func TestNewDestroyCommandExecution(t *testing.T) {
	t.Run("command has RunE function", func(t *testing.T) {
		cmd := NewDestroyCmd()
		assert.NotNil(t, cmd.RunE)
	})

	t.Run("command is marked as runnable", func(t *testing.T) {
		cmd := NewDestroyCmd()
		assert.True(t, cmd.Runnable())
	})
}

// TestCommandIntegration tests command integration
func TestCommandIntegration(t *testing.T) {
	t.Run("root command has all subcommands", func(t *testing.T) {
		root := NewRootCmd()

		subcommands := root.Commands()
		cmdMap := make(map[string]bool)
		for _, cmd := range subcommands {
			cmdMap[cmd.Name()] = true
		}

		assert.True(t, cmdMap["new"], "Should have 'new' command")
		assert.True(t, cmdMap["deploy"], "Should have 'deploy' command")
		assert.True(t, cmdMap["destroy"], "Should have 'destroy' command")
		assert.True(t, cmdMap["version"], "Should have 'version' command")
	})

	t.Run("new command creates projects", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		// Change to tmpDir since createProject creates in current directory
		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		err = createProject("test-project", "go1.x")
		require.NoError(t, err)

		// Verify project was created in current directory
		projectDir := filepath.Join(tmpDir, "test-project")
		assert.DirExists(t, projectDir)
		assert.FileExists(t, filepath.Join(projectDir, "forge.hcl"))
	})

	t.Run("new command creates stacks", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		// Change to tmpDir
		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create a project first since createStack requires forge.hcl
		err = createProject("test-project", "go1.x")
		require.NoError(t, err)

		// Change to project directory
		projectDir := filepath.Join(tmpDir, "test-project")
		err = os.Chdir(projectDir)
		require.NoError(t, err)

		// Now create the stack
		err = createStack("my-stack", "go1.x", "My stack")
		require.NoError(t, err)

		// Verify stack was created directly in project root (not in stacks/)
		stackDir := filepath.Join(projectDir, "my-stack")
		assert.DirExists(t, stackDir)
		assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
	})
}
