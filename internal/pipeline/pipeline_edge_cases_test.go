package pipeline

import (
	"context"
	"fmt"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunEdgeCases tests edge cases in pipeline execution
func TestRunEdgeCases(t *testing.T) {
	t.Run("handles state transformation through multiple stages", func(t *testing.T) {
		stage1 := func(ctx context.Context, s State) E.Either[error, State] {
			s.ProjectDir = "/stage1"
			return E.Right[error](s)
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, State] {
			s.ProjectDir = s.ProjectDir + "/stage2"
			return E.Right[error](s)
		}

		stage3 := func(ctx context.Context, s State) E.Either[error, State] {
			s.ProjectDir = s.ProjectDir + "/stage3"
			return E.Right[error](s)
		}

		pipeline := New(stage1, stage2, stage3)
		result := Run(pipeline, context.Background(), State{})

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Equal(t, "/stage1/stage2/stage3", finalState.ProjectDir)
	})

	t.Run("preserves artifacts across stages", func(t *testing.T) {
		initialState := State{
			ProjectDir: "/test",
			Artifacts: map[string]Artifact{
				"api": {Path: "/api.zip", Checksum: "abc123", Size: 1024},
			},
		}

		stage1 := func(ctx context.Context, s State) E.Either[error, State] {
			// Add new artifact
			s.Artifacts["worker"] = Artifact{Path: "/worker.zip", Checksum: "def456", Size: 2048}
			return E.Right[error](s)
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, State] {
			// Verify both artifacts exist
			if len(s.Artifacts) != 2 {
				return E.Left[State](fmt.Errorf("expected 2 artifacts, got %d", len(s.Artifacts)))
			}
			return E.Right[error](s)
		}

		pipeline := New(stage1, stage2)
		result := Run(pipeline, context.Background(), initialState)

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Len(t, finalState.Artifacts, 2)
		assert.Contains(t, finalState.Artifacts, "api")
		assert.Contains(t, finalState.Artifacts, "worker")
	})

	t.Run("handles config transformation", func(t *testing.T) {
		stage1 := func(ctx context.Context, s State) E.Either[error, State] {
			s.Config = map[string]string{"key": "value"}
			return E.Right[error](s)
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, State] {
			config, ok := s.Config.(map[string]string)
			if !ok {
				return E.Left[State](fmt.Errorf("invalid config type"))
			}
			config["key2"] = "value2"
			s.Config = config
			return E.Right[error](s)
		}

		pipeline := New(stage1, stage2)
		result := Run(pipeline, context.Background(), State{})

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		config := finalState.Config.(map[string]string)
		assert.Len(t, config, 2)
		assert.Equal(t, "value", config["key"])
		assert.Equal(t, "value2", config["key2"])
	})

	t.Run("error in middle stage stops execution", func(t *testing.T) {
		var stage1Called, stage2Called, stage3Called bool

		stage1 := func(ctx context.Context, s State) E.Either[error, State] {
			stage1Called = true
			return E.Right[error](s)
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, State] {
			stage2Called = true
			return E.Left[State](fmt.Errorf("stage 2 error"))
		}

		stage3 := func(ctx context.Context, s State) E.Either[error, State] {
			stage3Called = true
			return E.Right[error](s)
		}

		pipeline := New(stage1, stage2, stage3)
		result := Run(pipeline, context.Background(), State{})

		assert.True(t, E.IsLeft(result))
		assert.True(t, stage1Called)
		assert.True(t, stage2Called)
		assert.False(t, stage3Called, "Stage 3 should not execute after error")
	})

	t.Run("preserves outputs across stages", func(t *testing.T) {
		stage1 := func(ctx context.Context, s State) E.Either[error, State] {
			if s.Outputs == nil {
				s.Outputs = make(map[string]interface{})
			}
			s.Outputs["stage1"] = "output1"
			return E.Right[error](s)
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, State] {
			s.Outputs["stage2"] = "output2"
			return E.Right[error](s)
		}

		pipeline := New(stage1, stage2)
		result := Run(pipeline, context.Background(), State{})

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Len(t, finalState.Outputs, 2)
		assert.Equal(t, "output1", finalState.Outputs["stage1"])
		assert.Equal(t, "output2", finalState.Outputs["stage2"])
	})
}

// TestParallelEdgeCases tests edge cases in parallel execution
func TestParallelEdgeCases(t *testing.T) {
	t.Run("parallel stage preserves state modifications", func(t *testing.T) {
		stage1 := func(ctx context.Context, s State) E.Either[error, State] {
			if s.Artifacts == nil {
				s.Artifacts = make(map[string]Artifact)
			}
			s.Artifacts["api"] = Artifact{Path: "/api.zip"}
			return E.Right[error](s)
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, State] {
			if s.Outputs == nil {
				s.Outputs = make(map[string]interface{})
			}
			s.Outputs["url"] = "https://example.com"
			return E.Right[error](s)
		}

		parallelStage := Parallel(stage1, stage2)
		result := parallelStage(context.Background(), State{})

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		// Both modifications should be present
		assert.Len(t, finalState.Artifacts, 1)
		assert.Len(t, finalState.Outputs, 1)
	})

	t.Run("parallel stops on first error", func(t *testing.T) {
		var stage1Called, stage2Called bool

		stage1 := func(ctx context.Context, s State) E.Either[error, State] {
			stage1Called = true
			return E.Left[State](fmt.Errorf("error in stage 1"))
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, State] {
			stage2Called = true
			return E.Right[error](s)
		}

		parallelStage := Parallel(stage1, stage2)
		result := parallelStage(context.Background(), State{})

		assert.True(t, E.IsLeft(result))
		assert.True(t, stage1Called)
		// Since we run sequentially now, stage2 won't be called
		assert.False(t, stage2Called)
	})

	t.Run("parallel with single stage works", func(t *testing.T) {
		stage := func(ctx context.Context, s State) E.Either[error, State] {
			s.ProjectDir = "/test"
			return E.Right[error](s)
		}

		parallelStage := Parallel(stage)
		result := parallelStage(context.Background(), State{})

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Equal(t, "/test", finalState.ProjectDir)
	})
}

// TestChainEdgeCases tests edge cases in pipeline chaining
func TestChainEdgeCases(t *testing.T) {
	t.Run("chain preserves stage order", func(t *testing.T) {
		var order []int

		makeStage := func(id int) Stage {
			return func(ctx context.Context, s State) E.Either[error, State] {
				order = append(order, id)
				return E.Right[error](s)
			}
		}

		p1 := New(makeStage(1), makeStage(2))
		p2 := New(makeStage(3), makeStage(4))
		p3 := New(makeStage(5))

		chained := Chain(p1, p2, p3)
		result := Run(chained, context.Background(), State{})

		require.True(t, E.IsRight(result))
		assert.Equal(t, []int{1, 2, 3, 4, 5}, order)
	})

	t.Run("chain with single pipeline", func(t *testing.T) {
		stage := func(ctx context.Context, s State) E.Either[error, State] {
			s.ProjectDir = "/test"
			return E.Right[error](s)
		}

		p := New(stage)
		chained := Chain(p)
		result := Run(chained, context.Background(), State{})

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Equal(t, "/test", finalState.ProjectDir)
	})

	t.Run("chain handles error in any pipeline", func(t *testing.T) {
		successStage := func(ctx context.Context, s State) E.Either[error, State] {
			return E.Right[error](s)
		}

		errorStage := func(ctx context.Context, s State) E.Either[error, State] {
			return E.Left[State](fmt.Errorf("error in pipeline 2"))
		}

		p1 := New(successStage)
		p2 := New(errorStage)
		p3 := New(successStage)

		chained := Chain(p1, p2, p3)
		result := Run(chained, context.Background(), State{})

		assert.True(t, E.IsLeft(result))
	})
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	t.Run("stage respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stage := func(ctx context.Context, s State) E.Either[error, State] {
			select {
			case <-ctx.Done():
				return E.Left[State](ctx.Err())
			default:
				return E.Right[error](s)
			}
		}

		// Cancel before execution
		cancel()

		pipeline := New(stage)
		result := Run(pipeline, ctx, State{})

		assert.True(t, E.IsLeft(result))
	})

	t.Run("pipeline continues if context not cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		stage := func(ctx context.Context, s State) E.Either[error, State] {
			select {
			case <-ctx.Done():
				return E.Left[State](ctx.Err())
			default:
				s.ProjectDir = "/test"
				return E.Right[error](s)
			}
		}

		pipeline := New(stage)
		result := Run(pipeline, ctx, State{})

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Equal(t, "/test", finalState.ProjectDir)
	})
}

