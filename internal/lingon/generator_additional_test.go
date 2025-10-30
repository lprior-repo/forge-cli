package lingon

import (
	"testing"

	O "github.com/IBM/fp-go/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateEventBridgeRule(t *testing.T) {
	t.Run("generates basic EventBridge rule", func(t *testing.T) {
		config := EventBridgeConfig{
			Name:        "test-rule",
			Description: "Test EventBridge rule",
		}

		rule := generateEventBridgeRule("test-service", "events", config)

		assert.NotNil(t, rule)
		assert.Equal(t, "test-rule", rule.Name)
		assert.Equal(t, "test-rule", rule.Config.Name)
	})

	t.Run("uses service prefix when name is empty", func(t *testing.T) {
		config := EventBridgeConfig{
			Name: "",
		}

		rule := generateEventBridgeRule("my-service", "events", config)

		assert.Equal(t, "my-service-events", rule.Name)
		assert.Equal(t, "my-service-events", rule.Config.Name)
	})

	t.Run("preserves all config fields", func(t *testing.T) {
		config := EventBridgeConfig{
			Name:        "custom-rule",
			Description: "Custom rule",
		}

		rule := generateEventBridgeRule("test-service", "events", config)

		assert.Equal(t, "Custom rule", rule.Config.Description)
	})
}

func TestGenerateStateMachine(t *testing.T) {
	t.Run("generates basic state machine", func(t *testing.T) {
		config := StateMachineConfig{
			Name:       "test-sm",
			Definition: `{"StartAt": "HelloWorld"}`,
		}

		sm := generateStateMachine("test-service", "workflow", config)

		assert.NotNil(t, sm)
		assert.Equal(t, "test-sm", sm.Name)
		assert.Equal(t, "test-sm", sm.Config.Name)
	})

	t.Run("uses service prefix when name is empty", func(t *testing.T) {
		config := StateMachineConfig{
			Name: "",
		}

		sm := generateStateMachine("my-service", "workflow", config)

		assert.Equal(t, "my-service-workflow", sm.Name)
		assert.Equal(t, "my-service-workflow", sm.Config.Name)
	})

	t.Run("preserves all config fields", func(t *testing.T) {
		config := StateMachineConfig{
			Name:       "custom-sm",
			Definition: `{"StartAt": "Step1"}`,
		}

		sm := generateStateMachine("test-service", "workflow", config)

		assert.JSONEq(t, `{"StartAt": "Step1"}`, sm.Config.Definition)
	})
}

func TestGenerateSNSTopic(t *testing.T) {
	t.Run("generates basic SNS topic", func(t *testing.T) {
		config := TopicConfig{
			Name:        "test-topic",
			DisplayName: "Test Topic",
		}

		topic := generateSNSTopic("test-service", "notifications", config)

		assert.NotNil(t, topic)
		assert.Equal(t, "test-topic", topic.Name)
		assert.Equal(t, "test-topic", topic.Config.Name)
	})

	t.Run("uses service prefix when name is empty", func(t *testing.T) {
		config := TopicConfig{
			Name: "",
		}

		topic := generateSNSTopic("my-service", "notifications", config)

		assert.Equal(t, "my-service-notifications", topic.Name)
		assert.Equal(t, "my-service-notifications", topic.Config.Name)
	})

	t.Run("preserves all config fields", func(t *testing.T) {
		config := TopicConfig{
			Name:        "custom-topic",
			DisplayName: "Custom Topic",
		}

		topic := generateSNSTopic("test-service", "notifications", config)

		assert.Equal(t, "Custom Topic", topic.Config.DisplayName)
	})
}

func TestGenerateSQSQueue(t *testing.T) {
	t.Run("generates basic SQS queue", func(t *testing.T) {
		config := QueueConfig{
			Name:                     "test-queue",
			VisibilityTimeoutSeconds: 30,
		}

		queue := generateSQSQueue("test-service", "jobs", config)

		assert.NotNil(t, queue)
		assert.Equal(t, "test-queue", queue.Name)
		assert.Equal(t, "test-queue", queue.Config.Name)
	})

	t.Run("uses service prefix when name is empty", func(t *testing.T) {
		config := QueueConfig{
			Name: "",
		}

		queue := generateSQSQueue("my-service", "jobs", config)

		assert.Equal(t, "my-service-jobs", queue.Name)
		assert.Equal(t, "my-service-jobs", queue.Config.Name)
	})

	t.Run("preserves all config fields", func(t *testing.T) {
		config := QueueConfig{
			Name:                     "custom-queue",
			VisibilityTimeoutSeconds: 60,
		}

		queue := generateSQSQueue("test-service", "jobs", config)

		assert.Equal(t, 60, queue.Config.VisibilityTimeoutSeconds)
	})
}

