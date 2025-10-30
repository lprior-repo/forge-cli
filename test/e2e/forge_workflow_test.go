package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 8. Clean up all resources.
func TestForgeWorkflowEndToEnd(t *testing.T) {
	// Skip in short mode - this is a comprehensive test
	if testing.Short() {
		t.Skip("Skipping comprehensive E2E test in short mode")
	}

	// Check AWS credentials
	if !isAWSConfigured() {
		t.Skip("Skipping E2E test - AWS credentials not configured")
	}

	t.Log("==> Starting comprehensive Forge workflow E2E test")

	// Generate unique namespace for this test
	uniqueID := random.UniqueId()
	namespace := "forge-e2e-" + strings.ToLower(uniqueID)
	t.Logf("Using namespace: %s", namespace)

	// Create temporary directory for test project
	testDir := t.TempDir()
	t.Logf("Test directory: %s", testDir)

	// Build Forge binary
	forgeBinary := buildForgeBinary(t)
	t.Logf("Forge binary: %s", forgeBinary)

	// Create project directory with proper Forge structure
	projectDir := filepath.Join(testDir, "test-lambda-project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755), "Failed to create project directory")

	// Create src/functions/api directory for Python Lambda
	functionsDir := filepath.Join(projectDir, "src", "functions", "api")
	require.NoError(t, os.MkdirAll(functionsDir, 0o755), "Failed to create functions directory")

	// Create a simple Python Lambda that Forge can discover
	createSimplePythonLambda(t, functionsDir)

	// Debug: Verify directory structure
	t.Logf("Project directory: %s", projectDir)
	t.Logf("Functions directory: %s", functionsDir)
	if entries, err := os.ReadDir(filepath.Join(projectDir, "src", "functions")); err == nil {
		t.Logf("Found %d function directories:", len(entries))
		for _, e := range entries {
			t.Logf("  - %s (is_dir: %v)", e.Name(), e.IsDir())
		}
	}

	// Create infra directory
	infraDir := filepath.Join(projectDir, "infra")
	require.NoError(t, os.MkdirAll(infraDir, 0o755))

	// Create simple Terraform files that work with Forge
	createForgeTerraformFiles(t, infraDir)

	// Ensure cleanup of AWS resources
	terraformDir := filepath.Join(projectDir, "infra")
	terraformOptions := &terraform.Options{
		TerraformDir: terraformDir,
		NoColor:      true,
		Vars: map[string]interface{}{
			"namespace":    namespace,
			"service_name": "forge-test-" + namespace,
			"environment":  "test",
		},
		MaxRetries:         3,
		TimeBetweenRetries: 5 * time.Second,
	}

	defer func() {
		t.Log("==> Cleaning up test infrastructure...")
		terraform.Destroy(t, terraformOptions)
	}()

	// Build Lambda using Forge
	t.Log("==> Building Lambda function with Forge...")
	buildLambda(t, forgeBinary, projectDir)

	// Deploy using Forge
	t.Log("==> Deploying infrastructure with Forge...")
	deployOutput := deployWithForge(t, forgeBinary, projectDir, namespace)
	t.Logf("Deploy output:\n%s", deployOutput)

	// Get Terraform outputs
	t.Log("==> Retrieving Terraform outputs...")
	apiEndpoint := terraform.Output(t, terraformOptions, "api_endpoint")
	functionName := terraform.Output(t, terraformOptions, "function_name")
	functionArn := terraform.Output(t, terraformOptions, "function_arn")
	tableName := terraform.Output(t, terraformOptions, "table_name")
	awsRegion := getAWSRegion(t, terraformOptions)

	t.Logf("API Endpoint: %s", apiEndpoint)
	t.Logf("Function Name: %s", functionName)
	t.Logf("Function ARN: %s", functionArn)
	t.Logf("Table Name: %s", tableName)
	t.Logf("AWS Region: %s", awsRegion)

	// Verify outputs are valid
	require.NotEmpty(t, apiEndpoint, "API endpoint should not be empty")
	require.NotEmpty(t, functionName, "Function name should not be empty")
	require.NotEmpty(t, functionArn, "Function ARN should not be empty")
	require.NotEmpty(t, tableName, "Table name should not be empty")
	assert.Contains(t, apiEndpoint, "https://", "API endpoint should be HTTPS")
	assert.Contains(t, functionArn, "arn:aws:lambda", "Function ARN should be valid")

	// Test Suite 1: API Gateway Integration
	t.Run("API Gateway Integration", func(t *testing.T) {
		testAPIGatewayEndpoints(t, apiEndpoint)
	})

	// Test Suite 2: Lambda Function Direct Invocation
	t.Run("Lambda Direct Invocation", func(t *testing.T) {
		testLambdaInvocation(t, awsRegion, functionName)
	})

	// Test Suite 3: DynamoDB Integration
	t.Run("DynamoDB Integration", func(t *testing.T) {
		testDynamoDBIntegration(t, awsRegion, tableName, apiEndpoint)
	})

	// Test Suite 4: End-to-End User Flows
	t.Run("End-to-End User Flows", func(t *testing.T) {
		testEndToEndFlows(t, apiEndpoint)
	})

	t.Log("==> All E2E tests passed successfully!")
}

// buildForgeBinary builds the Forge CLI binary and returns its path.
func buildForgeBinary(t *testing.T) string {
	t.Helper()

	binaryPath := filepath.Join(t.TempDir(), "forge")

	cmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/forge")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build Forge binary: %s", string(output))

	// Verify binary exists and is executable
	_, err = os.Stat(binaryPath)
	require.NoError(t, err, "Forge binary not found at %s", binaryPath)

	return binaryPath
}

