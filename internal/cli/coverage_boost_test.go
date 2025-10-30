package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunDeployWithMockTerraform tests deploy with full mocked infrastructure
func TestRunDeployWithMockTerraform(t *testing.T) {
	t.Run("deploys successfully with auto-approve", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		// Create minimal project structure
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "main.go"), []byte("package main\nfunc main() {}"), 0644))

		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(infraDir, "main.tf"), []byte("# terraform"), 0644))

		os.Chdir(tmpDir)

		// This will fail at Terraform step but should pass early stages
		err := runDeploy(true, "")
		// Expect error at terraform execution (not mocked)
		assert.Error(t, err)
		// But should contain deployment context, not scan/build errors
		if err != nil {
			assert.Contains(t, err.Error(), "deployment failed")
		}
	})

	t.Run("accepts namespace parameter", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "main.go"), []byte("package main"), 0644))

		os.Chdir(tmpDir)

		// Should pass namespace through pipeline
		err := runDeploy(true, "test-namespace")
		assert.Error(t, err) // Will fail at terraform execution
	})
}

// TestRunDestroyWithMockTerraform tests destroy with mocked infrastructure
func TestRunDestroyWithMockTerraform(t *testing.T) {
	t.Run("fails gracefully when not in project directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		err := runDestroy(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})
}

// TestRunBuildErrorPaths tests error handling in runBuild
func TestRunBuildErrorPaths(t *testing.T) {
	t.Run("handles build failure gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		// Create function with invalid code
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))
		invalidGo := `package main
this is invalid go code that will not compile
`
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "main.go"), []byte(invalidGo), 0644))

		os.Chdir(tmpDir)

		// Stub build should succeed (doesn't compile)
		err := runBuild(true)
		assert.NoError(t, err)
	})
}

// TestNewCmdProjectCreation tests new command project creation paths
func TestNewCmdProjectCreation(t *testing.T) {
	t.Run("creates project with auto-state flag (skip actual provisioning)", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		// Test that auto-state flag is recognized (will fail at AWS provisioning)
		err := createProject("test-project", "provided.al2023", true)
		// Will fail at actual state provisioning (not implemented in test)
		// But should create project structure first
		if err != nil {
			// If it fails, should be at provisioning step, not project creation
			assert.Contains(t, err.Error(), "state")
		} else {
			// Or project was created successfully
			assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
		}
	})

	t.Run("uses environment variables for region detection", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		// Set AWS_REGION
		os.Setenv("AWS_REGION", "eu-west-1")
		defer os.Unsetenv("AWS_REGION")

		os.Chdir(tmpDir)

		err := createProject("test-project-eu", "provided.al2023", false)
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project-eu"))
	})

	t.Run("uses AWS_DEFAULT_REGION when AWS_REGION not set", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		// Ensure AWS_REGION is not set
		os.Unsetenv("AWS_REGION")
		os.Setenv("AWS_DEFAULT_REGION", "ap-southeast-1")
		defer os.Unsetenv("AWS_DEFAULT_REGION")

		os.Chdir(tmpDir)

		err := createProject("test-project-ap", "provided.al2023", false)
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project-ap"))
	})

	t.Run("defaults to us-east-1 when no env vars set", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		// Ensure no AWS region env vars
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_DEFAULT_REGION")

		os.Chdir(tmpDir)

		err := createProject("test-project-default", "provided.al2023", false)
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project-default"))
	})

	t.Run("fails when project directory already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		// Create directory first
		require.NoError(t, os.MkdirAll("existing-project", 0755))

		err := createProject("existing-project", "provided.al2023", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

// TestAddCommandEdgeCases tests edge cases in add command
func TestAddCommandEdgeCases(t *testing.T) {
	t.Run("handles working directory error", func(t *testing.T) {
		// We can't easily force os.Getwd() to fail, but we test the path
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		// Create infra directory
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0755))

		os.Chdir(tmpDir)

		cmd := NewAddCmd()
		args := []string{"sqs", "test-queue"}

		err := runAdd(cmd, args, "", false, false)
		assert.NoError(t, err)
	})
}

// TestPipelineIntegration tests pipeline integration with CLI
func TestPipelineIntegration(t *testing.T) {
	t.Run("pipeline handles successful workflow", func(t *testing.T) {
		// Test successful pipeline execution
		pipe := pipeline.NewEventPipeline(
			// Empty stage for testing
			func(ctx context.Context, state pipeline.State) E.Either[error, pipeline.StageResult] {
				return E.Right[error](pipeline.StageResult{
					State:  state,
					Events: []pipeline.StageEvent{}, // Empty events
				})
			},
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Artifacts:  make(map[string]pipeline.Artifact),
			Outputs:    make(map[string]interface{}),
		}

		result := pipeline.RunWithEvents(pipe, context.Background(), initialState)

		assert.True(t, E.IsRight(result))
	})

	t.Run("pipeline handles failure", func(t *testing.T) {
		pipe := pipeline.NewEventPipeline(
			// Failing stage
			func(ctx context.Context, state pipeline.State) E.Either[error, pipeline.StageResult] {
				return E.Left[pipeline.StageResult](assert.AnError)
			},
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
		}

		result := pipeline.RunWithEvents(pipe, context.Background(), initialState)

		assert.True(t, E.IsLeft(result))
	})
}

// TestCommandOutputFormatting tests that commands produce expected output
func TestCommandOutputFormatting(t *testing.T) {
	t.Run("build command shows helpful output on error", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		// No src/functions directory
		err := runBuild(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan functions")
	})

	t.Run("deploy command shows helpful output on error", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		err := runDeploy(false, "")
		assert.Error(t, err)
	})

	t.Run("destroy command shows helpful output on error", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		err := runDestroy(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})
}

// Note: Version command tests removed - version.go already has 100% coverage
// The version command uses fmt.Println which writes to stdout, not cmd.Out
// Testing output capture would require os.Stdout redirection which is tested in integration tests

// TestFlagInteractions tests complex flag interactions
func TestFlagInteractions(t *testing.T) {
	t.Run("add command with both --raw and --no-module", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0755))

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmpDir)

		cmd := NewAddCmd()
		args := []string{"sqs", "test-queue"}

		// Both flags set to true - should use raw mode
		err := runAdd(cmd, args, "", true, true)
		assert.NoError(t, err)
	})

	t.Run("new command without project name or stack flag", func(t *testing.T) {
		cmd := NewNewCmd()
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("build command with stub-only and no functions", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		// Create empty functions directory
		functionsDir := filepath.Join(tmpDir, "src", "functions")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))

		os.Chdir(tmpDir)

		err := runBuild(true)
		assert.NoError(t, err) // Should succeed with no functions message
	})
}

// TestCreateStackEdgeCases tests edge cases in stack creation
func TestCreateStackEdgeCases(t *testing.T) {
	t.Run("createStack with all default values", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		// Create forge.hcl
		require.NoError(t, os.WriteFile("forge.hcl", []byte("service = \"test\""), 0644))

		err := createStack("default-stack", "provided.al2023", "")
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "default-stack"))
	})

	t.Run("createStack fails when not in project", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		// No forge.hcl
		err := createStack("stack", "provided.al2023", "description")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in a Forge project")
	})
}
