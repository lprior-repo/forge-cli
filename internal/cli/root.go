package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	verbose bool
	region  string
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forge",
		Short: "Forge - Serverless infrastructure tool for AWS Lambda",
		Long: `Forge is a tool for building and deploying serverless applications on AWS Lambda.
It combines the power of Terraform with streamlined Lambda deployment workflows.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().StringVarP(&region, "region", "r", "", "AWS region (overrides forge.hcl)")

	// Add subcommands
	cmd.AddCommand(
		NewNewCmd(),
		NewBuildCmd(),
		NewDeployCmd(),
		NewDestroyCmd(),
		NewVersionCmd(),
	)

	return cmd
}

// Execute runs the root command
func Execute() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