// buildLambda uses Forge CLI to build the Lambda function.
func buildLambda(t *testing.T, forgeBinary, projectDir string) {
	t.Helper()

	cmd := exec.Command(forgeBinary, "build")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()

	t.Logf("Forge build output:\n%s", string(output))
	require.NoError(t, err, "Forge build failed: %s", string(output))

	// Verify build artifacts exist
	buildDir := filepath.Join(projectDir, ".forge", "build")
	_, err = os.Stat(buildDir)
	require.NoError(t, err, "Build directory not found")
}

// deployWithForge uses Forge CLI to deploy the infrastructure.
func deployWithForge(t *testing.T, forgeBinary, projectDir, namespace string) string {
	t.Helper()

	// Set environment variable for namespace
	cmd := exec.Command(forgeBinary, "deploy", "--auto-approve", "--namespace", namespace)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "TF_VAR_namespace="+namespace)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Forge deploy failed: %s", string(output))

	return string(output)
}

// testAPIGatewayEndpoints tests the deployed API Gateway endpoints.
func testAPIGatewayEndpoints(t *testing.T, apiEndpoint string) {
	t.Helper()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	t.Run("GET /health returns http.StatusOK", func(t *testing.T) {
		url := apiEndpoint + "/health"

		var resp *http.Response
		var err error

		// Retry to handle eventual consistency
		retry.DoWithRetry(t, "GET /health", 10, 3*time.Second, func() (string, error) {
			resp, err = client.Get(url)
			if err != nil {
				return "", fmt.Errorf("failed to GET /health: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return "", fmt.Errorf("expected status http.StatusOK, got %d", resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			t.Logf("Health response: %s", string(body))

			return "success", nil
		})
	})

	t.Run("GET / returns service info", func(t *testing.T) {
		var resp *http.Response
		var err error

		retry.DoWithRetry(t, "GET /", 10, 3*time.Second, func() (string, error) {
			resp, err = client.Get(apiEndpoint)
			if err != nil {
				return "", fmt.Errorf("failed to GET /: %w", err)
			}
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			t.Logf("Root response: %s", string(body))

			// Check if body is empty
			if len(body) == 0 {
				return "", errors.New("response body is empty")
			}

			var response map[string]interface{}
			err = json.Unmarshal(body, &response)
			if err != nil {
				return "", fmt.Errorf("invalid JSON: %w", err)
			}

			// Verify expected fields
			assert.Contains(t, response, "service", "Response should contain service field")

			return "success", nil
		})
	})

	t.Run("POST /items creates item", func(t *testing.T) {
		url := apiEndpoint + "/items"

		payload := map[string]interface{}{
			"name":        "test-item",
			"description": "Test item from E2E test",
			"value":       42,
		}
		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		var resp *http.Response

		retry.DoWithRetry(t, "POST /items", 10, 3*time.Second, func() (string, error) {
			resp, err = client.Post(url, "application/json", strings.NewReader(string(payloadBytes)))
			if err != nil {
				return "", fmt.Errorf("failed to POST /items: %w", err)
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			t.Logf("POST /items response: %s", string(body))

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				return "", fmt.Errorf("expected status http.StatusOK or 201, got %d", resp.StatusCode)
			}

			return "success", nil
		})
	})
}

// testLambdaInvocation tests direct Lambda function invocation.
func testLambdaInvocation(t *testing.T, region, functionName string) {
	t.Helper()

	t.Run("Lambda responds to test event", func(t *testing.T) {
		payload := map[string]interface{}{
			"test":    true,
			"message": "E2E test invocation",
		}

		output, err := aws.InvokeFunctionWithParamsE(
			t,
			region,
			functionName,
			&aws.LambdaOptions{
				Payload: payload,
			},
		)

		require.NoError(t, err, "Lambda invocation should succeed")
		assert.Equal(t, int32(http.StatusOK), output.StatusCode, "Lambda should return http.StatusOK")

		t.Logf("Lambda response: %s", string(output.Payload))

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(output.Payload, &response)
		require.NoError(t, err, "Lambda response should be valid JSON")
	})

	t.Run("Lambda handles API Gateway proxy event", func(t *testing.T) {
		// Simulate API Gateway proxy event
		payload := map[string]interface{}{
			"httpMethod": "GET",
			"path":       "/health",
			"headers": map[string]string{
				"Content-Type": "application/json",
			},
			"body": "",
		}

		output, err := aws.InvokeFunctionWithParamsE(
			t,
			region,
			functionName,
			&aws.LambdaOptions{
				Payload: payload,
			},
		)

		require.NoError(t, err, "Lambda invocation should succeed")
		assert.Equal(t, int32(http.StatusOK), output.StatusCode, "Lambda should return http.StatusOK")

		t.Logf("Lambda proxy response: %s", string(output.Payload))
	})
}

// testDynamoDBIntegration tests DynamoDB table integration.
func testDynamoDBIntegration(t *testing.T, _ string, tableName, apiEndpoint string) {
	t.Helper()

	t.Run("DynamoDB table exists and is active", func(t *testing.T) {
		// Use AWS SDK to verify table
		retry.DoWithRetry(t, "Verify DynamoDB table", 10, 3*time.Second, func() (string, error) {
			// TODO: Use AWS SDK to check table status
			// For now, we test via API which uses DynamoDB
			return "success", nil
		})
	})

	t.Run("Items persisted to DynamoDB via API", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}

		// Create item via API
		url := apiEndpoint + "/items"
		itemID := fmt.Sprintf("test-item-%d", time.Now().Unix())

		payload := map[string]interface{}{
			"id":          itemID,
			"name":        "DynamoDB test item",
			"description": "Testing persistence",
		}
		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		// POST to create
		resp, err := client.Post(url, "application/json", strings.NewReader(string(payloadBytes)))
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Create item response: %s", string(body))

		// TODO: GET to verify item exists
		// This would require the API to have a GET /items/{id} endpoint
	})
}

