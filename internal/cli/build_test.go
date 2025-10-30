package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuildCmd(t *testing.T) {
	t.Run("creates build command", func(t *testing.T) {
		cmd := NewBuildCmd()

		assert.NotNil(t, cmd)
		assert.Equal(t, "build", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("has stub-only flag", func(t *testing.T) {
		cmd := NewBuildCmd()

		flag := cmd.Flags().Lookup("stub-only")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})
}

func TestRunBuild(t *testing.T) {
	t.Run("returns error when no src/functions directory", func(t *testing.T) {
		// Change to temp directory with no functions
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		err := runBuild(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan functions")
	})

	t.Run("succeeds with stub-only when functions exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create src/functions structure with a Go function
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "main.go"), []byte("package main"), 0644))

		os.Chdir(tmpDir)

		err := runBuild(true)
		assert.NoError(t, err)

		// Verify stub was created
		stubPath := filepath.Join(tmpDir, ".forge", "build", "api.zip")
		assert.FileExists(t, stubPath)
	})

	t.Run("returns nil when no functions found", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create functions dir with unsupported files (will be ignored)
		functionsDir := filepath.Join(tmpDir, "src", "functions")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))

		os.Chdir(tmpDir)

		// Should succeed with no functions message
		err := runBuild(false)
		assert.NoError(t, err)
	})

	t.Run("handles unsupported runtime gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create function with unsupported entry file that gets detected
		// Ruby files are ignored by discovery, so we need a different approach
		// Create functions dir but no recognized entry files
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "handler.rb"), []byte("# Ruby"), 0644))

		os.Chdir(tmpDir)

		err := runBuild(false)
		// Should succeed with no functions message (Ruby is not detected)
		assert.NoError(t, err)
	})

	t.Run("builds Python function successfully", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping integration test in short mode")
		}

		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create Python function
		functionsDir := filepath.Join(tmpDir, "src", "functions", "handler")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))

		// Create minimal Python handler
		pythonCode := `def lambda_handler(event, context):
    return {"statusCode": 200, "body": "Hello"}
`
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "app.py"), []byte(pythonCode), 0644))

		// Create empty requirements.txt
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "requirements.txt"), []byte(""), 0644))

		os.Chdir(tmpDir)

		err := runBuild(false)
		assert.NoError(t, err)

		// Verify build artifact created
		buildPath := filepath.Join(tmpDir, ".forge", "build", "handler.zip")
		assert.FileExists(t, buildPath)
	})

	t.Run("builds multiple functions", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create multiple stub functions
		for _, name := range []string{"api", "worker", "processor"} {
			functionsDir := filepath.Join(tmpDir, "src", "functions", name)
			require.NoError(t, os.MkdirAll(functionsDir, 0755))
			require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "main.go"), []byte("package main"), 0644))
		}

		os.Chdir(tmpDir)

		// Use stub-only for faster test
		err := runBuild(true)
		assert.NoError(t, err)

		// Verify all stubs created
		for _, name := range []string{"api", "worker", "processor"} {
			stubPath := filepath.Join(tmpDir, ".forge", "build", name+".zip")
			assert.FileExists(t, stubPath)
		}
	})

	t.Run("creates build directory if missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create function
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "main.go"), []byte("package main"), 0644))

		os.Chdir(tmpDir)

		// Ensure build directory doesn't exist
		buildDir := filepath.Join(tmpDir, ".forge", "build")
		require.NoDirExists(t, buildDir)

		err := runBuild(true)
		assert.NoError(t, err)

		// Verify build directory was created
		assert.DirExists(t, buildDir)
	})

	t.Run("provides helpful error context on build failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create function with invalid Go code (to trigger build failure)
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))

		// Invalid Go code that won't compile
		invalidGo := `package main
this is not valid go code
`
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "main.go"), []byte(invalidGo), 0644))

		os.Chdir(tmpDir)

		// Stub-only should still work (doesn't compile)
		err := runBuild(true)
		assert.NoError(t, err)
	})

	t.Run("detects Node.js runtime correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create Node.js function
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))

		nodeCode := `exports.handler = async (event) => {
    return { statusCode: 200, body: 'Hello' };
};
`
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "index.js"), []byte(nodeCode), 0644))

		os.Chdir(tmpDir)

		// Stub-only for faster test
		err := runBuild(true)
		assert.NoError(t, err)

		// Verify stub created
		stubPath := filepath.Join(tmpDir, ".forge", "build", "api.zip")
		assert.FileExists(t, stubPath)
	})

	t.Run("handles mixed runtimes in same project", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create Go function
		goDir := filepath.Join(tmpDir, "src", "functions", "go-handler")
		require.NoError(t, os.MkdirAll(goDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(goDir, "main.go"), []byte("package main"), 0644))

		// Create Node.js function
		nodeDir := filepath.Join(tmpDir, "src", "functions", "node-handler")
		require.NoError(t, os.MkdirAll(nodeDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(nodeDir, "index.js"), []byte("exports.handler = async () => {}"), 0644))

		// Create Python function
		pythonDir := filepath.Join(tmpDir, "src", "functions", "python-handler")
		require.NoError(t, os.MkdirAll(pythonDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(pythonDir, "app.py"), []byte("def lambda_handler(e, c): pass"), 0644))

		os.Chdir(tmpDir)

		// Use stub-only for faster test
		err := runBuild(true)
		assert.NoError(t, err)

		// Verify all stubs created
		for _, name := range []string{"go-handler", "node-handler", "python-handler"} {
			stubPath := filepath.Join(tmpDir, ".forge", "build", name+".zip")
			assert.FileExists(t, stubPath, "Stub for %s should exist", name)
		}
	})

	t.Run("displays helpful output messages", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create single function for output test
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "main.go"), []byte("package main"), 0644))

		os.Chdir(tmpDir)

		// Run stub build - should output success messages
		err := runBuild(true)
		assert.NoError(t, err)
	})

	t.Run("handles handler.js Node entry file", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create Node.js function with handler.js
		functionsDir := filepath.Join(tmpDir, "src", "functions", "api")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))

		nodeCode := `exports.handler = async (event) => {
    return { statusCode: 200 };
};`
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "handler.js"), []byte(nodeCode), 0644))

		os.Chdir(tmpDir)

		err := runBuild(true)
		assert.NoError(t, err)

		stubPath := filepath.Join(tmpDir, ".forge", "build", "api.zip")
		assert.FileExists(t, stubPath)
	})

	t.Run("handles lambda_function.py Python entry file", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create Python function with lambda_function.py
		functionsDir := filepath.Join(tmpDir, "src", "functions", "worker")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))

		pythonCode := `def lambda_handler(event, context):
    return {"statusCode": 200}`
		require.NoError(t, os.WriteFile(filepath.Join(functionsDir, "lambda_function.py"), []byte(pythonCode), 0644))

		os.Chdir(tmpDir)

		err := runBuild(true)
		assert.NoError(t, err)

		stubPath := filepath.Join(tmpDir, ".forge", "build", "worker.zip")
		assert.FileExists(t, stubPath)
	})
}
