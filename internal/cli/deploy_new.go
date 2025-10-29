package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
	"github.com/lewis/forge/internal/build"
	"github.com/lewis/forge/internal/discovery"
	"github.com/lewis/forge/internal/terraform"
	"github.com/spf13/cobra"
)

// DeployConfig holds deployment configuration (immutable)
type DeployConfig struct {
	ProjectRoot string
	Namespace   string
	AutoApprove bool
}

// DeployState represents deployment state flowing through the pipeline
type DeployState struct {
	Config    DeployConfig
	Functions []discovery.Function
	Artifacts map[string]build.Artifact
	TFOutputs map[string]interface{}
}

// DeployStage is a functional pipeline stage
type DeployStage func(context.Context, DeployState) E.Either[error, DeployState]

// NewConventionDeployCmd creates the convention-based 'deploy' command
func NewConventionDeployCmd() *cobra.Command {
	var (
		autoApprove bool
		namespace   string
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy Lambda functions with Terraform",
		Long: `Build Lambda functions using convention-based discovery and deploy with Terraform.

Convention-based discovery:
  - Scans src/functions/* for Lambda functions
  - Detects runtime from entry files (main.go, index.js, app.py)
  - Builds to .forge/build/{name}.zip
  - Runs terraform init/plan/apply in infra/

Namespace support for ephemeral environments:
  forge deploy --namespace=pr-123
    → Sets TF_VAR_namespace=pr-123
    → All resources get pr-123- prefix
    → Isolated preview environment

Examples:
  forge deploy                    # Deploy to default environment
  forge deploy --namespace=pr-123 # Deploy to ephemeral PR environment
  forge deploy --auto-approve     # Skip interactive approval`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			cfg := DeployConfig{
				ProjectRoot: projectRoot,
				Namespace:   namespace,
				AutoApprove: autoApprove,
			}

			return runConventionDeploy(context.Background(), cfg)
		},
	}

	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip interactive approval")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace for ephemeral environments (e.g., pr-123)")

	return cmd
}

// runConventionDeploy executes the deployment using functional pipeline
func runConventionDeploy(ctx context.Context, cfg DeployConfig) error {
	// Create functional pipeline: Scan → Stub → Build → TF Init → TF Plan → TF Apply → Outputs
	pipeline := composePipeline(
		scanFunctionsStage,
		createStubsStage,
		buildFunctionsStage,
		terraformInitStage,
		terraformPlanStage,
		terraformApplyStage,
		terraformOutputsStage,
	)

	// Initial state
	initialState := DeployState{
		Config:    cfg,
		Functions: nil,
		Artifacts: make(map[string]build.Artifact),
		TFOutputs: make(map[string]interface{}),
	}

	// Run pipeline
	result := pipeline(ctx, initialState)

	// Handle result using Either monad
	return E.Fold(
		func(err error) error {
			return fmt.Errorf("deployment failed: %w", err)
		},
		func(finalState DeployState) error {
			fmt.Println("\n✓ Deployment successful")
			if finalState.Config.Namespace != "" {
				fmt.Printf("Namespace: %s\n", finalState.Config.Namespace)
			}
			if len(finalState.TFOutputs) > 0 {
				fmt.Println("\nOutputs:")
				for key, value := range finalState.TFOutputs {
					fmt.Printf("  %s = %v\n", key, value)
				}
			}
			return nil
		},
	)(result)
}

// composePipeline composes multiple stages into a single pipeline (Railway-oriented programming)
func composePipeline(stages ...DeployStage) DeployStage {
	return func(ctx context.Context, state DeployState) E.Either[error, DeployState] {
		result := E.Right[error](state)

		for _, stage := range stages {
			result = E.Chain(func(s DeployState) E.Either[error, DeployState] {
				return stage(ctx, s)
			})(result)

			// Short-circuit on error
			if E.IsLeft(result) {
				break
			}
		}

		return result
	}
}

// scanFunctionsStage discovers functions using convention
func scanFunctionsStage(ctx context.Context, state DeployState) E.Either[error, DeployState] {
	fmt.Println("==> Scanning for Lambda functions...")

	scanner := discovery.NewScanner(state.Config.ProjectRoot)
	functions, err := scanner.ScanFunctions()
	if err != nil {
		return E.Left[DeployState](fmt.Errorf("failed to scan functions: %w", err))
	}

	if len(functions) == 0 {
		return E.Left[DeployState](fmt.Errorf("no functions found in src/functions/"))
	}

	fmt.Printf("Found %d function(s):\n", len(functions))
	for _, fn := range functions {
		fmt.Printf("  - %s (%s)\n", fn.Name, fn.Runtime)
	}
	fmt.Println()

	state.Functions = functions
	return E.Right[error](state)
}

// createStubsStage creates stub zip files for Terraform initialization
func createStubsStage(ctx context.Context, state DeployState) E.Either[error, DeployState] {
	buildDir := filepath.Join(state.Config.ProjectRoot, ".forge", "build")

	count, err := discovery.CreateStubZips(state.Functions, buildDir)
	if err != nil {
		return E.Left[DeployState](fmt.Errorf("failed to create stub zips: %w", err))
	}

	if count > 0 {
		fmt.Printf("Created %d stub zip(s)\n\n", count)
	}

	return E.Right[error](state)
}

