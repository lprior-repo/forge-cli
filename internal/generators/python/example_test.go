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

// TestGenerateErrorPaths tests error handling in project generation.
func TestGenerateErrorPaths(t *testing.T) {
	t.Run("fails when directory cannot be created", func(t *testing.T) {
		tmpDir := t.TempDir()
		blockingFile := filepath.Join(tmpDir, "blocking")
		err := os.WriteFile(blockingFile, []byte("test"), 0o644)
		if err != nil {
			t.Fatal(err)
		}

		config := ProjectConfig{
			ServiceName:   "test",
			FunctionName:  "handler",
			Description:   "Test",
			PythonVersion: "3.13",
		}

		// Try to generate in a path where a file exists
		err = Generate(blockingFile, config)
		if err == nil {
			t.Error("Expected error when directory cannot be created")
		}
	})

	t.Run("fails when project files cannot be written", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "readonly-project")

		config := ProjectConfig{
			ServiceName:   "test",
			FunctionName:  "handler",
			Description:   "Test",
			PythonVersion: "3.13",
		}

		// Create the directory structure first
		err := createDirectoryStructure(projectDir, config)
		if err != nil {
			t.Fatal(err)
		}

		// Make directory read-only
		err = os.Chmod(projectDir, 0o555)
		if err != nil {
			t.Fatal(err)
		}

		// Try to generate files
		err = generateProjectFiles(projectDir, config)
		if err == nil {
			t.Error("Expected error when files cannot be written")
		}

		// Clean up
		_ = os.Chmod(projectDir, 0o755)
	})

	t.Run("fails when terraform directory cannot be created", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "readonly-terraform")

		config := ProjectConfig{
			ServiceName:   "test",
			FunctionName:  "handler",
			Description:   "Test",
			PythonVersion: "3.13",
		}

		// Create project directory
		err := os.MkdirAll(projectDir, 0o755)
		if err != nil {
			t.Fatal(err)
		}

		// Make directory read-only to prevent terraform/ creation
		err = os.Chmod(projectDir, 0o555)
		if err != nil {
			t.Fatal(err)
		}

		// Try to generate terraform files
		err = generateTerraformFiles(projectDir, config)
		if err == nil {
			t.Error("Expected error when terraform directory cannot be created")
		}

		// Clean up
		_ = os.Chmod(projectDir, 0o755)
	})
}

// TestPowertoolsHandlerVariations tests different HTTP methods.
func TestPowertoolsHandlerVariations(t *testing.T) {
	methods := []struct {
		httpMethod string
		expected   string
	}{
		{"GET", "get"},
		{"POST", "post"},
		{"PUT", "put"},
		{"DELETE", "delete"},
		{"PATCH", "post"}, // Default case
		{"", "post"},      // Empty default
	}

	for _, tt := range methods {
		t.Run(tt.httpMethod, func(t *testing.T) {
			config := ProjectConfig{
				ServiceName:   "test",
				Description:   "Test API",
				HTTPMethod:    tt.httpMethod,
				UsePowertools: true,
			}

			handler := generatePowertoolsHandler(config)
			if !contains(handler, "@app."+tt.expected) {
				t.Errorf("Expected handler to contain @app.%s, got:\n%s", tt.expected, handler)
			}
		})
	}
}

// TestDynamoDBConditionalGeneration tests DynamoDB file generation.
func TestDynamoDBConditionalGeneration(t *testing.T) {
	t.Run("generates DynamoDB files when enabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "with-dynamodb")

		config := ProjectConfig{
			ServiceName:   "test",
			FunctionName:  "handler",
			Description:   "Test",
			PythonVersion: "3.13",
			UseDynamoDB:   true,
			TableName:     "test-table",
		}

		err := Generate(projectDir, config)
		if err != nil {
			t.Fatal(err)
		}

		// Verify DynamoDB files exist
		dynamoFiles := []string{
			"service/dal/dynamodb_handler.py",
			"service/dal/models/db.py",
		}

		for _, file := range dynamoFiles {
			fullPath := filepath.Join(projectDir, file)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("Expected DynamoDB file not found: %s", file)
			}
		}
	})

	t.Run("does not generate DynamoDB files when disabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "without-dynamodb")

		config := ProjectConfig{
			ServiceName:   "test",
			FunctionName:  "handler",
			Description:   "Test",
			PythonVersion: "3.13",
			UseDynamoDB:   false,
		}

		err := Generate(projectDir, config)
		if err != nil {
			t.Fatal(err)
		}

		// Verify DynamoDB files do NOT exist
		dynamoFiles := []string{
			"service/dal/dynamodb_handler.py",
			"service/dal/models/db.py",
		}

		for _, file := range dynamoFiles {
			fullPath := filepath.Join(projectDir, file)
			if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
				t.Errorf("DynamoDB file should not exist: %s", file)
			}
		}
	})
}

