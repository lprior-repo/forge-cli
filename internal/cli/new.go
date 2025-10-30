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
		Short: "Create a new Forge project with zero configuration",
		Long: `
╭──────────────────────────────────────────────────────────────╮
│  🔨 Forge Project Generator                                 │
╰──────────────────────────────────────────────────────────────╯

Create a new serverless project with convention-over-configuration.
No YAML files, no config templates - just pure Terraform + smart defaults.

🎯 What you get:
  ✓ Convention-based project structure (src/functions/*)
  ✓ Auto-detected runtimes (Go, Python, Node.js)
  ✓ Production-ready Terraform templates
  ✓ Optional remote state with --auto-state
  ✓ Namespace support for PR preview environments

📦 Project Structure Created:
  my-app/
  ├── infra/              # Terraform infrastructure (edit freely!)
  │   ├── main.tf         # Lambda resources
  │   ├── variables.tf    # Input variables
  │   └── outputs.tf      # Output values
  └── src/
      └── functions/      # Lambda functions (auto-discovered)
          └── api/        # Example function
              └── main.go # Entry point

🚀 Quick Start Examples:

  # Minimal project (local state)
  forge new my-app

  # Production-ready project (remote state)
  forge new my-app --auto-state
    → Creates S3 bucket: my-app-terraform-state
    → Creates DynamoDB table: my-app-state-lock
    → Generates backend.tf with encryption
    → Ready for team collaboration!

  # Specialized Lambda projects
  forge new lambda my-api
    → Full Lambda + API Gateway setup
    → Python/Go/Node.js support
    → Fast uv-based builds

💡 Pro Tips:
  • Use --auto-state for team projects and CI/CD
  • State is namespace-aware for PR environments
  • Customize infra/ Terraform files as needed
  • No lock-in - it's just Terraform underneath

📖 Next Steps After Creation:
  1. cd my-app
  2. Add code to src/functions/
  3. forge build          # Build all functions
  4. forge deploy         # Deploy to AWS
  5. forge deploy --namespace=pr-123  # Ephemeral preview env
`,
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

	// Add lambda subcommand
	cmd.AddCommand(NewLambdaCmd())

	return cmd
}

// createProject creates a new Forge project
func createProject(name, defaultRuntime string, autoState bool) error {
	projectDir := filepath.Join(".", name)

	// Check if directory already exists
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", name)
	}

	// Generate project structure (pure functional - no OOP)
	// Detect AWS region from environment
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = "us-east-1"
	}

	opts := &scaffold.ProjectOptions{
		Name:   name,
		Region: region,
	}

	if err := scaffold.GenerateProject(projectDir, opts); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Success banner
	fmt.Println("")
	fmt.Println("╭────────────────────────────────────────────────────────────╮")
	fmt.Println("│                                                            │")
	fmt.Println("│  🎉 Success! Your Forge project is ready                  │")
	fmt.Println("│                                                            │")
	fmt.Println("╰────────────────────────────────────────────────────────────╯")
	fmt.Println("")
	fmt.Printf("✨ Created: %s/\n", name)
	fmt.Println("")
	fmt.Println("📁 Project Structure:")
	fmt.Printf("   %s/\n", name)
	fmt.Println("   ├── infra/              # Terraform infrastructure")
	fmt.Println("   │   ├── main.tf         # Lambda resources")
	fmt.Println("   │   ├── variables.tf    # Input variables")
	fmt.Println("   │   └── outputs.tf      # Output values")
	fmt.Println("   └── src/")
	fmt.Println("       └── functions/      # Your Lambda functions")
	fmt.Println("           └── api/        # Example function")
	fmt.Println("               └── main.go # Entry point")
	fmt.Println("")

	// Auto-provision state backend if requested
	if autoState {
		fmt.Println("🔄 Provisioning Terraform state backend...")
		fmt.Println("")
		if err := provisionStateBackend(projectDir, name, opts.Region); err != nil {
			fmt.Printf("⚠️  Warning: Failed to provision state backend: %v\n", err)
			fmt.Println("💡 You can manually set up state later or re-run:")
			fmt.Printf("   forge new --auto-state\n")
		} else {
			fmt.Println("✅ State backend provisioned successfully")
			fmt.Println("🔒 Your Terraform state is now encrypted and locked")
		}
		fmt.Println("")
	}

	// Next steps with clear visual hierarchy
	fmt.Println("🚀 Next Steps:")
	fmt.Println("")
	fmt.Println("   1. Navigate to your project:")
	fmt.Printf("      cd %s\n", name)
	fmt.Println("")
	fmt.Println("   2. Add your Lambda function code:")
	fmt.Println("      # Edit src/functions/api/main.go")
	fmt.Println("      # Or add new functions in src/functions/")
	fmt.Println("")
	fmt.Println("   3. Build your functions:")
	fmt.Println("      forge build")
	fmt.Println("")
	fmt.Println("   4. Deploy to AWS:")
	fmt.Println("      forge deploy")
	fmt.Println("")

	if !autoState {
		fmt.Println("💡 Pro Tip: For team projects, set up remote state:")
		fmt.Println("   forge new --auto-state")
		fmt.Println("   → Auto-creates S3 bucket + DynamoDB table")
		fmt.Println("   → Enables team collaboration & CI/CD")
		fmt.Println("")
	}

	fmt.Println("📚 Need Help?")
	fmt.Println("   forge --help        # All commands")
	fmt.Println("   forge build --help  # Build documentation")
	fmt.Println("   forge deploy --help # Deployment options")
	fmt.Println("")
	fmt.Println("✨ Happy building!")

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

	// Generate stack (pure functional - no OOP)
	opts := &scaffold.StackOptions{
		Name:        name,
		Runtime:     runtime,
		Description: desc,
	}
	if opts.Description == "" {
		_, _, _ = opts.Description, fmt.Sprintf, name
	}

	if err := scaffold.GenerateStack(projectRoot, opts); err != nil {
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
