package lingon

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewForgeStack tests Lingon stack creation from configuration
func TestNewForgeStack(t *testing.T) {
	t.Run("creates stack from minimal config", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"hello": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
				},
			},
		}

		stack, err := NewForgeStack(config)

		require.NoError(t, err)
		assert.Equal(t, "test-service", stack.Name)
		assert.Len(t, stack.Functions, 1)
		assert.Contains(t, stack.Functions, "hello")
	})

	t.Run("creates Lambda function resources", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src/api",
					},
					Environment: map[string]string{
						"TABLE_NAME": "users",
					},
				},
			},
		}

		stack, err := NewForgeStack(config)

		require.NoError(t, err)
		assert.NotNil(t, stack.Functions["api"])
		assert.NotNil(t, stack.Functions["api"].Function)
		assert.NotNil(t, stack.Functions["api"].Role)
	})

	t.Run("creates log group when logging configured", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
					Logs: CloudWatchLogsConfig{
						RetentionInDays: 7,
					},
				},
			},
		}

		stack, err := NewForgeStack(config)

		require.NoError(t, err)
		assert.NotNil(t, stack.Functions["api"].LogGroup)
	})

	t.Run("creates API Gateway resources", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
					HTTPRouting: &HTTPRoutingConfig{
						Method: "GET",
						Path:   "/hello",
					},
				},
			},
			APIGateway: &APIGatewayConfig{
				Name:         "test-api",
				ProtocolType: "HTTP",
			},
		}

		stack, err := NewForgeStack(config)

		require.NoError(t, err)
		assert.NotNil(t, stack.APIGateway)
		assert.NotNil(t, stack.APIGateway.API)
		assert.NotNil(t, stack.APIGateway.Stage)
		assert.Len(t, stack.APIGateway.Integrations, 1)
		assert.Len(t, stack.APIGateway.Routes, 1)
	})

	t.Run("creates DynamoDB table resources", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
				},
			},
			Tables: map[string]TableConfig{
				"users": {
					TableName:   "users",
					BillingMode: "PAY_PER_REQUEST",
					HashKey:     "userId",
					Attributes: []AttributeDefinition{
						{Name: "userId", Type: "S"},
					},
				},
			},
		}

		stack, err := NewForgeStack(config)

		require.NoError(t, err)
		assert.Len(t, stack.Tables, 1)
		assert.Contains(t, stack.Tables, "users")
		assert.NotNil(t, stack.Tables["users"].Table)
	})

	t.Run("fails with invalid config", func(t *testing.T) {
		config := ForgeConfig{
			Service: "", // Missing service name
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
				},
			},
		}

		_, err := NewForgeStack(config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service name is required")
	})

	t.Run("fails with invalid function config", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "", // Missing handler
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
				},
			},
		}

		_, err := NewForgeStack(config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler is required")
	})
}

