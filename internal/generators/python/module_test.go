package python_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lewis/forge/internal/generators/python"
)

// TestGenerateLambdaModule tests the pure Lambda module generation.
func TestGenerateLambdaModule(t *testing.T) {
	t.Run("generates module with basic configuration", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
			UseDynamoDB:   false,
		}

		module := python.GenerateLambdaModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.FunctionName)
		assert.NotNil(t, module.Handler)
		assert.NotNil(t, module.Runtime)
		assert.Equal(t, "python3.13", *module.Runtime)
		assert.Equal(t, "service.handlers.handle_request.lambda_handler", *module.Handler)
		assert.NotNil(t, module.MemorySize)
		assert.Equal(t, 512, *module.MemorySize)
		assert.NotNil(t, module.Timeout)
		assert.Equal(t, 30, *module.Timeout)
	})

	t.Run("includes DynamoDB environment variables when enabled", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
			UseDynamoDB:   true,
		}

		module := python.GenerateLambdaModule(config)

		require.NotNil(t, module)
		assert.Contains(t, module.EnvironmentVariables, "TABLE_NAME")
		assert.Equal(t, "${module.dynamodb_table.dynamodb_table_id}", module.EnvironmentVariables["TABLE_NAME"])
	})

	t.Run("includes idempotency table when enabled", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:    "test-service",
			FunctionName:   "handler",
			PythonVersion:  "3.13",
			UseDynamoDB:    true,
			UseIdempotency: true,
		}

		module := python.GenerateLambdaModule(config)

		require.NotNil(t, module)
		assert.Contains(t, module.EnvironmentVariables, "IDEMPOTENCY_TABLE_NAME")
		assert.Contains(t, module.EnvironmentVariables, "TABLE_NAME")
	})

	t.Run("configures IAM policy for DynamoDB when enabled", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
			UseDynamoDB:   true,
		}

		module := python.GenerateLambdaModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.AttachPolicyJSON)
		assert.True(t, *module.AttachPolicyJSON)
		assert.NotNil(t, module.PolicyJSON)
		assert.Contains(t, *module.PolicyJSON, "dynamodb:GetItem")
		assert.Contains(t, *module.PolicyJSON, "dynamodb:PutItem")
		assert.Contains(t, *module.PolicyJSON, "dynamodb:UpdateItem")
		assert.Contains(t, *module.PolicyJSON, "dynamodb:DeleteItem")
	})

	t.Run("enables X-Ray tracing", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
		}

		module := python.GenerateLambdaModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.TracingMode)
		assert.Equal(t, "Active", *module.TracingMode)
	})

	t.Run("configures CloudWatch Logs retention", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
		}

		module := python.GenerateLambdaModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.CloudwatchLogsRetentionInDays)
		assert.Equal(t, 7, *module.CloudwatchLogsRetentionInDays)
	})

	t.Run("uses local existing package", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
		}

		module := python.GenerateLambdaModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.LocalExistingPackage)
		assert.Equal(t, "${path.module}/../.build/lambda.zip", *module.LocalExistingPackage)
		assert.NotNil(t, module.CreatePackage)
		assert.False(t, *module.CreatePackage)
	})

	t.Run("enables module-managed IAM role", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
		}

		module := python.GenerateLambdaModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.CreateRole)
		assert.True(t, *module.CreateRole)
	})
}

