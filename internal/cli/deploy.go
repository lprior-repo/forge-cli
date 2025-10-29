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

	// Create Terraform executor (adapts to pipeline interface)
	tfPath := findTerraformPath()
	tfExec := terraform.NewExecutor(tfPath)
	exec := newTerraformAdapter(tfExec)

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
	result := deployPipeline.Run(ctx, initialState)

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

// terraformAdapter adapts terraform.Executor to pipeline.TerraformExecutor interface
type terraformAdapter struct {
	exec terraform.Executor
}

func newTerraformAdapter(exec terraform.Executor) pipeline.TerraformExecutor {
	return &terraformAdapter{exec: exec}
}

func (a *terraformAdapter) Init(ctx context.Context, dir string) error {
	return a.exec.Init(ctx, dir, terraform.Upgrade(false))
}

func (a *terraformAdapter) Plan(ctx context.Context, dir string) (bool, error) {
	return a.PlanWithVars(ctx, dir, nil)
}

func (a *terraformAdapter) PlanWithVars(ctx context.Context, dir string, vars map[string]string) (bool, error) {
	var opts []terraform.PlanOption
	opts = append(opts, terraform.PlanOut(dir+"/tfplan"))

	// Add variables as options
	for k, v := range vars {
		opts = append(opts, terraform.PlanVar(k, v))
	}

	return a.exec.Plan(ctx, dir, opts...)
}

func (a *terraformAdapter) Apply(ctx context.Context, dir string) error {
	return a.exec.Apply(ctx, dir,
		terraform.ApplyPlanFile(dir+"/tfplan"),
		terraform.AutoApprove(true), // Already approved in pipeline stage
	)
}

func (a *terraformAdapter) Output(ctx context.Context, dir string) (map[string]interface{}, error) {
	return a.exec.Output(ctx, dir)
}
