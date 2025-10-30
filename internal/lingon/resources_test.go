package lingon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateLambdaFunctionResourcesAliases tests Lambda alias creation.
func TestCreateLambdaFunctionResourcesAliases(t *testing.T) {
	t.Run("creates Lambda aliases with version references", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Aliases: []AliasConfig{
				{
					Name:            "prod",
					Description:     "Production alias",
					FunctionVersion: "1",
				},
				{
					Name:            "dev",
					Description:     "Development alias",
					FunctionVersion: "$LATEST",
				},
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.Len(t, resources.Aliases, 2)
		assert.Equal(t, "test-service-hello-prod", resources.Aliases[0].Name)
		assert.Equal(t, "test-service-hello-dev", resources.Aliases[1].Name)
	})

	t.Run("creates Lambda alias with weighted routing", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Aliases: []AliasConfig{
				{
					Name:            "live",
					Description:     "Live traffic with blue/green",
					FunctionVersion: "2",
					RoutingConfig: &AliasRoutingConfig{
						AdditionalVersionWeights: map[string]float64{
							"1": 0.1, // 10% to version 1
						},
					},
				},
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.Len(t, resources.Aliases, 1)
		assert.NotNil(t, resources.Aliases[0].Args.RoutingConfig)
	})
}

// TestCreateLambdaFunctionResourcesEventSources tests event source mapping creation.
func TestCreateLambdaFunctionResourcesEventSources(t *testing.T) {
	t.Run("creates DynamoDB stream event source mapping", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			EventSourceMappings: []EventSourceMappingConfig{
				{
					EventSourceArn:             "arn:aws:dynamodb:us-east-1:123456789012:table/test/stream/2021-01-01T00:00:00.000",
					StartingPosition:           "LATEST",
					BatchSize:                  100,
					ParallelizationFactor:      2,
					MaximumRecordAgeInSeconds:  3600,
					MaximumRetryAttempts:       3,
					BisectBatchOnFunctionError: true,
				},
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.Len(t, resources.EventSourceMappings, 1)
		assert.Equal(t, "test-service-hello-esm-0", resources.EventSourceMappings[0].Name)
	})

	t.Run("creates SQS event source mapping with scaling", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			EventSourceMappings: []EventSourceMappingConfig{
				{
					EventSourceArn: "arn:aws:sqs:us-east-1:123456789012:my-queue",
					BatchSize:      10,
					ScalingConfig: &ScalingConfig{
						MaximumConcurrency: 100,
					},
				},
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.Len(t, resources.EventSourceMappings, 1)
	})

	t.Run("creates event source with filter criteria", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			EventSourceMappings: []EventSourceMappingConfig{
				{
					EventSourceArn:   "arn:aws:kinesis:us-east-1:123456789012:stream/test",
					StartingPosition: "TRIM_HORIZON",
					FilterCriteria: &FilterCriteria{
						Filters: []FilterPattern{
							{Pattern: `{"eventType": ["ORDER_PLACED"]}`},
						},
					},
				},
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.Len(t, resources.EventSourceMappings, 1)
	})

	t.Run("creates event source with destination config", func(t *testing.T) {
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
					DestinationConfig: &EventSourceDestinationConfig{
						OnFailure: &DestinationConfig{
							Destination: "arn:aws:sqs:us-east-1:123456789012:dlq",
						},
					},
				},
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.Len(t, resources.EventSourceMappings, 1)
	})
}

// TestCreateLambdaFunctionResourcesEnhanced tests enhanced Lambda configurations.
func TestCreateLambdaFunctionResourcesEnhanced(t *testing.T) {
	t.Run("creates Lambda with VPC configuration", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			VPC: &VPCConfig{
				SubnetIds:        []string{"subnet-123", "subnet-456"},
				SecurityGroupIds: []string{"sg-789"},
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.VpcConfig)
	})

	t.Run("creates Lambda with layers", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Layers: []string{
				"arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1",
				"arn:aws:lambda:us-east-1:123456789012:layer:another-layer:2",
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.Layers)
	})

	t.Run("creates Lambda with dead letter config", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			DeadLetterConfig: &DeadLetterConfig{
				TargetArn: "arn:aws:sqs:us-east-1:123456789012:dlq",
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.DeadLetterConfig)
	})

	t.Run("creates Lambda with X-Ray tracing", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			TracingMode: "Active",
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.TracingConfig)
	})

	t.Run("creates Lambda with reserved concurrency", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			ReservedConcurrentExecutions: 10,
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		// Reserved concurrency is set in args
		assert.NotNil(t, resources.Function)
	})

	t.Run("creates Lambda with architectures", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Architectures: []string{"arm64"},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.Architectures)
	})

	t.Run("creates Lambda with ephemeral storage", func(t *testing.T) {
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

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.EphemeralStorage)
	})

	t.Run("creates Lambda with KMS encryption", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			KMSKeyArn: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.KmsKeyArn)
	})

	t.Run("creates Lambda with code signing", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			CodeSigningConfigArn: "arn:aws:lambda:us-east-1:123456789012:code-signing-config:csc-123",
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.CodeSigningConfigArn)
	})

	t.Run("creates Lambda with publish flag", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Publish: true,
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.Publish)
	})

	t.Run("creates Lambda with tags", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			Tags: map[string]string{
				"Environment": "production",
				"Team":        "platform",
			},
		}

		resources, err := createLambdaFunctionResources("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, resources.Function.Args.Tags)
	})
}

