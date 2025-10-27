package lingon

import (
	"context"
	"testing"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGenerator tests the generator constructor
func TestNewGenerator(t *testing.T) {
	t.Run("creates generator with generate function", func(t *testing.T) {
		gen := NewGenerator()

		assert.NotNil(t, gen.Generate)
	})
}

// TestValidateConfig tests configuration validation
func TestValidateConfig(t *testing.T) {
	t.Run("valid minimal config", func(t *testing.T) {
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

		err := validateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("fails when service name is missing", func(t *testing.T) {
		config := ForgeConfig{
			Service: "",
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

		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service name is required")
	})

	t.Run("fails when region is missing", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "",
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

		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider region is required")
	})

	t.Run("fails when no functions defined", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{},
		}

		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one function is required")
	})
}

// TestValidateFunction tests function validation
func TestValidateFunction(t *testing.T) {
	t.Run("valid function config", func(t *testing.T) {
		fn := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
		}

		err := validateFunction("test", fn)
		assert.NoError(t, err)
	})

	t.Run("fails when handler is missing", func(t *testing.T) {
		fn := FunctionConfig{
			Handler: "",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
		}

		err := validateFunction("test", fn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler is required")
	})

	t.Run("fails when runtime is missing", func(t *testing.T) {
		fn := FunctionConfig{
			Handler: "index.handler",
			Runtime: "",
			Source: SourceConfig{
				Path: "./src",
			},
		}

		err := validateFunction("test", fn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "runtime is required")
	})

	t.Run("fails when source is missing", func(t *testing.T) {
		fn := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source:  SourceConfig{},
		}

		err := validateFunction("test", fn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source path, S3 location, or filename is required")
	})

	t.Run("accepts S3 source", func(t *testing.T) {
		fn := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				S3Bucket: "my-bucket",
				S3Key:    "lambda.zip",
			},
		}

		err := validateFunction("test", fn)
		assert.NoError(t, err)
	})

	t.Run("accepts filename source", func(t *testing.T) {
		fn := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Filename: "./lambda.zip",
			},
		}

		err := validateFunction("test", fn)
		assert.NoError(t, err)
	})

	t.Run("validates runtime", func(t *testing.T) {
		validRuntimes := []string{
			"nodejs18.x", "nodejs20.x",
			"python3.9", "python3.10", "python3.11", "python3.12",
			"go1.x",
			"java11", "java17", "java21",
		}

		for _, runtime := range validRuntimes {
			fn := FunctionConfig{
				Handler: "index.handler",
				Runtime: runtime,
				Source: SourceConfig{
					Path: "./src",
				},
			}

			err := validateFunction("test", fn)
			assert.NoError(t, err, "Runtime %s should be valid", runtime)
		}
	})

	t.Run("fails for invalid runtime", func(t *testing.T) {
		fn := FunctionConfig{
			Handler: "index.handler",
			Runtime: "invalid-runtime",
			Source: SourceConfig{
				Path: "./src",
			},
		}

		err := validateFunction("test", fn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported runtime")
	})
}

// TestGenerateLambdaFunction tests Lambda function generation
func TestGenerateLambdaFunction(t *testing.T) {
	t.Run("generates basic lambda function", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
		}

		fn, err := generateLambdaFunction("test-service", "hello", config)

		require.NoError(t, err)
		assert.Equal(t, "test-service-hello", fn.Name)
		assert.Equal(t, config, fn.Config)
		assert.NotNil(t, fn.Role)
	})

	t.Run("generates IAM role", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
		}

		fn, err := generateLambdaFunction("test-service", "hello", config)

		require.NoError(t, err)
		assert.NotNil(t, fn.Role)
		assert.Equal(t, "test-service-hello-role", fn.Role.Name)
		assert.Contains(t, fn.Role.AssumeRolePolicy, "lambda.amazonaws.com")
		assert.Contains(t, fn.Role.ManagedPolicyArns, "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole")
	})

	t.Run("generates log group when configured", func(t *testing.T) {
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

		fn, err := generateLambdaFunction("test-service", "hello", config)

		require.NoError(t, err)
		assert.True(t, O.IsSome(fn.LogGroup))

		logGroup := O.Fold(
			func() *CloudWatchLogGroup { return nil },
			func(lg *CloudWatchLogGroup) *CloudWatchLogGroup { return lg },
		)(fn.LogGroup)

		assert.NotNil(t, logGroup)
		assert.Equal(t, "/aws/lambda/test-service-hello", logGroup.Name)
	})

	t.Run("generates function URL when configured", func(t *testing.T) {
		config := FunctionConfig{
			Handler: "index.handler",
			Runtime: "nodejs20.x",
			Source: SourceConfig{
				Path: "./src",
			},
			FunctionURL: &FunctionURLConfig{
				AuthorizationType: "NONE",
			},
		}

		fn, err := generateLambdaFunction("test-service", "hello", config)

		require.NoError(t, err)
		assert.True(t, O.IsSome(fn.FunctionURL))

		functionURL := O.Fold(
			func() *LambdaFunctionURL { return nil },
			func(furl *LambdaFunctionURL) *LambdaFunctionURL { return furl },
		)(fn.FunctionURL)

		assert.NotNil(t, functionURL)
		assert.Equal(t, "NONE", functionURL.AuthorizationType)
	})

	t.Run("generates event source mappings", func(t *testing.T) {
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
			},
		}

		fn, err := generateLambdaFunction("test-service", "hello", config)

		require.NoError(t, err)
		assert.Len(t, fn.EventSources, 1)
		assert.Equal(t, "test-service-hello", fn.EventSources[0].FunctionName)
		assert.Equal(t, "arn:aws:dynamodb:us-east-1:123456789012:table/test/stream/2021-01-01T00:00:00.000", fn.EventSources[0].EventSourceArn)
	})
}