// TestExportTerraform tests Terraform HCL generation
func TestExportTerraform(t *testing.T) {
	t.Run("exports minimal stack to HCL", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"hello": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
				},
			},
		}

		stack, err := NewForgeStack(config)
		require.NoError(t, err)

		hcl, err := stack.ExportTerraform()

		require.NoError(t, err)
		assert.NotNil(t, hcl)

		hclString := string(hcl)
		assert.Contains(t, hclString, "terraform {")
		assert.Contains(t, hclString, "provider \"aws\"")
		assert.Contains(t, hclString, "region = \"us-east-1\"")
		assert.Contains(t, hclString, "resource \"aws_lambda_function\"")
		assert.Contains(t, hclString, "resource \"aws_iam_role\"")
	})

	t.Run("exports Lambda function with environment variables", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
					Environment: map[string]string{
						"TABLE_NAME": "users",
						"STAGE":      "prod",
					},
				},
			},
		}

		stack, err := NewForgeStack(config)
		require.NoError(t, err)

		hcl, err := stack.ExportTerraform()

		require.NoError(t, err)
		hclString := string(hcl)
		assert.Contains(t, hclString, "environment {")
		assert.Contains(t, hclString, "TABLE_NAME")
		assert.Contains(t, hclString, "STAGE")
	})

	t.Run("exports Lambda function with log group", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
					Logs: CloudWatchLogsConfig{
						RetentionInDays: 7,
					},
				},
			},
		}

		stack, err := NewForgeStack(config)
		require.NoError(t, err)

		hcl, err := stack.ExportTerraform()

		require.NoError(t, err)
		hclString := string(hcl)
		assert.Contains(t, hclString, "resource \"aws_cloudwatch_log_group\"")
		assert.Contains(t, hclString, "retention_in_days = 7")
	})

	t.Run("exports API Gateway resources", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
					HTTPRouting: &HTTPRoutingConfig{
						Method: "GET",
						Path:   "/hello",
					},
				},
			},
			APIGateway: &APIGatewayConfig{
				Name:         "test-api",
				ProtocolType: "HTTP",
			},
		}

		stack, err := NewForgeStack(config)
		require.NoError(t, err)

		hcl, err := stack.ExportTerraform()

		require.NoError(t, err)
		hclString := string(hcl)
		assert.Contains(t, hclString, "resource \"aws_apigatewayv2_api\"")
		assert.Contains(t, hclString, "resource \"aws_apigatewayv2_stage\"")
		assert.Contains(t, hclString, "protocol_type = \"HTTP\"")
	})

	t.Run("exports DynamoDB table resources", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
				},
			},
			Tables: map[string]TableConfig{
				"users": {
					TableName:   "users",
					BillingMode: "PAY_PER_REQUEST",
					HashKey:     "userId",
					Attributes: []AttributeDefinition{
						{Name: "userId", Type: "S"},
						{Name: "email", Type: "S"},
					},
				},
			},
		}

		stack, err := NewForgeStack(config)
		require.NoError(t, err)

		hcl, err := stack.ExportTerraform()

		require.NoError(t, err)
		hclString := string(hcl)
		assert.Contains(t, hclString, "resource \"aws_dynamodb_table\"")
		assert.Contains(t, hclString, "billing_mode   = \"PAY_PER_REQUEST\"")
		assert.Contains(t, hclString, "hash_key       = \"userId\"")
		assert.Contains(t, hclString, "attribute {")
	})

	t.Run("exports outputs", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
				},
			},
			APIGateway: &APIGatewayConfig{
				Name:         "test-api",
				ProtocolType: "HTTP",
			},
		}

		stack, err := NewForgeStack(config)
		require.NoError(t, err)

		hcl, err := stack.ExportTerraform()

		require.NoError(t, err)
		hclString := string(hcl)
		assert.Contains(t, hclString, "output \"api_function_arn\"")
		assert.Contains(t, hclString, "output \"api_endpoint\"")
	})

	t.Run("exports complete stack", func(t *testing.T) {
		config := ForgeConfig{
			Service: "complete-app",
			Provider: ProviderConfig{
				Region: "us-west-2",
			},
			Functions: map[string]FunctionConfig{
				"api": {
					Handler: "index.handler",
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src/api",
					},
					HTTPRouting: &HTTPRoutingConfig{
						Method: "ANY",
						Path:   "/{proxy+}",
					},
					Environment: map[string]string{
						"TABLE_NAME": "users",
					},
					Logs: CloudWatchLogsConfig{
						RetentionInDays: 14,
					},
				},
				"processor": {
					Handler: "processor.handler",
					Runtime: "python3.12",
					Source: SourceConfig{
						Path: "./src/processor",
					},
					Logs: CloudWatchLogsConfig{
						RetentionInDays: 7,
					},
				},
			},
			APIGateway: &APIGatewayConfig{
				Name:         "complete-api",
				ProtocolType: "HTTP",
			},
			Tables: map[string]TableConfig{
				"users": {
					TableName:   "users",
					BillingMode: "PAY_PER_REQUEST",
					HashKey:     "userId",
					Attributes: []AttributeDefinition{
						{Name: "userId", Type: "S"},
					},
				},
			},
		}

		stack, err := NewForgeStack(config)
		require.NoError(t, err)

		hcl, err := stack.ExportTerraform()

		require.NoError(t, err)
		hclString := string(hcl)

		// Verify all components are present
		assert.Contains(t, hclString, "complete-app")
		assert.Contains(t, hclString, "us-west-2")

		// Count resources
		assert.Equal(t, 2, strings.Count(hclString, "resource \"aws_lambda_function\""))
		assert.Equal(t, 2, strings.Count(hclString, "resource \"aws_iam_role\""))
		assert.Equal(t, 2, strings.Count(hclString, "resource \"aws_cloudwatch_log_group\""))
		assert.Equal(t, 1, strings.Count(hclString, "resource \"aws_apigatewayv2_api\""))
		assert.Equal(t, 1, strings.Count(hclString, "resource \"aws_dynamodb_table\""))

		// Verify outputs
		assert.Contains(t, hclString, "output \"api_function_arn\"")
		assert.Contains(t, hclString, "output \"processor_function_arn\"")
		assert.Contains(t, hclString, "output \"api_endpoint\"")
	})
}

