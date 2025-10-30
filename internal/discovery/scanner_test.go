package discovery

import (
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanner_ScanFunctions(t *testing.T) {
	tests := []struct {
		name          string
		setupFiles    map[string]string // path -> content
		expectedFuncs []Function
		expectError   bool
	}{
		{
			name: "detects Go function with main.go",
			setupFiles: map[string]string{
				"src/functions/api/main.go": "package main\n\nfunc main() {}",
			},
			expectedFuncs: []Function{
				{
					Name:       "api",
					Runtime:    "provided.al2023",
					EntryPoint: "main.go",
				},
			},
		},
		{
			name: "detects Node.js function with index.js",
			setupFiles: map[string]string{
				"src/functions/worker/index.js": "exports.handler = async () => {};",
			},
			expectedFuncs: []Function{
				{
					Name:       "worker",
					Runtime:    "nodejs20.x",
					EntryPoint: "index.js",
				},
			},
		},
		{
			name: "detects Python function with app.py",
			setupFiles: map[string]string{
				"src/functions/processor/app.py": "def handler(event, context):\n    pass",
			},
			expectedFuncs: []Function{
				{
					Name:       "processor",
					Runtime:    "python3.13",
					EntryPoint: "app.py",
				},
			},
		},
		{
			name: "detects multiple functions of different runtimes",
			setupFiles: map[string]string{
				"src/functions/api/main.go":        "package main",
				"src/functions/worker/index.js":    "exports.handler = () => {}",
				"src/functions/processor/app.py":   "def handler(): pass",
			},
			expectedFuncs: []Function{
				{Name: "api", Runtime: "provided.al2023", EntryPoint: "main.go"},
				{Name: "processor", Runtime: "python3.13", EntryPoint: "app.py"},
				{Name: "worker", Runtime: "nodejs20.x", EntryPoint: "index.js"},
			},
		},
		{
			name: "skips directories without entry points",
			setupFiles: map[string]string{
				"src/functions/api/main.go":    "package main",
				"src/functions/invalid/foo.txt": "not a function",
			},
			expectedFuncs: []Function{
				{Name: "api", Runtime: "provided.al2023", EntryPoint: "main.go"},
			},
		},
		{
			name:        "returns error when src/functions does not exist",
			setupFiles:  map[string]string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp directory
			tmpDir := t.TempDir()

			// Create files
			for path, content := range tt.setupFiles {
				fullPath := filepath.Join(tmpDir, path)
				require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
			}

			// Scan functions (pure functional - no OOP)
			functions, err := ScanFunctions(tmpDir)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, functions, len(tt.expectedFuncs))

			// Sort for consistent comparison (functions are returned in directory order)
			// Check each expected function is present
			for _, expected := range tt.expectedFuncs {
				found := false
				for _, actual := range functions {
					if actual.Name == expected.Name {
						assert.Equal(t, expected.Runtime, actual.Runtime)
						assert.Equal(t, expected.EntryPoint, actual.EntryPoint)
						assert.Equal(t, filepath.Join(tmpDir, "src/functions", expected.Name), actual.Path)
						found = true
						break
					}
				}
				assert.True(t, found, "expected function %s not found", expected.Name)
			}
		})
	}
}

func TestScanner_detectRuntime(t *testing.T) {
	tests := []struct {
		name           string
		files          []string
		expectedRT     string
		expectedEntry  string
		expectError    bool
	}{
		{
			name:          "Go with main.go",
			files:         []string{"main.go"},
			expectedRT:    "provided.al2023",
			expectedEntry: "main.go",
		},
		{
			name:          "Go with other .go files",
			files:         []string{"handler.go", "util.go"},
			expectedRT:    "provided.al2023",
			expectedEntry: "*.go",
		},
		{
			name:          "Node.js with index.js",
			files:         []string{"index.js"},
			expectedRT:    "nodejs20.x",
			expectedEntry: "index.js",
		},
		{
			name:          "Node.js with index.mjs",
			files:         []string{"index.mjs"},
			expectedRT:    "nodejs20.x",
			expectedEntry: "index.mjs",
		},
		{
			name:          "Python with app.py",
			files:         []string{"app.py"},
			expectedRT:    "python3.13",
			expectedEntry: "app.py",
		},
		{
			name:          "Python with lambda_function.py",
			files:         []string{"lambda_function.py"},
			expectedRT:    "python3.13",
			expectedEntry: "lambda_function.py",
		},
		{
			name:        "no recognized entry point",
			files:       []string{"README.md"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create files
			for _, file := range tt.files {
				path := filepath.Join(tmpDir, file)
				require.NoError(t, os.WriteFile(path, []byte("test"), 0644))
			}

			// Pure functional - no OOP
			runtime, entryPoint, err := detectRuntime(tmpDir)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedRT, runtime)
			assert.Equal(t, tt.expectedEntry, entryPoint)
		})
	}
}

func TestFunction_ToBuildConfig(t *testing.T) {
	tests := []struct {
		name     string
		function Function
		buildDir string
		expected map[string]interface{}
	}{
		{
			name: "Go function",
			function: Function{
				Name:       "api",
				Path:       "/project/src/functions/api",
				Runtime:    "provided.al2023",
				EntryPoint: "main.go",
			},
			buildDir: "/project/.forge/build",
			expected: map[string]interface{}{
				"SourceDir":  "/project/src/functions/api",
				"OutputPath": "/project/.forge/build/api.zip",
				"Runtime":    "provided.al2023",
				"Handler":    "bootstrap",
			},
		},
		{
			name: "Node.js function",
			function: Function{
				Name:       "worker",
				Path:       "/project/src/functions/worker",
				Runtime:    "nodejs20.x",
				EntryPoint: "index.js",
			},
			buildDir: "/project/.forge/build",
			expected: map[string]interface{}{
				"SourceDir":  "/project/src/functions/worker",
				"OutputPath": "/project/.forge/build/worker.zip",
				"Runtime":    "nodejs20.x",
				"Handler":    "index.handler",
			},
		},
		{
			name: "Python function",
			function: Function{
				Name:       "processor",
				Path:       "/project/src/functions/processor",
				Runtime:    "python3.13",
				EntryPoint: "app.py",
			},
			buildDir: "/project/.forge/build",
			expected: map[string]interface{}{
				"SourceDir":  "/project/src/functions/processor",
				"OutputPath": "/project/.forge/build/processor.zip",
				"Runtime":    "python3.13",
				"Handler":    "handler",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToBuildConfig(tt.function, tt.buildDir)

			// Verify Either is Right (success)
			assert.True(t, E.IsRight(result), "ToBuildConfig should return Right for valid input")

			// Extract config from Either
			cfg := E.Fold(
				func(err error) build.Config { return build.Config{} },
				func(c build.Config) build.Config { return c },
			)(result)

			assert.Equal(t, tt.expected["SourceDir"], cfg.SourceDir)
			assert.Equal(t, tt.expected["OutputPath"], cfg.OutputPath)
			assert.Equal(t, tt.expected["Runtime"], cfg.Runtime)
			assert.Equal(t, tt.expected["Handler"], cfg.Handler)
			assert.NotNil(t, cfg.Env)
		})
	}
}
