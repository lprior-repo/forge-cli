package pipeline

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
)

// PURE: Returns events as data instead of printing to console.
func ConventionTerraformInitV2(exec TerraformExecutor) EventStage {
	return func(ctx context.Context, s State) E.Either[error, StageResult] {
		infraDir := filepath.Join(s.ProjectDir, "infra")

		// Build events
		events := []StageEvent{
			NewEvent(EventLevelInfo, "==> Initializing Terraform..."),
		}

		// Execute terraform init (I/O)
		if err := exec.Init(ctx, infraDir); err != nil {
			return E.Left[StageResult](fmt.Errorf("terraform init failed: %w", err))
		}

		events = append(events, NewEvent(EventLevelSuccess, "[terraform] Initialized"))

		return E.Right[error](StageResult{
			State:  s,
			Events: events,
		})
	}
}

// PURE: Returns events as data instead of printing to console.
func ConventionTerraformPlanV2(exec TerraformExecutor, namespace string) EventStage {
	return func(ctx context.Context, s State) E.Either[error, StageResult] {
		infraDir := filepath.Join(s.ProjectDir, "infra")

		// Build events
		events := []StageEvent{
			NewEvent(EventLevelInfo, "==> Planning infrastructure changes..."),
		}

		// Set namespace variable if provided
		var vars map[string]string
		if namespace != "" {
			vars = map[string]string{
				"namespace": namespace + "-",
			}
			events = append(events, NewEvent(EventLevelInfo, "Deploying to namespace: "+namespace))
		}

		// Execute terraform plan (I/O)
		hasChanges, err := exec.PlanWithVars(ctx, infraDir, vars)
		if err != nil {
			return E.Left[StageResult](fmt.Errorf("terraform plan failed: %w", err))
		}

		if hasChanges {
			events = append(events, NewEvent(EventLevelInfo, "[terraform] Changes detected"))
		} else {
			events = append(events, NewEvent(EventLevelInfo, "[terraform] No changes detected"))
		}

		return E.Right[error](StageResult{
			State:  s,
			Events: events,
		})
	}
}

// Returns true if approved, false if canceled.
type ApprovalFunc func() bool

// Takes an approval function to maintain purity - I/O happens at edges.
func ConventionTerraformApplyV2(exec TerraformExecutor, approvalFunc ApprovalFunc) EventStage {
	return func(ctx context.Context, s State) E.Either[error, StageResult] {
		// Build events
		events := []StageEvent{}

		// Request approval through function parameter (I/O at edge)
		if approvalFunc != nil && !approvalFunc() {
			return E.Left[StageResult](errors.New("deployment canceled by user"))
		}

		events = append(events, NewEvent(EventLevelInfo, "==> Applying infrastructure changes..."))

		infraDir := filepath.Join(s.ProjectDir, "infra")

		// Execute terraform apply (I/O)
		if err := exec.Apply(ctx, infraDir); err != nil {
			return E.Left[StageResult](fmt.Errorf("terraform apply failed: %w", err))
		}

		events = append(events, NewEvent(EventLevelSuccess, "[terraform] Applied successfully"))

		return E.Right[error](StageResult{
			State:  s,
			Events: events,
		})
	}
}

// PURE: Returns events as data instead of printing to console.
func ConventionTerraformOutputsV2(exec TerraformExecutor) EventStage {
	return func(ctx context.Context, s State) E.Either[error, StageResult] {
		infraDir := filepath.Join(s.ProjectDir, "infra")

		// Build events
		events := []StageEvent{}

		// Execute terraform output (I/O)
		outputs, err := exec.Output(ctx, infraDir)
		if err != nil {
			// Non-fatal - just warn
			events = append(events, NewEvent(EventLevelWarning, fmt.Sprintf("Failed to retrieve outputs: %v", err)))
			return E.Right[error](StageResult{
				State:  s,
				Events: events,
			})
		}

		// Create new state with outputs (immutable)
		newOutputs := outputs
		if newOutputs == nil {
			newOutputs = make(map[string]interface{})
		}

		newState := State{
			ProjectDir: s.ProjectDir,
			Artifacts:  s.Artifacts,
			Outputs:    newOutputs,
			Config:     s.Config,
		}

		if len(outputs) > 0 {
			events = append(events, NewEvent(EventLevelInfo, fmt.Sprintf("Captured %d output(s)", len(outputs))))
		}

		return E.Right[error](StageResult{
			State:  newState,
			Events: events,
		})
	}
}
