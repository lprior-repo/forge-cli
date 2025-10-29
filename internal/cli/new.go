package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lewis/forge/internal/scaffold"
	"github.com/lewis/forge/internal/state"
	"github.com/lewis/forge/internal/terraform"
	"github.com/spf13/cobra"
)

// NewNewCmd creates the 'new' command
func NewNewCmd() *cobra.Command {
	var (
		projectName string
		stackName   string
		runtime     string
		description string
		autoState   bool
	)

	cmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Forge project or stack",
		Long: `Create a new Forge project with initial configuration,
or add a new stack to an existing project.

Auto-state provisioning:
  forge new my-app --auto-state
    → Auto-provisions S3 bucket for Terraform state
    → Auto-provisions DynamoDB table for state locking
    → Generates backend.tf with namespace-aware configuration
    → Production-ready state management from day 1`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine if this is a new project or new stack
			isNewProject := len(args) > 0

			if isNewProject {
				projectName = args[0]
				return createProject(projectName, runtime, autoState)
			}

			// New stack in existing project
			if stackName == "" {
				return fmt.Errorf("--stack flag is required when creating a new stack")
			}

			return createStack(stackName, runtime, description)
		},
	}

	cmd.Flags().StringVar(&stackName, "stack", "", "Create a new stack in existing project")
	cmd.Flags().StringVar(&runtime, "runtime", "go1.x", "Runtime for the stack (go1.x, python3.11, nodejs20.x)")
	cmd.Flags().StringVar(&description, "description", "", "Stack description")
	cmd.Flags().BoolVar(&autoState, "auto-state", false, "Auto-provision S3 bucket and DynamoDB table for Terraform state")

	return cmd
}

// createProject creates a new Forge project
func createProject(name, defaultRuntime string, autoState bool) error {
	projectDir := filepath.Join(".", name)

	// Check if directory already exists
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", name)
	}

	// Create generator
	gen, err := scaffold.NewGenerator(projectDir)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	// Generate project structure
	opts := &scaffold.ProjectOptions{
		Name:   name,
		Region: region,
	}
	if opts.Region == "" {
		opts.Region = "us-east-1"
	}

	if err := gen.GenerateProject(opts); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	fmt.Printf("✓ Created Forge project: %s\n", name)

	// Auto-provision state backend if requested
	if autoState {
		fmt.Println("\nProvisioning Terraform state backend...")
		if err := provisionStateBackend(projectDir, name, opts.Region); err != nil {
			fmt.Printf("Warning: Failed to provision state backend: %v\n", err)
			fmt.Println("You can manually set up state later or run with --auto-state again")
		} else {
			fmt.Println("✓ State backend provisioned successfully")
		}
	}

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  # Add your Lambda function code to src/functions/\n")
	fmt.Printf("  forge build\n")
	fmt.Printf("  forge deploy\n")

	if !autoState {
		fmt.Println("\nOptional: Set up remote state")
		fmt.Printf("  forge new --auto-state  # Provision S3 + DynamoDB for state\n")
	}

	return nil
}

// provisionStateBackend provisions S3 bucket and DynamoDB table for Terraform state
// This is the imperative shell that orchestrates state backend provisioning
func provisionStateBackend(projectDir, projectName, region string) error {
	ctx := context.Background()

	// Create Terraform executor
	tfPath := findTerraformPath()
	tfExec := terraform.NewExecutor(tfPath)

	// Provision state backend (uses Railway-Oriented Programming internally)
	result, err := state.ProvisionStateBackendSync(ctx, projectDir, projectName, region, tfExec)
	if err != nil {
		return err
	}

	// Display results
	fmt.Printf("  S3 Bucket: %s\n", result.BucketName)
	fmt.Printf("  DynamoDB Table: %s\n", result.TableName)
	fmt.Printf("  Backend Config: %s\n", result.BackendTFPath)

	return nil
}

// createStack creates a new stack in the current project
func createStack(name, runtime, desc string) error {
	// Verify we're in a Forge project
	if _, err := os.Stat("forge.hcl"); os.IsNotExist(err) {
		return fmt.Errorf("not in a Forge project (forge.hcl not found)")
	}

	// Get current directory as project root
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create generator
	gen, err := scaffold.NewGenerator(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	// Generate stack
	opts := &scaffold.StackOptions{
		Name:        name,
		Runtime:     runtime,
		Description: desc,
	}
	if opts.Description == "" {
		_, _, _ = opts.Description, fmt.Sprintf, name
	}

	if err := gen.GenerateStack(opts); err != nil {
		return fmt.Errorf("failed to generate stack: %w", err)
	}

	fmt.Printf("Created stack: %s\n", name)
	fmt.Printf("Runtime: %s\n", runtime)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  # Edit your function code\n")
	fmt.Printf("  cd ..\n")
	fmt.Printf("  forge deploy %s\n", name)

	return nil
}
