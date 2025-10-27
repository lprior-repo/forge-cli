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
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Forge version %s\n", version)
		},
	}

	return cmd
}
