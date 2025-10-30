package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	A "github.com/IBM/fp-go/array"
	E "github.com/IBM/fp-go/either"
	"github.com/spf13/cobra"

	"github.com/lewis/forge/internal/build"
	"github.com/lewis/forge/internal/discovery"
	"github.com/lewis/forge/internal/ui"
)

// NewBuildCmd creates the 'build' command.
func NewBuildCmd() *cobra.Command {
	var stubOnly bool

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build Lambda functions with convention-based discovery",
		Long: `
╭──────────────────────────────────────────────────────────────╮
│  🔨 Forge Build System                                      │
╰──────────────────────────────────────────────────────────────╯

Build Lambda functions with zero configuration.
Automatically discovers functions, detects runtimes, and creates deployment packages.

🎯 Conventions (No Config Required):
  • Function name = directory name (e.g., src/functions/api → api)
  • Runtime auto-detection from entry files:
    - main.go or *.go        → Go (provided.al2023)
    - index.js/handler.js    → Node.js (nodejs20.x)
    - app.py/lambda_function → Python (python3.13)
  • Output: .forge/build/{name}.zip

📦 Build Process:
  1. Scans src/functions/* for function directories
  2. Detects runtime from entry file
  3. Runs runtime-specific builder (go build, npm install, pip)
  4. Creates deployment package with dependencies
  5. Generates SHA256 checksum for caching

🚀 Examples:

  # Build all functions in src/functions/
  forge build

  # Create stub zips only (for terraform init)
  forge build --stub-only

💡 Pro Tips:
  • Build artifacts are cached by checksum
  • Dependencies are bundled automatically
  • Stub zips allow Terraform to initialize before real build
  • Use --verbose to see detailed build output

📁 Expected Structure:
  src/functions/
  ├── api/          # Function: api
  │   └── main.go   # Runtime: Go
  └── worker/       # Function: worker
      └── index.js  # Runtime: Node.js
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(stubOnly)
		},
	}

	cmd.Flags().BoolVar(&stubOnly, "stub-only", false, "Create stub zips without building")

	return cmd
}

// runBuild executes the build process.
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

	// Build single function with UI feedback
	buildOne := func(index int, fn discovery.Function) E.Either[error, build.Artifact] {
		out.Step(index+1, len(functions), "Building "+fn.Name)

		// Get builder from registry and convert Option to Either
		builderEither := E.FromOption[build.BuildFunc](
			func() error {
				out.Error("Unsupported runtime: %s", fn.Runtime)
				out.Warning("Supported runtimes:")
				out.Print("  • provided.al2023, provided.al2 (Go)")
				out.Print("  • nodejs20.x, nodejs18.x (Node.js)")
				out.Print("  • python3.13, python3.12, python3.11 (Python)")
				return fmt.Errorf("unsupported runtime: %s", fn.Runtime)
			},
		)(build.GetBuilder(registry, fn.Runtime))

		// Chain the build operation
		return E.Chain(func(builder build.BuildFunc) E.Either[error, build.Artifact] {
			cfg := build.Config{
				SourceDir:  fn.Path,
				OutputPath: filepath.Join(buildDir, fn.Name),
				Handler:    fn.EntryPoint,
				Runtime:    fn.Runtime,
				Env:        make(map[string]string),
			}

			// Execute build and add error context
			return E.MapLeft[build.Artifact](func(err error) error {
				out.Error("Build failed for %s: %v", fn.Name, err)
				out.Warning("Debug tips:")
				out.Print("  • Check function source code in %s", fn.Path)
				out.Print("  • Verify dependencies are specified correctly")
				out.Print("  • Review build logs above for details")
				return fmt.Errorf("failed to build %s: %w", fn.Name, err)
			})(builder(ctx, cfg))
		})(builderEither)
	}

	// Build all functions functionally using indexed map and fold
	type indexedFunc struct {
		index int
		fn    discovery.Function
	}

	// Create indexed list
	indexed := A.MapWithIndex(func(i int, fn discovery.Function) indexedFunc {
		return indexedFunc{index: i, fn: fn}
	})(functions)

	// Build all and short-circuit on first error
	artifactsEither := A.Reduce(
		func(acc E.Either[error, []build.Artifact], item indexedFunc) E.Either[error, []build.Artifact] {
			return E.Chain(func(artifacts []build.Artifact) E.Either[error, []build.Artifact] {
				return E.Map[error](func(artifact build.Artifact) []build.Artifact {
					// Print success message
					sizeMB := float64(artifact.Size) / 1024 / 1024
					out.Success("%s: %s (%.2f MB, checksum: %s)",
						item.fn.Name,
						filepath.Base(artifact.Path),
						sizeMB,
						artifact.Checksum[:8],
					)
					return append(artifacts, artifact)
				})(buildOne(item.index, item.fn))
			})(acc)
		},
		E.Right[error]([]build.Artifact{}),
	)(indexed)

	// Handle final result
	return E.Fold(
		func(err error) error {
			return err
		},
		func(artifacts []build.Artifact) error {
			out.Print("")
			out.Success("All %d functions built successfully", len(artifacts))
			out.Dim("Output directory: %s", buildDir)
			return nil
		},
	)(artifactsEither)
}
