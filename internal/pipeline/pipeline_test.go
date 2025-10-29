package pipeline

import (
	"context"
	"fmt"
	"testing"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
	"github.com/lewis/forge/internal/stack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipelineCreation tests creating pipelines
func TestPipelineCreation(t *testing.T) {
	t.Run("New creates pipeline with stages", func(t *testing.T) {
		stage1 := func(ctx context.Context, s State) E.Either[error, State] {
			return E.Right[error](s)
		}
		stage2 := func(ctx context.Context, s State) E.Either[error, State] {
			return E.Right[error](s)
		}

		pipeline := New(stage1, stage2)

		assert.NotNil(t, pipeline)
		assert.Len(t, pipeline.stages, 2)
	})

	t.Run("Chain combines multiple pipelines", func(t *testing.T) {
		pipeline1 := New(
			func(ctx context.Context, s State) E.Either[error, State] {
				return E.Right[error](s)
			},
		)

		pipeline2 := New(
			func(ctx context.Context, s State) E.Either[error, State] {
				return E.Right[error](s)
			},
		)

		combined := Chain(pipeline1, pipeline2)

		assert.Len(t, combined.stages, 2, "Should combine stages from both pipelines")
	})
}

// TestPipelineExecution tests running pipelines
func TestPipelineExecution(t *testing.T) {
	t.Run("executes all stages in order", func(t *testing.T) {
		var execution []string

		pipeline := New(
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage1")
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage2")
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage3")
				return E.Right[error](s)
			},
		)

		result := pipeline.Run(context.Background(), State{})

		assert.True(t, E.IsRight(result), "Pipeline should succeed")
		assert.Equal(t, []string{"stage1", "stage2", "stage3"}, execution)
	})

	t.Run("stops on first error", func(t *testing.T) {
		var execution []string

		pipeline := New(
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage1")
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage2-error")
				return E.Left[State](fmt.Errorf("stage 2 failed"))
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage3-should-not-run")
				return E.Right[error](s)
			},
		)

		result := pipeline.Run(context.Background(), State{})

		assert.True(t, E.IsLeft(result), "Pipeline should fail")
		assert.Equal(t, []string{"stage1", "stage2-error"}, execution, "Should stop after error")
	})

	t.Run("passes state through stages", func(t *testing.T) {
		pipeline := New(
			func(ctx context.Context, s State) E.Either[error, State] {
				s.ProjectDir = "/project"
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				s.Stacks = []*stack.Stack{{Name: "api"}}
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				s.Artifacts = make(map[string]Artifact)
				s.Artifacts["api"] = Artifact{Path: "/build/api"}
				return E.Right[error](s)
			},
		)

		result := pipeline.Run(context.Background(), State{})

		require.True(t, E.IsRight(result))

		// Extract final state
		finalState := E.ToOption(result)
		state := O.GetOrElse(func() State { return State{} })(finalState)

		assert.Equal(t, "/project", state.ProjectDir)
		assert.Len(t, state.Stacks, 1)
		assert.Len(t, state.Artifacts, 1)
	})
}

// TestStageComposition tests composing individual stages
func TestStageComposition(t *testing.T) {
	t.Run("can compose stages manually", func(t *testing.T) {
		addProjectDir := func(ctx context.Context, s State) E.Either[error, State] {
			s.ProjectDir = "/test"
			return E.Right[error](s)
		}

		addStacks := func(ctx context.Context, s State) E.Either[error, State] {
			s.Stacks = []*stack.Stack{{Name: "api"}}
			return E.Right[error](s)
		}

		// Compose manually
		ctx := context.Background()
		initialState := State{}

		result1 := addProjectDir(ctx, initialState)
		if E.IsLeft(result1) {
			t.Fatal("Should not fail")
		}

		opt := E.ToOption(result1)
		state1 := O.GetOrElse(func() State { return State{} })(opt)
		result := addStacks(ctx, state1)

		assert.True(t, E.IsRight(result))
	})
}

// TestPipelineErrorHandling tests error scenarios
func TestPipelineErrorHandling(t *testing.T) {
	t.Run("propagates errors from any stage", func(t *testing.T) {
		testCases := []struct {
			name        string
			failStage   int
			totalStages int
		}{
			{"first stage fails", 0, 3},
			{"middle stage fails", 1, 3},
			{"last stage fails", 2, 3},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var stages []Stage
				for i := 0; i < tc.totalStages; i++ {
					stageNum := i
					stage := func(ctx context.Context, s State) E.Either[error, State] {
						if stageNum == tc.failStage {
							return E.Left[State](fmt.Errorf("stage %d failed", stageNum))
						}
						return E.Right[error](s)
					}
					stages = append(stages, stage)
				}

				pipeline := New(stages...)
				result := pipeline.Run(context.Background(), State{})

				assert.True(t, E.IsLeft(result), "Should fail when stage %d fails", tc.failStage)
			})
		}
	})

	t.Run("empty pipeline succeeds with initial state", func(t *testing.T) {
		pipeline := New()

		initialState := State{ProjectDir: "/test"}
		result := pipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsRight(result))

		// Extract final state
		opt := E.ToOption(result)
		finalState := O.GetOrElse(func() State { return State{} })(opt)
		assert.Equal(t, "/test", finalState.ProjectDir)
	})
}

