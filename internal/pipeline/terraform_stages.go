package pipeline

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/terraform"
)

// TerraformInit creates a stage that initializes all stacks
func TerraformInit(exec terraform.Executor) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		// Execute init for each stack
		for _, st := range s.Stacks {
			if err := exec.Init(ctx, st.Path); err != nil {
				return E.Left[State](fmt.Errorf("init failed for %s: %w", st.Name, err))
			}
		}

		return E.Right[error](s)
	}
}

// TerraformPlan creates a stage that plans all stacks
func TerraformPlan(exec terraform.Executor) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		hasAnyChanges := false

		for _, st := range s.Stacks {
			hasChanges, err := exec.Plan(ctx, st.Path)
			if err != nil {
				return E.Left[State](fmt.Errorf("plan failed for %s: %w", st.Name, err))
			}
			if hasChanges {
				hasAnyChanges = true
			}
		}

		if !hasAnyChanges {
			fmt.Println("No changes detected in any stack")
		}

		return E.Right[error](s)
	}
}

// TerraformApply creates a stage that applies all stacks
func TerraformApply(exec terraform.Executor, autoApprove bool) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		opts := []terraform.ApplyOption{
			terraform.AutoApprove(autoApprove),
		}

		for idx, st := range s.Stacks {
			fmt.Printf("[%d/%d] Applying %s...\n", idx+1, len(s.Stacks), st.Name)

			if err := exec.Apply(ctx, st.Path, opts...); err != nil {
				return E.Left[State](fmt.Errorf("apply failed for %s: %w", st.Name, err))
			}
		}

		return E.Right[error](s)
	}
}

// TerraformDestroy creates a stage that destroys all stacks
func TerraformDestroy(exec terraform.Executor, autoApprove bool) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		opts := []terraform.DestroyOption{
			terraform.DestroyAutoApprove(autoApprove),
		}

		// Reverse order for destroy (dependencies last)
		for i := len(s.Stacks) - 1; i >= 0; i-- {
			st := s.Stacks[i]
			fmt.Printf("[%d/%d] Destroying %s...\n", len(s.Stacks)-i, len(s.Stacks), st.Name)

			if err := exec.Destroy(ctx, st.Path, opts...); err != nil {
				return E.Left[State](fmt.Errorf("destroy failed for %s: %w", st.Name, err))
			}
		}

		return E.Right[error](s)
	}
}

// CaptureOutputs captures terraform outputs from all stacks
func CaptureOutputs(exec terraform.Executor) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		if s.Outputs == nil {
			s.Outputs = make(map[string]interface{})
		}

		for _, st := range s.Stacks {
			outputs, err := exec.Output(ctx, st.Path)
			if err != nil {
				return E.Left[State](fmt.Errorf("failed to get outputs for %s: %w", st.Name, err))
			}
			s.Outputs[st.Name] = outputs
		}

		return E.Right[error](s)
	}
}
