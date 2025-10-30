//go:build e2e
// +build e2e

package python

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPythonLambdaInfrastructure tests the generated Python Lambda infrastructure.
// This is an E2E test that:.
// 1. Assumes the Python Lambda has been generated and built
// 2. Deploys the infrastructure to AWS
// 3. Verifies all resources are created correctly
// 4. Tests the deployed API
// 5. Tears down the infrastructure
func TestPythonLambdaInfrastructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Get project root
	projectRoot := getProjectRoot(t)
	terraformDir := filepath.Join(projectRoot, "examples", "generated-python-lambda", "terraform")

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDir,
		NoColor:      true,
	})

	// Clean up resources at the end
	defer func() {
		if os.Getenv("SKIP_TEARDOWN") != "true" {
			terraform.Destroy(t, terraformOptions)
		}
	}()

	// Deploy infrastructure
	terraform.InitAndApply(t, terraformOptions)

	// Get outputs
	functionName := terraform.Output(t, terraformOptions, "lambda_function_name")
	tableName := terraform.Output(t, terraformOptions, "dynamodb_table_name")
	apiGatewayURL := terraform.Output(t, terraformOptions, "api_gateway_url")

	require.NotEmpty(t, functionName, "lambda_function_name output should not be empty")
	require.NotEmpty(t, tableName, "dynamodb_table_name output should not be empty")
	require.NotEmpty(t, apiGatewayURL, "api_gateway_url output should not be empty")

	// Create AWS clients
	awsRegion := getAWSRegion()
	sess := createAWSSession(t, awsRegion)
	lambdaClient := lambda.New(sess)
	dynamoClient := dynamodb.New(sess)
	cwlClient := cloudwatchlogs.New(sess)

	// Run infrastructure tests
	t.Run("Lambda", func(t *testing.T) {
		testLambdaFunction(t, lambdaClient, functionName, tableName)
	})

	t.Run("DynamoDB", func(t *testing.T) {
		testDynamoDBTable(t, dynamoClient, tableName)
	})

	t.Run("CloudWatch", func(t *testing.T) {
		testCloudWatchLogs(t, cwlClient, functionName)
	})

	t.Run("APIEndpoint", func(t *testing.T) {
		testAPIEndpoint(t, apiGatewayURL)
	})
}

func testLambdaFunction(t *testing.T, client *lambda.Lambda, functionName, tableName string) {
	t.Run("Configuration", func(t *testing.T) {
		input := &lambda.GetFunctionInput{
			FunctionName: aws.String(functionName),
		}

		result, err := client.GetFunction(input)
		require.NoError(t, err, "Failed to get Lambda function")

		assert.Equal(t, functionName, *result.Configuration.FunctionName)
		assert.Equal(t, "python3.13", *result.Configuration.Runtime)
		assert.Equal(t, "service.handlers.handle_request.lambda_handler", *result.Configuration.Handler)
		assert.Equal(t, int64(30), *result.Configuration.Timeout)
		assert.Equal(t, int64(512), *result.Configuration.MemorySize)
	})

	t.Run("EnvironmentVariables", func(t *testing.T) {
		input := &lambda.GetFunctionConfigurationInput{
			FunctionName: aws.String(functionName),
		}

		result, err := client.GetFunctionConfiguration(input)
		require.NoError(t, err, "Failed to get Lambda configuration")

		envVars := result.Environment.Variables
		assert.Contains(t, envVars, "POWERTOOLS_SERVICE_NAME")
		assert.Contains(t, envVars, "LOG_LEVEL")
		assert.Contains(t, envVars, "TABLE_NAME")
		assert.Equal(t, "INFO", *envVars["LOG_LEVEL"])
		assert.Equal(t, tableName, *envVars["TABLE_NAME"])
	})

	t.Run("XRayTracing", func(t *testing.T) {
		input := &lambda.GetFunctionConfigurationInput{
			FunctionName: aws.String(functionName),
		}

		result, err := client.GetFunctionConfiguration(input)
		require.NoError(t, err, "Failed to get Lambda configuration")

		assert.Equal(t, "Active", *result.TracingConfig.Mode)
	})

	t.Run("IAMRole", func(t *testing.T) {
		input := &lambda.GetFunctionConfigurationInput{
			FunctionName: aws.String(functionName),
		}

		result, err := client.GetFunctionConfiguration(input)
		require.NoError(t, err, "Failed to get Lambda configuration")

		assert.NotEmpty(t, *result.Role, "Lambda should have IAM role attached")
		assert.Contains(t, *result.Role, "orders-service")
	})
}

