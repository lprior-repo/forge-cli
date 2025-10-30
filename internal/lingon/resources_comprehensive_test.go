package lingon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLambdaFunctionResources(t *testing.T) {
	t.Run("creates basic Lambda function resources", func(t *testing.T) {
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
		assert.Equal(t, "test-service-hello", resources.Function.Name)
		assert.Equal(t, "test-service-hello-role", resources.Role.Name)
	})

	t.Run("creates function with environment variables", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Environment: map[string]string{
				"NODE_ENV":  "production",
				"LOG_LEVEL": "debug",
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "api", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.Environment)
	})

	t.Run("creates function with S3 source", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				S3Bucket: "my-bucket",
				S3Key:    "lambda.zip",
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "s3-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function)
	})

	t.Run("creates function with VPC config", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			VPC: &VPCConfig{
				SubnetIds:        []string{"subnet-123", "subnet-456"},
				SecurityGroupIds: []string{"sg-123"},
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "vpc-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.VpcConfig)
	})

	t.Run("creates function with layers", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Layers: []string{
				"arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1",
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "layered-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.Layers)
	})

	t.Run("creates function with dead letter config", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			DeadLetterConfig: &DeadLetterConfig{
				TargetArn: "arn:aws:sqs:us-east-1:123456789012:queue/my-dlq",
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "dlq-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.DeadLetterConfig)
	})

	t.Run("creates function with tracing", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			TracingMode: "Active",
		}

		resources, err := createLambdaFunctionResources("test-service", "traced-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.TracingConfig)
	})

	t.Run("creates function with reserved concurrent executions", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			ReservedConcurrentExecutions: 10,
		}

		resources, err := createLambdaFunctionResources("test-service", "concurrent-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.ReservedConcurrentExecutions)
	})

	t.Run("creates function with architectures", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Architectures: []string{"arm64"},
		}

		resources, err := createLambdaFunctionResources("test-service", "arm-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.Architectures)
	})

	t.Run("creates function with ephemeral storage", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			EphemeralStorage: &EphemeralStorageConfig{
				Size: 1024,
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "storage-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.EphemeralStorage)
	})

	t.Run("creates function with KMS key", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			KMSKeyArn: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
		}

		resources, err := createLambdaFunctionResources("test-service", "kms-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.KmsKeyArn)
	})

	t.Run("creates function with code signing config", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			CodeSigningConfigArn: "arn:aws:lambda:us-east-1:123456789012:code-signing-config:csc-1234",
		}

		resources, err := createLambdaFunctionResources("test-service", "signed-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.CodeSigningConfigArn)
	})

	t.Run("creates function with publish flag", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Publish: true,
		}

		resources, err := createLambdaFunctionResources("test-service", "published-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.Publish)
	})

	t.Run("creates function with tags", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Tags: map[string]string{
				"Environment": "production",
				"Team":        "backend",
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "tagged-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.Tags)
	})

	t.Run("creates log group when configured", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Logs: CloudWatchLogsConfig{
				RetentionInDays: 14,
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "logged-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.LogGroup)
		assert.NotNil(t, resources.LogGroup.Args.Name)
	})

	t.Run("creates log group with custom name", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Logs: CloudWatchLogsConfig{
				LogGroupName:    "/custom/log/group",
				RetentionInDays: 7,
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "custom-log-fn", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.LogGroup)
		assert.NotNil(t, resources.LogGroup.Args.Name)
	})

	t.Run("creates event source mappings", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			EventSourceMappings: []EventSourceMappingConfig{
				{
					EventSourceArn:   "arn:aws:dynamodb:us-east-1:123456789012:table/test/stream/2021-01-01T00:00:00.000",
					StartingPosition: "LATEST",
					BatchSize:        100,
				},
				{
					EventSourceArn:   "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
					StartingPosition: "TRIM_HORIZON",
					BatchSize:        500,
				},
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "esm-fn", config)

		require.NoError(t, err)
		assert.Len(t, resources.EventSourceMappings, 2)
	})

	t.Run("creates aliases", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Aliases: []AliasConfig{
				{
					Name:            "production",
					FunctionVersion: "1",
					Description:     "Production alias",
				},
				{
					Name:            "staging",
					FunctionVersion: "2",
					Description:     "Staging alias",
				},
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "alias-fn", config)

		require.NoError(t, err)
		assert.Len(t, resources.Aliases, 2)
	})

	t.Run("fails with invalid config", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "", // Missing handler
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
		}

		_, err := createLambdaFunctionResources("test-service", "invalid-fn", config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid function config")
	})
}

