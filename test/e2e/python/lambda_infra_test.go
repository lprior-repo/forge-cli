//go:build e2e
// +build e2e.

package python

import (
	"os"
	"path/filepath"
	"testing"

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
	apiEndpoint := terraform.Output(t, terraformOptions, "api_endpoint")

	require.NotEmpty(t, functionName, "lambda_function_name output should not be empty")
	require.NotEmpty(t, tableName, "dynamodb_table_name output should not be empty")
	require.NotEmpty(t, apiEndpoint, "api_endpoint output should not be empty")

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
		testAPIEndpoint(t, apiEndpoint)
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

func testAPIEndpoint(t *testing.T, apiEndpoint string) {
	// Basic validation - actual HTTP testing could be added here
	assert.NotEmpty(t, apiEndpoint)
	assert.Contains(t, apiEndpoint, "execute-api")
	assert.Contains(t, apiEndpoint, "/api/orders")

	// TODO: Add actual HTTP request test
	// This would require the Lambda to be fully deployed and functional
}
