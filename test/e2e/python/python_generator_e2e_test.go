//go:build e2e
// +build e2e

package python

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPythonGeneratorE2E tests the complete Python Lambda generator workflow using Forge commands.
// This test:
// 1. Builds the Forge CLI binary
// 2. Generates a Python Lambda project using `forge new lambda`
// 3. Builds the Lambda using `task build`
// 4. Deploys infrastructure using `task deploy` (terraform apply)
// 5. Tests API Gateway endpoints with HTTP requests
// 6. Verifies DynamoDB operations
// 7. Cleans up with `task destroy` (terraform destroy) - even on test failure
func TestPythonGeneratorE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Check AWS credentials
	if !isAWSConfigured(t) {
		t.Skip("Skipping E2E test - AWS credentials not configured")
	}

	t.Log("==> Starting Python Lambda Generator E2E Test")

	// Generate unique project name
	uniqueID := random.UniqueId()
	projectName := "forge-python-e2e-" + strings.ToLower(uniqueID)
	t.Logf("Project name: %s", projectName)

	// Create temporary directory for test
	testDir := t.TempDir()
	projectDir := filepath.Join(testDir, projectName)
	t.Logf("Test directory: %s", projectDir)

	// Build Forge binary
	t.Log("==> Building Forge CLI binary")
	forgeBinary := buildForgeBinary(t, testDir)
	t.Logf("Forge binary: %s", forgeBinary)

	// Generate Python Lambda project using Forge CLI
	t.Log("==> Generating Python Lambda project with forge new lambda")
	generatePythonProject(t, forgeBinary, testDir, projectName)

	// Set up cleanup that runs even on panic
	cleanupDone := false
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Test panicked: %v", r)
		}
		if !cleanupDone && os.Getenv("SKIP_TEARDOWN") != "true" {
			t.Log("==> Cleaning up infrastructure (defer)")
			destroyInfrastructure(t, projectDir)
		}
	}()

	// Build Lambda package
	t.Log("==> Building Lambda package with task build")
	buildLambdaPackage(t, projectDir)

	// Deploy infrastructure
	t.Log("==> Deploying infrastructure with task deploy")
	deployInfrastructure(t, projectDir)

	// Get Terraform outputs
	t.Log("==> Retrieving deployment outputs")
	apiEndpoint := getTerraformOutput(t, projectDir, "api_endpoint")
	functionName := getTerraformOutput(t, projectDir, "lambda_function_name")
	tableName := getTerraformOutput(t, projectDir, "dynamodb_table_name")

	t.Logf("API Endpoint: %s", apiEndpoint)
	t.Logf("Function Name: %s", functionName)
	t.Logf("Table Name: %s", tableName)

	// Validate outputs
	require.NotEmpty(t, apiEndpoint, "API endpoint should not be empty")
	require.NotEmpty(t, functionName, "Function name should not be empty")
	require.NotEmpty(t, tableName, "Table name should not be empty")

	// Test Suite 1: API Gateway Endpoints
	t.Run("API_Gateway_Endpoints", func(t *testing.T) {
		testAPIEndpoints(t, apiEndpoint)
	})

	// Test Suite 2: DynamoDB Integration
	t.Run("DynamoDB_Integration", func(t *testing.T) {
		testDynamoDBIntegration(t, apiEndpoint, tableName)
	})

	// Test Suite 3: End-to-End Workflows
	t.Run("E2E_Workflows", func(t *testing.T) {
		testEndToEndWorkflows(t, apiEndpoint, tableName)
	})

	// Clean up infrastructure
	t.Log("==> Destroying infrastructure with task destroy")
	destroyInfrastructure(t, projectDir)
	cleanupDone = true

	t.Log("==> All E2E tests passed successfully!")
}

// buildForgeBinary builds the Forge CLI binary.
func buildForgeBinary(t *testing.T, tempDir string) string {
	t.Helper()

	binaryPath := filepath.Join(tempDir, "forge")

	// Get project root (3 levels up from test/e2e/python)
	projectRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err, "Failed to determine project root")

	cmd := exec.Command("go", "build", "-o", binaryPath, filepath.Join(projectRoot, "cmd", "forge"))
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build Forge binary: %s", string(output))

	// Verify binary exists
	_, err = os.Stat(binaryPath)
	require.NoError(t, err, "Forge binary not found at %s", binaryPath)

	return binaryPath
}

