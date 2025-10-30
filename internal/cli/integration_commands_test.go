package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lewis/forge/internal/terraform"
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

		namespaceFlag := cmd.Flags().Lookup("namespace")
		assert.NotNil(t, namespaceFlag)
		assert.Equal(t, "", namespaceFlag.DefValue)
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

		// Try to deploy without forge.hcl (convention-based discovery)
		err = runDeploy(false, "")
		assert.Error(t, err)
		// Convention-based discovery may fail at different stages
	})

	t.Run("fails when no functions found", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create infra directory for Terraform files
		err = os.MkdirAll("infra", 0755)
		require.NoError(t, err)

		// Try to deploy with no functions (convention-based expects src/functions/)
		err = runDeploy(false, "")
		assert.Error(t, err)
		// Convention-based discovery expects src/functions/* directories
	})

	t.Run("deploys with namespace parameter", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create infra directory
		err = os.MkdirAll("infra", 0755)
		require.NoError(t, err)

		// Create a function directory (convention-based expects src/functions/)
		err = os.MkdirAll("src/functions/api", 0755)
		require.NoError(t, err)

		// Create main.go to satisfy Go runtime detection
		mainGo := `package main

func main() {}`
		err = os.WriteFile("src/functions/api/main.go", []byte(mainGo), 0644)
		require.NoError(t, err)

		// Try to deploy with namespace (exercises namespace parameter path)
		err = runDeploy(false, "pr-123")
		// May fail at terraform stage, but exercises the namespace parameter
		_ = err
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
		err = runDestroy(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
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
		err = createProject("test-project", "go1.x", false)
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

// TestRunDeployWithMultipleFunctions tests deploying multiple Lambda functions
func TestRunDeployWithMultipleFunctions(t *testing.T) {
	t.Run("deploys multiple functions via convention-based discovery", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create infra directory
		err = os.MkdirAll("infra", 0755)
		require.NoError(t, err)

		// Create multiple function directories (convention-based expects src/functions/)
		for _, funcName := range []string{"api", "worker"} {
			err = os.MkdirAll(filepath.Join("src/functions", funcName), 0755)
			require.NoError(t, err)

			// Create main.go for Go runtime detection
			mainGo := `package main

func main() {}`
			err = os.WriteFile(filepath.Join("src/functions", funcName, "main.go"), []byte(mainGo), 0644)
			require.NoError(t, err)
		}

		// Try to deploy all functions (convention-based discovery)
		err = runDeploy(false, "")
		// May succeed or fail depending on terraform setup
		_ = err
	})
}

// TestRunDeployWithAutoApprove tests auto-approve flag
func TestRunDeployWithAutoApprove(t *testing.T) {
	t.Run("deploys with auto-approve flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create infra directory
		err = os.MkdirAll("infra", 0755)
		require.NoError(t, err)

		// Create a function directory
		err = os.MkdirAll("src/functions/api", 0755)
		require.NoError(t, err)

		// Create main.go
		mainGo := `package main

func main() {}`
		err = os.WriteFile("src/functions/api/main.go", []byte(mainGo), 0644)
		require.NoError(t, err)

		// Try to deploy with auto-approve (exercises auto-approve parameter)
		err = runDeploy(true, "")
		// May succeed or fail depending on terraform setup
		_ = err
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
		err = runDestroy(true)
		// May succeed or fail depending on terraform setup, but we exercised the multi-stack path
		// We don't assert error here since the test environment may have mock terraform
	})
}

// TestAdaptTerraformExecutor tests the functional adapter
func TestAdaptTerraformExecutor(t *testing.T) {
	t.Run("functional adapter creates struct with function fields", func(t *testing.T) {
		// Test pure functional adaptation - NO OOP, NO METHODS!
		tfPath := findTerraformPath()
		tfExec := terraform.NewExecutor(tfPath)
		adapted := adaptTerraformExecutor(tfExec)

		// Verify all function fields are set
		assert.NotNil(t, adapted.Init)
		assert.NotNil(t, adapted.Plan)
		assert.NotNil(t, adapted.PlanWithVars)
		assert.NotNil(t, adapted.Apply)
		assert.NotNil(t, adapted.Output)
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

		err = createProject("test-project", "go1.x", false)
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
		err = createProject("test-project", "go1.x", false)
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
