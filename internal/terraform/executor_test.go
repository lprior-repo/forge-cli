package terraform

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockExecutor tests the mock executor (no terraform binary needed).
func TestMockExecutor(t *testing.T) {
	t.Run("Init should succeed with mock", func(t *testing.T) {
		exec := NewMockExecutor()
		err := exec.Init(t.Context(), "/tmp/test")
		assert.NoError(t, err)
	})

	t.Run("Plan should return hasChanges=true by default", func(t *testing.T) {
		exec := NewMockExecutor()
		hasChanges, err := exec.Plan(t.Context(), "/tmp/test")
		assert.NoError(t, err)
		assert.True(t, hasChanges, "Mock should return hasChanges=true by default")
	})

	t.Run("Apply should succeed with mock", func(t *testing.T) {
		exec := NewMockExecutor()
		err := exec.Apply(t.Context(), "/tmp/test")
		assert.NoError(t, err)
	})

	t.Run("Destroy should succeed with mock", func(t *testing.T) {
		exec := NewMockExecutor()
		err := exec.Destroy(t.Context(), "/tmp/test")
		assert.NoError(t, err)
	})

	t.Run("Output should return empty map by default", func(t *testing.T) {
		exec := NewMockExecutor()
		outputs, err := exec.Output(t.Context(), "/tmp/test")
		assert.NoError(t, err)
		assert.NotNil(t, outputs)
		assert.Empty(t, outputs)
	})

	t.Run("Validate should succeed with mock", func(t *testing.T) {
		exec := NewMockExecutor()
		err := exec.Validate(t.Context(), "/tmp/test")
		assert.NoError(t, err)
	})
}

// TestMockExecutorCustomBehavior tests customizing mock behavior.
func TestMockExecutorCustomBehavior(t *testing.T) {
	t.Run("can customize Init to return error", func(t *testing.T) {
		exec := Executor{
			Init: func(ctx context.Context, dir string, opts ...InitOption) error {
				return assert.AnError
			},
		}

		err := exec.Init(t.Context(), "/tmp/test")
		assert.Error(t, err)
	})

	t.Run("can customize Plan to return hasChanges=false", func(t *testing.T) {
		exec := Executor{
			Plan: func(ctx context.Context, dir string, opts ...PlanOption) (bool, error) {
				return false, nil
			},
		}

		hasChanges, err := exec.Plan(t.Context(), "/tmp/test")
		assert.NoError(t, err)
		assert.False(t, hasChanges, "Custom mock should return hasChanges=false")
	})

	t.Run("can track calls to operations", func(t *testing.T) {
		var initCalled, planCalled, applyCalled bool

		exec := Executor{
			Init: func(ctx context.Context, dir string, opts ...InitOption) error {
				initCalled = true
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...PlanOption) (bool, error) {
				planCalled = true
				return true, nil
			},
			Apply: func(ctx context.Context, dir string, opts ...ApplyOption) error {
				applyCalled = true
				return nil
			},
		}

		// Execute operations
		exec.Init(t.Context(), "/tmp/test")
		exec.Plan(t.Context(), "/tmp/test")
		exec.Apply(t.Context(), "/tmp/test")

		// Verify all were called
		assert.True(t, initCalled, "Init should have been called")
		assert.True(t, planCalled, "Plan should have been called")
		assert.True(t, applyCalled, "Apply should have been called")
	})

	t.Run("can verify options passed to operations", func(t *testing.T) {
		var receivedOpts []InitOption

		exec := Executor{
			Init: func(ctx context.Context, dir string, opts ...InitOption) error {
				receivedOpts = opts
				return nil
			},
		}

		// Call with options
		exec.Init(t.Context(), "/tmp/test", Upgrade(true), Backend(false))

		// Verify options were received
		assert.Len(t, receivedOpts, 2, "Should receive 2 options")
	})
}