// generatePythonProject uses forge new lambda to generate a project.
func generatePythonProject(t *testing.T, forgeBinary, testDir, projectName string) {
	t.Helper()

	cmd := exec.Command(
		forgeBinary,
		"new", "lambda", projectName,
		"--service", projectName,
		"--api-path", "/api/orders",
		"--method", "POST",
	)
	cmd.Dir = testDir
	output, err := cmd.CombinedOutput()
	t.Logf("forge new lambda output:\n%s", string(output))
	require.NoError(t, err, "forge new lambda failed: %s", string(output))

	// Verify project structure
	projectDir := filepath.Join(testDir, projectName)
	requiredDirs := []string{
		"service",
		"service/handlers",
		"service/logic",
		"service/dal",
		"service/models",
		"terraform",
		"tests",
	}

	for _, dir := range requiredDirs {
		path := filepath.Join(projectDir, dir)
		_, err := os.Stat(path)
		require.NoError(t, err, "Required directory should exist: %s", dir)
	}

	// Verify Taskfile exists
	taskfilePath := filepath.Join(projectDir, "Taskfile.yml")
	_, err = os.Stat(taskfilePath)
	require.NoError(t, err, "Taskfile.yml should exist")
}

// buildLambdaPackage builds the Lambda deployment package.
func buildLambdaPackage(t *testing.T, projectDir string) {
	t.Helper()

	cmd := exec.Command("task", "build")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	t.Logf("task build output:\n%s", string(output))
	require.NoError(t, err, "task build failed: %s", string(output))

	// Verify build artifact exists
	buildArtifact := filepath.Join(projectDir, ".build", "lambda.zip")
	_, err = os.Stat(buildArtifact)
	require.NoError(t, err, "Build artifact should exist: .build/lambda.zip")
}

// deployInfrastructure deploys the infrastructure using task deploy.
func deployInfrastructure(t *testing.T, projectDir string) {
	t.Helper()

	cmd := exec.Command("task", "deploy")
	cmd.Dir = projectDir
	// Set environment to avoid interactive prompts
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=1")

	output, err := cmd.CombinedOutput()
	t.Logf("task deploy output:\n%s", string(output))
	require.NoError(t, err, "task deploy failed: %s", string(output))
}

// destroyInfrastructure destroys the infrastructure.
func destroyInfrastructure(t *testing.T, projectDir string) {
	t.Helper()

	cmd := exec.Command("task", "destroy")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=1")

	output, err := cmd.CombinedOutput()
	t.Logf("task destroy output:\n%s", string(output))
	if err != nil {
		t.Logf("Warning: task destroy failed: %v", err)
		// Don't fail the test on cleanup errors
	}
}

// getTerraformOutput gets a Terraform output value.
func getTerraformOutput(t *testing.T, projectDir, outputName string) string {
	t.Helper()

	terraformDir := filepath.Join(projectDir, "terraform")
	cmd := exec.Command("terraform", "output", "-raw", outputName)
	cmd.Dir = terraformDir

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to get terraform output %s: %s", outputName, string(output))

	return strings.TrimSpace(string(output))
}

