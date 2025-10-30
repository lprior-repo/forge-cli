package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
	"github.com/lewis/forge/internal/build"
	"github.com/lewis/forge/internal/discovery"
	"github.com/lewis/forge/internal/ui"
	"github.com/spf13/cobra"
)

// NewBuildCmd creates the 'build' command
func NewBuildCmd() *cobra.Command {
	var (
		stubOnly bool
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build Lambda functions using convention-based discovery",
		Long: `Scans src/functions/* to discover Lambda functions and builds them.

Conventions:
  - Function name = directory name
  - Runtime detected from entry file:
    - main.go or *.go → Go (provided.al2023)
    - index.js/handler.js → Node.js (nodejs20.x)
    - app.py/lambda_function.py → Python (python3.13)
  - Output: .forge/build/{name}.zip

Examples:
  forge build              # Build all functions
  forge build --stub-only  # Create empty stub zips only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(stubOnly)
		},
	}

	cmd.Flags().BoolVar(&stubOnly, "stub-only", false, "Create stub zips without building")

	return cmd
}

// runBuild executes the build process
func runBuild(stubOnly bool) error {
	out := ui.DefaultOutput()

	ctx := context.Background()
	projectRoot, err := os.Getwd()
	if err != nil {
		out.Error("Failed to get current directory: %v", err)
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	out.Header("Building Lambda Functions")

	// Scan for functions (pure functional approach - no OOP)
	functions, err := discovery.ScanFunctions(projectRoot)
	if err != nil {
		out.Error("Failed to scan for functions: %v", err)
		out.Warning("Troubleshooting tips:")
		out.Print("  • Make sure you're in a Forge project directory")
		out.Print("  • Check that src/functions/ directory exists")
		out.Print("  • Verify function directories contain entry files")
		return fmt.Errorf("failed to scan functions: %w", err)
	}

	if len(functions) == 0 {
		out.Warning("No functions found in src/functions/")
		out.Print("")
		out.Info("To create a function:")
		out.Print("  1. mkdir -p src/functions/my-function")
		out.Print("  2. Create an entry file:")
		out.Print("     • main.go for Go")
		out.Print("     • index.js for Node.js")
		out.Print("     • app.py for Python")
		return nil
	}

	out.Info("Found %d function(s):", len(functions))
	for _, fn := range functions {
		out.Print("  • %s (%s)", fn.Name, fn.Runtime)
	}
	out.Print("")

	// Setup build directory
	buildDir := filepath.Join(projectRoot, ".forge", "build")

	// Create stub zips first (or only, if --stub-only)
	if stubOnly {
		count, err := discovery.CreateStubZips(functions, buildDir)
		if err != nil {
			out.Error("Failed to create stub zips: %v", err)
			return fmt.Errorf("failed to create stub zips: %w", err)
		}
		out.Success("Created %d stub zip(s)", count)
		out.Dim("Output: %s", buildDir)
		return nil
	}

	// Always ensure stubs exist before building (for terraform init)
	if _, err := discovery.CreateStubZips(functions, buildDir); err != nil {
		out.Error("Failed to create stub zips: %v", err)
		return fmt.Errorf("failed to create stub zips: %w", err)
	}

	// Create build registry
	registry := build.NewRegistry()

	// Build each function with progress tracking
	successCount := 0
	for i, fn := range functions {
		out.Step(i+1, len(functions), fmt.Sprintf("Building %s", fn.Name))

		// Get builder from registry
		builderOpt := build.GetBuilder(registry, fn.Runtime)
		if O.IsNone(builderOpt) {
			out.Error("Unsupported runtime: %s", fn.Runtime)
			out.Warning("Supported runtimes:")
			out.Print("  • provided.al2023, provided.al2 (Go)")
			out.Print("  • nodejs20.x, nodejs18.x (Node.js)")
			out.Print("  • python3.13, python3.12, python3.11 (Python)")
			return fmt.Errorf("unsupported runtime: %s", fn.Runtime)
		}

		// Extract builder
		builder := O.Fold(
			func() build.BuildFunc { return nil },
			func(b build.BuildFunc) build.BuildFunc { return b },
		)(builderOpt)

		// Convert to build config
		cfg := build.Config{
			SourceDir:  fn.Path,
			OutputPath: filepath.Join(buildDir, fn.Name),
			Handler:    fn.EntryPoint,
			Runtime:    fn.Runtime,
			Env:        make(map[string]string),
		}

		// Execute build
		result := builder(ctx, cfg)

		// Handle result
		if E.IsLeft(result) {
			err := E.Fold(
				func(e error) error { return e },
				func(a build.Artifact) error { return nil },
			)(result)
			out.Error("Build failed for %s: %v", fn.Name, err)
			out.Warning("Debug tips:")
			out.Print("  • Check function source code in %s", fn.Path)
			out.Print("  • Verify dependencies are specified correctly")
			out.Print("  • Review build logs above for details")
			return fmt.Errorf("failed to build %s: %w", fn.Name, err)
		}

		// Extract artifact
		artifact := E.Fold(
			func(e error) build.Artifact { return build.Artifact{} },
			func(a build.Artifact) build.Artifact { return a },
		)(result)

		sizeMB := float64(artifact.Size) / 1024 / 1024
		out.Success("%s: %s (%.2f MB, checksum: %s)",
			fn.Name,
			filepath.Base(artifact.Path),
			sizeMB,
			artifact.Checksum[:8],
		)
		successCount++
	}

	out.Print("")
	out.Success("All %d functions built successfully", successCount)
	out.Dim("Output directory: %s", buildDir)

	return nil
}
