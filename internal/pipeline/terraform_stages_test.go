package pipeline

import (
	"context"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/stack"
	"github.com/lewis/forge/internal/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTerraformInit tests the TerraformInit stage
func TestTerraformInit(t *testing.T) {
	t.Run("initializes all stacks", func(t *testing.T) {
		var initCalls []string

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				initCalls = append(initCalls, dir)
				return nil
			},
		}

		stage := TerraformInit(exec)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
				{Name: "stack2", Path: "path2"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsRight(result))
		assert.Equal(t, []string{"path1", "path2"}, initCalls)
	})

	t.Run("fails if init fails", func(t *testing.T) {
		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return assert.AnError
			},
		}

		stage := TerraformInit(exec)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsLeft(result))
	})
}

// TestTerraformPlan tests the TerraformPlan stage
func TestTerraformPlan(t *testing.T) {
	t.Run("plans all stacks", func(t *testing.T) {
		var planCalls []string

		exec := terraform.Executor{
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				planCalls = append(planCalls, dir)
				return true, nil
			},
		}

		stage := TerraformPlan(exec)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
				{Name: "stack2", Path: "path2"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsRight(result))
		assert.Equal(t, []string{"path1", "path2"}, planCalls)
	})

	t.Run("fails if plan fails", func(t *testing.T) {
		exec := terraform.Executor{
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return false, assert.AnError
			},
		}

		stage := TerraformPlan(exec)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsLeft(result))
	})

	t.Run("prints message when no changes", func(t *testing.T) {
		exec := terraform.Executor{
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return false, nil
			},
		}

		stage := TerraformPlan(exec)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsRight(result))
	})
}

// TestTerraformApply tests the TerraformApply stage
func TestTerraformApply(t *testing.T) {
	t.Run("applies all stacks with auto-approve", func(t *testing.T) {
		var applyCalls []string
		var receivedAutoApprove bool

		exec := terraform.Executor{
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				applyCalls = append(applyCalls, dir)
				cfg := terraform.ApplyConfig{}
				for _, opt := range opts {
					opt(&cfg)
				}
				receivedAutoApprove = cfg.AutoApprove
				return nil
			},
		}

		stage := TerraformApply(exec, true)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
				{Name: "stack2", Path: "path2"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsRight(result))
		assert.Equal(t, []string{"path1", "path2"}, applyCalls)
		assert.True(t, receivedAutoApprove)
	})

	t.Run("fails if apply fails", func(t *testing.T) {
		exec := terraform.Executor{
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				return assert.AnError
			},
		}

		stage := TerraformApply(exec, true)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsLeft(result))
	})
}

// TestTerraformDestroy tests the TerraformDestroy stage
func TestTerraformDestroy(t *testing.T) {
	t.Run("destroys all stacks in reverse order", func(t *testing.T) {
		var destroyCalls []string

		exec := terraform.Executor{
			Destroy: func(ctx context.Context, dir string, opts ...terraform.DestroyOption) error {
				destroyCalls = append(destroyCalls, dir)
				return nil
			},
		}

		stage := TerraformDestroy(exec, true)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
				{Name: "stack2", Path: "path2"},
				{Name: "stack3", Path: "path3"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsRight(result))
		// Should be in reverse order
		assert.Equal(t, []string{"path3", "path2", "path1"}, destroyCalls)
	})

	t.Run("fails if destroy fails", func(t *testing.T) {
		exec := terraform.Executor{
			Destroy: func(ctx context.Context, dir string, opts ...terraform.DestroyOption) error {
				return assert.AnError
			},
		}

		stage := TerraformDestroy(exec, true)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsLeft(result))
	})

	t.Run("passes auto-approve flag", func(t *testing.T) {
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

		stage := TerraformDestroy(exec, true)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsRight(result))
		assert.True(t, receivedAutoApprove)
	})
}

// TestCaptureOutputs tests the CaptureOutputs stage
func TestCaptureOutputs(t *testing.T) {
	t.Run("captures outputs from all stacks", func(t *testing.T) {
		exec := terraform.Executor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				return map[string]interface{}{
					"url": "https://example.com",
					"port": 8080,
				}, nil
			},
		}

		stage := CaptureOutputs(exec)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
				{Name: "stack2", Path: "path2"},
			},
		}

		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Contains(t, finalState.Outputs, "stack1")
		assert.Contains(t, finalState.Outputs, "stack2")

		stack1Outputs := finalState.Outputs["stack1"].(map[string]interface{})
		assert.Equal(t, "https://example.com", stack1Outputs["url"])
		assert.Equal(t, 8080, stack1Outputs["port"])
	})

	t.Run("fails if output fails", func(t *testing.T) {
		exec := terraform.Executor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				return nil, assert.AnError
			},
		}

		stage := CaptureOutputs(exec)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
			},
		}

		result := stage(context.Background(), state)

		assert.True(t, E.IsLeft(result))
	})

	t.Run("initializes outputs map if nil", func(t *testing.T) {
		exec := terraform.Executor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				return map[string]interface{}{"key": "value"}, nil
			},
		}

		stage := CaptureOutputs(exec)
		state := State{
			Stacks: []*stack.Stack{
				{Name: "stack1", Path: "path1"},
			},
			Outputs: nil, // Explicitly nil
		}

		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.NotNil(t, finalState.Outputs)
		assert.Contains(t, finalState.Outputs, "stack1")
	})
}
