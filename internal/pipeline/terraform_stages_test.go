package pipeline

import (
	"context"
	"fmt"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a mock executor with custom behavior
func newMockExecutor(
	initErr error,
	planResult bool, planErr error,
	applyErr error,
	destroyErr error,
	outputs map[string]interface{}, outputErr error,
) terraform.Executor {
	return terraform.Executor{
		Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
			return initErr
		},
		Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
			return planResult, planErr
		},
		Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
			return applyErr
		},
		Destroy: func(ctx context.Context, dir string, opts ...terraform.DestroyOption) error {
			return destroyErr
		},
		Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
			return outputs, outputErr
		},
	}
}

func TestTerraformInit(t *testing.T) {
	t.Run("executes init successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		exec := newMockExecutor(nil, false, nil, nil, nil, nil, nil)
		stage := TerraformInit(exec)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "Init should succeed")

		resultState := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Equal(t, "/project/infra", resultState.ProjectDir)
	})

	t.Run("returns error when init fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		initErr := fmt.Errorf("terraform not found")
		exec := newMockExecutor(initErr, false, nil, nil, nil, nil, nil)
		stage := TerraformInit(exec)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsLeft(result), "Init should fail")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "init failed")
		assert.Contains(t, err.Error(), "terraform not found")
	})

	t.Run("preserves state on success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		exec := newMockExecutor(nil, false, nil, nil, nil, nil, nil)
		stage := TerraformInit(exec)
		state := State{
			ProjectDir: "/project/infra",
			Artifacts: map[string]Artifact{
				"lambda": {Path: "/build/lambda.zip"},
			},
		}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "Init should succeed")

		resultState := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Contains(t, resultState.Artifacts, "lambda")
	})
}

func TestTerraformPlan(t *testing.T) {
	t.Run("executes plan successfully with changes", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		exec := newMockExecutor(nil, true, nil, nil, nil, nil, nil)
		stage := TerraformPlan(exec)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "Plan should succeed")
	})

	t.Run("executes plan successfully with no changes", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		exec := newMockExecutor(nil, false, nil, nil, nil, nil, nil)
		stage := TerraformPlan(exec)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "Plan should succeed even with no changes")
	})

	t.Run("returns error when plan fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		planErr := fmt.Errorf("invalid configuration")
		exec := newMockExecutor(nil, false, planErr, nil, nil, nil, nil)
		stage := TerraformPlan(exec)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsLeft(result), "Plan should fail")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "plan failed")
		assert.Contains(t, err.Error(), "invalid configuration")
	})
}

func TestTerraformApply(t *testing.T) {
	t.Run("executes apply successfully with auto-approve", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		exec := newMockExecutor(nil, false, nil, nil, nil, nil, nil)
		stage := TerraformApply(exec, true)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "Apply should succeed")
	})

	t.Run("executes apply successfully without auto-approve", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		exec := newMockExecutor(nil, false, nil, nil, nil, nil, nil)
		stage := TerraformApply(exec, false)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "Apply should succeed")
	})

	t.Run("returns error when apply fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		applyErr := fmt.Errorf("resource creation failed")
		exec := newMockExecutor(nil, false, nil, applyErr, nil, nil, nil)
		stage := TerraformApply(exec, true)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsLeft(result), "Apply should fail")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "apply failed")
		assert.Contains(t, err.Error(), "resource creation failed")
	})

	t.Run("preserves state on success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		exec := newMockExecutor(nil, false, nil, nil, nil, nil, nil)
		stage := TerraformApply(exec, true)
		state := State{
			ProjectDir: "/project/infra",
			Artifacts: map[string]Artifact{
				"lambda": {Path: "/build/lambda.zip"},
			},
		}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "Apply should succeed")

		resultState := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Contains(t, resultState.Artifacts, "lambda")
	})
}

func TestTerraformDestroy(t *testing.T) {
	t.Run("executes destroy successfully with auto-approve", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		exec := newMockExecutor(nil, false, nil, nil, nil, nil, nil)
		stage := TerraformDestroy(exec, true)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "Destroy should succeed")
	})

	t.Run("executes destroy successfully without auto-approve", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		exec := newMockExecutor(nil, false, nil, nil, nil, nil, nil)
		stage := TerraformDestroy(exec, false)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "Destroy should succeed")
	})

	t.Run("returns error when destroy fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		destroyErr := fmt.Errorf("resource still in use")
		exec := newMockExecutor(nil, false, nil, nil, destroyErr, nil, nil)
		stage := TerraformDestroy(exec, true)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsLeft(result), "Destroy should fail")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "destroy failed")
		assert.Contains(t, err.Error(), "resource still in use")
	})
}

