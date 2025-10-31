package python_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lewis/forge/internal/generators/python"
)

// TestPythonGeneratorIntegration tests the full Python generator with Terraform validation.
func TestPythonGeneratorIntegration(t *testing.T) {
	// Create temporary directory for test project
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-python-project")

	// Generate project with all features enabled
	config := python.ProjectConfig{
		ServiceName:    "test-service",
		FunctionName:   "test-function",
		Description:    "Test Python Lambda service",
		PythonVersion:  "3.13",
		UsePowertools:  true,
		UseIdempotency: true,
		UseDynamoDB:    true,
		TableName:      "test-table",
		APIPath:        "/api/test",
		HTTPMethod:     "POST",
	}

	// Generate the project
	err := python.Generate(projectPath, config)
	if err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify directory structure
	requiredDirs := []string{
		"service",
		"service/handlers",
		"service/logic",
		"service/dal",
		"service/models",
		"tests",
		"terraform",
	}

	for _, dir := range requiredDirs {
		path := filepath.Join(projectPath, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Required directory missing: %s", dir)
		}
	}

	// Verify required files
	requiredFiles := []string{
		"requirements.txt",
		"README.md",
		".gitignore",
		"Taskfile.yml",
		"terraform/main.tf",
		"terraform/variables.tf",
		"terraform/outputs.tf",
		"terraform/lambda.tf",
		"terraform/apigateway.tf",
		"terraform/dynamodb.tf",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(projectPath, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Required file missing: %s", file)
		}
	}

	// Verify requirements.txt contains expected dependencies
	requirementsPath := filepath.Join(projectPath, "requirements.txt")
	content, err := os.ReadFile(requirementsPath)
	if err != nil {
		t.Fatalf("Failed to read requirements.txt: %v", err)
	}

	requiredDeps := []string{
		"pydantic",
		"boto3",
		"aws-lambda-powertools",
		"mypy-boto3-dynamodb",
		"cachetools",
	}

	for _, dep := range requiredDeps {
		if !contains(string(content), dep) {
			t.Errorf("requirements.txt missing dependency: %s", dep)
		}
	}

	// Verify Terraform files use modules
	lambdaTf := filepath.Join(projectPath, "terraform/lambda.tf")
	lambdaContent, err := os.ReadFile(lambdaTf)
	if err != nil {
		t.Fatalf("Failed to read lambda.tf: %v", err)
	}

	if !contains(string(lambdaContent), "terraform-aws-modules/lambda/aws") {
		t.Error("lambda.tf should use terraform-aws-modules/lambda/aws")
	}

	// Run terraform init (if terraform is available)
	terraformPath := filepath.Join(projectPath, "terraform")
	if _, err := exec.LookPath("terraform"); err == nil {
		t.Log("Running terraform init...")
		cmd := exec.Command("terraform", "init")
		cmd.Dir = terraformPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Terraform init output: %s", output)
			t.Errorf("Terraform init failed: %v", err)
		} else {
			t.Log("✓ Terraform init successful")

			// Run terraform validate
			t.Log("Running terraform validate...")
			cmd = exec.Command("terraform", "validate")
			cmd.Dir = terraformPath
			output, err = cmd.CombinedOutput()
			if err != nil {
				t.Logf("Terraform validate output: %s", output)
				t.Errorf("Terraform validate failed: %v", err)
			} else {
				t.Log("✓ Terraform validate successful")
			}
		}
	} else {
		t.Log("Skipping Terraform validation (terraform not in PATH)")
	}
}

// TestPythonGeneratorMinimal tests generation without optional features.
func TestPythonGeneratorMinimal(t *testing.T) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "minimal-project")

	config := python.ProjectConfig{
		ServiceName:   "minimal",
		FunctionName:  "handler",
		Description:   "Minimal Lambda",
		PythonVersion: "3.13",
		APIPath:       "/",
		HTTPMethod:    "GET",
	}

	err := python.Generate(projectPath, config)
	if err != nil {
		t.Fatalf("Failed to generate minimal project: %v", err)
	}

	// Verify DynamoDB file is NOT generated
	dynamoPath := filepath.Join(projectPath, "terraform/dynamodb.tf")
	if _, err := os.Stat(dynamoPath); !os.IsNotExist(err) {
		t.Error("dynamodb.tf should not exist for minimal config")
	}

	// Verify requirements.txt only has minimal dependencies
	requirementsPath := filepath.Join(projectPath, "requirements.txt")
	content, err := os.ReadFile(requirementsPath)
	if err != nil {
		t.Fatalf("Failed to read requirements.txt: %v", err)
	}

	if contains(string(content), "aws-lambda-powertools") {
		t.Error("Minimal config should not include powertools")
	}

	if contains(string(content), "mypy-boto3-dynamodb") {
		t.Error("Minimal config should not include dynamodb types")
	}
}

// Helper function to check if string contains substring.
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
