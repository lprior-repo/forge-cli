package cli

import (
	"context"
	"fmt"
	"os"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/config"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/stack"
	"github.com/lewis/forge/internal/terraform"
	"github.com/spf13/cobra"
)

// NewDestroyCmd creates the 'destroy' command
func NewDestroyCmd() *cobra.Command {
	var autoApprove bool

	cmd := &cobra.Command{
		Use:   "destroy [stack-name]",
		Short: "Destroy infrastructure",
		Long: `Destroy infrastructure with Terraform.
If no stack name is provided, destroys all stacks in reverse dependency order.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var targetStack string
			if len(args) > 0 {
				targetStack = args[0]
			}
			return runDestroy(targetStack, autoApprove)
		},
	}

	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip interactive approval")

	return cmd
}

// runDestroy executes the destroy operation
func runDestroy(targetStack string, autoApprove bool) error {
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
	allStacks, err := detector.FindStacks()
	if err != nil {
		return fmt.Errorf("failed to find stacks: %w", err)
	}

	if len(allStacks) == 0 {
		return fmt.Errorf("no stacks found")
	}

	// Filter to target stack if specified
	var stacksToDestroy []*stack.Stack
	if targetStack != "" {
		found := false
		for _, st := range allStacks {
			if st.Name == targetStack {
				stacksToDestroy = []*stack.Stack{st}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("stack not found: %s", targetStack)
		}
	} else {
		stacksToDestroy = allStacks
	}

	// Build dependency graph and get reverse order for destroy
	graph, err := stack.NewGraph(stacksToDestroy)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	orderedStacks, err := graph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to sort stacks: %w", err)
	}

	// Reverse order for destroy (dependencies last)
	for i := 0; i < len(orderedStacks)/2; i++ {
		j := len(orderedStacks) - 1 - i
		orderedStacks[i], orderedStacks[j] = orderedStacks[j], orderedStacks[i]
	}

	fmt.Printf("Destroying %d stack(s) in order: ", len(orderedStacks))
	for i, st := range orderedStacks {
		if i > 0 {
			fmt.Print(" → ")
		}
		fmt.Print(st.Name)
	}
	fmt.Println()

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
		Stacks:     orderedStacks,
		Config:     cfg,
	}

	// Run pipeline
	result := destroyPipeline.Run(ctx, initialState)

	// Handle result
	if E.IsLeft(result) {
		err := E.Fold(
			func(e error) error { return e },
			func(s pipeline.State) error { return nil },
		)(result)
		return err
	}

	fmt.Println("\n✓ All stacks destroyed")
	return nil
}
