package cli

import (
	"context"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/stack"
	"github.com/lewis/forge/internal/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDestroyPipeline tests the destroy pipeline using functional approach
func TestDestroyPipeline(t *testing.T) {
	t.Run("destroys a single stack", func(t *testing.T) {
		exec := terraform.NewMockExecutor()

		destroyPipeline := pipeline.New(
			pipeline.TerraformDestroy(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api"},
			},
		}

		result := destroyPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsRight(result), "Destroy pipeline should succeed")
	})

	t.Run("destroys multiple stacks in reverse order", func(t *testing.T) {
		var executionOrder []string

		exec := terraform.Executor{
			Destroy: func(ctx context.Context, dir string, opts ...terraform.DestroyOption) error {
				executionOrder = append(executionOrder, dir)
				return nil
			},
		}

		destroyPipeline := pipeline.New(
			pipeline.TerraformDestroy(exec, true),
		)

		// Stacks should be destroyed in reverse dependency order
		// TerraformDestroy internally reverses the order
		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "database", Path: "stacks/database"},
				{Name: "api", Path: "stacks/api"},
				{Name: "frontend", Path: "stacks/frontend"},
			},
		}

		result := destroyPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsRight(result), "Multi-stack destroy should succeed")
		assert.Len(t, executionOrder, 3, "Should destroy all 3 stacks")
		// TerraformDestroy reverses: database, api, frontend -> frontend, api, database
		assert.Equal(t, "stacks/frontend", executionOrder[0])
		assert.Equal(t, "stacks/api", executionOrder[1])
		assert.Equal(t, "stacks/database", executionOrder[2])
	})

	t.Run("stops on destroy failure", func(t *testing.T) {
		var executionOrder []string

		exec := terraform.Executor{
			Destroy: func(ctx context.Context, dir string, opts ...terraform.DestroyOption) error {
				executionOrder = append(executionOrder, dir)
				// Fail on api stack
				if dir == "stacks/api" {
					return assert.AnError
				}
				return nil
			},
		}

		destroyPipeline := pipeline.New(
			pipeline.TerraformDestroy(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "database", Path: "stacks/database"},
				{Name: "api", Path: "stacks/api"},
				{Name: "frontend", Path: "stacks/frontend"},
			},
		}

		result := destroyPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsLeft(result), "Should fail on stack destroy error")
		// TerraformDestroy reverses order, so it tries: frontend, api (fails), database (not attempted)
		assert.Contains(t, executionOrder, "stacks/frontend")
		assert.Contains(t, executionOrder, "stacks/api")
		// Should NOT have attempted database since api failed
		assert.NotContains(t, executionOrder, "stacks/database")
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
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api"},
			},
		}

		result := destroyPipeline.Run(context.Background(), initialState)

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
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api"},
			},
		}

		result := destroyPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsRight(result), "Destroy should succeed")
		assert.False(t, receivedAutoApprove, "AutoApprove should be false")
	})
}

// TestDestroyEmptyState tests destroying with no stacks
func TestDestroyEmptyState(t *testing.T) {
	t.Run("succeeds with no stacks", func(t *testing.T) {
		exec := terraform.NewMockExecutor()

		destroyPipeline := pipeline.New(
			pipeline.TerraformDestroy(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks:     []*stack.Stack{},
		}

		result := destroyPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsRight(result), "Should succeed with empty stack list")
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
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api"},
			},
			Config: "test-config",
		}

		result := destroyPipeline.Run(context.Background(), initialState)

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
		Stacks: []*stack.Stack{
			{Name: "api", Path: "stacks/api"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destroyPipeline.Run(context.Background(), initialState)
	}
}