// TestGeneratorFunctions tests individual generator functions.
func TestGeneratorFunctions(t *testing.T) {
	t.Run("generatePyProjectToml with Powertools", func(t *testing.T) {
		config := ProjectConfig{
			ServiceName:    "test",
			Description:    "Test service",
			PythonVersion:  "3.13",
			UsePowertools:  true,
			UseDynamoDB:    true,
			UseIdempotency: true,
		}

		content := generatePyProjectToml(config)
		expectedStrings := []string{
			"aws-lambda-powertools",
			"mypy-boto3-dynamodb",
			"cachetools",
			"pydantic",
			"^3.13",
		}

		for _, expected := range expectedStrings {
			if !contains(content, expected) {
				t.Errorf("pyproject.toml missing: %s", expected)
			}
		}
	})

	t.Run("generatePyProjectToml without Powertools", func(t *testing.T) {
		config := ProjectConfig{
			ServiceName:    "test",
			Description:    "Test service",
			PythonVersion:  "3.13",
			UsePowertools:  false,
			UseDynamoDB:    false,
			UseIdempotency: false,
		}

		content := generatePyProjectToml(config)
		if contains(content, "aws-lambda-powertools") {
			t.Error("Basic pyproject.toml should not contain Powertools")
		}
		if contains(content, "mypy-boto3-dynamodb") {
			t.Error("Basic pyproject.toml should not contain DynamoDB")
		}
		if contains(content, "cachetools") {
			t.Error("Basic pyproject.toml should not contain cachetools")
		}
	})

	t.Run("generateReadme contains project info", func(t *testing.T) {
		config := ProjectConfig{
			ServiceName:   "my-service",
			Description:   "My service description",
			PythonVersion: "3.13",
		}

		content := generateReadme(config)
		expectedStrings := []string{
			"my-service",
			"My service description",
		}

		for _, expected := range expectedStrings {
			if !contains(content, expected) {
				t.Errorf("README missing: %s", expected)
			}
		}
	})

	t.Run("generateGitignore contains patterns", func(t *testing.T) {
		config := ProjectConfig{}
		content := generateGitignore(config)

		expectedPatterns := []string{
			"__pycache__",
			".pytest_cache",
			".coverage",
			"*.py[cod]",
			".env",
			".venv",
		}

		for _, pattern := range expectedPatterns {
			if !contains(content, pattern) {
				t.Errorf("Gitignore missing pattern: %s", pattern)
			}
		}
	})

	t.Run("generateMakefile contains targets", func(t *testing.T) {
		config := ProjectConfig{ServiceName: "test"}
		content := generateMakefile(config)

		expectedTargets := []string{
			"install:",
			"test:",
			"lint:",
			"format:",
		}

		for _, target := range expectedTargets {
			if !contains(content, target) {
				t.Errorf("Makefile missing target: %s", target)
			}
		}
	})

	t.Run("generateBasicHandler is simple", func(t *testing.T) {
		config := ProjectConfig{}
		content := generateBasicHandler(config)

		if contains(content, "aws_lambda_powertools") {
			t.Error("Basic handler should not use Powertools")
		}
		if !contains(content, "lambda_handler") {
			t.Error("Basic handler should have lambda_handler function")
		}
		if !contains(content, "json.loads") {
			t.Error("Basic handler should parse JSON")
		}
	})
}

// Helper function.
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
