package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/lewis/forge/internal/config"
	"github.com/lewis/forge/internal/stack"
	"github.com/lewis/forge/internal/terraform"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the 'init' command
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Terraform for all stacks",
		Long:  `Run terraform init on all stacks in the project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}

	return cmd
}

// runInit initializes all stacks
func runInit() error {
	ctx := context.Background()
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load config
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find all stacks
	detector := stack.NewDetector(projectRoot)
	stacks, err := detector.FindStacks()
	if err != nil {
		return fmt.Errorf("failed to find stacks: %w", err)
	}

	if len(stacks) == 0 {
		fmt.Println("No stacks found")
		return nil
	}

	fmt.Printf("Initializing %d stack(s)...\n", len(stacks))

	// Create terraform executor
	exec := terraform.NewExecutor("terraform")

	// Initialize each stack
	for _, st := range stacks {
		fmt.Printf("\n[%s] Initializing...\n", st.Name)

		// Set AWS region env var if needed
		if region != "" || cfg.Project.Region != "" {
			awsRegion := region
			if awsRegion == "" {
				awsRegion = cfg.Project.Region
			}
			os.Setenv("AWS_REGION", awsRegion)
		}

		if err := exec.Init(ctx, st.AbsPath); err != nil {
			return fmt.Errorf("failed to init %s: %w", st.Name, err)
		}

		fmt.Printf("[%s] ✓ Initialized\n", st.Name)
	}

	fmt.Println("\n✓ All stacks initialized")
	return nil
}