// TestGenerateIAMRole tests IAM role generation
func TestGenerateIAMRole(t *testing.T) {
	t.Run("generates default role", func(t *testing.T) {
		config := IAMConfig{}

		role := generateIAMRole("test-service", "hello", config)

		assert.Equal(t, "test-service-hello-role", role.Name)
		assert.Contains(t, role.AssumeRolePolicy, "lambda.amazonaws.com")
		assert.Contains(t, role.ManagedPolicyArns, "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole")
	})

	t.Run("uses custom role name", func(t *testing.T) {
		config := IAMConfig{
			RoleName: "custom-role",
		}

		role := generateIAMRole("test-service", "hello", config)

		assert.Equal(t, "custom-role", role.Name)
	})

	t.Run("uses custom assume role policy", func(t *testing.T) {
		customPolicy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}`

		config := IAMConfig{
			AssumeRolePolicy: customPolicy,
		}

		role := generateIAMRole("test-service", "hello", config)

		assert.Equal(t, customPolicy, role.AssumeRolePolicy)
	})

	t.Run("uses custom managed policies", func(t *testing.T) {
		config := IAMConfig{
			ManagedPolicyArns: []string{
				"arn:aws:iam::aws:policy/AmazonS3FullAccess",
				"arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess",
			},
		}

		role := generateIAMRole("test-service", "hello", config)

		assert.Equal(t, config.ManagedPolicyArns, role.ManagedPolicyArns)
	})

	t.Run("includes inline policies", func(t *testing.T) {
		config := IAMConfig{
			InlinePolicies: []InlinePolicy{
				{
					Name:   "s3-access",
					Policy: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:*","Resource":"*"}]}`,
				},
			},
		}

		role := generateIAMRole("test-service", "hello", config)

		assert.Len(t, role.InlinePolicies, 1)
		assert.Equal(t, "s3-access", role.InlinePolicies[0].Name)
	})
}

// TestGenerateAPIGateway tests API Gateway generation
func TestGenerateAPIGateway(t *testing.T) {
	t.Run("generates API Gateway with integrations", func(t *testing.T) {
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

		api, err := generateAPIGateway("test-service", config, functions)

		require.NoError(t, err)
		assert.Equal(t, "test-api", api.Name)
		assert.Len(t, api.Integrations, 1)
		assert.Len(t, api.Routes, 1)
	})

	t.Run("generates default stage", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
		}

		functions := map[string]FunctionConfig{}

		api, err := generateAPIGateway("test-service", config, functions)

		require.NoError(t, err)
		assert.Len(t, api.Stages, 1)
		assert.Contains(t, api.Stages, "default")
	})

	t.Run("generates custom stages", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			Stages: map[string]StageConfig{
				"production": {
					Name:       "production",
					AutoDeploy: true,
				},
				"staging": {
					Name:       "staging",
					AutoDeploy: true,
				},
			},
		}

		functions := map[string]FunctionConfig{}

		api, err := generateAPIGateway("test-service", config, functions)

		require.NoError(t, err)
		assert.Len(t, api.Stages, 2)
		assert.Contains(t, api.Stages, "production")
		assert.Contains(t, api.Stages, "staging")
	})

	t.Run("generates custom domain", func(t *testing.T) {
		config := APIGatewayConfig{
			Name:         "test-api",
			ProtocolType: "HTTP",
			Domain: &DomainConfig{
				DomainName:     "api.example.com",
				CertificateArn: "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			},
		}

		functions := map[string]FunctionConfig{}

		api, err := generateAPIGateway("test-service", config, functions)

		require.NoError(t, err)
		assert.True(t, O.IsSome(api.Domain))

		domain := O.Fold(
			func() *APIGatewayDomain { return nil },
			func(d *APIGatewayDomain) *APIGatewayDomain { return d },
		)(api.Domain)

		assert.NotNil(t, domain)
		assert.Equal(t, "api.example.com", domain.DomainName)
	})
}

