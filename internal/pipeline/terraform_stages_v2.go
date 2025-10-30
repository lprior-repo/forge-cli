package pipeline

import (
	"context"
	"fmt"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
)

// ConventionTerraformInitV2 creates an event-based stage that initializes Terraform
// PURE: Returns events as data instead of printing to console
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

// ConventionTerraformPlanV2 creates an event-based stage that plans infrastructure
// PURE: Returns events as data instead of printing to console
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
			events = append(events, NewEvent(EventLevelInfo, fmt.Sprintf("Deploying to namespace: %s", namespace)))
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

// ConventionTerraformApplyV2 creates an event-based stage that applies infrastructure
// PURE: Returns events as data instead of printing to console
func ConventionTerraformApplyV2(exec TerraformExecutor, autoApprove bool) EventStage {
	return func(ctx context.Context, s State) E.Either[error, StageResult] {
		// Build events
		events := []StageEvent{}

		// Request approval if not auto-approved
		if !autoApprove {
			// This is still an I/O operation (console input)
			// For now, we'll keep this as-is since it's user interaction
			fmt.Print("\nDo you want to apply these changes? (yes/no): ")
			var response string
			_, _ = fmt.Scanln(&response) // #nosec G104 - user input error is non-critical
			if response != "yes" {
				return E.Left[StageResult](fmt.Errorf("deployment canceled by user"))
			}
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

// ConventionTerraformOutputsV2 creates an event-based stage that captures Terraform outputs
// PURE: Returns events as data instead of printing to console
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