func TestGenerateS3Bucket(t *testing.T) {
	t.Run("generates basic S3 bucket", func(t *testing.T) {
		config := BucketConfig{
			Name: "test-bucket",
		}

		bucket := generateS3Bucket("test-service", "storage", config)

		assert.NotNil(t, bucket)
		assert.Equal(t, "test-bucket", bucket.Name)
		assert.Equal(t, "test-bucket", bucket.Config.Name)
	})

	t.Run("uses service prefix when name is empty", func(t *testing.T) {
		config := BucketConfig{
			Name: "",
		}

		bucket := generateS3Bucket("my-service", "storage", config)

		assert.Equal(t, "my-service-storage", bucket.Name)
		assert.Equal(t, "my-service-storage", bucket.Config.Name)
	})

	t.Run("preserves all config fields", func(t *testing.T) {
		config := BucketConfig{
			Name: "custom-bucket",
			Versioning: &VersioningConfig{
				Enabled: true,
			},
		}

		bucket := generateS3Bucket("test-service", "storage", config)

		assert.NotNil(t, bucket.Config.Versioning)
		assert.True(t, bucket.Config.Versioning.Enabled)
	})
}

func TestGenerateCloudWatchAlarm(t *testing.T) {
	t.Run("generates basic CloudWatch alarm", func(t *testing.T) {
		config := AlarmConfig{
			Name:               "test-alarm",
			MetricName:         "Errors",
			ComparisonOperator: "GreaterThanThreshold",
		}

		alarm := generateCloudWatchAlarm("test-service", "errors", config)

		assert.NotNil(t, alarm)
		assert.Equal(t, "test-alarm", alarm.Name)
		assert.Equal(t, "test-alarm", alarm.Config.Name)
	})

	t.Run("uses service prefix when name is empty", func(t *testing.T) {
		config := AlarmConfig{
			Name: "",
		}

		alarm := generateCloudWatchAlarm("my-service", "errors", config)

		assert.Equal(t, "my-service-errors", alarm.Name)
		assert.Equal(t, "my-service-errors", alarm.Config.Name)
	})

	t.Run("preserves all config fields", func(t *testing.T) {
		config := AlarmConfig{
			Name:               "custom-alarm",
			MetricName:         "Duration",
			ComparisonOperator: "GreaterThanThreshold",
		}

		alarm := generateCloudWatchAlarm("test-service", "duration", config)

		assert.Equal(t, "Duration", alarm.Config.MetricName)
		assert.Equal(t, "GreaterThanThreshold", alarm.Config.ComparisonOperator)
	})
}