func TestCaptureOutputs(t *testing.T) {
	t.Run("captures outputs successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		outputs := map[string]interface{}{
			"api_url":     "https://api.example.com",
			"bucket_name": "my-bucket",
		}
		exec := newMockExecutor(nil, false, nil, nil, nil, outputs, nil)
		stage := CaptureOutputs(exec)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "CaptureOutputs should succeed")

		resultState := E.GetOrElse(func(error) State { return State{} })(result)
		assert.NotNil(t, resultState.Outputs)
		assert.Contains(t, resultState.Outputs, "main")

		mainOutputs := resultState.Outputs["main"].(map[string]interface{})
		assert.Equal(t, "https://api.example.com", mainOutputs["api_url"])
		assert.Equal(t, "my-bucket", mainOutputs["bucket_name"])
	})

	t.Run("handles empty outputs", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		outputs := map[string]interface{}{}
		exec := newMockExecutor(nil, false, nil, nil, nil, outputs, nil)
		stage := CaptureOutputs(exec)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "CaptureOutputs should succeed with empty outputs")

		resultState := E.GetOrElse(func(error) State { return State{} })(result)
		assert.NotNil(t, resultState.Outputs)
		assert.Contains(t, resultState.Outputs, "main")
	})

	t.Run("returns error when output fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		outputErr := fmt.Errorf("no state file found")
		exec := newMockExecutor(nil, false, nil, nil, nil, nil, outputErr)
		stage := CaptureOutputs(exec)
		state := State{ProjectDir: "/project/infra"}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsLeft(result), "CaptureOutputs should fail")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "failed to get outputs")
		assert.Contains(t, err.Error(), "no state file found")
	})

	t.Run("preserves existing outputs", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		outputs := map[string]interface{}{
			"new_output": "value",
		}
		exec := newMockExecutor(nil, false, nil, nil, nil, outputs, nil)
		stage := CaptureOutputs(exec)
		state := State{
			ProjectDir: "/project/infra",
			Outputs: map[string]interface{}{
				"existing": "data",
			},
		}

		// Act
		result := stage(ctx, state)

		// Assert
		require.True(t, E.IsRight(result), "CaptureOutputs should succeed")

		resultState := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Contains(t, resultState.Outputs, "existing")
		assert.Contains(t, resultState.Outputs, "main")
	})
}

// Integration test for terraform stages in a pipeline
func TestTerraformStagesIntegration(t *testing.T) {
	t.Run("full terraform deployment pipeline", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		outputs := map[string]interface{}{
			"endpoint": "https://api.example.com",
		}
		exec := newMockExecutor(nil, true, nil, nil, nil, outputs, nil)

		// Build a full deployment pipeline
		pipeline := New(
			TerraformInit(exec),
			TerraformPlan(exec),
			TerraformApply(exec, true),
			CaptureOutputs(exec),
		)

		initial := State{ProjectDir: "/project/infra"}

		// Act
		result := Run(pipeline, ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Full terraform pipeline should succeed")

		// Verify outputs were captured
		state := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Contains(t, state.Outputs, "main")
	})

	t.Run("pipeline stops on init failure", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		initErr := fmt.Errorf("terraform init failed")
		exec := newMockExecutor(initErr, true, nil, nil, nil, nil, nil)

		pipeline := New(
			TerraformInit(exec),
			TerraformPlan(exec),
			TerraformApply(exec, true),
		)

		initial := State{ProjectDir: "/project/infra"}

		// Act
		result := Run(pipeline, ctx, initial)

		// Assert
		require.True(t, E.IsLeft(result), "Pipeline should fail on init")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "init failed")
	})

	t.Run("destroy pipeline", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		exec := newMockExecutor(nil, false, nil, nil, nil, nil, nil)

		pipeline := New(
			TerraformDestroy(exec, true),
		)

		initial := State{ProjectDir: "/project/infra"}

		// Act
		result := Run(pipeline, ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Destroy pipeline should succeed")
	})
}
