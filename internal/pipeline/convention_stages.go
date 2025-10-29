package pipeline

import (
	"context"
	"fmt"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
	"github.com/lewis/forge/internal/build"
	"github.com/lewis/forge/internal/discovery"
)

// ConventionScan creates a stage that scans for functions using convention-based discovery
func ConventionScan() Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		fmt.Println("==> Scanning for Lambda functions...")

		scanner := discovery.NewScanner(s.ProjectDir)
		functions, err := scanner.ScanFunctions()
		if err != nil {
			return E.Left[State](fmt.Errorf("failed to scan functions: %w", err))
		}

		if len(functions) == 0 {
			return E.Left[State](fmt.Errorf("no functions found in src/functions/"))
		}

		fmt.Printf("Found %d function(s):\n", len(functions))
		for _, fn := range functions {
			fmt.Printf("  - %s (%s)\n", fn.Name, fn.Runtime)
		}
		fmt.Println()

		// Store functions in state (reuse Config field as interface{})
		s.Config = functions
		return E.Right[error](s)
	}
}

// ConventionStubs creates a stage that generates stub zip files
func ConventionStubs() Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		// Extract functions from state
		functions, ok := s.Config.([]discovery.Function)
		if !ok {
			return E.Left[State](fmt.Errorf("invalid state: functions not found"))
		}

		buildDir := filepath.Join(s.ProjectDir, ".forge", "build")

		count, err := discovery.CreateStubZips(functions, buildDir)
		if err != nil {
			return E.Left[State](fmt.Errorf("failed to create stub zips: %w", err))
		}

		if count > 0 {
			fmt.Printf("Created %d stub zip(s)\n\n", count)
		}

		return E.Right[error](s)
	}
}

// ConventionBuild creates a stage that builds all discovered functions
func ConventionBuild() Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		fmt.Println("==> Building Lambda functions...")

		// Extract functions from state
		functions, ok := s.Config.([]discovery.Function)
		if !ok {
			return E.Left[State](fmt.Errorf("invalid state: functions not found"))
		}

		// Initialize artifacts map if needed
		if s.Artifacts == nil {
			s.Artifacts = make(map[string]Artifact)
		}

		registry := build.NewRegistry()
		buildDir := filepath.Join(s.ProjectDir, ".forge", "build")

		for _, fn := range functions {
			fmt.Printf("[%s] Building...\n", fn.Name)

			// Get builder from registry (returns Option)
			builderOpt := registry.Get(fn.Runtime)
			if O.IsNone(builderOpt) {
				return E.Left[State](fmt.Errorf("unsupported runtime: %s", fn.Runtime))
			}

			// Extract builder using Fold
			builder := O.Fold(
				func() build.BuildFunc { return nil },
				func(b build.BuildFunc) build.BuildFunc { return b },
			)(builderOpt)

			// Convert to build config
			cfg := fn.ToBuildConfig(buildDir)

			// Execute build (returns Either)
			result := builder(ctx, cfg)

			// Handle result using functional error handling
			if E.IsLeft(result) {
				err := E.Fold(
					func(e error) error { return e },
					func(a build.Artifact) error { return nil },
				)(result)
				return E.Left[State](fmt.Errorf("failed to build %s: %w", fn.Name, err))
			}

			// Extract artifact
			artifact := E.Fold(
				func(e error) build.Artifact { return build.Artifact{} },
				func(a build.Artifact) build.Artifact { return a },
			)(result)

			// Store artifact in state
			s.Artifacts[fn.Name] = Artifact{
				Path:     artifact.Path,
				Checksum: artifact.Checksum,
				Size:     artifact.Size,
			}

			sizeMB := float64(artifact.Size) / 1024 / 1024
			fmt.Printf("[%s] ✓ Built: %s (%.2f MB)\n", fn.Name, filepath.Base(artifact.Path), sizeMB)
		}

		fmt.Println()
		return E.Right[error](s)
	}
}

// ConventionTerraformInit creates a stage that initializes Terraform in infra/
func ConventionTerraformInit(exec TerraformExecutor) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		fmt.Println("==> Initializing Terraform...")

		infraDir := filepath.Join(s.ProjectDir, "infra")
		if err := exec.Init(ctx, infraDir); err != nil {
			return E.Left[State](fmt.Errorf("terraform init failed: %w", err))
		}

		return E.Right[error](s)
	}
}

// ConventionTerraformPlan creates a stage that plans infrastructure in infra/
func ConventionTerraformPlan(exec TerraformExecutor, namespace string) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		fmt.Println("==> Planning infrastructure changes...")

		infraDir := filepath.Join(s.ProjectDir, "infra")

		// Set namespace variable if provided
		var vars map[string]string
		if namespace != "" {
			vars = map[string]string{
				"namespace": namespace + "-",
			}
			fmt.Printf("Deploying to namespace: %s\n", namespace)
		}

		hasChanges, err := exec.PlanWithVars(ctx, infraDir, vars)
		if err != nil {
			return E.Left[State](fmt.Errorf("terraform plan failed: %w", err))
		}

		if !hasChanges {
			fmt.Println("No changes detected")
		}

		return E.Right[error](s)
	}
}

// ConventionTerraformApply creates a stage that applies infrastructure in infra/
func ConventionTerraformApply(exec TerraformExecutor, autoApprove bool) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		// Request approval if not auto-approved
		if !autoApprove {
			fmt.Print("\nDo you want to apply these changes? (yes/no): ")
			var response string
			fmt.Scanln(&response)
			if response != "yes" {
				return E.Left[State](fmt.Errorf("deployment cancelled by user"))
			}
		}

		fmt.Println("==> Applying infrastructure changes...")

		infraDir := filepath.Join(s.ProjectDir, "infra")
		if err := exec.Apply(ctx, infraDir); err != nil {
			return E.Left[State](fmt.Errorf("terraform apply failed: %w", err))
		}

		return E.Right[error](s)
	}
}

// ConventionTerraformOutputs creates a stage that captures Terraform outputs
func ConventionTerraformOutputs(exec TerraformExecutor) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		infraDir := filepath.Join(s.ProjectDir, "infra")

		outputs, err := exec.Output(ctx, infraDir)
		if err != nil {
			// Non-fatal - just warn
			fmt.Printf("Warning: failed to retrieve outputs: %v\n", err)
			return E.Right[error](s)
		}

		if s.Outputs == nil {
			s.Outputs = make(map[string]interface{})
		}
		s.Outputs = outputs

		return E.Right[error](s)
	}
}

// TerraformExecutor interface for convention-based stages
// This allows stages to work with simplified terraform operations
type TerraformExecutor interface {
	Init(ctx context.Context, dir string) error
	Plan(ctx context.Context, dir string) (bool, error)
	PlanWithVars(ctx context.Context, dir string, vars map[string]string) (bool, error)
	Apply(ctx context.Context, dir string) error
	Output(ctx context.Context, dir string) (map[string]interface{}, error)
}