// TestLambdaFunctionResourceCreation tests Lambda resource creation
func TestLambdaFunctionResourceCreation(t *testing.T) {
	t.Run("creates basic Lambda resources", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function)
		assert.NotNil(t, resources.Role)
	})

	t.Run("creates log group when configured", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Logs: CloudWatchLogsConfig{
				RetentionInDays: 7,
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.LogGroup)
	})

	t.Run("fails with invalid config", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "", // Missing handler
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
		}

		_, err := createLambdaFunctionResources("test-service", "hello", config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler is required")
	})
}

// TestAPIGatewayResourceCreation tests API Gateway resource creation
func TestAPIGatewayResourceCreation(t *testing.T) {
	t.Run("creates API Gateway resources", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
		}

		functions := map[string]FunctionConfig{
			"hello": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				HTTPRouting: &HTTPRoutingConfig{
					Method: "GET",
					Path:   "/hello",
				},
			},
		}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.API)
		assert.NotNil(t, resources.Stage)
		assert.Len(t, resources.Integrations, 1)
		assert.Len(t, resources.Routes, 1)
	})

	t.Run("creates resources for multiple functions", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
		}

		functions := map[string]FunctionConfig{
			"hello": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				HTTPRouting: &HTTPRoutingConfig{
					Method: "GET",
					Path:   "/hello",
				},
			},
			"users": {
				Handler: "users.handler",
				Runtime: "nodejs20.x",
				HTTPRouting: &HTTPRoutingConfig{
					Method: "POST",
					Path:   "/users",
				},
			},
		}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.Len(t, resources.Integrations, 2)
		assert.Len(t, resources.Routes, 2)
	})
}

// TestDynamoDBTableResourceCreation tests DynamoDB resource creation
func TestDynamoDBTableResourceCreation(t *testing.T) {
	t.Run("creates DynamoDB table resources", func(t *testing.T) {
		config := TableConfig{
			TableName:   "users",
			BillingMode: "PAY_PER_REQUEST",
			HashKey:     "userId",
			Attributes: []AttributeDefinition{
				{Name: "userId", Type: "S"},
			},
		}

		resources, err := createDynamoDBTableResources("test-service", "users", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Table)
	})

	t.Run("uses service prefix when table name is empty", func(t *testing.T) {
		config := TableConfig{
			TableName:   "",
			BillingMode: "PAY_PER_REQUEST",
			HashKey:     "userId",
			Attributes: []AttributeDefinition{
				{Name: "userId", Type: "S"},
			},
		}

		resources, err := createDynamoDBTableResources("test-service", "users", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Table)
		assert.Contains(t, resources.Table.LocalName(), "test-service-users")
	})
}