func TestCreateAPIGatewayResources(t *testing.T) {
	t.Run("creates basic API Gateway", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
		}
		functions := map[string]FunctionConfig{}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.API)
		assert.NotNil(t, resources.Stage)
		assert.Equal(t, "api", resources.API.Name)
	})

	t.Run("uses service name when API name is empty", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "",
			ProtocolType: "HTTP",
		}
		functions := map[string]FunctionConfig{}

		resources, err := createAPIGatewayResources("my-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.API.Args.Name)
	})

	t.Run("creates API with description", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			Description:  "Test API Gateway",
		}
		functions := map[string]FunctionConfig{}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.API.Args.Description)
	})

	t.Run("creates API with CORS configuration", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			CORS: &CORSConfig{
				AllowOrigins:     []string{"*"},
				AllowMethods:     []string{"GET", "POST"},
				AllowHeaders:     []string{"Content-Type"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
		}
		functions := map[string]FunctionConfig{}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.API.Args.CorsConfiguration)
	})

	t.Run("creates API with disabled execute endpoint", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:                      "test-api",
			ProtocolType:              "HTTP",
			DisableExecuteApiEndpoint: true,
		}
		functions := map[string]FunctionConfig{}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.API.Args.DisableExecuteApiEndpoint)
	})

	t.Run("creates API with tags", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			Tags: map[string]string{
				"Environment": "production",
			},
		}
		functions := map[string]FunctionConfig{}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.API.Args.Tags)
	})

	t.Run("creates stage with access logs", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			AccessLogs: &AccessLogsConfig{
				DestinationArn: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/apigateway/test",
				Format:         "$context.requestId",
			},
		}
		functions := map[string]FunctionConfig{}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.Stage.Args.AccessLogSettings)
	})

	t.Run("creates stage with default route settings", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			DefaultRouteSettings: &RouteSettings{
				ThrottlingBurstLimit:   1000,
				ThrottlingRateLimit:    500,
				DetailedMetricsEnabled: true,
				LoggingLevel:           "INFO",
				DataTraceEnabled:       true,
			},
		}
		functions := map[string]FunctionConfig{}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.Stage.Args.DefaultRouteSettings)
	})

	t.Run("creates authorizers", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			Authorizers: map[string]AuthorizerConfig{
				"jwt-auth": {
					Type:           "JWT",
					IdentitySource: []string{"$request.header.Authorization"},
					JWTConfiguration: &JWTConfiguration{
						Audience: []string{"https://example.com"},
						Issuer:   "https://issuer.example.com",
					},
				},
				"lambda-auth": {
					Type:           "REQUEST",
					AuthorizerURI:  "arn:aws:lambda:us-east-1:123456789012:function:authorizer",
					IdentitySource: []string{"$request.header.Authorization"},
				},
			},
		}
		functions := map[string]FunctionConfig{}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.Len(t, resources.Authorizers, 2)
		assert.Contains(t, resources.Authorizers, "jwt-auth")
		assert.Contains(t, resources.Authorizers, "lambda-auth")
	})

	t.Run("creates integrations and routes for HTTP functions", func(t *testing.T) {
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
				Handler: "index.handler",
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
		assert.Len(t, resources.Permissions, 2) // Lambda permissions
	})

	t.Run("creates routes with authorizers", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			Authorizers: map[string]AuthorizerConfig{
				"jwt-auth": {
					Type:           "JWT",
					IdentitySource: []string{"$request.header.Authorization"},
				},
			},
		}
		functions := map[string]FunctionConfig{
			"protected": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				HTTPRouting: &HTTPRoutingConfig{
					Method:            "GET",
					Path:              "/protected",
					AuthorizerId:      "jwt-auth",
					AuthorizationType: "JWT",
				},
			},
		}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.Len(t, resources.Routes, 1)
	})

	t.Run("skips functions without HTTP routing", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
		}
		functions := map[string]FunctionConfig{
			"worker": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				// No HTTPRouting
			},
		}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.Empty(t, resources.Integrations)
		assert.Empty(t, resources.Routes)
	})
}

