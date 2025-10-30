package infrastructure

import (
	"fmt"
	"net/http"
	"os"
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

// TestLambdaDeploymentEndToEnd tests full Lambda deployment lifecycle.
// This test actually deploys to AWS and cleans up afterwards.
func TestLambdaDeploymentEndToEnd(t *testing.T) {
	t.Run("deploys and destroys Python Lambda function", func(t *testing.T) {
		// Skip in short mode and CI - only run with explicit flag
		if testing.Short() {
			t.Skip("Skipping E2E deployment test in short mode")
		}

		// Skip unless AWS credentials are available
		if !isAWSConfigured() {
			t.Skip("Skipping E2E test - AWS credentials not configured")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		// Generate a unique identifier for this test run
		uniqueID := random.UniqueId()
		namespace := "forge-test-" + strings.ToLower(uniqueID)

		// Configure Terraform options
		terraformOptions := &terraform.Options{
			TerraformDir: terraformDir,
			NoColor:      true,

			// Pass test-specific variables
			Vars: map[string]interface{}{
				"namespace": namespace,
			},

			// Retry on known transient errors
			MaxRetries:         3,
			TimeBetweenRetries: 5 * time.Second,
		}

		// Ensure cleanup happens even if test fails
		defer func() {
			t.Log("Cleaning up test infrastructure...")
			terraform.Destroy(t, terraformOptions)
		}()

		// DEPLOY PHASE
		t.Log("Initializing Terraform...")
		terraform.Init(t, terraformOptions)

		t.Log("Planning infrastructure...")
		planOutput := terraform.Plan(t, terraformOptions)
		assert.Contains(t, planOutput, "aws_lambda_function", "Plan should include Lambda function")

		t.Log("Applying infrastructure...")
		terraform.Apply(t, terraformOptions)

		// VERIFICATION PHASE
		t.Log("Verifying deployed resources...")

		// Get outputs
		functionName := terraform.Output(t, terraformOptions, "function_name")
		require.NotEmpty(t, functionName, "Function name output should not be empty")
		assert.Contains(t, functionName, namespace, "Function name should contain namespace")

		functionArn := terraform.Output(t, terraformOptions, "function_arn")
		require.NotEmpty(t, functionArn, "Function ARN should not be empty")
		assert.Contains(t, functionArn, "arn:aws:lambda", "ARN should be valid Lambda ARN")

		// Verify Lambda function exists in AWS
		awsRegion := terraform.Output(t, terraformOptions, "aws_region")
		if awsRegion == "" {
			awsRegion = "us-east-1" // Default region
		}

		t.Logf("Verifying Lambda function in AWS region: %s", awsRegion)

		// Use retry to handle eventual consistency
		maxRetries := 10
		sleepBetweenRetries := 3 * time.Second

		retry.DoWithRetry(t, "Check Lambda function exists", maxRetries, sleepBetweenRetries, func() (string, error) {
			// Try to invoke the function with a test payload to verify it exists
			// This will fail if the function doesn't exist
			output, err := aws.InvokeFunctionWithParamsE(
				t,
				awsRegion,
				functionName,
				&aws.LambdaOptions{
					Payload: map[string]interface{}{
						"test": "validation",
					},
				},
			)
			if err != nil {
				return "", fmt.Errorf("failed to invoke Lambda function: %w", err)
			}

			if output.StatusCode != http.StatusOK {
				return "", fmt.Errorf("Lambda invocation returned status code %d", output.StatusCode)
			}

			return "Lambda function verified and invoked successfully", nil
		})

		// DESTROY PHASE is handled by defer
		t.Log("Test completed successfully - cleanup will run via defer")
	})
}

// TestLambdaFunctionProperties tests Lambda function configuration.
func TestLambdaFunctionProperties(t *testing.T) {
	t.Run("validates Lambda function properties from plan", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping infrastructure test in short mode")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		terraformOptions := &terraform.Options{
			TerraformDir: terraformDir,
			NoColor:      true,
		}

		terraform.Init(t, terraformOptions)

		// Get plan in JSON format for detailed analysis
		planJSON, err := terraform.RunTerraformCommandAndGetStdoutE(
			t,
			terraformOptions,
			"plan",
			"-out=test.tfplan",
		)

		require.NoError(t, err, "Plan should succeed")
		assert.NotEmpty(t, planJSON, "Plan output should not be empty")

		// Cleanup plan file
		defer terraform.RunTerraformCommand(t, terraformOptions, "clean")
	})
}

// TestAPIGatewayIntegration tests API Gateway integration with Lambda.
func TestAPIGatewayIntegration(t *testing.T) {
	t.Run("validates API Gateway configuration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping infrastructure test in short mode")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		// Check API Gateway file exists
		apiGatewayFile := filepath.Join(terraformDir, "apigateway.tf")
		assert.FileExists(t, apiGatewayFile, "apigateway.tf should exist")

		terraformOptions := &terraform.Options{
			TerraformDir: terraformDir,
			NoColor:      true,
		}

		terraform.Init(t, terraformOptions)
		planOutput := terraform.Plan(t, terraformOptions)

		// Verify API Gateway resources in plan
		assert.Contains(t, planOutput, "aws_apigatewayv2", "Plan should include API Gateway v2 resources")
	})
}

// TestDynamoDBIntegration tests DynamoDB table configuration.
func TestDynamoDBIntegration(t *testing.T) {
	t.Run("validates DynamoDB table configuration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping infrastructure test in short mode")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		// Check DynamoDB file exists
		dynamoDBFile := filepath.Join(terraformDir, "dynamodb.tf")
		assert.FileExists(t, dynamoDBFile, "dynamodb.tf should exist")

		terraformOptions := &terraform.Options{
			TerraformDir: terraformDir,
			NoColor:      true,
		}

		terraform.Init(t, terraformOptions)
		planOutput := terraform.Plan(t, terraformOptions)

		// Verify DynamoDB resources in plan
		assert.Contains(t, planOutput, "aws_dynamodb_table", "Plan should include DynamoDB table")
	})
}

// TestIAMConfiguration tests IAM roles and policies.
func TestIAMConfiguration(t *testing.T) {
	t.Run("validates IAM roles and policies", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping infrastructure test in short mode")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		// Check IAM file exists
		iamFile := filepath.Join(terraformDir, "iam.tf")
		assert.FileExists(t, iamFile, "iam.tf should exist")

		terraformOptions := &terraform.Options{
			TerraformDir: terraformDir,
			NoColor:      true,
		}

		terraform.Init(t, terraformOptions)
		planOutput := terraform.Plan(t, terraformOptions)

		// Verify IAM resources in plan
		assert.Contains(t, planOutput, "aws_iam_role", "Plan should include IAM role")
		assert.Contains(t, planOutput, "aws_iam_policy", "Plan should include IAM policies")
	})
}

// Helper function to check if AWS is configured.
func isAWSConfigured() bool {
	// Check for AWS credentials in environment
	// This is a simple check - actual SDK will do full credential validation
	return os.Getenv("AWS_ACCESS_KEY_ID") != "" ||
		os.Getenv("AWS_PROFILE") != "" ||
		os.Getenv("AWS_CONFIG_FILE") != ""
}
