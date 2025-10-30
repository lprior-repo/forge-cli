package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags.
	verbose bool
	region  string
)

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forge",
		Short: "Forge - Convention-over-configuration Lambda deployment",
		Long: `
╭─────────────────────────────────────────────────────────────────╮
│                                                                 │
│   ████████╗ ███████╗ ██████╗  ██████╗  ███████╗              │
│   ██╔════╝ ██╔═══██║ ██╔══██╗ ██╔════╝ ██╔════╝               │
│   █████╗   ██║   ██║ ██████╔╝ ██║  ███╗█████╗                 │
│   ██╔══╝   ██║   ██║ ██╔══██╗ ██║   ██║██╔══╝                 │
│   ██║      ╚██████╔╝ ██║  ██║ ╚██████╔╝███████╗               │
│   ╚═╝       ╚═════╝  ╚═╝  ╚═╝  ╚═════╝ ╚══════╝               │
│                                                                 │
│   Convention-over-configuration Lambda deployment              │
│   Terraform + Serverless = Zero Config                         │
│                                                                 │
╰─────────────────────────────────────────────────────────────────╯

Forge combines the power of Terraform with zero-config Lambda workflows.

🎯 What makes Forge different:
  • No forge.yaml, serverless.yml, or config files - just conventions
  • Full Terraform control - edit .tf files directly when needed
  • Built-in PR preview environments - test changes in isolation
  • Auto-detect runtimes from code structure (Go, Python, Node.js)
  • Production-ready from day 1 - state management, CI/CD ready

🚀 Quick Start:
  forge new my-app --auto-state    # Create project with remote state
  cd my-app
  forge deploy                      # Build + deploy in one command

📖 Philosophy:
  Convention over configuration (Omakase)
  Pure Terraform power with zero lock-in
  No magic, maximum control

🔗 Learn more: https://github.com/lewis/forge
`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			// Show the long description as a welcome message
			fmt.Println(cmd.Long)
			fmt.Println("\nRun 'forge --help' to see available commands")
			fmt.Println("Run 'forge new --help' to get started")
		},
	}

	// Global flags
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().StringVarP(&region, "region", "r", "", "AWS region (overrides forge.hcl)")

	// Add subcommands
	cmd.AddCommand(
		NewNewCmd(),
		NewAddCmd(),
		NewBuildCmd(),
		NewDeployCmd(),
		NewDestroyCmd(),
		NewVersionCmd(),
	)

	return cmd
}

// Execute runs the root command.
func Execute() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
