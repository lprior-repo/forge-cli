package python

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratePythonLambdaProject(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "my-lambda-service")

	// Configure project
	config := ProjectConfig{
		ServiceName:    "my-lambda-service",
		FunctionName:   "api-handler",
		Description:    "Example Lambda service with Powertools",
		PythonVersion:  "3.13",
		UsePowertools:  true,
		UseIdempotency: true,
		UseDynamoDB:    true,
		TableName:      "orders-table",
		APIPath:        "/api/orders",
		HTTPMethod:     "POST",
	}

	// Generate project
	err := Generate(projectDir, config)
	if err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify key files exist
	expectedFiles := []string{
		"pyproject.toml",
		"README.md",
		".gitignore",
		"Makefile",
		"service/handlers/handle_request.py",
		"service/handlers/models/env_vars.py",
		"service/handlers/utils/observability.py",
		"service/handlers/utils/rest_api.py",
		"service/models/input.py",
		"service/models/output.py",
		"service/logic/business_logic.py",
		"service/dal/dynamodb_handler.py",
		"service/dal/models/db.py",
		"terraform/main.tf",
		"terraform/variables.tf",
		"terraform/outputs.tf",
		"terraform/lambda.tf",
		"terraform/iam.tf",
		"terraform/dynamodb.tf",
		"terraform/apigateway.tf",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(projectDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file not found: %s", file)
		}
	}

	// Verify content of key files
	t.Run("pyproject.toml contains correct dependencies", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectDir, "pyproject.toml"))
		if err != nil {
			t.Fatal(err)
		}

		expectedStrings := []string{
			"aws-lambda-powertools",
			"pydantic",
			"mypy-boto3-dynamodb",
			"cachetools",
		}

		contentStr := string(content)
		for _, expected := range expectedStrings {
			if !contains(contentStr, expected) {
				t.Errorf("pyproject.toml missing expected string: %s", expected)
			}
		}
	})

	t.Run("handler contains Powertools decorators", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectDir, "service/handlers/handle_request.py"))
		if err != nil {
			t.Fatal(err)
		}

		expectedStrings := []string{
			"@logger.inject_lambda_context",
			"@metrics.log_metrics",
			"@tracer.capture_lambda_handler",
			"aws_lambda_powertools",
		}

		contentStr := string(content)
		for _, expected := range expectedStrings {
			if !contains(contentStr, expected) {
				t.Errorf("Handler missing expected string: %s", expected)
			}
		}
	})

	t.Log("✓ Python Lambda project generated successfully")
	t.Logf("  Project location: %s", projectDir)
}

func TestGenerateBasicPythonLambda(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "basic-lambda")

	// Configure basic project without Powertools
	config := ProjectConfig{
		ServiceName:    "basic-lambda",
		FunctionName:   "handler",
		Description:    "Basic Lambda service",
		PythonVersion:  "3.13",
		UsePowertools:  false,
		UseIdempotency: false,
		UseDynamoDB:    false,
		APIPath:        "/api/hello",
		HTTPMethod:     "GET",
	}

	// Generate project
	err := Generate(projectDir, config)
	if err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify handler is basic (no Powertools)
	content, err := os.ReadFile(filepath.Join(projectDir, "service/handlers/handle_request.py"))
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if contains(contentStr, "aws_lambda_powertools") {
		t.Error("Basic handler should not contain Powertools imports")
	}

	if !contains(contentStr, "lambda_handler") {
		t.Error("Handler should contain lambda_handler function")
	}

	t.Log("✓ Basic Python Lambda project generated successfully")
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
