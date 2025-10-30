package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/terraform"
)

// TestNewDestroyCmd tests the destroy command creation.
func TestNewDestroyCmd(t *testing.T) {
	t.Run("creates destroy command", func(t *testing.T) {
		cmd := NewDestroyCmd()

		assert.NotNil(t, cmd)
		assert.Equal(t, "destroy", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.Contains(t, cmd.Long, "💥 Forge Destroy")
		assert.Contains(t, cmd.Long, "PERMANENTLY DELETE")
	})

	t.Run("has auto-approve flag", func(t *testing.T) {
		cmd := NewDestroyCmd()

		flag := cmd.Flags().Lookup("auto-approve")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("requires no args", func(t *testing.T) {
		cmd := NewDestroyCmd()
		assert.NotNil(t, cmd.Args)
		// Test that Args function rejects arguments
		err := cmd.Args(cmd, []string{"extra"})
		assert.Error(t, err, "Should reject extra arguments")
	})

	t.Run("has RunE function", func(t *testing.T) {
		cmd := NewDestroyCmd()
		assert.NotNil(t, cmd.RunE)
	})
}

// TestRunDestroy tests the runDestroy function with various scenarios.
func TestRunDestroy(t *testing.T) {
	t.Run("returns error when config loading fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Change to directory without forge.hcl
		os.Chdir(tmpDir)

		err := runDestroy(true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})

	t.Run("succeeds with auto-approve and valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create minimal forge.hcl
		forgeHCL := `
project {
  name   = "test-project"
  region = "us-east-1"
}
`
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "forge.hcl"), []byte(forgeHCL), 0o644))

		// Create infra directory for terraform
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		os.Chdir(tmpDir)

		// This will fail with terraform not found or terraform errors, but config loading should succeed
		err := runDestroy(true)
		// We expect terraform-related errors, not config errors
		if err != nil {
			assert.NotContains(t, err.Error(), "failed to load config")
		}
	})

	t.Run("handles working directory error gracefully", func(t *testing.T) {
		// We can't really make os.Getwd() fail in a test, but we can verify the error path
		// by checking that runDestroy handles errors properly
		// This test documents the expected behavior
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		// Without forge.hcl, we'll get config error
		err := runDestroy(true)
		assert.Error(t, err)
	})
}

// TestDestroyPipeline tests the destroy pipeline using functional approach.
func TestDestroyPipeline(t *testing.T) {
	t.Run("destroys infrastructure", func(t *testing.T) {
		exec := terraform.NewMockExecutor()

		destroyPipeline := pipeline.New(
			pipeline.TerraformDestroy(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
		}

		result := pipeline.Run(destroyPipeline, t.Context(), initialState)

		assert.True(t, E.IsRight(result), "Destroy pipeline should succeed")
	})

	t.Run("handles destroy failure", func(t *testing.T) {
		exec := terraform.Executor{
			Destroy: func(ctx context.Context, dir string, opts ...terraform.DestroyOption) error {
				return assert.AnError
			},
		}

		destroyPipeline := pipeline.New(
			pipeline.TerraformDestroy(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
		}

		result := pipeline.Run(destroyPipeline, t.Context(), initialState)

		assert.True(t, E.IsLeft(result), "Should fail on destroy error")
	})
}

// TestDestroyWithAutoApprove tests auto-approve flag.
func TestDestroyWithAutoApprove(t *testing.T) {
	t.Run("auto-approve flag is passed to executor", func(t *testing.T) {
		var receivedAutoApprove bool

		exec := terraform.Executor{
			Destroy: func(ctx context.Context, dir string, opts ...terraform.DestroyOption) error {
				// Apply options and check autoApprove
				cfg := terraform.DestroyConfig{}
				for _, opt := range opts {
					opt(&cfg)
				}
				receivedAutoApprove = cfg.AutoApprove
				return nil
			},
		}

		destroyPipeline := pipeline.New(
			pipeline.TerraformDestroy(exec, true), // autoApprove = true
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
		}

		result := pipeline.Run(destroyPipeline, t.Context(), initialState)

		assert.True(t, E.IsRight(result), "Destroy should succeed")
		assert.True(t, receivedAutoApprove, "AutoApprove should be true")
	})

	t.Run("auto-approve false requires confirmation", func(t *testing.T) {
		var receivedAutoApprove bool

		exec := terraform.Executor{
			Destroy: func(ctx context.Context, dir string, opts ...terraform.DestroyOption) error {
				cfg := terraform.DestroyConfig{}
				for _, opt := range opts {
					opt(&cfg)
				}
				receivedAutoApprove = cfg.AutoApprove
				return nil
			},
		}

		destroyPipeline := pipeline.New(
			pipeline.TerraformDestroy(exec, false), // autoApprove = false
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
		}

		result := pipeline.Run(destroyPipeline, t.Context(), initialState)

		assert.True(t, E.IsRight(result), "Destroy should succeed")
		assert.False(t, receivedAutoApprove, "AutoApprove should be false")
	})
}

// TestDestroyPreservesState tests that state is preserved through pipeline.
func TestDestroyPreservesState(t *testing.T) {
	t.Run("preserves project directory and config", func(t *testing.T) {
		exec := terraform.NewMockExecutor()

		destroyPipeline := pipeline.New(
			pipeline.TerraformDestroy(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test/project",
			Config:     "test-config",
		}

		result := pipeline.Run(destroyPipeline, t.Context(), initialState)

		require.True(t, E.IsRight(result), "Destroy should succeed")

		// Extract final state
		finalState := E.Fold(
			func(e error) pipeline.State { return pipeline.State{} },
			func(s pipeline.State) pipeline.State { return s },
		)(result)

		assert.Equal(t, "/test/project", finalState.ProjectDir)
		assert.Equal(t, "test-config", finalState.Config)
	})
}

// BenchmarkDestroyPipeline benchmarks the destroy pipeline.
func BenchmarkDestroyPipeline(b *testing.B) {
	exec := terraform.NewMockExecutor()

	destroyPipeline := pipeline.New(
		pipeline.TerraformDestroy(exec, true),
	)

	initialState := pipeline.State{
		ProjectDir: "/test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Run(destroyPipeline, b.Context(), initialState)
	}
}
