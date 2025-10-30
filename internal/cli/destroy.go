package cli

import (
	"context"
	"fmt"
	"os"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/config"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/terraform"
	"github.com/spf13/cobra"
)

// NewDestroyCmd creates the 'destroy' command
func NewDestroyCmd() *cobra.Command {
	var autoApprove bool

	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy infrastructure",
		Long:  `Destroy infrastructure with Terraform in the current directory.`,
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

	fmt.Println("Destroying infrastructure...")

	if !autoApprove {
		fmt.Print("\nThis will destroy all resources. Continue? (yes/no): ")
		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("Destroy cancelled")
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

	// Handle result
	if E.IsLeft(result) {
		err := E.Fold(
			func(e error) error { return e },
			func(s pipeline.State) error { return nil },
		)(result)
		return err
	}

	fmt.Println("\n✓ Infrastructure destroyed")
	return nil
}