func TestGenerateStackWithAllResources(t *testing.T) {
	t.Run("generates stack with EventBridge rules", func(t *testing.T) {
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
			EventBridge: map[string]EventBridgeConfig{
				"schedule": {
					Name:        "daily-schedule",
					Description: "Daily trigger",
				},
			},
		}

		stack, err := generateStack(config)

		require.NoError(t, err)
		assert.Len(t, stack.EventBridgeRules, 1)
		assert.Contains(t, stack.EventBridgeRules, "schedule")
	})

	t.Run("generates stack with Step Functions state machines", func(t *testing.T) {
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
			StateMachines: map[string]StateMachineConfig{
				"workflow": {
					Name:       "test-workflow",
					Definition: `{"StartAt": "HelloWorld"}`,
				},
			},
		}

		stack, err := generateStack(config)

		require.NoError(t, err)
		assert.Len(t, stack.StateMachines, 1)
		assert.Contains(t, stack.StateMachines, "workflow")
	})

	t.Run("generates stack with SNS topics", func(t *testing.T) {
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
			Topics: map[string]TopicConfig{
				"notifications": {
					Name:        "user-notifications",
					DisplayName: "User Notifications",
				},
			},
		}

		stack, err := generateStack(config)

		require.NoError(t, err)
		assert.Len(t, stack.Topics, 1)
		assert.Contains(t, stack.Topics, "notifications")
	})

	t.Run("generates stack with SQS queues", func(t *testing.T) {
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
			Queues: map[string]QueueConfig{
				"jobs": {
					Name:                     "job-queue",
					VisibilityTimeoutSeconds: 30,
				},
			},
		}

		stack, err := generateStack(config)

		require.NoError(t, err)
		assert.Len(t, stack.Queues, 1)
		assert.Contains(t, stack.Queues, "jobs")
	})

	t.Run("generates stack with S3 buckets", func(t *testing.T) {
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
			Buckets: map[string]BucketConfig{
				"storage": {
					Name: "my-storage-bucket",
					Versioning: &VersioningConfig{
						Enabled: true,
					},
				},
			},
		}

		stack, err := generateStack(config)

		require.NoError(t, err)
		assert.Len(t, stack.Buckets, 1)
		assert.Contains(t, stack.Buckets, "storage")
	})

	t.Run("generates stack with CloudWatch alarms", func(t *testing.T) {
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
			Alarms: map[string]AlarmConfig{
				"errors": {
					Name:               "high-error-rate",
					MetricName:         "Errors",
					ComparisonOperator: "GreaterThanThreshold",
				},
			},
		}

		stack, err := generateStack(config)

		require.NoError(t, err)
		assert.Len(t, stack.Alarms, 1)
		assert.Contains(t, stack.Alarms, "errors")
	})

	t.Run("generates stack with all resource types", func(t *testing.T) {
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
			APIGateway: &APIGatewayConfig{
				Name:         "test-api",
				ProtocolType: "HTTP",
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
			EventBridge: map[string]EventBridgeConfig{
				"schedule": {
					Name: "daily-schedule",
				},
			},
			StateMachines: map[string]StateMachineConfig{
				"workflow": {
					Name:       "test-workflow",
					Definition: `{"StartAt": "HelloWorld"}`,
				},
			},
			Topics: map[string]TopicConfig{
				"notifications": {
					Name: "user-notifications",
				},
			},
			Queues: map[string]QueueConfig{
				"jobs": {
					Name: "job-queue",
				},
			},
			Buckets: map[string]BucketConfig{
				"storage": {
					Name: "my-storage-bucket",
				},
			},
			Alarms: map[string]AlarmConfig{
				"errors": {
					Name:       "high-error-rate",
					MetricName: "Errors",
				},
			},
		}

		stack, err := generateStack(config)

		require.NoError(t, err)
		assert.Equal(t, "test-service", stack.Service)
		assert.Len(t, stack.Functions, 1)
		assert.True(t, O.IsSome(stack.APIGateway))
		assert.Len(t, stack.Tables, 1)
		assert.Len(t, stack.EventBridgeRules, 1)
		assert.Len(t, stack.StateMachines, 1)
		assert.Len(t, stack.Topics, 1)
		assert.Len(t, stack.Queues, 1)
		assert.Len(t, stack.Buckets, 1)
		assert.Len(t, stack.Alarms, 1)
	})

	t.Run("generates empty stack without optional resources", func(t *testing.T) {
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

		stack, err := generateStack(config)

		require.NoError(t, err)
		assert.Len(t, stack.Functions, 1)
		assert.True(t, O.IsNone(stack.APIGateway))
		assert.Empty(t, stack.Tables)
		assert.Empty(t, stack.EventBridgeRules)
		assert.Empty(t, stack.StateMachines)
		assert.Empty(t, stack.Topics)
		assert.Empty(t, stack.Queues)
		assert.Empty(t, stack.Buckets)
		assert.Empty(t, stack.Alarms)
	})
}