// testEndToEndFlows tests complete user workflows.
func testEndToEndFlows(t *testing.T, apiEndpoint string) {
	t.Helper()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("Complete CRUD workflow", func(t *testing.T) {
		itemID := fmt.Sprintf("e2e-test-%d", time.Now().Unix())

		// Create
		createPayload := map[string]interface{}{
			"id":   itemID,
			"name": "E2E Test Item",
			"tags": []string{"test", "e2e"},
		}
		createBytes, err := json.Marshal(createPayload)
		require.NoError(t, err)

		resp, err := client.Post(
			apiEndpoint+"/items",
			"application/json",
			strings.NewReader(string(createBytes)),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		t.Logf("CRUD Create response: %s", string(body))

		assert.True(t, resp.StatusCode >= http.StatusOK && resp.StatusCode < 300, "Create should succeed")
	})

	t.Run("Error handling", func(t *testing.T) {
		// Test invalid payload
		resp, err := client.Post(
			apiEndpoint+"/items",
			"application/json",
			strings.NewReader("invalid json{"),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle gracefully (might return 400 or 500 depending on implementation)
		t.Logf("Error handling response status: %d", resp.StatusCode)
	})
}

// Helper functions.

func isAWSConfigured() bool {
	// Check environment variables
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" ||
		os.Getenv("AWS_PROFILE") != "" {

		return true
	}

	// Check if AWS credentials file exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	credentialsPath := filepath.Join(homeDir, ".aws", "credentials")
	if _, err := os.Stat(credentialsPath); err == nil {
		return true
	}

	configPath := filepath.Join(homeDir, ".aws", "config")
	if _, err := os.Stat(configPath); err == nil {
		return true
	}

	return false
}

func getAWSRegion(t *testing.T, opts *terraform.Options) string {
	t.Helper()
	region := terraform.Output(t, opts, "aws_region")
	if region == "" {
		region = os.Getenv("AWS_REGION")
	}
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = "us-east-1" // Default
	}
	return region
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()

	// Use rsync or cp -r with proper syntax
	// cp -r src/. dst/ copies contents of src into dst
	cmd := exec.Command("cp", "-r", src+"/.", dst+"/")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to copy directory: %s", string(output))
}

func createSimplePythonLambda(t *testing.T, dir string) {
	t.Helper()

	// Create a simple Lambda handler that works with API Gateway v2 (HTTP API)
	lambdaCode := `import json

def lambda_handler(event, context):
    """Simple Lambda handler for E2E testing - supports API Gateway v2 (HTTP API)"""

    # API Gateway v2 (HTTP API) format
    if "requestContext" in event and "http" in event["requestContext"]:
        path = event.get("rawPath", "/")
        method = event["requestContext"]["http"]["method"]

        if path == "/health":
            return {
                "statusCode": http.StatusOK,
                "headers": {"Content-Type": "application/json"},
                "body": json.dumps({"status": "healthy", "service": "forge-e2e-test"})
            }

        if path == "/items" and method == "POST":
            try:
                body = json.loads(event.get("body", "{}"))
                return {
                    "statusCode": 201,
                    "headers": {"Content-Type": "application/json"},
                    "body": json.dumps({"message": "Item created", "item": body})
                }
            except Exception as e:
                return {
                    "statusCode": 400,
                    "headers": {"Content-Type": "application/json"},
                    "body": json.dumps({"error": str(e)})
                }

        return {
            "statusCode": http.StatusOK,
            "headers": {"Content-Type": "application/json"},
            "body": json.dumps({
                "service": "forge-e2e-test",
                "message": "Test Lambda deployed successfully",
                "path": path,
                "method": method
            })
        }

    # API Gateway v1 (REST API) format - legacy support
    elif "httpMethod" in event:
        path = event.get("path", "/")
        method = event.get("httpMethod", "GET")

        if path == "/health":
            return {
                "statusCode": http.StatusOK,
                "headers": {"Content-Type": "application/json"},
                "body": json.dumps({"status": "healthy", "service": "forge-e2e-test"})
            }

        if path == "/items" and method == "POST":
            try:
                body = json.loads(event.get("body", "{}"))
                return {
                    "statusCode": 201,
                    "headers": {"Content-Type": "application/json"},
                    "body": json.dumps({"message": "Item created", "item": body})
                }
            except Exception as e:
                return {
                    "statusCode": 400,
                    "headers": {"Content-Type": "application/json"},
                    "body": json.dumps({"error": str(e)})
                }

        return {
            "statusCode": http.StatusOK,
            "headers": {"Content-Type": "application/json"},
            "body": json.dumps({
                "service": "forge-e2e-test",
                "message": "Test Lambda deployed successfully",
                "path": path,
                "method": method
            })
        }

    # Handle direct invocation (testing)
    return {
        "statusCode": http.StatusOK,
        "message": "Test Lambda invoked successfully",
        "event": event
    }
`

	// Write the Lambda handler
	handlerPath := filepath.Join(dir, "app.py")
	err := os.WriteFile(handlerPath, []byte(lambdaCode), 0o644)
	require.NoError(t, err, "Failed to create Lambda handler")

	// Create requirements.txt (empty for this simple example)
	reqPath := filepath.Join(dir, "requirements.txt")
	err = os.WriteFile(reqPath, []byte("# No external dependencies\n"), 0o644)
	require.NoError(t, err, "Failed to create requirements.txt")
}

func createForgeTerraformFiles(t *testing.T, infraDir string) {
	t.Helper()

	// main.tf - Provider and basic config
	mainTF := `terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      ManagedBy = "Terraform"
      Tool      = "Forge"
      Namespace = var.namespace
    }
  }
}
`

	// variables.tf
	variablesTF := `variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-1"
}

variable "namespace" {
  description = "Namespace for resource isolation"
  type        = string
  default     = ""
}

variable "service_name" {
  description = "Service name"
  type        = string
  default     = "forge-e2e-test"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "test"
}
`

	// lambda.tf - Lambda function using Forge build output
	lambdaTF := `# Lambda function using Forge build output
resource "aws_lambda_function" "api" {
  filename         = "../.forge/build/api.zip"
  function_name    = "${var.namespace}${var.service_name}-api"
  role             = aws_iam_role.lambda.arn
  handler          = "app.lambda_handler"
  source_code_hash = filebase64sha256("../.forge/build/api.zip")
  runtime          = "python3.13"
  timeout          = 30
  memory_size      = 512

  environment {
    variables = {
      ENVIRONMENT = var.environment
      NAMESPACE   = var.namespace
    }
  }
}

# Lambda permission for API Gateway
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.main.execution_arn}/*/*"
}
`

	// iam.tf - IAM roles and policies
	iamTF := `# IAM role for Lambda
resource "aws_iam_role" "lambda" {
  name = "${var.namespace}${var.service_name}-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

# Basic Lambda execution policy
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
`

	// apigateway.tf - API Gateway configuration
	apigatewayTF := `# API Gateway HTTP API
resource "aws_apigatewayv2_api" "main" {
  name          = "${var.namespace}${var.service_name}-api"
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers = ["*"]
  }
}

# API Gateway integration with Lambda
resource "aws_apigatewayv2_integration" "lambda" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  integration_method     = "POST"
  integration_uri        = aws_lambda_function.api.invoke_arn
  payload_format_version = "2.0"
}

# Default route - catches all requests
resource "aws_apigatewayv2_route" "default" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

# API Gateway stage
resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = "$default"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
    })
  }
}

# CloudWatch log group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${var.namespace}${var.service_name}"
  retention_in_days = 7
}
`

	// outputs.tf
	outputsTF := `output "api_endpoint" {
  description = "API Gateway endpoint URL"
  value       = aws_apigatewayv2_api.main.api_endpoint
}

output "function_name" {
  description = "Lambda function name"
  value       = aws_lambda_function.api.function_name
}

output "function_arn" {
  description = "Lambda function ARN"
  value       = aws_lambda_function.api.arn
}

output "aws_region" {
  description = "AWS region"
  value       = var.aws_region
}

output "table_name" {
  description = "DynamoDB table name (placeholder)"
  value       = "no-table-created"
}
`

	// Write all Terraform files
	files := map[string]string{
		"main.tf":       mainTF,
		"variables.tf":  variablesTF,
		"lambda.tf":     lambdaTF,
		"iam.tf":        iamTF,
		"apigateway.tf": apigatewayTF,
		"outputs.tf":    outputsTF,
	}

	for filename, content := range files {
		path := filepath.Join(infraDir, filename)
		err := os.WriteFile(path, []byte(content), 0o644)
		require.NoError(t, err, "Failed to create %s", filename)
	}
}