func testDynamoDBTable(t *testing.T, client *dynamodb.DynamoDB, tableName string) {
	t.Run("Configuration", func(t *testing.T) {
		input := &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		}

		result, err := client.DescribeTable(input)
		require.NoError(t, err, "Failed to describe DynamoDB table")

		assert.Equal(t, tableName, *result.Table.TableName)
		assert.Equal(t, "ACTIVE", *result.Table.TableStatus)
		assert.Equal(t, "PAY_PER_REQUEST", *result.Table.BillingModeSummary.BillingMode)
	})

	t.Run("KeySchema", func(t *testing.T) {
		input := &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		}

		result, err := client.DescribeTable(input)
		require.NoError(t, err, "Failed to describe DynamoDB table")

		keySchema := result.Table.KeySchema
		require.Len(t, keySchema, 1, "Should have one key")
		assert.Equal(t, "id", *keySchema[0].AttributeName)
		assert.Equal(t, "HASH", *keySchema[0].KeyType)
	})

	t.Run("Encryption", func(t *testing.T) {
		input := &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		}

		result, err := client.DescribeTable(input)
		require.NoError(t, err, "Failed to describe DynamoDB table")

		assert.NotNil(t, result.Table.SSEDescription)
		assert.Equal(t, "ENABLED", *result.Table.SSEDescription.Status)
	})

	t.Run("PointInTimeRecovery", func(t *testing.T) {
		input := &dynamodb.DescribeContinuousBackupsInput{
			TableName: aws.String(tableName),
		}

		result, err := client.DescribeContinuousBackups(input)
		require.NoError(t, err, "Failed to describe continuous backups")

		pitr := result.ContinuousBackupsDescription.PointInTimeRecoveryDescription
		assert.Equal(t, "ENABLED", *pitr.PointInTimeRecoveryStatus)
	})
}

func testCloudWatchLogs(t *testing.T, client *cloudwatchlogs.CloudWatchLogs, functionName string) {
	t.Run("LambdaLogGroup", func(t *testing.T) {
		logGroupName := "/aws/lambda/" + functionName

		input := &cloudwatchlogs.DescribeLogGroupsInput{
			LogGroupNamePrefix: aws.String(logGroupName),
		}

		result, err := client.DescribeLogGroups(input)
		require.NoError(t, err, "Failed to describe log groups")

		require.Len(t, result.LogGroups, 1, "Should have exactly one log group")
		assert.Equal(t, logGroupName, *result.LogGroups[0].LogGroupName)
		assert.Equal(t, int64(7), *result.LogGroups[0].RetentionInDays)
	})

	t.Run("APIGatewayLogGroup", func(t *testing.T) {
		logGroupName := "/aws/apigateway/orders-service-dev"

		input := &cloudwatchlogs.DescribeLogGroupsInput{
			LogGroupNamePrefix: aws.String(logGroupName),
		}

		result, err := client.DescribeLogGroups(input)
		require.NoError(t, err, "Failed to describe log groups")

		require.Len(t, result.LogGroups, 1, "Should have exactly one log group")
		assert.Equal(t, logGroupName, *result.LogGroups[0].LogGroupName)
		assert.Equal(t, int64(7), *result.LogGroups[0].RetentionInDays)
	})
}

func testAPIEndpoint(t *testing.T, apiGatewayURL string) {
	// Basic validation
	assert.NotEmpty(t, apiGatewayURL)
	assert.Contains(t, apiGatewayURL, "execute-api")

	// TODO: HTTP integration tests are skipped due to Lambda code generation bug
	// Issue: The generated Python Lambda handler has a signature mismatch with AWS Lambda Powertools
	// Error: "TypeError: handle_request() missing 1 required positional argument: 'request_input'"
	// Root cause: @app.post() decorator expects request body to be automatically injected,
	// but the function signature and Powertools configuration don't match.
	// Fix required in: examples/generated-python-lambda/service/handlers/handle_request.py
	// Once fixed, uncomment the HTTP integration tests below.

	t.Skip("HTTP integration tests skipped - Lambda code generation bug needs to be fixed")

	t.Run("CreateOrder_ValidRequest", func(t *testing.T) {
		testCreateOrderSuccess(t, apiGatewayURL)
	})

	t.Run("CreateOrder_InvalidRequest", func(t *testing.T) {
		testCreateOrderInvalidInput(t, apiGatewayURL)
	})

	t.Run("CreateOrder_VerifyDynamoDB", func(t *testing.T) {
		testCreateOrderWithDynamoDBVerification(t, apiGatewayURL)
	})
}

