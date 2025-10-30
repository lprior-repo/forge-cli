// Package discovery provides convention-based function and resource discovery.
// It scans project directories to detect Lambda functions, their runtimes,
// and generates build configurations following SAM-like conventions.
package discovery

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	E "github.com/IBM/fp-go/either"

	"github.com/lewis/forge/internal/build"
)

const (
	// RuntimeGo is the Go Lambda runtime (provided.al2023).
	RuntimeGo = "provided.al2023"
	// RuntimeNode is the Node.js Lambda runtime.
	RuntimeNode = "nodejs20.x"
	// RuntimePython is the Python Lambda runtime.
	RuntimePython = "python3.13"
)

type (
	// Function represents a discovered Lambda function.
	Function struct {
		Name       string // Function name (directory name)
		Path       string // Absolute path to function source
		Runtime    string // Detected runtime
		EntryPoint string // Entry file (main.go, index.js, app.py, etc.)
	}
)

// Pure functional approach - no methods, no state.
func ScanFunctions(projectRoot string) ([]Function, error) {
	functionsDir := filepath.Join(projectRoot, "src", "functions")

	// Check if functions directory exists
	if _, err := os.Stat(functionsDir); os.IsNotExist(err) {
		return nil, errors.New("src/functions directory not found")
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
		runtime, entryPoint, err := detectRuntime(functionPath)
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

// Pure function - no methods, takes path as parameter.
func detectRuntime(functionPath string) (string, string, error) {
	// Go: main.go or *.go files
	if fileExists(functionPath, "main.go") {
		return RuntimeGo, "main.go", nil
	}
	if hasGoFiles(functionPath) {
		return RuntimeGo, "*.go", nil
	}

	// Node.js: index.js, index.mjs, or handler.js
	if fileExists(functionPath, "index.js") {
		return RuntimeNode, "index.js", nil
	}
	if fileExists(functionPath, "index.mjs") {
		return RuntimeNode, "index.mjs", nil
	}
	if fileExists(functionPath, "handler.js") {
		return RuntimeNode, "handler.js", nil
	}

	// Python: app.py, lambda_function.py, or handler.py
	if fileExists(functionPath, "app.py") {
		return RuntimePython, "app.py", nil
	}
	if fileExists(functionPath, "lambda_function.py") {
		return RuntimePython, "lambda_function.py", nil
	}
	if fileExists(functionPath, "handler.py") {
		return RuntimePython, "handler.py", nil
	}

	return "", "", errors.New("no recognized entry point found")
}

// Pure function - no methods.
func fileExists(dir, filename string) bool {
	path := filepath.Join(dir, filename)
	_, err := os.Stat(path)
	return err == nil
}

// Pure function - no methods.
func hasGoFiles(dir string) bool {
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

// PURE: Calculation with validation.
func ToBuildConfig(f Function, buildDir string) E.Either[error, build.Config] {
	// Validate function name
	if f.Name == "" {
		return E.Left[build.Config](errors.New("function name cannot be empty"))
	}

	// Validate runtime
	if f.Runtime == "" {
		return E.Left[build.Config](errors.New("function runtime cannot be empty"))
	}

	// Validate path
	if f.Path == "" {
		return E.Left[build.Config](errors.New("function path cannot be empty"))
	}

	// Validate buildDir
	if buildDir == "" {
		return E.Left[build.Config](errors.New("build directory cannot be empty"))
	}

	outputPath := filepath.Join(buildDir, f.Name+".zip")

	// Determine handler based on runtime (safe with strings.HasPrefix)
	handler := determineHandler(f.Runtime)

	return E.Right[error](build.Config{
		SourceDir:  f.Path,
		OutputPath: outputPath,
		Runtime:    f.Runtime,
		Handler:    handler,
		Env:        make(map[string]string),
	})
}

// PURE: Calculation - safe string matching.
func determineHandler(runtime string) string {
	switch {
	case strings.HasPrefix(runtime, "nodejs"):
		return "index.handler"
	case strings.HasPrefix(runtime, "python"):
		return "handler"
	default:
		return "bootstrap"
	}
}
