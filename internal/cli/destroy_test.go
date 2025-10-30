package cli

import (
	"context"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDestroyPipeline tests the destroy pipeline using functional approach
func TestDestroyPipeline(t *testing.T) {
	t.Run("destroys infrastructure", func(t *testing.T) {
		exec := terraform.NewMockExecutor()

		destroyPipeline := pipeline.New(
			pipeline.TerraformDestroy(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
		}

		result := pipeline.Run(destroyPipeline, context.Background(), initialState)

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

		result := pipeline.Run(destroyPipeline, context.Background(), initialState)

		assert.True(t, E.IsLeft(result), "Should fail on destroy error")
	})
}

// TestDestroyWithAutoApprove tests auto-approve flag
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

		result := pipeline.Run(destroyPipeline, context.Background(), initialState)

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

		result := pipeline.Run(destroyPipeline, context.Background(), initialState)

		assert.True(t, E.IsRight(result), "Destroy should succeed")
		assert.False(t, receivedAutoApprove, "AutoApprove should be false")
	})
}

// TestDestroyPreservesState tests that state is preserved through pipeline
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

		result := pipeline.Run(destroyPipeline, context.Background(), initialState)

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

// BenchmarkDestroyPipeline benchmarks the destroy pipeline
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
		pipeline.Run(destroyPipeline, context.Background(), initialState)
	}
}
