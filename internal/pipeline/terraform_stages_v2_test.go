package pipeline

import (
	"context"
	"fmt"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConventionTerraformInitV2 tests the event-based Terraform init stage
func TestConventionTerraformInitV2(t *testing.T) {
	t.Run("initializes Terraform successfully with events", func(t *testing.T) {
		called := false
		exec := TerraformExecutor{
			Init: func(ctx context.Context, dir string) error {
				called = true
				return nil
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformInitV2(exec)
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result), "Should successfully initialize")
		assert.True(t, called, "Init should be called")

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Verify events
		assert.NotEmpty(t, stageResult.Events, "Should have events")
		assert.Equal(t, EventLevelInfo, stageResult.Events[0].Level)
		assert.Contains(t, stageResult.Events[0].Message, "Initializing Terraform")

		// Should have success event
		hasSuccessEvent := false
		for _, event := range stageResult.Events {
			if event.Level == EventLevelSuccess && event.Message == "[terraform] Initialized" {
				hasSuccessEvent = true
				break
			}
		}
		assert.True(t, hasSuccessEvent, "Should have success event")
	})

	t.Run("returns error when init fails", func(t *testing.T) {
		exec := TerraformExecutor{
			Init: func(ctx context.Context, dir string) error {
				return fmt.Errorf("init failed")
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformInitV2(exec)
		result := stage(context.Background(), state)

		require.True(t, E.IsLeft(result), "Should return error")

		err := E.Fold(
			func(e error) error { return e },
			func(r StageResult) error { return nil },
		)(result)

		assert.Contains(t, err.Error(), "terraform init failed")
	})

	t.Run("preserves state on success", func(t *testing.T) {
		exec := TerraformExecutor{
			Init: func(ctx context.Context, dir string) error {
				return nil
			},
		}

		state := State{
			ProjectDir: "/test/project",
			Artifacts: map[string]Artifact{
				"test": {Path: "/test/path"},
			},
		}

		stage := ConventionTerraformInitV2(exec)
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// State should be preserved
		assert.Equal(t, state.ProjectDir, stageResult.State.ProjectDir)
		assert.Equal(t, state.Artifacts, stageResult.State.Artifacts)
	})
}

// TestConventionTerraformPlanV2 tests the event-based Terraform plan stage
func TestConventionTerraformPlanV2(t *testing.T) {
	t.Run("plans infrastructure successfully without namespace", func(t *testing.T) {
		called := false
		exec := TerraformExecutor{
			PlanWithVars: func(ctx context.Context, dir string, vars map[string]string) (bool, error) {
				called = true
				assert.Nil(t, vars, "Vars should be nil without namespace")
				return true, nil
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformPlanV2(exec, "")
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))
		assert.True(t, called, "PlanWithVars should be called")

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Verify events
		assert.NotEmpty(t, stageResult.Events)
		assert.Contains(t, stageResult.Events[0].Message, "Planning infrastructure")

		// Should have changes detected event
		hasChangesEvent := false
		for _, event := range stageResult.Events {
			if event.Message == "[terraform] Changes detected" {
				hasChangesEvent = true
				break
			}
		}
		assert.True(t, hasChangesEvent, "Should have changes detected event")
	})

	t.Run("plans infrastructure with namespace", func(t *testing.T) {
		exec := TerraformExecutor{
			PlanWithVars: func(ctx context.Context, dir string, vars map[string]string) (bool, error) {
				assert.NotNil(t, vars, "Vars should not be nil with namespace")
				assert.Equal(t, "staging-", vars["namespace"])
				return true, nil
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformPlanV2(exec, "staging")
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Should have namespace event
		hasNamespaceEvent := false
		for _, event := range stageResult.Events {
			if event.Message == "Deploying to namespace: staging" {
				hasNamespaceEvent = true
				break
			}
		}
		assert.True(t, hasNamespaceEvent, "Should have namespace event")
	})

	t.Run("returns error when plan fails", func(t *testing.T) {
		exec := TerraformExecutor{
			PlanWithVars: func(ctx context.Context, dir string, vars map[string]string) (bool, error) {
				return false, fmt.Errorf("plan failed")
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformPlanV2(exec, "")
		result := stage(context.Background(), state)

		require.True(t, E.IsLeft(result))

		err := E.Fold(
			func(e error) error { return e },
			func(r StageResult) error { return nil },
		)(result)

		assert.Contains(t, err.Error(), "terraform plan failed")
	})

	t.Run("handles no changes detected", func(t *testing.T) {
		exec := TerraformExecutor{
			PlanWithVars: func(ctx context.Context, dir string, vars map[string]string) (bool, error) {
				return false, nil // No changes
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformPlanV2(exec, "")
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Should have no changes event
		hasNoChangesEvent := false
		for _, event := range stageResult.Events {
			if event.Message == "[terraform] No changes detected" {
				hasNoChangesEvent = true
				break
			}
		}
		assert.True(t, hasNoChangesEvent, "Should have no changes event")
	})
}

// TestConventionTerraformApplyV2 tests the event-based Terraform apply stage
func TestConventionTerraformApplyV2(t *testing.T) {
	t.Run("executes apply successfully with auto-approve", func(t *testing.T) {
		called := false
		exec := TerraformExecutor{
			Apply: func(ctx context.Context, dir string) error {
				called = true
				return nil
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformApplyV2(exec, true) // auto-approve
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))
		assert.True(t, called, "Apply should be called")

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Verify events
		assert.NotEmpty(t, stageResult.Events)

		// Should have applying event
		hasApplyingEvent := false
		for _, event := range stageResult.Events {
			if event.Message == "==> Applying infrastructure changes..." {
				hasApplyingEvent = true
				break
			}
		}
		assert.True(t, hasApplyingEvent, "Should have applying event")

		// Should have success event
		hasSuccessEvent := false
		for _, event := range stageResult.Events {
			if event.Level == EventLevelSuccess && event.Message == "[terraform] Applied successfully" {
				hasSuccessEvent = true
				break
			}
		}
		assert.True(t, hasSuccessEvent, "Should have success event")
	})

	t.Run("returns error when apply fails", func(t *testing.T) {
		exec := TerraformExecutor{
			Apply: func(ctx context.Context, dir string) error {
				return fmt.Errorf("apply failed")
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformApplyV2(exec, true)
		result := stage(context.Background(), state)

		require.True(t, E.IsLeft(result))

		err := E.Fold(
			func(e error) error { return e },
			func(r StageResult) error { return nil },
		)(result)

		assert.Contains(t, err.Error(), "terraform apply failed")
	})

	t.Run("preserves state on success", func(t *testing.T) {
		exec := TerraformExecutor{
			Apply: func(ctx context.Context, dir string) error {
				return nil
			},
		}

		originalArtifacts := map[string]Artifact{
			"test": {Path: "/test/path"},
		}

		state := State{
			ProjectDir: "/test/project",
			Artifacts:  originalArtifacts,
		}

		stage := ConventionTerraformApplyV2(exec, true)
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		assert.Equal(t, originalArtifacts, stageResult.State.Artifacts)
	})
}

// TestConventionTerraformOutputsV2 tests the event-based Terraform outputs stage
func TestConventionTerraformOutputsV2(t *testing.T) {
	t.Run("captures outputs successfully", func(t *testing.T) {
		expectedOutputs := map[string]interface{}{
			"api_url":      "https://example.com/api",
			"function_arn": "arn:aws:lambda:us-east-1:123456789:function:test",
		}

		exec := TerraformExecutor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				return expectedOutputs, nil
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformOutputsV2(exec)
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Verify outputs in state
		assert.Equal(t, expectedOutputs, stageResult.State.Outputs)

		// Verify events
		hasOutputEvent := false
		for _, event := range stageResult.Events {
			if event.Message == "Captured 2 output(s)" {
				hasOutputEvent = true
				break
			}
		}
		assert.True(t, hasOutputEvent, "Should have output count event")
	})

	t.Run("handles empty outputs", func(t *testing.T) {
		exec := TerraformExecutor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				return map[string]interface{}{}, nil
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformOutputsV2(exec)
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Empty outputs should still be set
		assert.NotNil(t, stageResult.State.Outputs)
		assert.Empty(t, stageResult.State.Outputs)

		// No output count event for empty outputs
		for _, event := range stageResult.Events {
			assert.NotContains(t, event.Message, "Captured")
		}
	})

	t.Run("returns warning event when output retrieval fails", func(t *testing.T) {
		exec := TerraformExecutor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				return nil, fmt.Errorf("output failed")
			},
		}

		state := State{ProjectDir: "/test/project"}
		stage := ConventionTerraformOutputsV2(exec)
		result := stage(context.Background(), state)

		// Should succeed despite output failure (non-fatal)
		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Should have warning event
		hasWarningEvent := false
		for _, event := range stageResult.Events {
			if event.Level == EventLevelWarning {
				assert.Contains(t, event.Message, "Failed to retrieve outputs")
				hasWarningEvent = true
				break
			}
		}
		assert.True(t, hasWarningEvent, "Should have warning event")
	})

	t.Run("preserves existing state", func(t *testing.T) {
		exec := TerraformExecutor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				return map[string]interface{}{"key": "value"}, nil
			},
		}

		originalArtifacts := map[string]Artifact{
			"test": {Path: "/test/path"},
		}

		state := State{
			ProjectDir: "/test/project",
			Artifacts:  originalArtifacts,
			Config:     "test-config",
		}

		stage := ConventionTerraformOutputsV2(exec)
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Other state fields should be preserved
		assert.Equal(t, state.ProjectDir, stageResult.State.ProjectDir)
		assert.Equal(t, originalArtifacts, stageResult.State.Artifacts)
		assert.Equal(t, "test-config", stageResult.State.Config)
	})
}

// TestTerraformPipelineV2Integration tests full Terraform pipeline with events
func TestTerraformPipelineV2Integration(t *testing.T) {
	t.Run("full terraform deployment pipeline with event collection", func(t *testing.T) {
		initCalled := false
		planCalled := false
		applyCalled := false
		outputCalled := false

		exec := TerraformExecutor{
			Init: func(ctx context.Context, dir string) error {
				initCalled = true
				return nil
			},
			PlanWithVars: func(ctx context.Context, dir string, vars map[string]string) (bool, error) {
				planCalled = true
				return true, nil
			},
			Apply: func(ctx context.Context, dir string) error {
				applyCalled = true
				return nil
			},
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				outputCalled = true
				return map[string]interface{}{"url": "https://example.com"}, nil
			},
		}

		pipeline := NewEventPipeline(
			ConventionTerraformInitV2(exec),
			ConventionTerraformPlanV2(exec, ""),
			ConventionTerraformApplyV2(exec, true),
			ConventionTerraformOutputsV2(exec),
		)

		initialState := State{
			ProjectDir: "/test/project",
			Artifacts:  make(map[string]Artifact),
		}

		result := RunWithEvents(pipeline, context.Background(), initialState)

		require.True(t, E.IsRight(result))
		assert.True(t, initCalled, "Init should be called")
		assert.True(t, planCalled, "Plan should be called")
		assert.True(t, applyCalled, "Apply should be called")
		assert.True(t, outputCalled, "Output should be called")

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Should have collected events from all stages
		assert.NotEmpty(t, stageResult.Events)

		// Count events from each stage
		eventCounts := map[string]int{
			"init":   0,
			"plan":   0,
			"apply":  0,
			"output": 0,
		}

		for _, event := range stageResult.Events {
			if event.Message == "[terraform] Initialized" {
				eventCounts["init"]++
			}
			if event.Message == "[terraform] Changes detected" {
				eventCounts["plan"]++
			}
			if event.Message == "[terraform] Applied successfully" {
				eventCounts["apply"]++
			}
			if event.Message == "Captured 1 output(s)" {
				eventCounts["output"]++
			}
		}

		assert.Greater(t, eventCounts["init"], 0, "Should have init events")
		assert.Greater(t, eventCounts["plan"], 0, "Should have plan events")
		assert.Greater(t, eventCounts["apply"], 0, "Should have apply events")
		assert.Greater(t, eventCounts["output"], 0, "Should have output events")

		// Verify final state has outputs
		assert.NotEmpty(t, stageResult.State.Outputs)
	})

	t.Run("pipeline stops on init failure", func(t *testing.T) {
		planCalled := false

		exec := TerraformExecutor{
			Init: func(ctx context.Context, dir string) error {
				return fmt.Errorf("init failed")
			},
			PlanWithVars: func(ctx context.Context, dir string, vars map[string]string) (bool, error) {
				planCalled = true
				return false, nil
			},
		}

		pipeline := NewEventPipeline(
			ConventionTerraformInitV2(exec),
			ConventionTerraformPlanV2(exec, ""),
		)

		initialState := State{ProjectDir: "/test/project"}

		result := RunWithEvents(pipeline, context.Background(), initialState)

		require.True(t, E.IsLeft(result))
		assert.False(t, planCalled, "Plan should not be called after init fails")
	})
}
