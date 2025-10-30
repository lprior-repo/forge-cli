package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

// NewVersionCmd creates the 'version' command
func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long: `
╭──────────────────────────────────────────────────────────────╮
│  ℹ️  Forge Version Info                                     │
╰──────────────────────────────────────────────────────────────╯

Display version information and build details.
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("")
			fmt.Println("╭─────────────────────────────────────╮")
			fmt.Println("│  Forge - Serverless Infrastructure  │")
			fmt.Println("╰─────────────────────────────────────╯")
			fmt.Println("")
			fmt.Printf("  Version:  %s\n", version)
			fmt.Println("  License:  MIT")
			fmt.Println("  Repo:     https://github.com/lewis/forge")
			fmt.Println("")
			fmt.Println("💡 Check for updates:")
			fmt.Println("   git pull origin master")
			fmt.Println("")
			fmt.Println("📚 Documentation:")
			fmt.Println("   forge --help")
			fmt.Println("")
		},
	}

	return cmd
}
