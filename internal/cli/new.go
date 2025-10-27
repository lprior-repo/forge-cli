package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lewis/forge/internal/scaffold"
	"github.com/spf13/cobra"
)

// NewNewCmd creates the 'new' command
func NewNewCmd() *cobra.Command {
	var (
		projectName string
		stackName   string
		runtime     string
		description string
	)

	cmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Forge project or stack",
		Long: `Create a new Forge project with initial configuration,
or add a new stack to an existing project.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine if this is a new project or new stack
			isNewProject := len(args) > 0

			if isNewProject {
				projectName = args[0]
				return createProject(projectName, runtime)
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

	return cmd
}

// createProject creates a new Forge project
func createProject(name, defaultRuntime string) error {
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

	fmt.Printf("Created Forge project: %s\n", name)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  forge new --stack api --runtime %s\n", defaultRuntime)
	fmt.Printf("  forge deploy\n")

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
		opts.Description = fmt.Sprintf("%s Lambda function", name)
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