// TestStateImmutability tests state transformation behavior
func TestStateImmutability(t *testing.T) {
	t.Run("state fields are passed by value but maps by reference", func(t *testing.T) {
		original := State{
			ProjectDir: "/original",
			Artifacts: map[string]Artifact{
				"api": {Path: "/original.zip"},
			},
		}

		stage := func(ctx context.Context, s State) E.Either[error, State] {
			// Modifying ProjectDir (string) doesn't affect original
			s.ProjectDir = "/modified"
			// Modifying map DOES affect original (maps are reference types)
			s.Artifacts["worker"] = Artifact{Path: "/worker.zip"}
			return E.Right[error](s)
		}

		pipeline := New(stage)
		_ = Run(pipeline, context.Background(), original)

		// ProjectDir is unchanged (strings are value types)
		assert.Equal(t, "/original", original.ProjectDir)
		// But Artifacts map IS modified (maps are reference types in Go)
		// This is expected Go behavior - we're not doing deep copies
		assert.Len(t, original.Artifacts, 2)
		assert.Contains(t, original.Artifacts, "worker")
	})

	t.Run("each stage gets a copy of state", func(t *testing.T) {
		var state1, state2 State

		stage1 := func(ctx context.Context, s State) E.Either[error, State] {
			state1 = s
			s.ProjectDir = "/stage1"
			return E.Right[error](s)
		}

		stage2 := func(ctx context.Context, s State) E.Either[error, State] {
			state2 = s
			return E.Right[error](s)
		}

		initial := State{ProjectDir: "/initial"}
		pipeline := New(stage1, stage2)
		_ = Run(pipeline, context.Background(), initial)

		assert.Equal(t, "/initial", state1.ProjectDir)
		assert.Equal(t, "/stage1", state2.ProjectDir)
	})
}