// TestCreateAPIGatewayResourcesCORS tests CORS configuration.
func TestCreateAPIGatewayResourcesCORS(t *testing.T) {
	t.Run("creates API Gateway with CORS configuration", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			CORS: &CORSConfig{
				AllowOrigins:     []string{"https://example.com"},
				AllowMethods:     []string{"GET", "POST"},
				AllowHeaders:     []string{"Content-Type", "Authorization"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
		}

		functions := map[string]FunctionConfig{
			"hello": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				Source: SourceConfig{
					Path: "./src",
				},
			},
		}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.API.Args.CorsConfiguration)
	})
}

// TestCreateAPIGatewayResourcesAuthorizers tests authorizer configuration.
func TestCreateAPIGatewayResourcesAuthorizers(t *testing.T) {
	t.Run("creates API Gateway with JWT authorizer", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			Authorizers: map[string]AuthorizerConfig{
				"cognito": {
					Type:           "JWT",
					IdentitySource: []string{"$request.header.Authorization"},
					JWTConfiguration: &JWTConfiguration{
						Issuer:   "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_123456789",
						Audience: []string{"client-id-123"},
					},
				},
			},
		}

		functions := map[string]FunctionConfig{
			"hello": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				Source: SourceConfig{
					Path: "./src",
				},
				HTTPRouting: &HTTPRoutingConfig{
					Path:   "/hello",
					Method: "GET",
				},
			},
		}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.Len(t, resources.Authorizers, 1)
		assert.NotNil(t, resources.Authorizers["cognito"])
	})

	t.Run("creates API Gateway with Lambda REQUEST authorizer", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			Authorizers: map[string]AuthorizerConfig{
				"lambda-auth": {
					Type:                           "REQUEST",
					AuthorizerURI:                  "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:123456789012:function:authorizer/invocations",
					AuthorizerPayloadFormatVersion: "2.0",
					AuthorizerResultTtlInSeconds:   300,
					IdentitySource:                 []string{"$request.header.Authorization"},
				},
			},
		}

		functions := map[string]FunctionConfig{
			"hello": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				Source: SourceConfig{
					Path: "./src",
				},
				HTTPRouting: &HTTPRoutingConfig{
					Path:   "/hello",
					Method: "GET",
				},
			},
		}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.Len(t, resources.Authorizers, 1)
	})
}

// TestCreateAPIGatewayResourcesThrottling tests throttling configuration.
func TestCreateAPIGatewayResourcesThrottling(t *testing.T) {
	t.Run("creates API Gateway with default route throttling", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			DefaultRouteSettings: &RouteSettings{
				ThrottlingBurstLimit: 500,
				ThrottlingRateLimit:  1000,
			},
		}

		functions := map[string]FunctionConfig{
			"hello": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				Source: SourceConfig{
					Path: "./src",
				},
			},
		}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.Stage.Args.DefaultRouteSettings)
	})
}

// TestCreateAPIGatewayResourcesAccessLogs tests access logging configuration.
func TestCreateAPIGatewayResourcesAccessLogs(t *testing.T) {
	t.Run("creates API Gateway with access logs", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			AccessLogs: &AccessLogsConfig{
				DestinationArn: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/apigateway/test-api",
				Format:         "$requestId",
			},
		}

		functions := map[string]FunctionConfig{
			"hello": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				Source: SourceConfig{
					Path: "./src",
				},
			},
		}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.NotNil(t, resources.Stage.Args.AccessLogSettings)
	})
}

// TestCreateAPIGatewayResourcesPermissions tests Lambda permission creation.
func TestCreateAPIGatewayResourcesPermissions(t *testing.T) {
	t.Run("creates Lambda permissions for API Gateway invocation", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
		}

		functions := map[string]FunctionConfig{
			"hello": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				Source: SourceConfig{
					Path: "./src",
				},
				HTTPRouting: &HTTPRoutingConfig{
					Path:   "/hello",
					Method: "GET",
				},
			},
			"goodbye": {
				Handler: "index.handler",
				Runtime: "nodejs20.x",
				Source: SourceConfig{
					Path: "./src",
				},
				HTTPRouting: &HTTPRoutingConfig{
					Path:   "/goodbye",
					Method: "POST",
				},
			},
		}

		resources, err := createAPIGatewayResources("test-service", config, functions)

		require.NoError(t, err)
		assert.Len(t, resources.Permissions, 2)
	})
}