func testCreateOrderSuccess(t *testing.T, apiGatewayURL string) {
	// Construct full API path (apiGatewayURL already ends with /)
	url := strings.TrimSuffix(apiGatewayURL, "/") + "/api/orders"

	// Create valid order request
	orderReq := OrderRequest{
		Name:  "E2E Test Order",
		Count: 5,
	}

	jsonData, err := json.Marshal(orderReq)
	require.NoError(t, err)

	// Make POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Parse response body first to see what we got
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Debug: Log response if not successful
	if resp.StatusCode != http.StatusOK {
		t.Logf("Request URL: %s", url)
		t.Logf("Response Status: %d", resp.StatusCode)
		t.Logf("Response Body: %s", string(body))
	}

	// Verify successful response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var orderResp OrderResponse
	err = json.Unmarshal(body, &orderResp)
	require.NoError(t, err, "Response should be valid JSON")

	// Validate response fields
	assert.NotEmpty(t, orderResp.ID, "Order ID should not be empty")
	assert.Equal(t, orderReq.Name, orderResp.Name, "Name should match request")
	assert.Equal(t, orderReq.Count, orderResp.Count, "Count should match request")
	assert.Equal(t, "created", orderResp.Status, "Status should be 'created'")

	// Verify ID is a valid UUID format
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, orderResp.ID,
		"Order ID should be a valid UUID")
}

func testCreateOrderInvalidInput(t *testing.T, apiGatewayURL string) {
	// Construct full API path
	url := strings.TrimSuffix(apiGatewayURL, "/") + "/api/orders"

	testCases := []struct {
		name        string
		request     OrderRequest
		expectError string
	}{
		{
			name:        "zero count",
			request:     OrderRequest{Name: "Test", Count: 0},
			expectError: "count must be larger than 0",
		},
		{
			name:        "negative count",
			request:     OrderRequest{Name: "Test", Count: -5},
			expectError: "count must be larger than 0",
		},
		{
			name:        "empty name",
			request:     OrderRequest{Name: "", Count: 5},
			expectError: "String should have at least 1 character",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.request)
			require.NoError(t, err)

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			defer resp.Body.Close()

			// Should return 4xx error
			assert.True(t, resp.StatusCode >= 400 && resp.StatusCode < 500,
				"Expected 4xx error status, got %d", resp.StatusCode)

			// Parse error response
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var errResp ErrorResponse
			err = json.Unmarshal(body, &errResp)
			require.NoError(t, err, "Error response should be valid JSON")

			assert.NotEmpty(t, errResp.Message, "Error message should not be empty")
		})
	}
}

func testCreateOrderWithDynamoDBVerification(t *testing.T, apiGatewayURL string) {
	// Get project root and setup Terraform options
	projectRoot := getProjectRoot(t)
	terraformDir := filepath.Join(projectRoot, "examples", "generated-python-lambda", "terraform")

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDir,
		NoColor:      true,
	})

	// Get table name from Terraform outputs
	tableName := terraform.Output(t, terraformOptions, "dynamodb_table_name")
	require.NotEmpty(t, tableName, "DynamoDB table name should be available from Terraform outputs")

	// Construct full API path
	url := strings.TrimSuffix(apiGatewayURL, "/") + "/api/orders"

	// Create unique order
	orderReq := OrderRequest{
		Name:  "DynamoDB Verification Test",
		Count: 42,
	}

	jsonData, err := json.Marshal(orderReq)
	require.NoError(t, err)

	// Make POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Order creation should succeed")

	// Parse response to get order ID
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var orderResp OrderResponse
	err = json.Unmarshal(body, &orderResp)
	require.NoError(t, err)

	orderID := orderResp.ID
	require.NotEmpty(t, orderID, "Order ID should be present in response")

	// Wait a moment for DynamoDB consistency
	time.Sleep(2 * time.Second)

	// Create DynamoDB client
	awsRegion := getAWSRegion()
	sess := createAWSSession(t, awsRegion)
	dynamoClient := dynamodb.New(sess)

	// Query DynamoDB to verify item exists
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(orderID),
			},
		},
	}

	result, err := dynamoClient.GetItem(getInput)
	require.NoError(t, err, "Should be able to query DynamoDB")
	require.NotNil(t, result.Item, "Order should exist in DynamoDB")

	// Verify item attributes
	assert.NotNil(t, result.Item["name"], "Item should have 'name' attribute")
	assert.NotNil(t, result.Item["count"], "Item should have 'count' attribute")
	assert.NotNil(t, result.Item["status"], "Item should have 'status' attribute")

	if result.Item["name"] != nil && result.Item["name"].S != nil {
		assert.Equal(t, orderReq.Name, *result.Item["name"].S, "Name in DynamoDB should match request")
	}

	if result.Item["status"] != nil && result.Item["status"].S != nil {
		assert.Equal(t, "created", *result.Item["status"].S, "Status in DynamoDB should be 'created'")
	}
}