func TestExportToTerraformEdgeCases(t *testing.T) {
	t.Run("exports stack with only functions", func(t *testing.T) {
		stack := &Stack{
			Service: "minimal-service",
			Provider: ProviderConfig{
				Region: "us-west-2",
			},
			Functions: map[string]*LambdaFunction{
				"simple": {
					Name: "minimal-service-simple",
					Config: FunctionConfig{
						Handler: "index.handler",
						Runtime: "python3.11",
						Source: SourceConfig{
							Path: "./src/simple",
						},
					},
					Role: &IAMRole{
						Name:              "minimal-service-simple-role",
						AssumeRolePolicy:  `{"Version":"2012-10-17"}`,
						ManagedPolicyArns: []string{"arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"},
					},
					LogGroup:     O.None[*CloudWatchLogGroup](),
					FunctionURL:  O.None[*LambdaFunctionURL](),
					EventSources: []EventSourceMapping{},
				},
			},
			APIGateway:       O.None[*APIGateway](),
			Tables:           make(map[string]*DynamoDBTable),
			EventBridgeRules: make(map[string]*EventBridgeRule),
			StateMachines:    make(map[string]*StepFunctionsStateMachine),
			Topics:           make(map[string]*SNSTopic),
			Queues:           make(map[string]*SQSQueue),
			Buckets:          make(map[string]*S3Bucket),
			Alarms:           make(map[string]*CloudWatchAlarm),
		}

		code, err := exportToTerraform(stack)

		require.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Contains(t, string(code), "minimal-service")
		assert.Contains(t, string(code), "us-west-2")
	})

	t.Run("exports stack with API Gateway", func(t *testing.T) {
		apiGateway := &APIGateway{
			Name: "test-api",
			Config: APIGatewayConfig{
				Name:         "test-api",
				ProtocolType: "HTTP",
			},
			Integrations: make(map[string]*APIGatewayIntegration),
			Routes:       make(map[string]*APIGatewayRoute),
			Stages:       make(map[string]*APIGatewayStage),
			Domain:       O.None[*APIGatewayDomain](),
		}

		stack := &Stack{
			Service: "api-service",
			Provider: ProviderConfig{
				Region: "eu-west-1",
			},
			Functions: map[string]*LambdaFunction{
				"api": {
					Name: "api-service-api",
					Config: FunctionConfig{
						Handler: "index.handler",
						Runtime: "nodejs20.x",
						Source: SourceConfig{
							Path: "./src/api",
						},
					},
					Role: &IAMRole{
						Name:              "api-service-api-role",
						AssumeRolePolicy:  `{"Version":"2012-10-17"}`,
						ManagedPolicyArns: []string{"arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"},
					},
					LogGroup:     O.None[*CloudWatchLogGroup](),
					FunctionURL:  O.None[*LambdaFunctionURL](),
					EventSources: []EventSourceMapping{},
				},
			},
			APIGateway:       O.Some(apiGateway),
			Tables:           make(map[string]*DynamoDBTable),
			EventBridgeRules: make(map[string]*EventBridgeRule),
			StateMachines:    make(map[string]*StepFunctionsStateMachine),
			Topics:           make(map[string]*SNSTopic),
			Queues:           make(map[string]*SQSQueue),
			Buckets:          make(map[string]*S3Bucket),
			Alarms:           make(map[string]*CloudWatchAlarm),
		}

		code, err := exportToTerraform(stack)

		require.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Contains(t, string(code), "api-service")
	})

	t.Run("exports stack with DynamoDB tables", func(t *testing.T) {
		stack := &Stack{
			Service: "data-service",
			Provider: ProviderConfig{
				Region: "ap-southeast-1",
			},
			Functions: map[string]*LambdaFunction{
				"processor": {
					Name: "data-service-processor",
					Config: FunctionConfig{
						Handler: "main.handler",
						Runtime: "python3.11",
						Source: SourceConfig{
							Path: "./src/processor",
						},
					},
					Role: &IAMRole{
						Name:              "data-service-processor-role",
						AssumeRolePolicy:  `{"Version":"2012-10-17"}`,
						ManagedPolicyArns: []string{"arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"},
					},
					LogGroup:     O.None[*CloudWatchLogGroup](),
					FunctionURL:  O.None[*LambdaFunctionURL](),
					EventSources: []EventSourceMapping{},
				},
			},
			APIGateway: O.None[*APIGateway](),
			Tables: map[string]*DynamoDBTable{
				"users": {
					Name: "data-service-users",
					Config: TableConfig{
						TableName:   "data-service-users",
						BillingMode: "PAY_PER_REQUEST",
						HashKey:     "userId",
						Attributes: []AttributeDefinition{
							{Name: "userId", Type: "S"},
						},
					},
				},
			},
			EventBridgeRules: make(map[string]*EventBridgeRule),
			StateMachines:    make(map[string]*StepFunctionsStateMachine),
			Topics:           make(map[string]*SNSTopic),
			Queues:           make(map[string]*SQSQueue),
			Buckets:          make(map[string]*S3Bucket),
			Alarms:           make(map[string]*CloudWatchAlarm),
		}

		code, err := exportToTerraform(stack)

		require.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Contains(t, string(code), "data-service")
	})
}

func TestValidateRuntimeEdgeCases(t *testing.T) {
	t.Run("validates all supported runtimes", func(t *testing.T) {
		supportedRuntimes := []string{
			"nodejs18.x", "nodejs20.x",
			"python3.9", "python3.10", "python3.11", "python3.12",
			"go1.x",
			"java11", "java17", "java21",
			"dotnet6", "dotnet7", "dotnet8",
			"ruby3.2", "ruby3.3",
			"provided.al2", "provided.al2023",
		}

		for _, runtime := range supportedRuntimes {
			t.Run(runtime, func(t *testing.T) {
				config := FunctionConfig{
					Handler: "index.handler",
					Runtime: runtime,
					Source: SourceConfig{
						Path: "./src",
					},
				}

				err := validateFunction("test", config)
				assert.NoError(t, err, "Runtime %s should be valid", runtime)
			})
		}
	})

	t.Run("rejects unsupported runtimes", func(t *testing.T) {
		unsupportedRuntimes := []string{
			"nodejs16.x", // Old version
			"python3.8",  // Old version
			"go2.x",      // Non-existent
			"java8",      // Old version
			"dotnet5",    // Old version
			"ruby2.7",    // Old version
			"invalid",    // Invalid
			"",           // Empty
		}

		for _, runtime := range unsupportedRuntimes {
			t.Run(runtime, func(t *testing.T) {
				config := FunctionConfig{
					Handler: "index.handler",
					Runtime: runtime,
					Source: SourceConfig{
						Path: "./src",
					},
				}

				err := validateFunction("test", config)
				assert.Error(t, err, "Runtime %s should be invalid", runtime)
			})
		}
	})
}
