package stack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// Stack represents a deployable unit (e.g., a Lambda function)
type Stack struct {
	Name         string
	Path         string   // Relative to project root
	AbsPath      string   // Absolute path
	Dependencies []string // Paths to other stacks (relative to project root)
	Runtime      string   // go, python3.11, nodejs20.x, etc.
	Handler      string   // Path to handler code
	Description  string
}

// Metadata is the HCL structure for stack.forge.hcl
type Metadata struct {
	Stack StackBlock `hcl:"stack,block"`
}

// StackBlock represents the stack configuration block
type StackBlock struct {
	Name        string   `hcl:"name"`
	Description string   `hcl:"description,optional"`
	After       []string `hcl:"after,optional"`
	Runtime     string   `hcl:"runtime"`
	Handler     string   `hcl:"handler,optional"`
}

// Detector finds all stacks in a project
type Detector struct {
	projectRoot string
}

// NewDetector creates a new stack detector
func NewDetector(projectRoot string) *Detector {
	return &Detector{projectRoot: projectRoot}
}

// FindStacks walks the directory tree and finds all stack.forge.hcl files
func (d *Detector) FindStacks() ([]*Stack, error) {
	var stacks []*Stack

	err := filepath.Walk(d.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for stack.forge.hcl files
		if !info.IsDir() && info.Name() == "stack.forge.hcl" {
			stack, err := d.loadStack(path)
			if err != nil {
				return fmt.Errorf("failed to load stack at %s: %w", path, err)
			}
			stacks = append(stacks, stack)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return stacks, nil
}

// loadStack loads a stack from a stack.forge.hcl file
func (d *Detector) loadStack(metadataPath string) (*Stack, error) {
	var metadata Metadata
	err := hclsimple.DecodeFile(metadataPath, nil, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HCL: %w", err)
	}

	// Get the directory containing the stack.forge.hcl
	stackDir := filepath.Dir(metadataPath)
	relPath, err := filepath.Rel(d.projectRoot, stackDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	// Normalize dependencies to be relative to project root
	var deps []string
	for _, dep := range metadata.Stack.After {
		// If dep is relative, resolve it relative to the stack directory
		if !filepath.IsAbs(dep) {
			absDepPath := filepath.Join(stackDir, dep)
			relDep, err := filepath.Rel(d.projectRoot, absDepPath)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve dependency %s: %w", dep, err)
			}
			deps = append(deps, filepath.Clean(relDep))
		} else {
			deps = append(deps, dep)
		}
	}

	// Default handler to the stack directory if not specified
	handler := metadata.Stack.Handler
	if handler == "" {
		handler = relPath
	}

	return &Stack{
		Name:         metadata.Stack.Name,
		Path:         relPath,
		AbsPath:      stackDir,
		Dependencies: deps,
		Runtime:      metadata.Stack.Runtime,
		Handler:      handler,
		Description:  metadata.Stack.Description,
	}, nil
}

// Validate ensures the stack configuration is valid
func (s *Stack) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("stack name is required")
	}
	if s.Runtime == "" {
		return fmt.Errorf("stack runtime is required")
	}
	if !isValidRuntime(s.Runtime) {
		return fmt.Errorf("unsupported runtime: %s", s.Runtime)
	}
	return nil
}

// isValidRuntime checks if the runtime is supported
func isValidRuntime(runtime string) bool {
	validRuntimes := []string{
		"go1.x",
		"python3.11", "python3.12", "python3.13",
		"nodejs20.x", "nodejs18.x",
		"provided.al2", "provided.al2023",
	}

	for _, valid := range validRuntimes {
		if runtime == valid {
			return true
		}
	}

	return false
}

// GetBuildTarget returns the build target based on runtime
func (s *Stack) GetBuildTarget() string {
	if strings.HasPrefix(s.Runtime, "go") {
		return "bootstrap"
	}
	return "lambda.zip"
}

// NeedsBuild determines if this stack requires a build step
func (s *Stack) NeedsBuild() bool {
	// Go requires compilation
	if strings.HasPrefix(s.Runtime, "go") {
		return true
	}
	// Python and Node require packaging
	if strings.HasPrefix(s.Runtime, "python") || strings.HasPrefix(s.Runtime, "nodejs") {
		return true
	}
	return false
}