// TestGenerateDynamoDBTable tests DynamoDB table generation
func TestGenerateDynamoDBTable(t *testing.T) {
	t.Run("generates basic table", func(t *testing.T) {
		config := TableConfig{
			TableName:   "users",
			BillingMode: "PAY_PER_REQUEST",
			HashKey:     "userId",
			Attributes: []AttributeDefinition{
				{Name: "userId", Type: "S"},
			},
		}

		table := generateDynamoDBTable("test-service", "users", config)

		assert.Equal(t, "users", table.Name)
		assert.Equal(t, "users", table.Config.TableName)
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

		table := generateDynamoDBTable("test-service", "users", config)

		assert.Equal(t, "test-service-users", table.Name)
	})
}

// TestGenerateStack tests complete stack generation
func TestGenerateStack(t *testing.T) {
	t.Run("generates complete stack", func(t *testing.T) {
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
			Tables: map[string]TableConfig{
				"users": {
					BillingMode: "PAY_PER_REQUEST",
					HashKey:     "userId",
					Attributes: []AttributeDefinition{
						{Name: "userId", Type: "S"},
					},
				},
			},
		}

		stack, err := generateStack(config)

		require.NoError(t, err)
		assert.Equal(t, "test-service", stack.Service)
		assert.Len(t, stack.Functions, 1)
		assert.Len(t, stack.Tables, 1)
		assert.Contains(t, stack.Functions, "hello")
		assert.Contains(t, stack.Tables, "users")
	})

	t.Run("succeeds with empty functions", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{},
		}

		stack, err := generateStack(config)

		require.NoError(t, err)
		assert.Equal(t, "test-service", stack.Service)
		assert.Len(t, stack.Functions, 0)
	})

	t.Run("fails with invalid function config", func(t *testing.T) {
		config := ForgeConfig{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{
				"invalid": {
					Handler: "",  // Missing handler
					Runtime: "nodejs20.x",
					Source: SourceConfig{
						Path: "./src",
					},
				},
			},
		}

		_, err := generateStack(config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler is required")
	})
}

// TestGenerateFunc tests the main generator function
func TestGenerateFunc(t *testing.T) {
	t.Run("generates terraform code from valid config", func(t *testing.T) {
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

		gen := NewGenerator()
		result := gen.Generate(context.Background(), config)

		assert.True(t, E.IsRight(result))

		code := E.Fold(
			func(err error) []byte { return nil },
			func(code []byte) []byte { return code },
		)(result)

		assert.NotNil(t, code)
		assert.Contains(t, string(code), "test-service")
		assert.Contains(t, string(code), "us-east-1")
	})

	t.Run("fails with invalid config", func(t *testing.T) {
		config := ForgeConfig{
			Service: "",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]FunctionConfig{},
		}

		gen := NewGenerator()
		result := gen.Generate(context.Background(), config)

		assert.True(t, E.IsLeft(result))

		err := E.Fold(
			func(err error) error { return err },
			func(code []byte) error { return nil },
		)(result)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid configuration")
	})
}

// TestExportToTerraform tests Terraform export
func TestExportToTerraform(t *testing.T) {
	t.Run("exports stack to terraform", func(t *testing.T) {
		stack := &Stack{
			Service: "test-service",
			Provider: ProviderConfig{
				Region: "us-east-1",
			},
			Functions: map[string]*LambdaFunction{
				"hello": {
					Name: "test-service-hello",
					Config: FunctionConfig{
						Handler: "index.handler",
						Runtime: "nodejs20.x",
					},
				},
			},
			APIGateway: O.None[*APIGateway](),
			Tables:     make(map[string]*DynamoDBTable),
			EventBridgeRules: make(map[string]*EventBridgeRule),
			StateMachines: make(map[string]*StepFunctionsStateMachine),
			Topics:     make(map[string]*SNSTopic),
			Queues:     make(map[string]*SQSQueue),
			Buckets:    make(map[string]*S3Bucket),
			Alarms:     make(map[string]*CloudWatchAlarm),
		}

		code, err := exportToTerraform(stack)

		require.NoError(t, err)
		assert.NotNil(t, code)
		assert.Contains(t, string(code), "test-service")
		assert.Contains(t, string(code), "us-east-1")
		assert.Contains(t, string(code), "terraform")
		assert.Contains(t, string(code), "provider \"aws\"")
	})
}
