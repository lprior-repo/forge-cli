// Package pipeline provides convention-based Lambda function deployment stages
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	A "github.com/IBM/fp-go/array"
	E "github.com/IBM/fp-go/either"

	"github.com/lewis/forge/internal/build"
	"github.com/lewis/forge/internal/discovery"
)

// ConventionScan creates a stage that scans for functions using convention-based discovery.
func ConventionScan() Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		fmt.Println("==> Scanning for Lambda functions...")

		// Pure functional call - no OOP
		functions, err := discovery.ScanFunctions(s.ProjectDir)
		if err != nil {
			return E.Left[State](fmt.Errorf("failed to scan functions: %w", err))
		}

		if len(functions) == 0 {
			return E.Left[State](errors.New("no functions found in src/functions/"))
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

// ConventionStubs creates a stage that generates stub zip files.
func ConventionStubs() Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		// Extract functions from state
		functions, ok := s.Config.([]discovery.Function)
		if !ok {
			return E.Left[State](errors.New("invalid state: functions not found"))
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

// BuildResult represents the result of building a function.
type BuildResult struct {
	name     string
	artifact Artifact
}

// buildFunction is a helper that builds a single function and returns Either.
func buildFunction(ctx context.Context, registry build.Registry, buildDir string) func(discovery.Function) E.Either[error, BuildResult] {
	return func(fn discovery.Function) E.Either[error, BuildResult] {
		fmt.Printf("[%s] Building...\n", fn.Name)

		// Get builder from registry and convert Option to Either
		builderEither := E.FromOption[build.BuildFunc](
			func() error { return fmt.Errorf("unsupported runtime: %s", fn.Runtime) },
		)(build.GetBuilder(registry, fn.Runtime))

		// Chain the build operation with config validation
		return E.Chain(func(builder build.BuildFunc) E.Either[error, BuildResult] {
			// ToBuildConfig now returns Either for validation
			return E.Chain(func(cfg build.Config) E.Either[error, BuildResult] {
				return E.Chain(func(artifact build.Artifact) E.Either[error, BuildResult] {
					sizeMB := float64(artifact.Size) / 1024 / 1024
					fmt.Printf("[%s] ✓ Built: %s (%.2f MB)\n", fn.Name, filepath.Base(artifact.Path), sizeMB)

					return E.Right[error](BuildResult{
						name: fn.Name,
						artifact: Artifact{
							Path:     artifact.Path,
							Checksum: artifact.Checksum,
							Size:     artifact.Size,
						},
					})
				})(builder(ctx, cfg))
			})(discovery.ToBuildConfig(fn, buildDir))
		})(builderEither)
	}
}

// ConventionBuild creates a stage that builds all discovered functions.
func ConventionBuild() Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		fmt.Println("==> Building Lambda functions...")

		// Extract functions from state
		functions, ok := s.Config.([]discovery.Function)
		if !ok {
			return E.Left[State](errors.New("invalid state: functions not found"))
		}

		registry := build.NewRegistry()
		buildDir := filepath.Join(s.ProjectDir, ".forge", "build")

		// Use functional fold to build artifacts map
		artifactsEither := A.Reduce(
			func(acc E.Either[error, map[string]Artifact], fn discovery.Function) E.Either[error, map[string]Artifact] {
				return E.Chain(func(artifacts map[string]Artifact) E.Either[error, map[string]Artifact] {
					return E.Map[error](func(result BuildResult) map[string]Artifact {
						// Immutable update - create new map
						newArtifacts := make(map[string]Artifact, len(artifacts)+1)
						for k, v := range artifacts {
							newArtifacts[k] = v
						}
						newArtifacts[result.name] = result.artifact
						return newArtifacts
					})(buildFunction(ctx, registry, buildDir)(fn))
				})(acc)
			},
			E.Right[error](s.Artifacts),
		)(functions)

		// Return new state with updated artifacts
		fmt.Println()
		return E.Map[error](func(artifacts map[string]Artifact) State {
			return State{
				ProjectDir: s.ProjectDir,
				Artifacts:  artifacts,
				Outputs:    s.Outputs,
				Config:     s.Config,
			}
		})(artifactsEither)
	}
}

// ConventionTerraformInit creates a stage that initializes Terraform in infra/.
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

// ConventionTerraformPlan creates a stage that plans infrastructure in infra/.
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

// ConventionTerraformApply creates a stage that applies infrastructure in infra/.
func ConventionTerraformApply(exec TerraformExecutor, autoApprove bool) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		// Request approval if not auto-approved
		if !autoApprove {
			fmt.Print("\nDo you want to apply these changes? (yes/no): ")
			var response string
			_, _ = fmt.Scanln(&response) // #nosec G104 - user input error is non-critical
			if response != "yes" {
				return E.Left[State](errors.New("deployment canceled by user"))
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

// ConventionTerraformOutputs creates a stage that captures Terraform outputs.
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

// TerraformInitFunc initializes Terraform in a directory.
type TerraformInitFunc func(ctx context.Context, dir string) error

// TerraformPlanFunc plans infrastructure changes.
type TerraformPlanFunc func(ctx context.Context, dir string) (bool, error)

// TerraformPlanWithVarsFunc plans infrastructure changes with variables.
type TerraformPlanWithVarsFunc func(ctx context.Context, dir string, vars map[string]string) (bool, error)

// TerraformApplyFunc applies infrastructure changes.
type TerraformApplyFunc func(ctx context.Context, dir string) error

// TerraformOutputFunc retrieves Terraform outputs.
type TerraformOutputFunc func(ctx context.Context, dir string) (map[string]interface{}, error)

// This follows functional programming - functions as first-class values.
type TerraformExecutor struct {
	Init         TerraformInitFunc
	Plan         TerraformPlanFunc
	PlanWithVars TerraformPlanWithVarsFunc
	Apply        TerraformApplyFunc
	Output       TerraformOutputFunc
}
