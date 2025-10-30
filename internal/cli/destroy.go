package cli

import (
	"context"
	"fmt"
	"os"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/config"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/terraform"
	"github.com/lewis/forge/internal/ui"
	"github.com/spf13/cobra"
)

// NewDestroyCmd creates the 'destroy' command
func NewDestroyCmd() *cobra.Command {
	var autoApprove bool

	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy infrastructure and clean up AWS resources",
		Long: `
╭──────────────────────────────────────────────────────────────╮
│  💥 Forge Destroy                                           │
╰──────────────────────────────────────────────────────────────╯

Safely tear down all AWS resources managed by Terraform.
Includes confirmation prompts to prevent accidental deletion.

⚠️  Warning: This Action is Destructive
  This command will PERMANENTLY DELETE all infrastructure defined
  in your infra/ directory, including:
  • Lambda functions
  • API Gateways
  • DynamoDB tables
  • S3 buckets (if configured for deletion)
  • IAM roles and policies
  • CloudWatch log groups

🛡️  Safety Features:
  • Interactive confirmation required by default
  • Shows resource plan before destruction
  • Requires --auto-approve to skip confirmation
  • Dry-run with 'terraform plan -destroy' first

🚀 Examples:

  # Interactive destroy with confirmation
  forge destroy

  # Non-interactive (CI/CD, cleanup scripts)
  forge destroy --auto-approve

  # Destroy specific namespace (PR cleanup)
  forge destroy --namespace=pr-123 --auto-approve

💡 Pro Tips:
  • Always review the plan before confirming
  • Use namespaces to destroy only preview environments
  • Backup important data before destroying
  • Consider using 'terraform state' commands for partial cleanup

📋 Recommended Workflow:
  1. Review what will be destroyed:
     cd infra && terraform plan -destroy

  2. If satisfied, run:
     forge destroy

  3. Confirm when prompted
`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDestroy(autoApprove)
		},
	}

	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip interactive approval")

	return cmd
}

// runDestroy executes the destroy operation
func runDestroy(autoApprove bool) error {
	out := ui.DefaultOutput()
	prompter := ui.NewPrompter(os.Stdin, os.Stdout)

	ctx := context.Background()
	projectRoot, err := os.Getwd()
	if err != nil {
		out.Error("Failed to get current directory: %v", err)
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load config
	cfg, err := config.Load(projectRoot)
	if err != nil {
		out.Error("Failed to load configuration: %v", err)
		out.Warning("Make sure you're in a Forge project directory")
		out.Print("  • Check that forge.hcl exists")
		out.Print("  • Run 'forge new' to create a new project")
		return fmt.Errorf("failed to load config: %w", err)
	}

	out.Header("Destroying Infrastructure")
	out.Warning("This will destroy all AWS resources managed by this project")

	// Require explicit confirmation for destructive action
	if !autoApprove {
		if !prompter.ConfirmDestruction(
			"You are about to PERMANENTLY DELETE all infrastructure",
			projectRoot,
		) {
			out.Info("Destroy canceled")
			return nil
		}
	}

	// Create functional executor
	tfPath := findTerraformPath()
	exec := terraform.NewExecutor(tfPath)

	// Create destroy pipeline
	destroyPipeline := pipeline.New(
		pipeline.TerraformDestroy(exec, true),
	)

	// Initial state
	initialState := pipeline.State{
		ProjectDir: projectRoot,
		Config:     cfg,
	}

	// Run pipeline
	result := pipeline.Run(destroyPipeline, ctx, initialState)

	// Handle result using functional pattern
	return E.Fold(
		func(err error) error {
			out.Error("Destroy failed: %v", err)
			out.Print("")
			out.Warning("Troubleshooting tips:")
			out.Print("  • Check Terraform state is accessible")
			out.Print("  • Verify AWS credentials are valid")
			out.Print("  • Review .terraform/ directory for issues")
			out.Print("  • Try running 'terraform destroy' manually in infra/")
			return fmt.Errorf("destroy failed: %w", err)
		},
		func(finalState pipeline.State) error {
			out.Success("Infrastructure destroyed successfully")
			out.Print("")
			out.Dim("All AWS resources have been removed")
			return nil
		},
	)(result)
}
