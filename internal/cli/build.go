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
	ctx := context.Background()
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Scan for functions (pure functional approach - no OOP)
	functions, err := discovery.ScanFunctions(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to scan functions: %w", err)
	}

	if len(functions) == 0 {
		fmt.Println("No functions found in src/functions/")
		return nil
	}

	fmt.Printf("Found %d function(s):\n", len(functions))
	for _, fn := range functions {
		fmt.Printf("  - %s (%s)\n", fn.Name, fn.Runtime)
	}
	fmt.Println()

	// Setup build directory
	buildDir := filepath.Join(projectRoot, ".forge", "build")

	// Create stub zips first (or only, if --stub-only)
	if stubOnly {
		count, err := discovery.CreateStubZips(functions, buildDir)
		if err != nil {
			return fmt.Errorf("failed to create stub zips: %w", err)
		}
		fmt.Printf("Created %d stub zip(s) in %s\n", count, buildDir)
		return nil
	}

	// Always ensure stubs exist before building (for terraform init)
	if _, err := discovery.CreateStubZips(functions, buildDir); err != nil {
		return fmt.Errorf("failed to create stub zips: %w", err)
	}

	// Create build registry
	registry := build.NewRegistry()

	// Build each function
	for _, fn := range functions {
		fmt.Printf("[%s] Building...\n", fn.Name)

		// Get builder from registry
		builderOpt := build.GetBuilder(registry, fn.Runtime)
		if O.IsNone(builderOpt) {
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
			return fmt.Errorf("failed to build %s: %w", fn.Name, err)
		}

		// Extract artifact
		artifact := E.Fold(
			func(e error) build.Artifact { return build.Artifact{} },
			func(a build.Artifact) build.Artifact { return a },
		)(result)

		sizeMB := float64(artifact.Size) / 1024 / 1024
		fmt.Printf("[%s] ✓ Built: %s (%.2f MB, %s)\n",
			fn.Name,
			filepath.Base(artifact.Path),
			sizeMB,
			artifact.Checksum[:8],
		)
	}

	fmt.Printf("\n✓ All functions built successfully\n")
	fmt.Printf("Output directory: %s\n", buildDir)

	return nil
}
