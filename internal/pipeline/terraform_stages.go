package pipeline

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/either"

	"github.com/lewis/forge/internal/terraform"
)

// TerraformInit creates a stage that initializes terraform in the project directory.
func TerraformInit(exec terraform.Executor) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		if err := exec.Init(ctx, s.ProjectDir); err != nil {
			return E.Left[State](fmt.Errorf("init failed: %w", err))
		}

		return E.Right[error](s)
	}
}

// TerraformPlan creates a stage that plans terraform in the project directory.
func TerraformPlan(exec terraform.Executor) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		hasChanges, err := exec.Plan(ctx, s.ProjectDir)
		if err != nil {
			return E.Left[State](fmt.Errorf("plan failed: %w", err))
		}

		if !hasChanges {
			fmt.Println("No changes detected")
		}

		return E.Right[error](s)
	}
}

// TerraformApply creates a stage that applies terraform in the project directory.
func TerraformApply(exec terraform.Executor, autoApprove bool) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		opts := []terraform.ApplyOption{
			terraform.AutoApprove(autoApprove),
		}

		fmt.Println("Applying infrastructure...")

		if err := exec.Apply(ctx, s.ProjectDir, opts...); err != nil {
			return E.Left[State](fmt.Errorf("apply failed: %w", err))
		}

		return E.Right[error](s)
	}
}

// TerraformDestroy creates a stage that destroys terraform infrastructure.
func TerraformDestroy(exec terraform.Executor, autoApprove bool) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		opts := []terraform.DestroyOption{
			terraform.DestroyAutoApprove(autoApprove),
		}

		if err := exec.Destroy(ctx, s.ProjectDir, opts...); err != nil {
			return E.Left[State](fmt.Errorf("destroy failed: %w", err))
		}

		return E.Right[error](s)
	}
}

// CaptureOutputs captures terraform outputs from the project directory.
func CaptureOutputs(exec terraform.Executor) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		if s.Outputs == nil {
			s.Outputs = make(map[string]interface{})
		}

		outputs, err := exec.Output(ctx, s.ProjectDir)
		if err != nil {
			return E.Left[State](fmt.Errorf("failed to get outputs: %w", err))
		}
		s.Outputs["main"] = outputs

		return E.Right[error](s)
	}
}
