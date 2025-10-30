package state

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lewis/forge/internal/terraform"
)

// TestWriteStateBootstrap tests writing bootstrap Terraform files.
func TestWriteStateBootstrap(t *testing.T) {
	t.Run("writes bootstrap files successfully", func(t *testing.T) {
		tmpDir := t.TempDir()

		resources := GenerateStateResources("test-project", "us-east-1", "")

		bootstrapDir, err := WriteStateBootstrap(tmpDir, resources)
		require.NoError(t, err, "Should write bootstrap files successfully")

		// Verify bootstrap directory was created
		assert.DirExists(t, bootstrapDir, "Bootstrap directory should exist")
		assert.Equal(t, filepath.Join(tmpDir, ".forge", "bootstrap"), bootstrapDir)

		// Verify bootstrap.tf was created
		bootstrapPath := filepath.Join(bootstrapDir, "bootstrap.tf")
		assert.FileExists(t, bootstrapPath, "bootstrap.tf should exist")

		// Verify content
		content, err := os.ReadFile(bootstrapPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "aws_s3_bucket")
		assert.Contains(t, string(content), "aws_dynamodb_table")
		assert.Contains(t, string(content), "forge-state-test-project")
	})

	t.Run("creates bootstrap directory if it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		resources := GenerateStateResources("test", "us-west-2", "")

		bootstrapDir, err := WriteStateBootstrap(tmpDir, resources)
		require.NoError(t, err)

		assert.DirExists(t, bootstrapDir)
	})

	t.Run("overwrites existing bootstrap files", func(t *testing.T) {
		tmpDir := t.TempDir()

		resources1 := GenerateStateResources("project1", "us-east-1", "")
		resources2 := GenerateStateResources("project2", "us-west-2", "")

		// Write first time
		_, err := WriteStateBootstrap(tmpDir, resources1)
		require.NoError(t, err)

		// Write second time with different resources
		bootstrapDir, err := WriteStateBootstrap(tmpDir, resources2)
		require.NoError(t, err)

		// Verify new content
		content, err := os.ReadFile(filepath.Join(bootstrapDir, "bootstrap.tf"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "project2")
		assert.Contains(t, string(content), "us-west-2")
	})
}

// TestWriteBackendConfig tests writing backend.tf.
func TestWriteBackendConfig(t *testing.T) {
	t.Run("writes backend.tf successfully", func(t *testing.T) {
		tmpDir := t.TempDir()

		config := GenerateBackendConfig("test-project", "us-east-1", "")

		backendPath, err := WriteBackendConfig(tmpDir, config)
		require.NoError(t, err, "Should write backend config successfully")

		// Verify backend.tf was created in infra/
		expectedPath := filepath.Join(tmpDir, "infra", "backend.tf")
		assert.Equal(t, expectedPath, backendPath)
		assert.FileExists(t, backendPath, "backend.tf should exist")

		// Verify content
		content, err := os.ReadFile(backendPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "backend \"s3\"")
		assert.Contains(t, string(content), "bucket")
		assert.Contains(t, string(content), "dynamodb_table")
	})

	t.Run("creates infra directory if it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		config := GenerateBackendConfig("test", "us-west-2", "")

		backendPath, err := WriteBackendConfig(tmpDir, config)
		require.NoError(t, err)

		infraDir := filepath.Dir(backendPath)
		assert.DirExists(t, infraDir)
	})

	t.Run("includes namespace in state key", func(t *testing.T) {
		tmpDir := t.TempDir()

		config := GenerateBackendConfig("test", "us-east-1", "staging")

		backendPath, err := WriteBackendConfig(tmpDir, config)
		require.NoError(t, err)

		content, err := os.ReadFile(backendPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "staging/terraform.tfstate")
	})

	t.Run("overwrites existing backend.tf", func(t *testing.T) {
		tmpDir := t.TempDir()

		config1 := GenerateBackendConfig("project1", "us-east-1", "")
		config2 := GenerateBackendConfig("project2", "us-west-2", "prod")

		// Write first time
		_, err := WriteBackendConfig(tmpDir, config1)
		require.NoError(t, err)

		// Write second time
		backendPath, err := WriteBackendConfig(tmpDir, config2)
		require.NoError(t, err)

		// Verify new content
		content, err := os.ReadFile(backendPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "project2")
		assert.Contains(t, string(content), "us-west-2")
		assert.Contains(t, string(content), "prod/terraform.tfstate")
	})
}