// TestGenerateLambdaModuleHCL tests HCL string generation.
func TestGenerateLambdaModuleHCL(t *testing.T) {
	t.Run("generates valid HCL", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
		}

		module := python.GenerateLambdaModule(config)
		hcl := python.GenerateLambdaModuleHCL(module)

		assert.Contains(t, hcl, "module \"lambda_function\"")
		assert.Contains(t, hcl, "source")
		assert.Contains(t, hcl, "version")
		assert.Contains(t, hcl, "function_name")
		assert.Contains(t, hcl, "handler")
		assert.Contains(t, hcl, "runtime")
		assert.Contains(t, hcl, "python3.13")
		assert.Contains(t, hcl, "memory_size")
		assert.Contains(t, hcl, "timeout")
		assert.Contains(t, hcl, "create_package = false")
		assert.Contains(t, hcl, "local_existing_package")
	})

	t.Run("includes environment variables", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
			UseDynamoDB:   true,
		}

		module := python.GenerateLambdaModule(config)
		hcl := python.GenerateLambdaModuleHCL(module)

		assert.Contains(t, hcl, "environment_variables")
		assert.Contains(t, hcl, "POWERTOOLS_SERVICE_NAME")
		assert.Contains(t, hcl, "LOG_LEVEL")
		assert.Contains(t, hcl, "TABLE_NAME")
	})

	t.Run("includes tracing configuration", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
		}

		module := python.GenerateLambdaModule(config)
		hcl := python.GenerateLambdaModuleHCL(module)

		assert.Contains(t, hcl, "tracing_mode = \"Active\"")
	})

	t.Run("includes CloudWatch Logs retention", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
		}

		module := python.GenerateLambdaModule(config)
		hcl := python.GenerateLambdaModuleHCL(module)

		assert.Contains(t, hcl, "cloudwatch_logs_retention_in_days = 7")
	})

	t.Run("includes IAM policy when DynamoDB enabled", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
			UseDynamoDB:   true,
		}

		module := python.GenerateLambdaModule(config)
		hcl := python.GenerateLambdaModuleHCL(module)

		assert.Contains(t, hcl, "attach_policy_json = true")
		assert.Contains(t, hcl, "policy_json = <<-EOT")
		assert.Contains(t, hcl, "dynamodb:GetItem")
	})

	t.Run("includes tags", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName:   "test-service",
			FunctionName:  "handler",
			PythonVersion: "3.13",
		}

		module := python.GenerateLambdaModule(config)
		hcl := python.GenerateLambdaModuleHCL(module)

		assert.Contains(t, hcl, "tags = {")
		assert.Contains(t, hcl, "ManagedBy   = \"Terraform\"")
		assert.Contains(t, hcl, "Generator   = \"Forge\"")
		assert.Contains(t, hcl, "Service     = var.service_name")
		assert.Contains(t, hcl, "Environment = var.environment")
	})
}

// TestGenerateAPIGatewayModule tests API Gateway module generation.
func TestGenerateAPIGatewayModule(t *testing.T) {
	t.Run("generates HTTP API module", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
			Description: "Test API",
			HTTPMethod:  "POST",
			APIPath:     "/api/test",
		}

		module := python.GenerateAPIGatewayModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.ProtocolType)
		assert.Equal(t, "HTTP", *module.ProtocolType)
		assert.NotNil(t, module.Description)
		assert.Equal(t, "Test API", *module.Description)
	})

	t.Run("configures CORS", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
			Description: "Test API",
			HTTPMethod:  "POST",
			APIPath:     "/api/test",
		}

		module := python.GenerateAPIGatewayModule(config)

		require.NotNil(t, module)
		require.NotNil(t, module.CORSConfiguration)
		assert.Contains(t, module.CORSConfiguration.AllowOrigins, "*")
		assert.Contains(t, module.CORSConfiguration.AllowMethods, "GET")
		assert.Contains(t, module.CORSConfiguration.AllowMethods, "POST")
		assert.Contains(t, module.CORSConfiguration.AllowHeaders, "content-type")
	})

	t.Run("configures default stage with auto-deploy", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
			Description: "Test API",
			HTTPMethod:  "POST",
			APIPath:     "/api/test",
		}

		module := python.GenerateAPIGatewayModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.CreateStage)
		assert.True(t, *module.CreateStage)
		assert.NotNil(t, module.StageName)
		assert.Equal(t, "$default", *module.StageName)
		assert.NotNil(t, module.AutoDeploy)
		assert.True(t, *module.AutoDeploy)
	})

	t.Run("configures Lambda integration", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
			Description: "Test API",
			HTTPMethod:  "POST",
			APIPath:     "/api/test",
		}

		module := python.GenerateAPIGatewayModule(config)

		require.NotNil(t, module)
		require.NotEmpty(t, module.Integrations)
		integration, exists := module.Integrations["lambda"]
		assert.True(t, exists)
		assert.Equal(t, "AWS_PROXY", integration.IntegrationType)
		assert.NotNil(t, integration.IntegrationMethod)
		assert.Equal(t, "POST", *integration.IntegrationMethod)
	})

	t.Run("configures route with correct HTTP method", func(t *testing.T) {
		tests := []struct {
			method      string
			expectedKey string
		}{
			{"GET", "GET /api/test"},
			{"POST", "POST /api/test"},
			{"PUT", "PUT /api/test"},
			{"DELETE", "DELETE /api/test"},
		}

		for _, tt := range tests {
			t.Run(tt.method, func(t *testing.T) {
				config := python.ProjectConfig{
					ServiceName: "test-service",
					Description: "Test API",
					HTTPMethod:  tt.method,
					APIPath:     "/api/test",
				}

				module := python.GenerateAPIGatewayModule(config)

				require.NotNil(t, module)
				require.NotEmpty(t, module.Routes)
				route, exists := module.Routes["main"]
				assert.True(t, exists)
				assert.Equal(t, tt.expectedKey, route.RouteKey)
			})
		}
	})
}

