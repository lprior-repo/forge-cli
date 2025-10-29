// Package discovery provides convention-based function and resource discovery.
// It scans project directories to detect Lambda functions, their runtimes,
// and generates build configurations following SAM-like conventions.
package discovery

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lewis/forge/internal/build"
)

const (
	// RuntimeGo is the Go Lambda runtime (provided.al2023)
	RuntimeGo = "provided.al2023"
	// RuntimeNode is the Node.js Lambda runtime
	RuntimeNode = "nodejs20.x"
	// RuntimePython is the Python Lambda runtime
	RuntimePython = "python3.13"
)

// Function represents a discovered Lambda function
type Function struct {
	Name       string // Function name (directory name)
	Path       string // Absolute path to function source
	Runtime    string // Detected runtime
	EntryPoint string // Entry file (main.go, index.js, app.py, etc.)
}

// Scanner discovers functions following convention-over-configuration
type Scanner struct {
	projectRoot string
}

// NewScanner creates a new function scanner
func NewScanner(projectRoot string) *Scanner {
	return &Scanner{
		projectRoot: projectRoot,
	}
}

// ScanFunctions discovers all functions in src/functions/*
// Returns a slice of discovered functions
func (s *Scanner) ScanFunctions() ([]Function, error) {
	functionsDir := filepath.Join(s.projectRoot, "src", "functions")

	// Check if functions directory exists
	if _, err := os.Stat(functionsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("src/functions directory not found")
	}

	// Read all subdirectories
	entries, err := os.ReadDir(functionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read functions directory: %w", err)
	}

	functions := make([]Function, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		functionPath := filepath.Join(functionsDir, entry.Name())

		// Detect runtime by checking for entry files
		runtime, entryPoint, err := s.detectRuntime(functionPath)
		if err != nil {
			// Skip directories without recognizable entry points
			continue
		}

		functions = append(functions, Function{
			Name:       entry.Name(),
			Path:       functionPath,
			Runtime:    runtime,
			EntryPoint: entryPoint,
		})
	}

	return functions, nil
}

// detectRuntime determines the runtime by checking for entry point files
// Returns (runtime, entryPoint, error)
func (s *Scanner) detectRuntime(functionPath string) (string, string, error) {
	// Go: main.go or *.go files
	if s.fileExists(functionPath, "main.go") {
		return RuntimeGo, "main.go", nil
	}
	if s.hasGoFiles(functionPath) {
		return RuntimeGo, "*.go", nil
	}

	// Node.js: index.js, index.mjs, or handler.js
	if s.fileExists(functionPath, "index.js") {
		return RuntimeNode, "index.js", nil
	}
	if s.fileExists(functionPath, "index.mjs") {
		return RuntimeNode, "index.mjs", nil
	}
	if s.fileExists(functionPath, "handler.js") {
		return RuntimeNode, "handler.js", nil
	}

	// Python: app.py, lambda_function.py, or handler.py
	if s.fileExists(functionPath, "app.py") {
		return RuntimePython, "app.py", nil
	}
	if s.fileExists(functionPath, "lambda_function.py") {
		return RuntimePython, "lambda_function.py", nil
	}
	if s.fileExists(functionPath, "handler.py") {
		return RuntimePython, "handler.py", nil
	}

	return "", "", fmt.Errorf("no recognized entry point found")
}

// fileExists checks if a file exists in the given directory
func (s *Scanner) fileExists(dir, filename string) bool {
	path := filepath.Join(dir, filename)
	_, err := os.Stat(path)
	return err == nil
}

// hasGoFiles checks if directory contains any .go files
func (s *Scanner) hasGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".go" {
			return true
		}
	}
	return false
}

// ToBuildConfig converts a Function to a build.Config
func (f *Function) ToBuildConfig(buildDir string) build.Config {
	outputPath := filepath.Join(buildDir, f.Name+".zip")

	// Determine handler based on runtime
	handler := "bootstrap"
	if f.Runtime[:6] == "nodejs" {
		handler = "index.handler"
	} else if f.Runtime[:6] == "python" {
		handler = "handler"
	}

	return build.Config{
		SourceDir:  f.Path,
		OutputPath: outputPath,
		Runtime:    f.Runtime,
		Handler:    handler,
		Env:        make(map[string]string),
	}
}