// buildFunctionsStage builds all Lambda functions
func buildFunctionsStage(ctx context.Context, state DeployState) E.Either[error, DeployState] {
	fmt.Println("==> Building Lambda functions...")

	registry := build.NewRegistry()
	buildDir := filepath.Join(state.Config.ProjectRoot, ".forge", "build")

	for _, fn := range state.Functions {
		fmt.Printf("[%s] Building...\n", fn.Name)

		// Get builder from registry (returns Option)
		builderOpt := registry.Get(fn.Runtime)
		if O.IsNone(builderOpt) {
			return E.Left[DeployState](fmt.Errorf("unsupported runtime: %s", fn.Runtime))
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

		// Handle result
		if E.IsLeft(result) {
			err := E.Fold(
				func(e error) error { return e },
				func(a build.Artifact) error { return nil },
			)(result)
			return E.Left[DeployState](fmt.Errorf("failed to build %s: %w", fn.Name, err))
		}

		// Extract artifact
		artifact := E.Fold(
			func(e error) build.Artifact { return build.Artifact{} },
			func(a build.Artifact) build.Artifact { return a },
		)(result)

		// Store artifact
		state.Artifacts[fn.Name] = artifact

		sizeMB := float64(artifact.Size) / 1024 / 1024
		fmt.Printf("[%s] ✓ Built: %s (%.2f MB)\n", fn.Name, filepath.Base(artifact.Path), sizeMB)
	}

	fmt.Println()
	return E.Right[error](state)
}

// terraformInitStage initializes Terraform
func terraformInitStage(ctx context.Context, state DeployState) E.Either[error, DeployState] {
	fmt.Println("==> Initializing Terraform...")

	infraDir := filepath.Join(state.Config.ProjectRoot, "infra")
	if _, err := os.Stat(infraDir); os.IsNotExist(err) {
		return E.Left[DeployState](fmt.Errorf("infra/ directory not found"))
	}

	exec := terraform.NewExecutor(findTerraformPath())
	opts := terraform.InitOptions{
		WorkingDir: infraDir,
		Upgrade:    false,
	}

	if err := exec.Init(ctx, opts); err != nil {
		return E.Left[DeployState](fmt.Errorf("terraform init failed: %w", err))
	}

	return E.Right[error](state)
}

// terraformPlanStage creates Terraform plan
func terraformPlanStage(ctx context.Context, state DeployState) E.Either[error, DeployState] {
	fmt.Println("==> Planning infrastructure changes...")

	infraDir := filepath.Join(state.Config.ProjectRoot, "infra")
	exec := terraform.NewExecutor(findTerraformPath())

	// Set namespace variable if provided
	var tfVars map[string]string
	if state.Config.Namespace != "" {
		tfVars = map[string]string{
			"namespace": state.Config.Namespace + "-",
		}
		fmt.Printf("Deploying to namespace: %s\n", state.Config.Namespace)
	}

	opts := terraform.PlanOptions{
		WorkingDir: infraDir,
		Out:        filepath.Join(infraDir, "tfplan"),
		Vars:       tfVars,
	}

	planResult, err := exec.Plan(ctx, opts)
	if err != nil {
		return E.Left[DeployState](fmt.Errorf("terraform plan failed: %w", err))
	}

	if planResult.Changes == 0 {
		fmt.Println("No changes detected")
		return E.Right[error](state)
	}

	fmt.Printf("\nPlan: %d to add, %d to change, %d to destroy\n",
		planResult.Add, planResult.Change, planResult.Destroy)

	return E.Right[error](state)
}

// terraformApplyStage applies Terraform changes
func terraformApplyStage(ctx context.Context, state DeployState) E.Either[error, DeployState] {
	infraDir := filepath.Join(state.Config.ProjectRoot, "infra")
	planFile := filepath.Join(infraDir, "tfplan")

	// Check if plan has changes (plan file exists)
	if _, err := os.Stat(planFile); os.IsNotExist(err) {
		// No plan file means no changes
		return E.Right[error](state)
	}

	// Request approval if not auto-approved
	if !state.Config.AutoApprove {
		fmt.Print("\nDo you want to apply these changes? (yes/no): ")
		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			return E.Left[DeployState](fmt.Errorf("deployment cancelled by user"))
		}
	}

	fmt.Println("==> Applying infrastructure changes...")

	exec := terraform.NewExecutor(findTerraformPath())
	opts := terraform.ApplyOptions{
		WorkingDir:  infraDir,
		PlanFile:    planFile,
		AutoApprove: true, // Already approved above
	}

	if err := exec.Apply(ctx, opts); err != nil {
		return E.Left[DeployState](fmt.Errorf("terraform apply failed: %w", err))
	}

	return E.Right[error](state)
}

// terraformOutputsStage captures Terraform outputs
func terraformOutputsStage(ctx context.Context, state DeployState) E.Either[error, DeployState] {
	infraDir := filepath.Join(state.Config.ProjectRoot, "infra")
	exec := terraform.NewExecutor(findTerraformPath())

	opts := terraform.OutputOptions{
		WorkingDir: infraDir,
	}

	outputs, err := exec.Outputs(ctx, opts)
	if err != nil {
		// Non-fatal - just warn and continue
		fmt.Printf("Warning: failed to retrieve outputs: %v\n", err)
		return E.Right[error](state)
	}

	state.TFOutputs = outputs
	return E.Right[error](state)
}
