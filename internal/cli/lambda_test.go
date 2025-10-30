package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLambdaCmd tests lambda command creation
func TestNewLambdaCmd(t *testing.T) {
	t.Run("creates lambda command", func(t *testing.T) {
		cmd := NewLambdaCmd()
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Use, "lambda")
		assert.Contains(t, cmd.Short, "Lambda")
	})

	t.Run("has runtime flag", func(t *testing.T) {
		cmd := NewLambdaCmd()
		flag := cmd.Flags().Lookup("runtime")
		assert.NotNil(t, flag)
		assert.Equal(t, "python", flag.DefValue)
	})

	t.Run("has service flag", func(t *testing.T) {
		cmd := NewLambdaCmd()
		flag := cmd.Flags().Lookup("service")
		assert.NotNil(t, flag)
	})

	t.Run("has powertools flag", func(t *testing.T) {
		cmd := NewLambdaCmd()
		flag := cmd.Flags().Lookup("powertools")
		assert.NotNil(t, flag)
		assert.Equal(t, "true", flag.DefValue)
	})

	t.Run("has dynamodb flag", func(t *testing.T) {
		cmd := NewLambdaCmd()
		flag := cmd.Flags().Lookup("dynamodb")
		assert.NotNil(t, flag)
		assert.Equal(t, "true", flag.DefValue)
	})
}

// TestCreatePythonLambda tests Python Lambda project creation
func TestCreatePythonLambda(t *testing.T) {
	t.Run("creates Python Lambda project structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectName := "my-function"

		opts := LambdaProjectOptions{
			Runtime:        "python",
			ServiceName:    projectName,
			FunctionName:   "handler",
			Description:    "Test function",
			UsePowertools:  true,
			UseIdempotency: true,
			UseDynamoDB:    true,
			TableName:      "test-table",
			APIPath:        "/api/test",
			HTTPMethod:     "POST",
		}

		projectDir := filepath.Join(tmpDir, projectName)
		err := createPythonLambda(projectDir, projectName, opts)
		require.NoError(t, err)

		// Check service directory structure (actual Python generator output)
		serviceDir := filepath.Join(projectDir, "service")
		assert.DirExists(t, serviceDir)

		// Check handler exists
		handlerPy := filepath.Join(serviceDir, "handlers", "handle_request.py")
		assert.FileExists(t, handlerPy)

		// Check content includes handler and powertools
		content, err := os.ReadFile(handlerPy)
		require.NoError(t, err)
		assert.Contains(t, string(content), "lambda_handler")
	})

	t.Run("creates infrastructure directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectName := "test-func"

		opts := LambdaProjectOptions{
			Runtime:       "python",
			ServiceName:   projectName,
			FunctionName:  "handler",
			UsePowertools: true,
		}

		projectDir := filepath.Join(tmpDir, projectName)
		err := createPythonLambda(projectDir, projectName, opts)
		require.NoError(t, err)

		// Check for Terraform files in terraform/ subdirectory
		terraformDir := filepath.Join(projectDir, "terraform")
		assert.DirExists(t, terraformDir)

		mainTf := filepath.Join(terraformDir, "main.tf")
		assert.FileExists(t, mainTf)
	})

	t.Run("creates pyproject.toml with dependencies", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectName := "test-func"

		opts := LambdaProjectOptions{
			Runtime:       "python",
			ServiceName:   projectName,
			FunctionName:  "handler",
			UsePowertools: true,
		}

		projectDir := filepath.Join(tmpDir, projectName)
		err := createPythonLambda(projectDir, projectName, opts)
		require.NoError(t, err)

		// Python generator uses pyproject.toml, not requirements.txt
		pyprojectFile := filepath.Join(projectDir, "pyproject.toml")
		assert.FileExists(t, pyprojectFile)

		content, err := os.ReadFile(pyprojectFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "aws-lambda-powertools")
	})

	t.Run("fails with invalid directory", func(t *testing.T) {
		opts := LambdaProjectOptions{
			Runtime:      "python3.13",
			ServiceName:  "test",
			FunctionName: "handler",
		}

		err := createPythonLambda("/nonexistent/path/that/does/not/exist", "test", opts)
		assert.Error(t, err)
	})
}

