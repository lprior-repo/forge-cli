package cli

import (
	"context"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/build"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/stack"
	"github.com/lewis/forge/internal/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeployPipeline tests the deploy pipeline using functional approach
func TestDeployPipeline(t *testing.T) {
	t.Run("builds a complete deploy pipeline", func(t *testing.T) {
		// Mock executor
		exec := terraform.NewMockExecutor()

		// Create pipeline stages
		deployPipeline := pipeline.New(
			pipeline.TerraformInit(exec),
			pipeline.TerraformPlan(exec),
			pipeline.TerraformApply(exec, true),
		)

		// Initial state
		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api"},
			},
		}

		// Run pipeline
		result := deployPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsRight(result), "Deploy pipeline should succeed")
	})

	t.Run("stops on terraform init failure", func(t *testing.T) {
		// Mock executor with failing Init
		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return assert.AnError
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return true, nil
			},
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				return nil
			},
		}

		deployPipeline := pipeline.New(
			pipeline.TerraformInit(exec),
			pipeline.TerraformPlan(exec),
			pipeline.TerraformApply(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api"},
			},
		}

		result := deployPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsLeft(result), "Should fail on init error")
	})

	t.Run("stops on terraform plan failure", func(t *testing.T) {
		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return false, assert.AnError
			},
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				return nil
			},
		}

		deployPipeline := pipeline.New(
			pipeline.TerraformInit(exec),
			pipeline.TerraformPlan(exec),
			pipeline.TerraformApply(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api"},
			},
		}

		result := deployPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsLeft(result), "Should fail on plan error")
	})

	t.Run("stops on terraform apply failure", func(t *testing.T) {
		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return true, nil
			},
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				return assert.AnError
			},
		}

		deployPipeline := pipeline.New(
			pipeline.TerraformInit(exec),
			pipeline.TerraformPlan(exec),
			pipeline.TerraformApply(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api"},
			},
		}

		result := deployPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsLeft(result), "Should fail on apply error")
	})
}

// TestDeployWithBuild tests deploy pipeline with build stage
func TestDeployWithBuild(t *testing.T) {
	t.Run("builds artifacts before deployment", func(t *testing.T) {
		exec := terraform.NewMockExecutor()

		// Mock build function
		mockBuild := func(ctx context.Context, cfg build.Config) E.Either[error, build.Artifact] {
			return E.Right[error](build.Artifact{
				Path:     cfg.SourceDir + "/bootstrap",
				Checksum: "abc123",
				Size:     1024,
			})
		}

		// Build stage that populates artifacts
		buildStage := func(ctx context.Context, s pipeline.State) E.Either[error, pipeline.State] {
			if s.Artifacts == nil {
				s.Artifacts = make(map[string]pipeline.Artifact)
			}

			for _, st := range s.Stacks {
				cfg := build.Config{
					SourceDir: st.Path,
					Runtime:   st.Runtime,
				}

				result := mockBuild(ctx, cfg)
				if E.IsLeft(result) {
					// Extract error using Fold
					err := E.Fold(
						func(e error) error { return e },
						func(a build.Artifact) error { return nil },
					)(result)
					return E.Left[pipeline.State](err)
				}

				// Convert to pipeline artifact using Fold
				artifact := E.Fold(
					func(e error) build.Artifact { return build.Artifact{} },
					func(a build.Artifact) build.Artifact { return a },
				)(result)

				s.Artifacts[st.Name] = pipeline.Artifact{
					Path:     artifact.Path,
					Checksum: artifact.Checksum,
					Size:     artifact.Size,
				}
			}

			return E.Right[error](s)
		}

		// Complete pipeline: Build → Init → Plan → Apply
		deployPipeline := pipeline.New(
			buildStage,
			pipeline.TerraformInit(exec),
			pipeline.TerraformPlan(exec),
			pipeline.TerraformApply(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api", Runtime: "go1.x"},
			},
		}

		result := deployPipeline.Run(context.Background(), initialState)

		require.True(t, E.IsRight(result), "Complete pipeline should succeed")

		// Verify artifacts were created using Fold
		finalState := E.Fold(
			func(e error) pipeline.State { return pipeline.State{} },
			func(s pipeline.State) pipeline.State { return s },
		)(result)

		assert.Contains(t, finalState.Artifacts, "api")
		assert.Equal(t, "stacks/api/bootstrap", finalState.Artifacts["api"].Path)
	})

	t.Run("fails if build fails", func(t *testing.T) {
		exec := terraform.NewMockExecutor()

		// Mock build function that fails
		mockBuild := func(ctx context.Context, cfg build.Config) E.Either[error, build.Artifact] {
			return E.Left[build.Artifact](assert.AnError)
		}

		buildStage := func(ctx context.Context, s pipeline.State) E.Either[error, pipeline.State] {
			for _, st := range s.Stacks {
				cfg := build.Config{
					SourceDir: st.Path,
					Runtime:   st.Runtime,
				}

				result := mockBuild(ctx, cfg)
				if E.IsLeft(result) {
					err := E.Fold(
						func(e error) error { return e },
						func(a build.Artifact) error { return nil },
					)(result)
					return E.Left[pipeline.State](err)
				}
			}

			return E.Right[error](s)
		}

		deployPipeline := pipeline.New(
			buildStage,
			pipeline.TerraformInit(exec),
			pipeline.TerraformPlan(exec),
			pipeline.TerraformApply(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api", Runtime: "go1.x"},
			},
		}

		result := deployPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsLeft(result), "Should fail if build fails")
	})
}

