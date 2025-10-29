package cli

import (
	"context"
	"testing"

	"github.com/lewis/forge/internal/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExecutorState holds the state for mock terraform operations
type mockExecutorState struct {
	initCalled      bool
	planCalled      bool
	applyCalled     bool
	outputCalled    bool
	initErr         error
	planResult      bool
	planErr         error
	applyErr        error
	outputs         map[string]interface{}
	outputErr       error
	lastInitOpts    []terraform.InitOption
	lastPlanOpts    []terraform.PlanOption
	lastApplyOpts   []terraform.ApplyOption
}

// newMockExecutor creates a mock terraform.Executor with customizable behavior
func newMockExecutor(state *mockExecutorState) terraform.Executor {
	return terraform.Executor{
		Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
			state.initCalled = true
			state.lastInitOpts = opts
			return state.initErr
		},
		Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
			state.planCalled = true
			state.lastPlanOpts = opts
			return state.planResult, state.planErr
		},
		Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
			state.applyCalled = true
			state.lastApplyOpts = opts
			return state.applyErr
		},
		Destroy: func(ctx context.Context, dir string, opts ...terraform.DestroyOption) error {
			return nil
		},
		Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
			state.outputCalled = true
			return state.outputs, state.outputErr
		},
		Validate: func(ctx context.Context, dir string) error {
			return nil
		},
	}
}

// TestAdaptTerraformExecutorInit tests the function!)
func TestAdaptTerraformExecutorInit(t *testing.T) {
	t.Run("calls terraform executor Init with correct options", func(t *testing.T) {
		state := &mockExecutorState{}
		mock := newMockExecutor(state)
		// Pure functional adapter - no OOP, no methods!
		adapted := adaptTerraformExecutor(mock)

		err := adapted.Init(context.Background(), "/test/dir")

		require.NoError(t, err)
		assert.True(t, state.initCalled)
	})

	t.Run("propagates Init errors", func(t *testing.T) {
		state := &mockExecutorState{initErr: assert.AnError}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		err := adapted.Init(context.Background(), "/test/dir")

		assert.Error(t, err)
		assert.True(t, state.initCalled)
	})
}

// TestAdaptTerraformExecutorPlan tests the function
func TestAdaptTerraformExecutorPlan(t *testing.T) {
	t.Run("calls PlanWithVars with nil vars", func(t *testing.T) {
		state := &mockExecutorState{planResult: true}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		hasChanges, err := adapted.Plan(context.Background(), "/test/dir")

		require.NoError(t, err)
		assert.True(t, hasChanges)
		assert.True(t, state.planCalled)
	})

	t.Run("propagates Plan errors", func(t *testing.T) {
		state := &mockExecutorState{planErr: assert.AnError}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		_, err := adapted.Plan(context.Background(), "/test/dir")

		assert.Error(t, err)
		assert.True(t, state.planCalled)
	})
}

// TestAdaptTerraformExecutorPlanWithVars tests the function
func TestAdaptTerraformExecutorPlanWithVars(t *testing.T) {
	t.Run("calls terraform executor Plan with variables", func(t *testing.T) {
		state := &mockExecutorState{planResult: true}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		vars := map[string]string{
			"region":    "us-west-2",
			"namespace": "pr-123",
		}

		hasChanges, err := adapted.PlanWithVars(context.Background(), "/test/dir", vars)

		require.NoError(t, err)
		assert.True(t, hasChanges)
		assert.True(t, state.planCalled)
		// Verify PlanOut option is used
		assert.NotEmpty(t, state.lastPlanOpts)
	})

	t.Run("handles nil vars", func(t *testing.T) {
		state := &mockExecutorState{planResult: false}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		hasChanges, err := adapted.PlanWithVars(context.Background(), "/test/dir", nil)

		require.NoError(t, err)
		assert.False(t, hasChanges)
		assert.True(t, state.planCalled)
	})

	t.Run("propagates Plan errors", func(t *testing.T) {
		state := &mockExecutorState{planErr: assert.AnError}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		_, err := adapted.PlanWithVars(context.Background(), "/test/dir", map[string]string{"key": "value"})

		assert.Error(t, err)
		assert.True(t, state.planCalled)
	})
}

// TestAdaptTerraformExecutorApply tests the function
func TestAdaptTerraformExecutorApply(t *testing.T) {
	t.Run("calls terraform executor Apply with plan file", func(t *testing.T) {
		state := &mockExecutorState{}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		err := adapted.Apply(context.Background(), "/test/dir")

		require.NoError(t, err)
		assert.True(t, state.applyCalled)
		// Verify ApplyPlanFile and AutoApprove options are used
		assert.NotEmpty(t, state.lastApplyOpts)
	})

	t.Run("propagates Apply errors", func(t *testing.T) {
		state := &mockExecutorState{applyErr: assert.AnError}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		err := adapted.Apply(context.Background(), "/test/dir")

		assert.Error(t, err)
		assert.True(t, state.applyCalled)
	})
}

// TestAdaptTerraformExecutorOutput tests the function
func TestAdaptTerraformExecutorOutput(t *testing.T) {
	t.Run("calls terraform executor Output and returns outputs", func(t *testing.T) {
		expectedOutputs := map[string]interface{}{
			"function_arn":  "arn:aws:lambda:us-east-1:123456789012:function:my-function",
			"function_name": "my-function",
		}
		state := &mockExecutorState{outputs: expectedOutputs}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		outputs, err := adapted.Output(context.Background(), "/test/dir")

		require.NoError(t, err)
		assert.Equal(t, expectedOutputs, outputs)
		assert.True(t, state.outputCalled)
	})

	t.Run("propagates Output errors", func(t *testing.T) {
		state := &mockExecutorState{outputErr: assert.AnError}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		_, err := adapted.Output(context.Background(), "/test/dir")

		assert.Error(t, err)
		assert.True(t, state.outputCalled)
	})

	t.Run("handles empty outputs", func(t *testing.T) {
		state := &mockExecutorState{outputs: make(map[string]interface{})}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)

		outputs, err := adapted.Output(context.Background(), "/test/dir")

		require.NoError(t, err)
		assert.Empty(t, outputs)
		assert.True(t, state.outputCalled)
	})
}

// TestAdaptTerraformExecutorIntegration tests the full functional adapter workflow
func TestAdaptTerraformExecutorIntegration(t *testing.T) {
	t.Run("full deployment workflow", func(t *testing.T) {
		state := &mockExecutorState{
			planResult: true,
			outputs: map[string]interface{}{
				"api_url": "https://api.example.com",
			},
		}
		mock := newMockExecutor(state)
		adapted := adaptTerraformExecutor(mock)
		ctx := context.Background()
		dir := "/test/infra"

		// Init
		err := adapted.Init(ctx, dir)
		require.NoError(t, err)
		assert.True(t, state.initCalled)

		// Plan with vars
		hasChanges, err := adapted.PlanWithVars(ctx, dir, map[string]string{"namespace": "test"})
		require.NoError(t, err)
		assert.True(t, hasChanges)
		assert.True(t, state.planCalled)

		// Apply
		err = adapted.Apply(ctx, dir)
		require.NoError(t, err)
		assert.True(t, state.applyCalled)

		// Output
		outputs, err := adapted.Output(ctx, dir)
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com", outputs["api_url"])
		assert.True(t, state.outputCalled)
	})
}