// TestApplyBootstrap tests applying bootstrap Terraform.
func TestApplyBootstrap(t *testing.T) {
	t.Run("executes terraform init and apply successfully", func(t *testing.T) {
		tmpDir := t.TempDir()

		var initCalled bool
		var planCalled bool
		var applyCalled bool

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				initCalled = true
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				planCalled = true
				return true, nil // Has changes
			},
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				applyCalled = true
				return nil
			},
		}

		err := ApplyBootstrap(t.Context(), tmpDir, exec)
		require.NoError(t, err)

		assert.True(t, initCalled, "Should call terraform init")
		assert.True(t, planCalled, "Should call terraform plan")
		assert.True(t, applyCalled, "Should call terraform apply")
	})

	t.Run("skips apply when no changes detected", func(t *testing.T) {
		tmpDir := t.TempDir()

		var applyCalled bool

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return false, nil // No changes
			},
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				applyCalled = true
				return nil
			},
		}

		err := ApplyBootstrap(t.Context(), tmpDir, exec)
		require.NoError(t, err)

		assert.False(t, applyCalled, "Should not call apply when no changes")
	})

	t.Run("returns error when init fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return assert.AnError
			},
		}

		err := ApplyBootstrap(t.Context(), tmpDir, exec)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "terraform init failed")
	})

	t.Run("returns error when plan fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return false, assert.AnError
			},
		}

		err := ApplyBootstrap(t.Context(), tmpDir, exec)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "terraform plan failed")
	})

	t.Run("returns error when apply fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return true, nil
			},
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				return assert.AnError
			},
		}

		err := ApplyBootstrap(t.Context(), tmpDir, exec)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "terraform apply failed")
	})
}

// TestCleanupBootstrap tests cleanup of bootstrap directory.
func TestCleanupBootstrap(t *testing.T) {
	t.Run("removes bootstrap directory successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		bootstrapDir := filepath.Join(tmpDir, ".forge", "bootstrap")

		// Create directory with some files
		require.NoError(t, os.MkdirAll(bootstrapDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(bootstrapDir, "test.txt"), []byte("test"), 0o644))

		err := CleanupBootstrap(bootstrapDir)
		require.NoError(t, err)

		// Verify directory was removed
		_, err = os.Stat(bootstrapDir)
		assert.True(t, os.IsNotExist(err), "Bootstrap directory should be removed")
	})

	t.Run("succeeds when directory doesn't exist", func(t *testing.T) {
		err := CleanupBootstrap("/non/existent/path")
		assert.NoError(t, err, "Cleanup should succeed even if directory doesn't exist")
	})
}

// TestProvisionStateBackend tests the complete provisioning workflow.
func TestProvisionStateBackend(t *testing.T) {
	t.Run("provisions state backend successfully", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return true, nil
			},
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				return nil
			},
		}

		result := ProvisionStateBackend(t.Context(), tmpDir, "test-project", "us-east-1", exec)

		require.True(t, E.IsRight(result), "Provisioning should succeed")

		provisionResult := E.Fold(
			func(e error) ProvisionResult { return ProvisionResult{} },
			func(r ProvisionResult) ProvisionResult { return r },
		)(result)

		assert.Equal(t, "forge-state-test-project", provisionResult.BucketName)
		assert.Equal(t, "forge_locks_test_project", provisionResult.TableName)
		assert.Contains(t, provisionResult.BackendTFPath, "infra/backend.tf")
		assert.True(t, provisionResult.BootstrapApplied)

		// Verify files were created
		assert.FileExists(t, provisionResult.BackendTFPath)
		assert.DirExists(t, filepath.Join(tmpDir, ".forge", "bootstrap"))
	})

	t.Run("returns error when bootstrap apply fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return assert.AnError
			},
		}

		result := ProvisionStateBackend(t.Context(), tmpDir, "test-project", "us-east-1", exec)

		assert.True(t, E.IsLeft(result), "Should fail when bootstrap apply fails")
	})

	t.Run("handles different regions", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := terraform.Executor{
			Init:  func(ctx context.Context, dir string, opts ...terraform.InitOption) error { return nil },
			Plan:  func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) { return true, nil },
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error { return nil },
		}

		result := ProvisionStateBackend(t.Context(), tmpDir, "test", "eu-west-1", exec)

		require.True(t, E.IsRight(result))

		// Verify backend config has correct region
		backendPath := filepath.Join(tmpDir, "infra", "backend.tf")
		content, err := os.ReadFile(backendPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "eu-west-1")
	})
}

// TestProvisionStateBackendSync tests the synchronous wrapper.
func TestProvisionStateBackendSync(t *testing.T) {
	t.Run("returns result on success", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := terraform.Executor{
			Init:  func(ctx context.Context, dir string, opts ...terraform.InitOption) error { return nil },
			Plan:  func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) { return true, nil },
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error { return nil },
		}

		result, err := ProvisionStateBackendSync(t.Context(), tmpDir, "test-project", "us-east-1", exec)

		require.NoError(t, err)
		assert.Equal(t, "forge-state-test-project", result.BucketName)
		assert.Equal(t, "forge_locks_test_project", result.TableName)
		assert.True(t, result.BootstrapApplied)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return assert.AnError
			},
		}

		result, err := ProvisionStateBackendSync(t.Context(), tmpDir, "test-project", "us-east-1", exec)

		assert.Error(t, err)
		assert.Equal(t, ProvisionResult{}, result)
	})
}