// TestGenerateAPIGatewayModuleHCL tests API Gateway HCL generation.
func TestGenerateAPIGatewayModuleHCL(t *testing.T) {
	t.Run("generates valid HCL", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
			Description: "Test API",
			HTTPMethod:  "POST",
			APIPath:     "/api/test",
		}

		module := python.GenerateAPIGatewayModule(config)
		hcl := python.GenerateAPIGatewayModuleHCL(module, "lambda_function")

		assert.Contains(t, hcl, "module \"api_gateway\"")
		assert.Contains(t, hcl, "protocol_type = \"HTTP\"")
		assert.Contains(t, hcl, "cors_configuration")
		assert.Contains(t, hcl, "routes")
		assert.Contains(t, hcl, "POST /api/test")
	})

	t.Run("includes CloudWatch Log Group", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
			Description: "Test API",
			HTTPMethod:  "POST",
			APIPath:     "/api/test",
		}

		module := python.GenerateAPIGatewayModule(config)
		hcl := python.GenerateAPIGatewayModuleHCL(module, "lambda_function")

		assert.Contains(t, hcl, "resource \"aws_cloudwatch_log_group\" \"api_gateway\"")
		assert.Contains(t, hcl, "retention_in_days = 7")
	})

	t.Run("includes Lambda permission", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
			Description: "Test API",
			HTTPMethod:  "POST",
			APIPath:     "/api/test",
		}

		module := python.GenerateAPIGatewayModule(config)
		hcl := python.GenerateAPIGatewayModuleHCL(module, "lambda_function")

		assert.Contains(t, hcl, "resource \"aws_lambda_permission\" \"api_gateway\"")
		assert.Contains(t, hcl, "action        = \"lambda:InvokeFunction\"")
		assert.Contains(t, hcl, "principal     = \"apigateway.amazonaws.com\"")
	})

	t.Run("references Lambda module correctly", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
			Description: "Test API",
			HTTPMethod:  "POST",
			APIPath:     "/api/test",
		}

		module := python.GenerateAPIGatewayModule(config)
		hcl := python.GenerateAPIGatewayModuleHCL(module, "my_lambda")

		assert.Contains(t, hcl, "module.my_lambda.lambda_function_invoke_arn")
		assert.Contains(t, hcl, "module.my_lambda.lambda_function_name")
	})
}

