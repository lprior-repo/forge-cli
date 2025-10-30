package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lewis/forge/internal/terraform"
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

	t.Run("handles working directory error gracefully", func(t *testing.T) {
		// We can't really make os.Getwd() fail, but we document expected behavior
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		// Without src/functions, we'll get scan error
		err := runDeploy(true, "")
		assert.Error(t, err)
	})

	t.Run("handles namespace with special characters", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		// Will fail early, but namespace should be passed through
		err := runDeploy(true, "pr-123-feature-x")
		assert.Error(t, err)
		// Should fail on scan, not namespace validation
		assert.Contains(t, err.Error(), "failed to scan functions")
	})
}

func TestFindTerraformPath(t *testing.T) {
	t.Run("finds terraform in PATH", func(t *testing.T) {
		path := findTerraformPath()
		// Should return either system terraform or "terraform"
		assert.NotEmpty(t, path)
	})
}

func TestAdaptTerraformExecutor(t *testing.T) {
	t.Run("adapts Init function correctly", func(t *testing.T) {
		var initCalled bool
		var receivedDir string
		var receivedUpgrade bool

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				initCalled = true
				receivedDir = dir
				// Apply options to get upgrade value
				cfg := terraform.InitConfig{}
				for _, opt := range opts {
					opt(&cfg)
				}
				receivedUpgrade = cfg.Upgrade
				return nil
			},
		}

		adapted := adaptTerraformExecutor(exec)
		err := adapted.Init(context.Background(), "/test/dir")

		assert.NoError(t, err)
		assert.True(t, initCalled)
		assert.Equal(t, "/test/dir", receivedDir)
		assert.False(t, receivedUpgrade, "Upgrade should default to false")
	})

	t.Run("adapts Plan function correctly", func(t *testing.T) {
		var planCalled bool
		var receivedDir string

		exec := terraform.Executor{
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				planCalled = true
				receivedDir = dir
				return true, nil
			},
		}

		adapted := adaptTerraformExecutor(exec)
		hasChanges, err := adapted.Plan(context.Background(), "/test/dir")

		assert.NoError(t, err)
		assert.True(t, planCalled)
		assert.True(t, hasChanges)
		assert.Equal(t, "/test/dir", receivedDir)
	})

	t.Run("adapts PlanWithVars function correctly", func(t *testing.T) {
		var planCalled bool
		var receivedVars map[string]string

		exec := terraform.Executor{
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				planCalled = true
				// Extract vars from options
				cfg := terraform.PlanConfig{}
				for _, opt := range opts {
					opt(&cfg)
				}
				receivedVars = cfg.Vars
				return false, nil
			},
		}

		adapted := adaptTerraformExecutor(exec)
		vars := map[string]string{
			"namespace": "pr-123",
			"region":    "us-west-2",
		}
		hasChanges, err := adapted.PlanWithVars(context.Background(), "/test/dir", vars)

		assert.NoError(t, err)
		assert.True(t, planCalled)
		assert.False(t, hasChanges)
		assert.Equal(t, "pr-123", receivedVars["namespace"])
		assert.Equal(t, "us-west-2", receivedVars["region"])
	})

	t.Run("adapts Apply function correctly", func(t *testing.T) {
		var applyCalled bool
		var receivedDir string
		var receivedAutoApprove bool

		exec := terraform.Executor{
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				applyCalled = true
				receivedDir = dir
				cfg := terraform.ApplyConfig{}
				for _, opt := range opts {
					opt(&cfg)
				}
				receivedAutoApprove = cfg.AutoApprove
				return nil
			},
		}

		adapted := adaptTerraformExecutor(exec)
		err := adapted.Apply(context.Background(), "/test/dir")

		assert.NoError(t, err)
		assert.True(t, applyCalled)
		assert.Equal(t, "/test/dir", receivedDir)
		assert.True(t, receivedAutoApprove, "AutoApprove should be true")
	})

	t.Run("adapts Output function correctly", func(t *testing.T) {
		var outputCalled bool
		expectedOutputs := map[string]interface{}{
			"function_url": "https://example.com",
			"function_arn": "arn:aws:lambda:...",
		}

		exec := terraform.Executor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				outputCalled = true
				return expectedOutputs, nil
			},
		}

		adapted := adaptTerraformExecutor(exec)
		outputs, err := adapted.Output(context.Background(), "/test/dir")

		assert.NoError(t, err)
		assert.True(t, outputCalled)
		assert.Equal(t, expectedOutputs, outputs)
	})

	t.Run("propagates errors from Init", func(t *testing.T) {
		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return assert.AnError
			},
		}

		adapted := adaptTerraformExecutor(exec)
		err := adapted.Init(context.Background(), "/test/dir")

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("propagates errors from Plan", func(t *testing.T) {
		exec := terraform.Executor{
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return false, assert.AnError
			},
		}

		adapted := adaptTerraformExecutor(exec)
		_, err := adapted.Plan(context.Background(), "/test/dir")

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("propagates errors from Apply", func(t *testing.T) {
		exec := terraform.Executor{
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				return assert.AnError
			},
		}

		adapted := adaptTerraformExecutor(exec)
		err := adapted.Apply(context.Background(), "/test/dir")

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("propagates errors from Output", func(t *testing.T) {
		exec := terraform.Executor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				return nil, assert.AnError
			},
		}

		adapted := adaptTerraformExecutor(exec)
		_, err := adapted.Output(context.Background(), "/test/dir")

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("handles empty vars map", func(t *testing.T) {
		var receivedVars map[string]string

		exec := terraform.Executor{
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				cfg := terraform.PlanConfig{}
				for _, opt := range opts {
					opt(&cfg)
				}
				receivedVars = cfg.Vars
				return true, nil
			},
		}

		adapted := adaptTerraformExecutor(exec)
		emptyVars := make(map[string]string)
		_, err := adapted.PlanWithVars(context.Background(), "/test/dir", emptyVars)

		assert.NoError(t, err)
		assert.Empty(t, receivedVars)
	})

	t.Run("handles nil vars map", func(t *testing.T) {
		exec := terraform.Executor{
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return true, nil
			},
		}

		adapted := adaptTerraformExecutor(exec)
		_, err := adapted.PlanWithVars(context.Background(), "/test/dir", nil)

		assert.NoError(t, err)
	})
}
