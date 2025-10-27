package cli

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
	"github.com/lewis/forge/internal/build"
	"github.com/lewis/forge/internal/config"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/stack"
	"github.com/lewis/forge/internal/terraform"
	"github.com/spf13/cobra"
)

// NewDeployCmd creates the 'deploy' command
func NewDeployCmd() *cobra.Command {
	var (
		autoApprove bool
		parallel    bool
	)

	cmd := &cobra.Command{
		Use:   "deploy [stack-name]",
		Short: "Build and deploy stacks",
		Long: `Build Lambda functions and deploy infrastructure with Terraform.
If no stack name is provided, deploys all stacks in dependency order.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var targetStack string
			if len(args) > 0 {
				targetStack = args[0]
			}
			return runDeploy(targetStack, autoApprove, parallel)
		},
	}

	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip interactive approval")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Deploy independent stacks in parallel")

	return cmd
}

// runDeploy executes the deployment using functional pipeline
func runDeploy(targetStack string, autoApprove, parallel bool) error {
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
	var stacksToDeploy []*stack.Stack
	if targetStack != "" {
		found := false
		for _, st := range allStacks {
			if st.Name == targetStack {
				stacksToDeploy = []*stack.Stack{st}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("stack not found: %s", targetStack)
		}
	} else {
		stacksToDeploy = allStacks
	}

	// Build dependency graph and get deployment order
	graph, err := stack.NewGraph(stacksToDeploy)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	orderedStacks, err := graph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to sort stacks: %w", err)
	}

	fmt.Printf("Deploying %d stack(s) in order: ", len(orderedStacks))
	for i, st := range orderedStacks {
		if i > 0 {
			fmt.Print(" → ")
		}
		fmt.Print(st.Name)
	}
	fmt.Println()

	// Create functional executor
	tfPath := findTerraformPath()
	exec := terraform.NewExecutor(tfPath)

	// Build stage using functional approach
	buildStage := createBuildStage()

	// Create deployment pipeline: Build → Init → Plan → Apply → Outputs
	deployPipeline := pipeline.New(
		buildStage,
		pipeline.TerraformInit(exec),
		pipeline.TerraformPlan(exec),
		pipeline.TerraformApply(exec, autoApprove),
		pipeline.CaptureOutputs(exec),
	)

	// Initial state
	initialState := pipeline.State{
		ProjectDir: projectRoot,
		Stacks:     orderedStacks,
		Config:     cfg,
	}

	// Run pipeline
	result := deployPipeline.Run(ctx, initialState)

	// Handle result using Either
	if E.IsLeft(result) {
		err := E.Fold(
			func(e error) error { return e },
			func(s pipeline.State) error { return nil },
		)(result)
		return err
	}

	fmt.Println("\n✓ All stacks deployed successfully")
	return nil
}

// createBuildStage creates a pipeline stage for building Lambda functions
func createBuildStage() pipeline.Stage {
	return func(ctx context.Context, s pipeline.State) E.Either[error, pipeline.State] {
		// Initialize artifacts map if needed
		if s.Artifacts == nil {
			s.Artifacts = make(map[string]pipeline.Artifact)
		}

		// Create build registry
		registry := build.NewRegistry()

		// Build each stack that needs building
		for _, st := range s.Stacks {
			if !st.NeedsBuild() {
				continue
			}

			fmt.Printf("[%s] Building %s function...\n", st.Name, st.Runtime)

			// Get builder from registry
			builderOpt := registry.Get(st.Runtime)
			if O.IsNone(builderOpt) {
				return E.Left[pipeline.State](fmt.Errorf("unsupported runtime: %s", st.Runtime))
			}

			// Extract builder using Fold
			builder := O.Fold(
				func() build.BuildFunc { return nil },
				func(b build.BuildFunc) build.BuildFunc { return b },
			)(builderOpt)

			// Prepare build config
			outputPath := filepath.Join(st.AbsPath, st.GetBuildTarget())
			if st.GetBuildTarget() == "lambda.zip" {
				outputPath = filepath.Join(st.AbsPath, "lambda.zip")
			} else {
				outputPath = filepath.Join(st.AbsPath, "bootstrap")
			}

			cfg := build.Config{
				SourceDir:  st.AbsPath,
				OutputPath: outputPath,
				Runtime:    st.Runtime,
				Handler:    st.Handler,
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
				return E.Left[pipeline.State](fmt.Errorf("failed to build %s: %w", st.Name, err))
			}

			// Extract artifact
			artifact := E.Fold(
				func(e error) build.Artifact { return build.Artifact{} },
				func(a build.Artifact) build.Artifact { return a },
			)(result)

			// For Go, zip the bootstrap
			if st.GetBuildTarget() == "bootstrap" {
				zipPath := filepath.Join(st.AbsPath, "bootstrap.zip")
				if err := zipFile(artifact.Path, zipPath); err != nil {
					return E.Left[pipeline.State](fmt.Errorf("failed to zip bootstrap: %w", err))
				}
				fmt.Printf("[%s] Built: %s (%.2f MB)\n", st.Name, "bootstrap.zip", float64(artifact.Size)/1024/1024)
			} else {
				fmt.Printf("[%s] Built: %s (%.2f MB)\n", st.Name, filepath.Base(artifact.Path), float64(artifact.Size)/1024/1024)
			}

			// Store artifact in state
			s.Artifacts[st.Name] = pipeline.Artifact{
				Path:     artifact.Path,
				Checksum: artifact.Checksum,
				Size:     artifact.Size,
			}
		}

		return E.Right[error](s)
	}
}

// findTerraformPath finds the terraform binary
func findTerraformPath() string {
	// For now, assume terraform is in PATH
	// TODO: Add logic to find terraform binary
	return "terraform"
}


// zipFile creates a zip archive of a single file
func zipFile(sourcePath, destPath string) error {
	// Create zip file
	zipFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add file to zip
	w, err := zipWriter.Create(filepath.Base(sourcePath))
	if err != nil {
		return err
	}

	// Copy file contents
	f, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}