// TestGenerateDynamoDBModule tests DynamoDB module generation.
func TestGenerateDynamoDBModule(t *testing.T) {
	t.Run("generates module with default table name", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.Name)
		assert.Contains(t, *module.Name, "test-service")
	})

	t.Run("uses custom table name", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
			TableName:   "custom-table",
		}

		module := python.GenerateDynamoDBModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.Name)
		assert.Equal(t, "custom-table", *module.Name)
	})

	t.Run("configures PAY_PER_REQUEST billing", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.BillingMode)
		assert.Equal(t, "PAY_PER_REQUEST", *module.BillingMode)
	})

	t.Run("configures primary key", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.HashKey)
		assert.Equal(t, "id", *module.HashKey)
		require.Len(t, module.Attributes, 1)
		assert.Equal(t, "id", module.Attributes[0].Name)
		assert.Equal(t, "S", module.Attributes[0].Type)
	})

	t.Run("enables point-in-time recovery", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.PointInTimeRecoveryEnabled)
		assert.True(t, *module.PointInTimeRecoveryEnabled)
	})

	t.Run("enables server-side encryption", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.ServerSideEncryptionEnabled)
		assert.True(t, *module.ServerSideEncryptionEnabled)
	})

	t.Run("configures TTL", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)

		require.NotNil(t, module)
		assert.NotNil(t, module.TTLEnabled)
		assert.False(t, *module.TTLEnabled)
		assert.NotNil(t, module.TTLAttributeName)
		assert.Equal(t, "ttl", *module.TTLAttributeName)
	})
}

// TestGenerateDynamoDBModuleHCL tests DynamoDB HCL generation.
func TestGenerateDynamoDBModuleHCL(t *testing.T) {
	t.Run("generates valid HCL", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)
		hcl := python.GenerateDynamoDBModuleHCL(module, "lambda_function")

		assert.Contains(t, hcl, "module \"dynamodb_table\"")
		assert.Contains(t, hcl, "billing_mode = \"PAY_PER_REQUEST\"")
		assert.Contains(t, hcl, "hash_key     = \"id\"")
		assert.Contains(t, hcl, "attributes")
	})

	t.Run("includes point-in-time recovery", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)
		hcl := python.GenerateDynamoDBModuleHCL(module, "lambda_function")

		assert.Contains(t, hcl, "point_in_time_recovery_enabled = true")
	})

	t.Run("includes server-side encryption", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)
		hcl := python.GenerateDynamoDBModuleHCL(module, "lambda_function")

		assert.Contains(t, hcl, "server_side_encryption_enabled = true")
	})

	t.Run("includes IAM policy for Lambda", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)
		hcl := python.GenerateDynamoDBModuleHCL(module, "lambda_function")

		assert.Contains(t, hcl, "resource \"aws_iam_policy\" \"dynamodb_access\"")
		assert.Contains(t, hcl, "dynamodb:GetItem")
		assert.Contains(t, hcl, "dynamodb:PutItem")
		assert.Contains(t, hcl, "dynamodb:UpdateItem")
		assert.Contains(t, hcl, "dynamodb:DeleteItem")
		assert.Contains(t, hcl, "dynamodb:Query")
		assert.Contains(t, hcl, "dynamodb:Scan")
	})

	t.Run("includes IAM role policy attachment", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)
		hcl := python.GenerateDynamoDBModuleHCL(module, "my_lambda")

		assert.Contains(t, hcl, "resource \"aws_iam_role_policy_attachment\" \"lambda_dynamodb\"")
		assert.Contains(t, hcl, "role       = module.my_lambda.lambda_role_name")
		assert.Contains(t, hcl, "policy_arn = aws_iam_policy.dynamodb_access.arn")
	})

	t.Run("includes tags", func(t *testing.T) {
		config := python.ProjectConfig{
			ServiceName: "test-service",
		}

		module := python.GenerateDynamoDBModule(config)
		hcl := python.GenerateDynamoDBModuleHCL(module, "lambda_function")

		// Count tag occurrences
		tagCount := strings.Count(hcl, "tags = {")
		assert.Positive(t, tagCount, "Should have at least one tags block")
		assert.Contains(t, hcl, "ManagedBy   = \"Terraform\"")
		assert.Contains(t, hcl, "Generator   = \"Forge\"")
	})
}
