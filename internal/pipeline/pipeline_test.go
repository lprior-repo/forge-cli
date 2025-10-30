package pipeline

import (
	"context"
	"fmt"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers - pure functions for testing.
func successStage(name string) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		if s.Outputs == nil {
			s.Outputs = make(map[string]interface{})
		}
		s.Outputs[name] = "executed"
		return E.Right[error](s)
	}
}

func errorStage(errMsg string) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		return E.Left[State](fmt.Errorf("%s", errMsg))
	}
}

func addArtifactStage(name, path string) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		if s.Artifacts == nil {
			s.Artifacts = make(map[string]Artifact)
		}
		s.Artifacts[name] = Artifact{
			Path:     path,
			Checksum: "abc123",
			Size:     1024,
		}
		return E.Right[error](s)
	}
}

func TestNew(t *testing.T) {
	t.Run("creates empty pipeline with no stages", func(t *testing.T) {
		// Arrange & Act
		p := New()

		// Assert
		assert.NotNil(t, p)
		assert.Empty(t, p.stages)
	})

	t.Run("creates pipeline with single stage", func(t *testing.T) {
		// Arrange
		stage := successStage("test")

		// Act
		p := New(stage)

		// Assert
		assert.NotNil(t, p)
		assert.Len(t, p.stages, 1)
	})

	t.Run("creates pipeline with multiple stages", func(t *testing.T) {
		// Arrange
		stage1 := successStage("stage1")
		stage2 := successStage("stage2")
		stage3 := successStage("stage3")

		// Act
		p := New(stage1, stage2, stage3)

		// Assert
		assert.NotNil(t, p)
		assert.Len(t, p.stages, 3)
	})
}

func TestRun(t *testing.T) {
	t.Run("executes empty pipeline successfully", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		p := New()
		initial := State{ProjectDir: "/test"}

		// Act
		result := Run(p, ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Pipeline should succeed with no stages")

		state := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Equal(t, "/test", state.ProjectDir)
	})

	t.Run("executes single stage successfully", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		p := New(successStage("build"))
		initial := State{ProjectDir: "/project"}

		// Act
		result := Run(p, ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Pipeline should succeed")

		state := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Equal(t, "executed", state.Outputs["build"])
	})

	t.Run("executes multiple stages in order", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		p := New(
			successStage("scan"),
			successStage("build"),
			successStage("deploy"),
		)
		initial := State{ProjectDir: "/project"}

		// Act
		result := Run(p, ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Pipeline should succeed")

		state := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Equal(t, "executed", state.Outputs["scan"])
		assert.Equal(t, "executed", state.Outputs["build"])
		assert.Equal(t, "executed", state.Outputs["deploy"])
	})

	t.Run("short-circuits on first error", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		p := New(
			successStage("scan"),
			errorStage("build failed"),
			successStage("deploy"), // Should not execute
		)
		initial := State{ProjectDir: "/project"}

		// Act
		result := Run(p, ctx, initial)

		// Assert
		require.True(t, E.IsLeft(result), "Pipeline should fail")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "build failed")
	})

	t.Run("propagates state through stages", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		p := New(
			addArtifactStage("lambda", "/build/lambda.zip"),
			successStage("deploy"),
		)
		initial := State{ProjectDir: "/project"}

		// Act
		result := Run(p, ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Pipeline should succeed")

		state := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Contains(t, state.Artifacts, "lambda")
		assert.Equal(t, "/build/lambda.zip", state.Artifacts["lambda"].Path)
		assert.Equal(t, "executed", state.Outputs["deploy"])
	})

	t.Run("handles error in first stage", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		p := New(
			errorStage("scan failed"),
			successStage("build"),
		)
		initial := State{ProjectDir: "/project"}

		// Act
		result := Run(p, ctx, initial)

		// Assert
		require.True(t, E.IsLeft(result), "Pipeline should fail immediately")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "scan failed")
	})

	t.Run("handles error in last stage", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		p := New(
			successStage("scan"),
			successStage("build"),
			errorStage("deploy failed"),
		)
		initial := State{ProjectDir: "/project"}

		// Act
		result := Run(p, ctx, initial)

		// Assert
		require.True(t, E.IsLeft(result), "Pipeline should fail on last stage")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "deploy failed")
	})
}