// testAPIEndpoints tests the API Gateway endpoints.
func testAPIEndpoints(t *testing.T, apiEndpoint string) {
	t.Helper()

	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("POST_CreateOrder_Success", func(t *testing.T) {
		// Wait for API to be ready
		time.Sleep(5 * time.Second)

		url := apiEndpoint + "/api/orders"
		orderReq := OrderRequest{
			Name:  "E2E Test Order",
			Count: 10,
		}

		jsonData, err := json.Marshal(orderReq)
		require.NoError(t, err)

		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		t.Logf("Response status: %d", resp.StatusCode)
		t.Logf("Response body: %s", string(body))

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

		var orderResp OrderResponse
		err = json.Unmarshal(body, &orderResp)
		require.NoError(t, err, "Response should be valid JSON")

		assert.NotEmpty(t, orderResp.ID, "Order ID should not be empty")
		assert.Equal(t, orderReq.Name, orderResp.Name)
		assert.Equal(t, orderReq.Count, orderResp.Count)
		assert.Equal(t, "created", orderResp.Status)
	})

	t.Run("POST_CreateOrder_InvalidInput", func(t *testing.T) {
		url := apiEndpoint + "/api/orders"

		testCases := []struct {
			name    string
			request OrderRequest
		}{
			{"zero_count", OrderRequest{Name: "Test", Count: 0}},
			{"negative_count", OrderRequest{Name: "Test", Count: -1}},
			{"empty_name", OrderRequest{Name: "", Count: 5}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				jsonData, err := json.Marshal(tc.request)
				require.NoError(t, err)

				resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
				require.NoError(t, err)
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				t.Logf("Error case %s - Status: %d, Body: %s", tc.name, resp.StatusCode, string(body))

				// Should return 4xx error
				assert.True(t, resp.StatusCode >= 400 && resp.StatusCode < 500,
					"Expected 4xx error, got %d", resp.StatusCode)
			})
		}
	})
}

// testDynamoDBIntegration tests DynamoDB integration.
func testDynamoDBIntegration(t *testing.T, apiEndpoint, tableName string) {
	t.Helper()

	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("VerifyItemInDynamoDB", func(t *testing.T) {
		// Create order via API
		url := apiEndpoint + "/api/orders"
		orderReq := OrderRequest{
			Name:  "DynamoDB Test Order",
			Count: 42,
		}

		jsonData, err := json.Marshal(orderReq)
		require.NoError(t, err)

		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Order creation should succeed")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var orderResp OrderResponse
		err = json.Unmarshal(body, &orderResp)
		require.NoError(t, err)

		orderID := orderResp.ID
		require.NotEmpty(t, orderID, "Order ID should be present")

		// Wait for DynamoDB consistency
		time.Sleep(2 * time.Second)

		// Verify in DynamoDB
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(getAWSRegion()),
		})
		require.NoError(t, err)

		dynamoClient := dynamodb.New(sess)

		getInput := &dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {S: aws.String(orderID)},
			},
		}

		result, err := dynamoClient.GetItem(getInput)
		require.NoError(t, err, "Should be able to query DynamoDB")
		require.NotNil(t, result.Item, "Order should exist in DynamoDB")

		// Verify attributes
		assert.NotNil(t, result.Item["name"])
		assert.NotNil(t, result.Item["count"])
		assert.NotNil(t, result.Item["status"])

		if result.Item["name"] != nil && result.Item["name"].S != nil {
			assert.Equal(t, orderReq.Name, *result.Item["name"].S)
		}
	})
}

// testEndToEndWorkflows tests complete user workflows.
func testEndToEndWorkflows(t *testing.T, apiEndpoint, tableName string) {
	t.Helper()

	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("CreateMultipleOrders", func(t *testing.T) {
		url := apiEndpoint + "/api/orders"

		orders := []OrderRequest{
			{Name: "Order 1", Count: 1},
			{Name: "Order 2", Count: 2},
			{Name: "Order 3", Count: 3},
		}

		for i, order := range orders {
			jsonData, err := json.Marshal(order)
			require.NoError(t, err)

			resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode,
				"Order %d creation should succeed", i+1)

			body, _ := io.ReadAll(resp.Body)
			var orderResp OrderResponse
			err = json.Unmarshal(body, &orderResp)
			require.NoError(t, err)

			assert.NotEmpty(t, orderResp.ID)
			assert.Equal(t, order.Name, orderResp.Name)
			assert.Equal(t, order.Count, orderResp.Count)
		}
	})
}

// Helper functions

func isAWSConfigured(t *testing.T) bool {
	t.Helper()

	// Check environment variables
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" || os.Getenv("AWS_PROFILE") != "" {
		return true
	}

	// Check credentials file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	credPath := filepath.Join(homeDir, ".aws", "credentials")
	if _, err := os.Stat(credPath); err == nil {
		return true
	}

	configPath := filepath.Join(homeDir, ".aws", "config")
	if _, err := os.Stat(configPath); err == nil {
		return true
	}

	return false
}

func getAWSRegion() string {
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region
	}
	return "us-east-1"
}
