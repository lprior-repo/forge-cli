package cli

import (
	"context"
	"fmt"
	"os"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/terraform"
	"github.com/spf13/cobra"
)

// NewDeployCmd creates the 'deploy' command using convention-based discovery
func NewDeployCmd() *cobra.Command {
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
			return runDeploy(autoApprove, namespace)
		},
	}

	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip interactive approval")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace for ephemeral environments (e.g., pr-123)")

	return cmd
}

// runDeploy executes the deployment using functional pipeline composition
func runDeploy(autoApprove bool, namespace string) error {
	ctx := context.Background()
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create Terraform executor using pure functional composition
	tfPath := findTerraformPath()
	tfExec := terraform.NewExecutor(tfPath)
	exec := adaptTerraformExecutor(tfExec)

	// Compose functional pipeline:
	// Scan → Stubs → Build → TF Init → TF Plan → TF Apply → TF Outputs
	deployPipeline := pipeline.New(
		pipeline.ConventionScan(),
		pipeline.ConventionStubs(),
		pipeline.ConventionBuild(),
		pipeline.ConventionTerraformInit(exec),
		pipeline.ConventionTerraformPlan(exec, namespace),
		pipeline.ConventionTerraformApply(exec, autoApprove),
		pipeline.ConventionTerraformOutputs(exec),
	)

	// Initial state (immutable)
	initialState := pipeline.State{
		ProjectDir: projectRoot,
		Stacks:     nil, // Not used in convention mode
		Artifacts:  make(map[string]pipeline.Artifact),
		Outputs:    make(map[string]interface{}),
		Config:     nil, // Will hold discovered functions
	}

	// Run pipeline (returns Either[error, State])
	result := pipeline.Run(deployPipeline, ctx, initialState)

	// Handle result using functional pattern
	return E.Fold(
		func(err error) error {
			return fmt.Errorf("deployment failed: %w", err)
		},
		func(finalState pipeline.State) error {
			fmt.Println("\n✓ Deployment successful")
			if namespace != "" {
				fmt.Printf("Namespace: %s\n", namespace)
			}
			if len(finalState.Outputs) > 0 {
				fmt.Println("\nOutputs:")
				for key, value := range finalState.Outputs {
					fmt.Printf("  %s = %v\n", key, value)
				}
			}
			return nil
		},
	)(result)
}

// findTerraformPath finds the terraform binary
func findTerraformPath() string {
	// For now, assume terraform is in PATH
	// TODO: Add logic to find terraform binary
	return "terraform"
}

// adaptTerraformExecutor adapts terraform.Executor to pipeline.TerraformExecutor
// Pure functional approach: returns a struct with function fields, NO METHODS!
func adaptTerraformExecutor(exec terraform.Executor) pipeline.TerraformExecutor {
	return pipeline.TerraformExecutor{
		// Init function - no mutation, pure composition
		Init: func(ctx context.Context, dir string) error {
			return exec.Init(ctx, dir, terraform.Upgrade(false))
		},

		// Plan function - calls PlanWithVars with nil vars
		Plan: func(ctx context.Context, dir string) (bool, error) {
			opts := []terraform.PlanOption{terraform.PlanOut(dir + "/tfplan")}
			return exec.Plan(ctx, dir, opts...)
		},

		// PlanWithVars function - adds variable options functionally
		PlanWithVars: func(ctx context.Context, dir string, vars map[string]string) (bool, error) {
			opts := []terraform.PlanOption{terraform.PlanOut(dir + "/tfplan")}
			// Functional transformation: vars → options
			for k, v := range vars {
				opts = append(opts, terraform.PlanVar(k, v))
			}
			return exec.Plan(ctx, dir, opts...)
		},

		// Apply function - pure function call, no state
		Apply: func(ctx context.Context, dir string) error {
			return exec.Apply(ctx, dir,
				terraform.ApplyPlanFile(dir+"/tfplan"),
				terraform.AutoApprove(true),
			)
		},

		// Output function - pure retrieval
		Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
			return exec.Output(ctx, dir)
		},
	}
}
