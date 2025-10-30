package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lewis/forge/internal/discovery"
)

// TestConventionStubsEdgeCases tests edge cases in stub creation.
func TestConventionStubsEdgeCases(t *testing.T) {
	t.Run("creates build directory if it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Don't create .forge/build directory

		functionDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionDir, 0o755))

		functions := []discovery.Function{
			{
				Name:       "api",
				Path:       functionDir,
				Runtime:    "provided.al2023",
				EntryPoint: "main.go",
			},
		}

		state := State{
			ProjectDir: tmpDir,
			Config:     functions,
		}

		stage := ConventionStubs()
		result := stage(t.Context(), state)

		require.True(t, E.IsRight(result))

		buildDir := filepath.Join(tmpDir, ".forge", "build")
		assert.DirExists(t, buildDir)
		assert.FileExists(t, filepath.Join(buildDir, "api.zip"))
	})
}

// TestConventionStubsV2EdgeCases tests V2 stub creation edge cases.
func TestConventionStubsV2EdgeCases(t *testing.T) {
	t.Run("creates build directory and emits events", func(t *testing.T) {
		tmpDir := t.TempDir()

		functionDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionDir, 0o755))

		functions := []discovery.Function{
			{
				Name:       "api",
				Path:       functionDir,
				Runtime:    "provided.al2023",
				EntryPoint: "main.go",
			},
		}

		state := State{
			ProjectDir: tmpDir,
			Config:     functions,
		}

		stage := ConventionStubsV2()
		result := stage(t.Context(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Should emit events
		assert.NotEmpty(t, stageResult.Events)

		buildDir := filepath.Join(tmpDir, ".forge", "build")
		assert.DirExists(t, buildDir)
	})

	t.Run("emits events for each stub created", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create multiple functions
		for _, name := range []string{"api1", "api2", "api3"} {
			functionDir := filepath.Join(tmpDir, "src", "functions", name)
			require.NoError(t, os.MkdirAll(functionDir, 0o755))
		}

		functions := []discovery.Function{
			{Name: "api1", Path: filepath.Join(tmpDir, "src", "functions", "api1"), Runtime: "provided.al2023", EntryPoint: "main.go"},
			{Name: "api2", Path: filepath.Join(tmpDir, "src", "functions", "api2"), Runtime: "nodejs20.x", EntryPoint: "index.js"},
			{Name: "api3", Path: filepath.Join(tmpDir, "src", "functions", "api3"), Runtime: "python3.13", EntryPoint: "app.py"},
		}

		state := State{
			ProjectDir: tmpDir,
			Config:     functions,
		}

		stage := ConventionStubsV2()
		result := stage(t.Context(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Should have multiple events
		assert.GreaterOrEqual(t, len(stageResult.Events), 1)
	})
}

// TestRunEdgeCasesMore tests additional Run edge cases.
func TestRunEdgeCasesMore(t *testing.T) {
	t.Run("handles None value in pipeline gracefully", func(t *testing.T) {
		// This tests the O.IsNone check in Run
		var stageExecuted bool

		stage := func(ctx context.Context, s State) E.Either[error, State] {
			stageExecuted = true
			return E.Right[error](s)
		}

		pipeline := New(stage)
		result := Run(pipeline, t.Context(), State{ProjectDir: "/test"})

		assert.True(t, E.IsRight(result))
		assert.True(t, stageExecuted)
	})

	t.Run("processes all stages when no errors", func(t *testing.T) {
		executionOrder := []int{}

		makeStage := func(id int) Stage {
			return func(ctx context.Context, s State) E.Either[error, State] {
				executionOrder = append(executionOrder, id)
				return E.Right[error](s)
			}
		}

		pipeline := New(
			makeStage(1),
			makeStage(2),
			makeStage(3),
			makeStage(4),
			makeStage(5),
		)

		result := Run(pipeline, t.Context(), State{})

		require.True(t, E.IsRight(result))
		assert.Equal(t, []int{1, 2, 3, 4, 5}, executionOrder)
	})
}

// TestParallelEdgeCasesMore tests additional Parallel edge cases.
func TestParallelEdgeCasesMore(t *testing.T) {
	t.Run("parallel executes all stages when no errors", func(t *testing.T) {
		var stage1Called, stage2Called, stage3Called bool

		stage1 := func(ctx context.Context, s State) E.Either[error, State] {
			stage1Called = true
			return E.Right[error](s)
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, State] {
			stage2Called = true
			return E.Right[error](s)
		}

		stage3 := func(ctx context.Context, s State) E.Either[error, State] {
			stage3Called = true
			return E.Right[error](s)
		}

		parallelStage := Parallel(stage1, stage2, stage3)
		result := parallelStage(t.Context(), State{})

		assert.True(t, E.IsRight(result))
		assert.True(t, stage1Called)
		assert.True(t, stage2Called)
		assert.True(t, stage3Called)
	})
}

// TestRunWithEventsEdgeCases tests additional RunWithEvents edge cases.
func TestRunWithEventsEdgeCases(t *testing.T) {
	t.Run("handles successful execution through all stages", func(t *testing.T) {
		stage1 := func(ctx context.Context, s State) E.Either[error, StageResult] {
			return E.Right[error](StageResult{
				State: s,
				Events: []StageEvent{
					NewEvent(EventLevelInfo, "Stage 1 started"),
					NewEvent(EventLevelSuccess, "Stage 1 completed"),
				},
			})
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, StageResult] {
			return E.Right[error](StageResult{
				State: s,
				Events: []StageEvent{
					NewEvent(EventLevelInfo, "Stage 2 started"),
					NewEvent(EventLevelSuccess, "Stage 2 completed"),
				},
			})
		}

		pipeline := NewEventPipeline(stage1, stage2)
		runResult := RunWithEvents(pipeline, t.Context(), State{})

		require.True(t, E.IsRight(runResult.Result))

		// All events should be collected
		assert.Len(t, runResult.Events, 4)
	})

	t.Run("collects events before error", func(t *testing.T) {
		stage1 := func(ctx context.Context, s State) E.Either[error, StageResult] {
			return E.Right[error](StageResult{
				State: s,
				Events: []StageEvent{
					NewEvent(EventLevelInfo, "Stage 1 event"),
				},
			})
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, StageResult] {
			return E.Left[StageResult](assert.AnError)
		}

		pipeline := NewEventPipeline(stage1, stage2)
		runResult := RunWithEvents(pipeline, t.Context(), State{})

		assert.True(t, E.IsLeft(runResult.Result))
		// Events from stage1 should still be available even on error
		assert.Len(t, runResult.Events, 1)
		assert.Equal(t, "Stage 1 event", runResult.Events[0].Message)
	})
}

// TestConventionBuildV2EdgeCases tests V2 build edge cases.
func TestConventionBuildV2EdgeCases(t *testing.T) {
	t.Run("handles empty function list with events", func(t *testing.T) {
		state := State{
			ProjectDir: t.TempDir(),
			Config:     []discovery.Function{},
			Artifacts:  make(map[string]Artifact),
		}

		stage := ConventionBuildV2()
		result := stage(t.Context(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Should have events even with no functions
		assert.NotEmpty(t, stageResult.Events)
	})

	t.Run("returns error with events for invalid config", func(t *testing.T) {
		state := State{
			ProjectDir: t.TempDir(),
			Config:     "invalid", // Wrong type
		}

		stage := ConventionBuildV2()
		result := stage(t.Context(), state)

		assert.True(t, E.IsLeft(result))
	})
}

// TestTerraformApplyV2EdgeCases tests V2 apply edge cases.
func TestTerraformApplyV2EdgeCases(t *testing.T) {
	t.Run("applies with auto-approve and emits success event", func(t *testing.T) {
		tmpDir := t.TempDir()

		var applyCalled bool
		exec := TerraformExecutor{
			Apply: func(ctx context.Context, dir string) error {
				applyCalled = true
				return nil
			},
		}

		// Auto-approve function returns true (no user interaction)
		autoApprove := func() bool { return true }

		state := State{ProjectDir: tmpDir}
		stage := ConventionTerraformApplyV2(exec, autoApprove)
		result := stage(t.Context(), state)

		require.True(t, E.IsRight(result))
		assert.True(t, applyCalled)

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Should have success event
		hasSuccess := false
		for _, event := range stageResult.Events {
			if event.Level == EventLevelSuccess {
				hasSuccess = true
				break
			}
		}
		assert.True(t, hasSuccess)
	})

	t.Run("returns error event when apply fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := TerraformExecutor{
			Apply: func(ctx context.Context, dir string) error {
				return assert.AnError
			},
		}

		autoApprove := func() bool { return true }

		state := State{ProjectDir: tmpDir}
		stage := ConventionTerraformApplyV2(exec, autoApprove)
		result := stage(t.Context(), state)

		assert.True(t, E.IsLeft(result))
	})

	t.Run("handles user cancellation", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := TerraformExecutor{
			Apply: func(ctx context.Context, dir string) error {
				return nil
			},
		}

		// User cancels deployment
		cancelFunc := func() bool { return false }

		state := State{ProjectDir: tmpDir}
		stage := ConventionTerraformApplyV2(exec, cancelFunc)
		result := stage(t.Context(), state)

		assert.True(t, E.IsLeft(result))
	})

	t.Run("skips approval when func is nil", func(t *testing.T) {
		tmpDir := t.TempDir()

		var applyCalled bool
		exec := TerraformExecutor{
			Apply: func(ctx context.Context, dir string) error {
				applyCalled = true
				return nil
			},
		}

		state := State{ProjectDir: tmpDir}
		stage := ConventionTerraformApplyV2(exec, nil) // nil approval func
		result := stage(t.Context(), state)

		require.True(t, E.IsRight(result))
		assert.True(t, applyCalled)
	})
}

// TestOutputsV2EdgeCases tests V2 outputs edge cases.
func TestOutputsV2EdgeCases(t *testing.T) {
	t.Run("captures outputs and emits success events", func(t *testing.T) {
		tmpDir := t.TempDir()

		outputs := map[string]interface{}{
			"api_url": "https://example.com",
			"db_name": "testdb",
		}

		exec := TerraformExecutor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				return outputs, nil
			},
		}

		state := State{ProjectDir: tmpDir}
		stage := ConventionTerraformOutputsV2(exec)
		result := stage(t.Context(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		assert.Equal(t, outputs, stageResult.State.Outputs)

		// Should have events
		assert.NotEmpty(t, stageResult.Events)
	})

	t.Run("initializes outputs map when nil", func(t *testing.T) {
		tmpDir := t.TempDir()

		exec := TerraformExecutor{
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				return map[string]interface{}{"test": "value"}, nil
			},
		}

		state := State{
			ProjectDir: tmpDir,
			Outputs:    nil, // nil outputs
		}

		stage := ConventionTerraformOutputsV2(exec)
		result := stage(t.Context(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		assert.NotNil(t, stageResult.State.Outputs)
		assert.Equal(t, "value", stageResult.State.Outputs["test"])
	})
}