// TestPipelineWithContext tests context propagation
func TestPipelineWithContext(t *testing.T) {
	t.Run("context is passed to all stages", func(t *testing.T) {
		type ctxKey string
		key := ctxKey("test")

		var receivedValues []string

		pipeline := New(
			func(ctx context.Context, s State) E.Either[error, State] {
				if val, ok := ctx.Value(key).(string); ok {
					receivedValues = append(receivedValues, val)
				}
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				if val, ok := ctx.Value(key).(string); ok {
					receivedValues = append(receivedValues, val)
				}
				return E.Right[error](s)
			},
		)

		ctx := context.WithValue(context.Background(), key, "test-value")
		Run(pipeline, ctx, State{})

		assert.Len(t, receivedValues, 2)
		assert.Equal(t, "test-value", receivedValues[0])
		assert.Equal(t, "test-value", receivedValues[1])
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		var stageRan bool
		pipeline := New(
			func(ctx context.Context, s State) E.Either[error, State] {
				if ctx.Err() != nil {
					return E.Left[State](ctx.Err())
				}
				stageRan = true
				return E.Right[error](s)
			},
		)

		result := Run(pipeline, ctx, State{})

		// Stage should detect cancellation
		assert.True(t, E.IsLeft(result) || !stageRan)
	})
}

// TestStateMutation tests that state is properly threaded through pipeline
func TestStateMutation(t *testing.T) {
	t.Run("each stage can modify state", func(t *testing.T) {
		pipeline := New(
			func(ctx context.Context, s State) E.Either[error, State] {
				s.ProjectDir = "/project"
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				// Should see modification from previous stage
				assert.Equal(t, "/project", s.ProjectDir)
				s.Stacks = []*stack.Stack{{Name: "api"}}
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				// Should see modifications from both previous stages
				assert.Equal(t, "/project", s.ProjectDir)
				assert.Len(t, s.Stacks, 1)
				s.Outputs = map[string]interface{}{"result": "success"}
				return E.Right[error](s)
			},
		)

		result := pipeline.Run(context.Background(), State{})
		assert.True(t, E.IsRight(result))
	})
}

// BenchmarkPipeline benchmarks pipeline execution
func BenchmarkPipeline(b *testing.B) {
	b.Run("3 stages", func(b *testing.B) {
		pipeline := New(
			func(ctx context.Context, s State) E.Either[error, State] {
				s.ProjectDir = "/test"
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				s.Stacks = []*stack.Stack{{Name: "api"}}
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				s.Outputs = make(map[string]interface{})
				return E.Right[error](s)
			},
		)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pipeline.Run(context.Background(), State{})
		}
	})

	b.Run("10 stages", func(b *testing.B) {
		var stages []Stage
		for i := 0; i < 10; i++ {
			stages = append(stages, func(ctx context.Context, s State) E.Either[error, State] {
				return E.Right[error](s)
			})
		}

		pipeline := New(stages...)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pipeline.Run(context.Background(), State{})
		}
	})
}

// TestParallel tests the Parallel stage composition function
func TestParallel(t *testing.T) {
	t.Run("executes all stages successfully", func(t *testing.T) {
		var execution []string

		stage := Parallel(
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage1")
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage2")
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage3")
				return E.Right[error](s)
			},
		)

		result := stage(context.Background(), State{})

		assert.True(t, E.IsRight(result), "Parallel should succeed")
		assert.Len(t, execution, 3, "All stages should execute")
	})

	t.Run("stops on first error", func(t *testing.T) {
		var execution []string

		stage := Parallel(
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage1")
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage2-error")
				return E.Left[State](fmt.Errorf("stage 2 failed"))
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				execution = append(execution, "stage3-should-not-run")
				return E.Right[error](s)
			},
		)

		result := stage(context.Background(), State{})

		assert.True(t, E.IsLeft(result), "Parallel should fail")
		assert.Equal(t, []string{"stage1", "stage2-error"}, execution, "Should stop after error")
	})

	t.Run("passes state through stages", func(t *testing.T) {
		stage := Parallel(
			func(ctx context.Context, s State) E.Either[error, State] {
				s.ProjectDir = "/project"
				return E.Right[error](s)
			},
			func(ctx context.Context, s State) E.Either[error, State] {
				assert.Equal(t, "/project", s.ProjectDir)
				s.Stacks = []*stack.Stack{{Name: "api"}}
				return E.Right[error](s)
			},
		)

		result := stage(context.Background(), State{})

		require.True(t, E.IsRight(result))
		finalState := E.ToOption(result)
		state := O.GetOrElse(func() State { return State{} })(finalState)

		assert.Equal(t, "/project", state.ProjectDir)
		assert.Len(t, state.Stacks, 1)
	})

	t.Run("works with empty stage list", func(t *testing.T) {
		stage := Parallel()
		initialState := State{ProjectDir: "/test"}
		result := stage(context.Background(), initialState)

		assert.True(t, E.IsRight(result))
		opt := E.ToOption(result)
		finalState := O.GetOrElse(func() State { return State{} })(opt)
		assert.Equal(t, "/test", finalState.ProjectDir)
	})
}