// TestMultiStackDeploy tests deploying multiple stacks
func TestMultiStackDeploy(t *testing.T) {
	t.Run("deploys multiple stacks in order", func(t *testing.T) {
		var executionOrder []string

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return true, nil
			},
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				executionOrder = append(executionOrder, dir)
				return nil
			},
		}

		deployPipeline := pipeline.New(
			pipeline.TerraformInit(exec),
			pipeline.TerraformPlan(exec),
			pipeline.TerraformApply(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "database", Path: "stacks/database"},
				{Name: "api", Path: "stacks/api"},
				{Name: "frontend", Path: "stacks/frontend"},
			},
		}

		result := deployPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsRight(result), "Multi-stack deploy should succeed")
		assert.Len(t, executionOrder, 3, "Should deploy all 3 stacks")
	})

	t.Run("stops on first stack failure", func(t *testing.T) {
		var executionOrder []string

		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return true, nil
			},
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				executionOrder = append(executionOrder, dir)
				// Fail on second stack
				if dir == "stacks/api" {
					return assert.AnError
				}
				return nil
			},
		}

		deployPipeline := pipeline.New(
			pipeline.TerraformInit(exec),
			pipeline.TerraformPlan(exec),
			pipeline.TerraformApply(exec, true),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "database", Path: "stacks/database"},
				{Name: "api", Path: "stacks/api"},
				{Name: "frontend", Path: "stacks/frontend"},
			},
		}

		result := deployPipeline.Run(context.Background(), initialState)

		assert.True(t, E.IsLeft(result), "Should fail on stack failure")
		// Should have deployed database and attempted api (but failed)
		assert.Contains(t, executionOrder, "stacks/database")
		assert.Contains(t, executionOrder, "stacks/api")
		// Should NOT have attempted frontend
		assert.NotContains(t, executionOrder, "stacks/frontend")
	})
}

// TestDeployWithOutputCapture tests capturing outputs after deployment
func TestDeployWithOutputCapture(t *testing.T) {
	t.Run("captures outputs from all stacks", func(t *testing.T) {
		exec := terraform.Executor{
			Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
				return nil
			},
			Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
				return true, nil
			},
			Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
				return nil
			},
			Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
				// Return mock outputs based on directory
				if dir == "stacks/api" {
					return map[string]interface{}{
						"api_url": "https://api.example.com",
					}, nil
				}
				return map[string]interface{}{}, nil
			},
		}

		deployPipeline := pipeline.New(
			pipeline.TerraformInit(exec),
			pipeline.TerraformPlan(exec),
			pipeline.TerraformApply(exec, true),
			pipeline.CaptureOutputs(exec),
		)

		initialState := pipeline.State{
			ProjectDir: "/test",
			Stacks: []*stack.Stack{
				{Name: "api", Path: "stacks/api"},
			},
		}

		result := deployPipeline.Run(context.Background(), initialState)

		require.True(t, E.IsRight(result), "Deploy with output capture should succeed")

		// Verify outputs were captured using Fold
		finalState := E.Fold(
			func(e error) pipeline.State { return pipeline.State{} },
			func(s pipeline.State) pipeline.State { return s },
		)(result)

		assert.Contains(t, finalState.Outputs, "api")
		outputs := finalState.Outputs["api"].(map[string]interface{})
		assert.Equal(t, "https://api.example.com", outputs["api_url"])
	})
}

// BenchmarkDeployPipeline benchmarks the deploy pipeline
func BenchmarkDeployPipeline(b *testing.B) {
	exec := terraform.NewMockExecutor()

	deployPipeline := pipeline.New(
		pipeline.TerraformInit(exec),
		pipeline.TerraformPlan(exec),
		pipeline.TerraformApply(exec, true),
	)

	initialState := pipeline.State{
		ProjectDir: "/test",
		Stacks: []*stack.Stack{
			{Name: "api", Path: "stacks/api"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deployPipeline.Run(context.Background(), initialState)
	}
}