// TestArtifactManipulation tests artifact map operations
func TestArtifactManipulation(t *testing.T) {
	t.Run("nil artifacts map is initialized", func(t *testing.T) {
		stage := func(ctx context.Context, s State) E.Either[error, State] {
			if s.Artifacts == nil {
				s.Artifacts = make(map[string]Artifact)
			}
			s.Artifacts["test"] = Artifact{Path: "/test.zip"}
			return E.Right[error](s)
		}

		pipeline := New(stage)
		result := Run(pipeline, context.Background(), State{})

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.NotNil(t, finalState.Artifacts)
		assert.Len(t, finalState.Artifacts, 1)
	})

	t.Run("artifacts can be updated", func(t *testing.T) {
		initial := State{
			Artifacts: map[string]Artifact{
				"api": {Path: "/old.zip", Checksum: "old", Size: 100},
			},
		}

		stage := func(ctx context.Context, s State) E.Either[error, State] {
			s.Artifacts["api"] = Artifact{Path: "/new.zip", Checksum: "new", Size: 200}
			return E.Right[error](s)
		}

		pipeline := New(stage)
		result := Run(pipeline, context.Background(), initial)

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Equal(t, "/new.zip", finalState.Artifacts["api"].Path)
		assert.Equal(t, "new", finalState.Artifacts["api"].Checksum)
		assert.Equal(t, int64(200), finalState.Artifacts["api"].Size)
	})

	t.Run("artifacts can be deleted", func(t *testing.T) {
		initial := State{
			Artifacts: map[string]Artifact{
				"api":    {Path: "/api.zip"},
				"worker": {Path: "/worker.zip"},
			},
		}

		stage := func(ctx context.Context, s State) E.Either[error, State] {
			delete(s.Artifacts, "worker")
			return E.Right[error](s)
		}

		pipeline := New(stage)
		result := Run(pipeline, context.Background(), initial)

		require.True(t, E.IsRight(result))
		finalState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Len(t, finalState.Artifacts, 1)
		assert.Contains(t, finalState.Artifacts, "api")
		assert.NotContains(t, finalState.Artifacts, "worker")
	})
}