// TestFunctionalOptions tests the functional options pattern.
func TestFunctionalOptions(t *testing.T) {
	t.Run("InitOptions", func(t *testing.T) {
		t.Run("Upgrade option sets Upgrade field", func(t *testing.T) {
			cfg := applyInitOptions(Upgrade(true))
			assert.True(t, cfg.Upgrade)
		})

		t.Run("Backend option sets Backend field", func(t *testing.T) {
			cfg := applyInitOptions(Backend(false))
			assert.False(t, cfg.Backend)
		})

		t.Run("Reconfigure option sets Reconfigure field", func(t *testing.T) {
			cfg := applyInitOptions(Reconfigure(true))
			assert.True(t, cfg.Reconfigure)
		})

		t.Run("multiple options compose correctly", func(t *testing.T) {
			cfg := applyInitOptions(
				Upgrade(true),
				Backend(false),
				Reconfigure(true),
			)

			assert.True(t, cfg.Upgrade)
			assert.False(t, cfg.Backend)
			assert.True(t, cfg.Reconfigure)
		})

		t.Run("default values", func(t *testing.T) {
			cfg := applyInitOptions()
			assert.False(t, cfg.Upgrade, "Upgrade defaults to false")
			assert.True(t, cfg.Backend, "Backend defaults to true")
			assert.False(t, cfg.Reconfigure, "Reconfigure defaults to false")
		})
	})

	t.Run("PlanOptions", func(t *testing.T) {
		t.Run("PlanOut sets Out field", func(t *testing.T) {
			cfg := applyPlanOptions(PlanOut("plan.out"))
			assert.Equal(t, "plan.out", cfg.Out)
		})

		t.Run("PlanDestroy sets Destroy field", func(t *testing.T) {
			cfg := applyPlanOptions(PlanDestroy(true))
			assert.True(t, cfg.Destroy)
		})

		t.Run("PlanVarFile sets VarFile field", func(t *testing.T) {
			cfg := applyPlanOptions(PlanVarFile("terraform.tfvars"))
			assert.Equal(t, "terraform.tfvars", cfg.VarFile)
		})
	})

	t.Run("ApplyOptions", func(t *testing.T) {
		t.Run("AutoApprove sets AutoApprove field", func(t *testing.T) {
			cfg := applyApplyOptions(AutoApprove(true))
			assert.True(t, cfg.AutoApprove)
		})

		t.Run("ApplyVarFile sets VarFile field", func(t *testing.T) {
			cfg := applyApplyOptions(ApplyVarFile("terraform.tfvars"))
			assert.Equal(t, "terraform.tfvars", cfg.VarFile)
		})

		t.Run("ApplyPlanFile sets PlanFile field", func(t *testing.T) {
			cfg := applyApplyOptions(ApplyPlanFile("plan.out"))
			assert.Equal(t, "plan.out", cfg.PlanFile)
		})
	})

	t.Run("DestroyOptions", func(t *testing.T) {
		t.Run("DestroyAutoApprove sets AutoApprove field", func(t *testing.T) {
			cfg := applyDestroyOptions(DestroyAutoApprove(true))
			assert.True(t, cfg.AutoApprove)
		})

		t.Run("DestroyVarFile sets VarFile field", func(t *testing.T) {
			cfg := applyDestroyOptions(DestroyVarFile("terraform.tfvars"))
			assert.Equal(t, "terraform.tfvars", cfg.VarFile)
		})
	})
}

// TestExecutorComposition tests composing executor operations.
func TestExecutorComposition(t *testing.T) {
	t.Run("Init → Plan → Apply workflow", func(t *testing.T) {
		var workflow []string

		exec := Executor{
			Init: func(ctx context.Context, dir string, opts ...InitOption) error {
				workflow = append(workflow, "init")
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...PlanOption) (bool, error) {
				workflow = append(workflow, "plan")
				return true, nil
			},
			Apply: func(ctx context.Context, dir string, opts ...ApplyOption) error {
				workflow = append(workflow, "apply")
				return nil
			},
		}

		ctx := t.Context()
		dir := "/tmp/test"

		// Execute workflow
		err := exec.Init(ctx, dir)
		require.NoError(t, err)

		hasChanges, err := exec.Plan(ctx, dir)
		require.NoError(t, err)
		require.True(t, hasChanges)

		err = exec.Apply(ctx, dir, AutoApprove(true))
		require.NoError(t, err)

		// Verify workflow order
		assert.Equal(t, []string{"init", "plan", "apply"}, workflow)
	})

	t.Run("can skip Apply if no changes", func(t *testing.T) {
		exec := Executor{
			Plan: func(ctx context.Context, dir string, opts ...PlanOption) (bool, error) {
				return false, nil // No changes
			},
		}

		hasChanges, err := exec.Plan(t.Context(), "/tmp/test")
		require.NoError(t, err)

		if hasChanges {
			t.Fatal("Should not apply when no changes")
		}
	})
}

// TestExecutorErrorHandling tests error handling in executor.
func TestExecutorErrorHandling(t *testing.T) {
	t.Run("propagates Init errors", func(t *testing.T) {
		exec := Executor{
			Init: func(ctx context.Context, dir string, opts ...InitOption) error {
				return assert.AnError
			},
		}

		err := exec.Init(t.Context(), "/tmp/test")
		assert.Error(t, err)
	})

	t.Run("propagates Plan errors", func(t *testing.T) {
		exec := Executor{
			Plan: func(ctx context.Context, dir string, opts ...PlanOption) (bool, error) {
				return false, assert.AnError
			},
		}

		_, err := exec.Plan(t.Context(), "/tmp/test")
		assert.Error(t, err)
	})

	t.Run("propagates Apply errors", func(t *testing.T) {
		exec := Executor{
			Apply: func(ctx context.Context, dir string, opts ...ApplyOption) error {
				return assert.AnError
			},
		}

		err := exec.Apply(t.Context(), "/tmp/test")
		assert.Error(t, err)
	})
}

// BenchmarkMockExecutor benchmarks the mock executor.
func BenchmarkMockExecutor(b *testing.B) {
	exec := NewMockExecutor()
	ctx := b.Context()

	b.Run("Init", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			exec.Init(ctx, "/tmp/test")
		}
	})

	b.Run("Plan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			exec.Plan(ctx, "/tmp/test")
		}
	})

	b.Run("Apply", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			exec.Apply(ctx, "/tmp/test")
		}
	})
}
