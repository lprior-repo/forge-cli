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
		assert.Equal(t, "lambda", cmd.Use)
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
			Runtime:        "python3.13",
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

		err := createPythonLambda(tmpDir, projectName, opts)
		require.NoError(t, err)

		// Check function directory structure
		functionDir := filepath.Join(tmpDir, projectName, "src", "functions", "handler")
		assert.DirExists(t, functionDir)

		// Check app.py exists
		appPy := filepath.Join(functionDir, "app.py")
		assert.FileExists(t, appPy)

		// Check content includes handler
		content, err := os.ReadFile(appPy)
		require.NoError(t, err)
		assert.Contains(t, string(content), "lambda_handler")
		assert.Contains(t, string(content), "aws_lambda_powertools")
	})

	t.Run("creates infrastructure directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectName := "test-func"

		opts := LambdaProjectOptions{
			Runtime:       "python3.13",
			ServiceName:   projectName,
			FunctionName:  "handler",
			UsePowertools: true,
		}

		err := createPythonLambda(tmpDir, projectName, opts)
		require.NoError(t, err)

		infraDir := filepath.Join(tmpDir, projectName, "infra")
		assert.DirExists(t, infraDir)

		// Check for Terraform files
		mainTf := filepath.Join(infraDir, "main.tf")
		assert.FileExists(t, mainTf)
	})

	t.Run("creates requirements.txt with dependencies", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectName := "test-func"

		opts := LambdaProjectOptions{
			Runtime:       "python3.13",
			ServiceName:   projectName,
			FunctionName:  "handler",
			UsePowertools: true,
		}

		err := createPythonLambda(tmpDir, projectName, opts)
		require.NoError(t, err)

		functionDir := filepath.Join(tmpDir, projectName, "src", "functions", "handler")
		reqFile := filepath.Join(functionDir, "requirements.txt")
		assert.FileExists(t, reqFile)

		content, err := os.ReadFile(reqFile)
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
}