// TestCreateLambdaProject tests Lambda project orchestration
func TestCreateLambdaProject(t *testing.T) {
	t.Run("creates Python Lambda project", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		opts := LambdaProjectOptions{
			Runtime:       "python",
			ServiceName:   "test-service",
			FunctionName:  "handler",
			UsePowertools: true,
		}

		err := createLambdaProject("test-function", opts)
		require.NoError(t, err)

		// Verify project directory exists
		assert.DirExists(t, filepath.Join(tmpDir, "test-function"))
	})

	t.Run("returns error for unsupported Go runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		opts := LambdaProjectOptions{
			Runtime:      "go",
			ServiceName:  "test-service",
			FunctionName: "main",
		}

		err := createLambdaProject("test-function", opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("returns error for unsupported Node.js runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		opts := LambdaProjectOptions{
			Runtime:      "nodejs",
			ServiceName:  "test-service",
			FunctionName: "handler",
		}

		err := createLambdaProject("test-function", opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("handles unsupported runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		opts := LambdaProjectOptions{
			Runtime:      "rust",
			ServiceName:  "test-service",
			FunctionName: "handler",
		}

		err := createLambdaProject("test-function", opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported runtime")
	})

	t.Run("fails when directory already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		// Create directory first
		projectDir := filepath.Join(tmpDir, "existing-project")
		require.NoError(t, os.MkdirAll(projectDir, 0755))

		opts := LambdaProjectOptions{
			Runtime:      "python",
			ServiceName:  "test",
			FunctionName: "handler",
		}

		err := createLambdaProject("existing-project", opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("sets default service name from project name", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		opts := LambdaProjectOptions{
			Runtime:      "python",
			ServiceName:  "", // Empty - should use project name
			FunctionName: "handler",
		}

		err := createLambdaProject("my-test-project", opts)
		require.NoError(t, err)

		// Verify project was created
		assert.DirExists(t, filepath.Join(tmpDir, "my-test-project"))
	})

	t.Run("sets default description", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		opts := LambdaProjectOptions{
			Runtime:      "python",
			ServiceName:  "myservice",
			FunctionName: "handler",
			Description:  "", // Empty - should use default
		}

		err := createLambdaProject("test-project", opts)
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
	})

	t.Run("sets default table name", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		opts := LambdaProjectOptions{
			Runtime:      "python",
			ServiceName:  "my_service",
			FunctionName: "handler",
			TableName:    "", // Empty - should generate default
			UseDynamoDB:  true,
		}

		err := createLambdaProject("test-project", opts)
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
	})
}

// TestLambdaCmdExecution tests command execution with various flags
func TestLambdaCmdExecution(t *testing.T) {
	t.Run("executes with Python runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{"test-project", "--runtime", "python"})

		err := cmd.Execute()
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
	})

	t.Run("fails with invalid runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{"test-project", "--runtime", "ruby"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid runtime")
	})

	t.Run("fails with Go runtime (not implemented)", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{"test-project", "--runtime", "go"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("fails with Node.js runtime (not implemented)", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{"test-project", "--runtime", "nodejs"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("accepts custom service name", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{"test-project", "--service", "custom-service"})

		err := cmd.Execute()
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
	})

	t.Run("accepts custom function name", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{"test-project", "--function", "custom-handler"})

		err := cmd.Execute()
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
	})

	t.Run("respects powertools flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{"test-project", "--powertools=false"})

		err := cmd.Execute()
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
	})

	t.Run("respects dynamodb flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{"test-project", "--dynamodb=false"})

		err := cmd.Execute()
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
	})

	t.Run("accepts custom table name", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{"test-project", "--table", "custom-table"})

		err := cmd.Execute()
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
	})

	t.Run("accepts custom API path and method", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		os.Chdir(tmpDir)

		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{"test-project", "--api-path", "/custom/path", "--method", "GET"})

		err := cmd.Execute()
		require.NoError(t, err)

		assert.DirExists(t, filepath.Join(tmpDir, "test-project"))
	})

	t.Run("requires project name argument", func(t *testing.T) {
		cmd := NewLambdaCmd()
		cmd.SetArgs([]string{}) // No project name

		err := cmd.Execute()
		assert.Error(t, err)
	})
}

// TestLambdaCmdFlags tests all flag configurations
func TestLambdaCmdFlags(t *testing.T) {
	t.Run("has all required flags", func(t *testing.T) {
		cmd := NewLambdaCmd()

		requiredFlags := []string{
			"runtime", "service", "function", "description",
			"powertools", "idempotency", "dynamodb", "table",
			"api-path", "method",
		}

		for _, flagName := range requiredFlags {
			flag := cmd.Flags().Lookup(flagName)
			assert.NotNil(t, flag, "Flag %s should exist", flagName)
		}
	})

	t.Run("flag defaults are correct", func(t *testing.T) {
		cmd := NewLambdaCmd()

		tests := []struct {
			name     string
			expected string
		}{
			{"runtime", "python"},
			{"function", "handler"},
			{"powertools", "true"},
			{"idempotency", "true"},
			{"dynamodb", "true"},
			{"api-path", "/api/orders"},
			{"method", "POST"},
		}

		for _, tt := range tests {
			flag := cmd.Flags().Lookup(tt.name)
			assert.Equal(t, tt.expected, flag.DefValue, "Flag %s default should be %s", tt.name, tt.expected)
		}
	})
}