func TestCreateDynamoDBTableResources(t *testing.T) {
	t.Run("creates basic DynamoDB table", func(t *testing.T) {
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
		assert.Equal(t, "users", resources.Table.Name)
		assert.NotNil(t, resources.Table.Args.Name)
	})

	t.Run("uses service prefix when table name is empty", func(t *testing.T) {
		config := TableConfig{
			TableName:   "",
			BillingMode: "PAY_PER_REQUEST",
			HashKey:     "id",
			Attributes: []AttributeDefinition{
				{Name: "id", Type: "S"},
			},
		}

		resources, err := createDynamoDBTableResources("my-service", "items", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Table.Args.Name)
	})

	t.Run("creates table with multiple attributes", func(t *testing.T) {
		config := TableConfig{
			TableName:   "orders",
			BillingMode: "PROVISIONED",
			HashKey:     "orderId",
			Attributes: []AttributeDefinition{
				{Name: "orderId", Type: "S"},
				{Name: "customerId", Type: "S"},
				{Name: "orderDate", Type: "N"},
			},
		}

		resources, err := createDynamoDBTableResources("test-service", "orders", config)

		require.NoError(t, err)
		assert.Len(t, resources.Table.Args.Attribute, 3)
	})
}

func TestPlaceholderResource(t *testing.T) {
	t.Run("creates placeholder resource", func(t *testing.T) {
		resource := createPlaceholderResource("iam_role", "test-role")

		assert.NotNil(t, resource)
		assert.Equal(t, "iam_role", resource.resourceType)
		assert.Equal(t, "test-role", resource.resourceName)
	})

	t.Run("placeholder resource Type returns resource type", func(t *testing.T) {
		resource := createPlaceholderResource("lambda_function", "test-fn")

		assert.Equal(t, "lambda_function", resource.Type())
	})

	t.Run("placeholder resource LocalName returns resource name", func(t *testing.T) {
		resource := createPlaceholderResource("dynamodb_table", "test-table")

		assert.Equal(t, "test-table", resource.LocalName())
	})

	t.Run("placeholder resource Configuration returns attributes", func(t *testing.T) {
		resource := createPlaceholderResource("s3_bucket", "test-bucket")

		assert.NotNil(t, resource.Configuration())
	})

	t.Run("placeholder resource Dependencies returns empty dependencies", func(t *testing.T) {
		resource := createPlaceholderResource("sqs_queue", "test-queue")

		deps := resource.Dependencies()
		assert.NotNil(t, deps)
	})

	t.Run("placeholder resource LifecycleManagement returns nil", func(t *testing.T) {
		resource := createPlaceholderResource("sns_topic", "test-topic")

		assert.Nil(t, resource.LifecycleManagement())
	})

	t.Run("placeholder resource ImportState returns nil", func(t *testing.T) {
		resource := createPlaceholderResource("api_gateway", "test-api")

		err := resource.ImportState(nil)
		assert.NoError(t, err)
	})

	t.Run("placeholder resource Arn returns ARN reference", func(t *testing.T) {
		resource := createPlaceholderResource("iam_role", "test-role")

		arn := resource.Arn()
		assert.Contains(t, arn, "aws_iam_role")
		assert.Contains(t, arn, "test-role")
		assert.Contains(t, arn, "arn")
	})

	t.Run("placeholder resource Name returns name reference", func(t *testing.T) {
		resource := createPlaceholderResource("lambda_function", "test-fn")

		name := resource.Name()
		assert.Contains(t, name, "aws_lambda_function")
		assert.Contains(t, name, "test-fn")
		assert.Contains(t, name, "name")
	})

	t.Run("placeholder resource ID returns ID reference", func(t *testing.T) {
		resource := createPlaceholderResource("dynamodb_table", "test-table")

		id := resource.ID()
		assert.Contains(t, id, "aws_dynamodb_table")
		assert.Contains(t, id, "test-table")
		assert.Contains(t, id, "id")
	})
}