func TestChain(t *testing.T) {
	t.Run("chains empty pipelines", func(t *testing.T) {
		// Arrange
		p1 := New()
		p2 := New()

		// Act
		chained := Chain(p1, p2)

		// Assert
		assert.NotNil(t, chained)
		assert.Empty(t, chained.stages)
	})

	t.Run("chains single pipeline", func(t *testing.T) {
		// Arrange
		p := New(successStage("build"))

		// Act
		chained := Chain(p)

		// Assert
		assert.NotNil(t, chained)
		assert.Len(t, chained.stages, 1)
	})

	t.Run("chains multiple pipelines preserving order", func(t *testing.T) {
		// Arrange
		p1 := New(successStage("scan"), successStage("build"))
		p2 := New(successStage("test"))
		p3 := New(successStage("deploy"))

		// Act
		chained := Chain(p1, p2, p3)

		// Assert
		assert.NotNil(t, chained)
		assert.Len(t, chained.stages, 4)

		// Verify execution order
		ctx := t.Context()
		initial := State{ProjectDir: "/project"}
		result := Run(chained, ctx, initial)

		require.True(t, E.IsRight(result), "Chained pipeline should succeed")

		state := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Equal(t, "executed", state.Outputs["scan"])
		assert.Equal(t, "executed", state.Outputs["build"])
		assert.Equal(t, "executed", state.Outputs["test"])
		assert.Equal(t, "executed", state.Outputs["deploy"])
	})

	t.Run("chains pipelines with mixed empty and non-empty", func(t *testing.T) {
		// Arrange
		p1 := New()
		p2 := New(successStage("build"))
		p3 := New()

		// Act
		chained := Chain(p1, p2, p3)

		// Assert
		assert.NotNil(t, chained)
		assert.Len(t, chained.stages, 1)
	})
}

func TestParallel(t *testing.T) {
	t.Run("executes no stages successfully", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		parallel := Parallel()
		initial := State{ProjectDir: "/project"}

		// Act
		result := parallel(ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Parallel with no stages should succeed")
	})

	t.Run("executes single stage successfully", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		parallel := Parallel(successStage("build"))
		initial := State{ProjectDir: "/project"}

		// Act
		result := parallel(ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Parallel should succeed")

		state := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Equal(t, "executed", state.Outputs["build"])
	})

	t.Run("executes multiple stages", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		parallel := Parallel(
			successStage("lint"),
			successStage("test"),
			successStage("build"),
		)
		initial := State{ProjectDir: "/project"}

		// Act
		result := parallel(ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Parallel should succeed")

		state := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Equal(t, "executed", state.Outputs["lint"])
		assert.Equal(t, "executed", state.Outputs["test"])
		assert.Equal(t, "executed", state.Outputs["build"])
	})

	t.Run("stops on first error", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		parallel := Parallel(
			successStage("lint"),
			errorStage("test failed"),
			successStage("build"),
		)
		initial := State{ProjectDir: "/project"}

		// Act
		result := parallel(ctx, initial)

		// Assert
		require.True(t, E.IsLeft(result), "Parallel should fail on error")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "test failed")
	})

	t.Run("accumulates state from all stages", func(t *testing.T) {
		// Arrange
		ctx := t.Context()
		parallel := Parallel(
			addArtifactStage("api", "/build/api.zip"),
			addArtifactStage("worker", "/build/worker.zip"),
		)
		initial := State{ProjectDir: "/project"}

		// Act
		result := parallel(ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Parallel should succeed")

		state := E.GetOrElse(func(error) State { return State{} })(result)
		assert.Contains(t, state.Artifacts, "api")
		assert.Contains(t, state.Artifacts, "worker")
	})
}

// Integration test combining New, Run, Chain, and Parallel.
func TestPipelineIntegration(t *testing.T) {
	t.Run("complex pipeline with chaining and parallel stages", func(t *testing.T) {
		// Arrange
		ctx := t.Context()

		// Build phase
		buildPipeline := New(
			successStage("scan"),
			addArtifactStage("lambda", "/build/lambda.zip"),
		)

		// Test phase (parallel)
		testStage := Parallel(
			successStage("unit-tests"),
			successStage("integration-tests"),
		)

		// Deploy phase
		deployPipeline := New(testStage, successStage("deploy"))

		// Chain all phases
		fullPipeline := Chain(buildPipeline, deployPipeline)

		initial := State{ProjectDir: "/project"}

		// Act
		result := Run(fullPipeline, ctx, initial)

		// Assert
		require.True(t, E.IsRight(result), "Full pipeline should succeed")

		state := E.GetOrElse(func(error) State { return State{} })(result)

		// Verify all stages executed
		assert.Equal(t, "executed", state.Outputs["scan"])
		assert.Contains(t, state.Artifacts, "lambda")
		assert.Equal(t, "executed", state.Outputs["unit-tests"])
		assert.Equal(t, "executed", state.Outputs["integration-tests"])
		assert.Equal(t, "executed", state.Outputs["deploy"])
	})

	t.Run("pipeline fails and stops at error stage", func(t *testing.T) {
		// Arrange
		ctx := t.Context()

		pipeline := New(
			successStage("scan"),
			successStage("build"),
			errorStage("tests failed"),
			successStage("deploy"), // Should not execute
		)

		initial := State{ProjectDir: "/project"}

		// Act
		result := Run(pipeline, ctx, initial)

		// Assert
		require.True(t, E.IsLeft(result), "Pipeline should fail")

		err := E.Fold(
			func(e error) error { return e },
			func(State) error { return nil },
		)(result)
		assert.Contains(t, err.Error(), "tests failed")
	})
}
