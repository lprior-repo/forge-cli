package cli

import (
	"context"
	"fmt"
	"os"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/terraform"
	"github.com/lewis/forge/internal/ui"
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
		Long: `
╭──────────────────────────────────────────────────────────────╮
│  🚀 Forge Deploy                                            │
╰──────────────────────────────────────────────────────────────╯

Build Lambda functions and deploy infrastructure with Terraform.
One command to go from code to production AWS resources.

🎯 What It Does:
  1. Scans src/functions/* for Lambda functions
  2. Auto-detects runtimes (Go, Python, Node.js)
  3. Builds deployment packages
  4. Runs terraform init/plan/apply in infra/
  5. Outputs deployed URLs and resources

🌟 Namespace Support (PR Previews):
  Deploy to isolated ephemeral environments for testing:

  forge deploy --namespace=pr-123
    → Sets TF_VAR_namespace=pr-123
    → All resources prefixed: my-app-pr-123-*
    → Completely isolated AWS environment
    → Perfect for PR preview deployments

🚀 Examples:

  # Deploy to production (default environment)
  forge deploy

  # Deploy to ephemeral PR environment
  forge deploy --namespace=pr-123

  # Non-interactive deployment (CI/CD)
  forge deploy --auto-approve

  # Deploy to specific region
  forge deploy --region=us-west-2

💡 Pro Tips:
  • Use namespaces for PR preview environments
  • Each namespace has isolated Terraform state
  • Combine with GitHub Actions for automatic PR deploys
  • Use --auto-approve in CI/CD pipelines

📋 Requirements:
  • Terraform installed (terraform version)
  • AWS credentials configured
  • infra/ directory with Terraform config
  • src/functions/ with Lambda code
`,
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
	out := ui.DefaultOutput()

	ctx := context.Background()
	projectRoot, err := os.Getwd()
	if err != nil {
		out.Error("Failed to get current directory: %v", err)
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	out.Header("Deploying Lambda Functions")
	if namespace != "" {
		out.Info("Deploying to namespace: %s", namespace)
	}

	// Create Terraform executor using pure functional composition
	tfPath := findTerraformPath()
	tfExec := terraform.NewExecutor(tfPath)
	tfExecutor := adaptTerraformExecutor(tfExec)

	// Compose functional pipeline using event-based stages:
	// Scan → Stubs → Build → TF Init → TF Plan → TF Apply → TF Outputs
	// Event-based stages return events as data instead of printing
	deployPipeline := pipeline.NewEventPipeline(
		pipeline.ConventionScanV2(),
		pipeline.ConventionStubsV2(),
		pipeline.ConventionBuildV2(),
		pipeline.ConventionTerraformInitV2(tfExecutor),
		pipeline.ConventionTerraformPlanV2(tfExecutor, namespace),
		pipeline.ConventionTerraformApplyV2(tfExecutor, autoApprove),
		pipeline.ConventionTerraformOutputsV2(tfExecutor),
	)

	// Initial state (immutable)
	initialState := pipeline.State{
		ProjectDir: projectRoot,
		Artifacts:  make(map[string]pipeline.Artifact),
		Outputs:    make(map[string]interface{}),
		Config:     nil, // Will hold discovered functions
	}

	// Run event-based pipeline (returns Either[error, StageResult])
	result := pipeline.RunWithEvents(deployPipeline, ctx, initialState)

	// Handle result using functional pattern with StageResult
	return E.Fold(
		func(err error) error {
			out.Error("Deployment failed: %v", err)
			out.Print("")
			out.Warning("Troubleshooting tips:")
			out.Print("  • Check that Terraform is installed: terraform version")
			out.Print("  • Verify AWS credentials are configured: aws sts get-caller-identity")
			out.Print("  • Review function build logs in .forge/build/")
			out.Print("  • Run 'forge build' separately to test builds")
			return fmt.Errorf("deployment failed: %w", err)
		},
		func(stageResult pipeline.StageResult) error {
			// Print all collected events
			pipeline.PrintEvents(stageResult.Events)

			out.Success("Deployment completed successfully")
			if namespace != "" {
				out.Info("Namespace: %s", namespace)
			}
			if len(stageResult.State.Outputs) > 0 {
				out.Header("Terraform Outputs")
				for key, value := range stageResult.State.Outputs {
					out.Print("  %s = %v", key, value)
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
