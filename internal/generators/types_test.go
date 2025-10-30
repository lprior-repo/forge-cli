package generators

import (
	"context"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
)

// TestResourceTypes tests resource type constants
func TestResourceTypes(t *testing.T) {
	t.Run("all resource types are defined", func(t *testing.T) {
		assert.Equal(t, ResourceType("lambda"), ResourceLambda)
		assert.Equal(t, ResourceType("sqs"), ResourceSQS)
		assert.Equal(t, ResourceType("sns"), ResourceSNS)
		assert.Equal(t, ResourceType("dynamodb"), ResourceDynamoDB)
		assert.Equal(t, ResourceType("apigw"), ResourceAPIGateway)
		assert.Equal(t, ResourceType("eventbridge"), ResourceEventBridge)
		assert.Equal(t, ResourceType("sfn"), ResourceStepFunctions)
		assert.Equal(t, ResourceType("s3"), ResourceS3)
		assert.Equal(t, ResourceType("cognito"), ResourceCognito)
	})
}

// TestResourceIntent tests intent data structure
func TestResourceIntent(t *testing.T) {
	t.Run("creates intent with required fields", func(t *testing.T) {
		intent := ResourceIntent{
			Type:      ResourceSQS,
			Name:      "orders_queue",
			ToFunc:    "process_orders",
			UseModule: true,
			Flags:     map[string]string{"fifo": "true"},
		}

		assert.Equal(t, ResourceSQS, intent.Type)
		assert.Equal(t, "orders_queue", intent.Name)
		assert.Equal(t, "process_orders", intent.ToFunc)
		assert.True(t, intent.UseModule)
		assert.Equal(t, "true", intent.Flags["fifo"])
	})

	t.Run("intent without integration", func(t *testing.T) {
		intent := ResourceIntent{
			Type: ResourceDynamoDB,
			Name: "users_table",
		}

		assert.Equal(t, ResourceDynamoDB, intent.Type)
		assert.Empty(t, intent.ToFunc)
	})
}

// TestProjectState tests project state data structure
func TestProjectState(t *testing.T) {
	t.Run("creates empty project state", func(t *testing.T) {
		state := ProjectState{
			ProjectRoot: "/project",
			Functions:   make(map[string]FunctionInfo),
			Queues:      make(map[string]QueueInfo),
			Tables:      make(map[string]TableInfo),
			APIs:        make(map[string]APIInfo),
			Topics:      make(map[string]TopicInfo),
			InfraFiles:  []string{},
		}

		assert.Equal(t, "/project", state.ProjectRoot)
		assert.Empty(t, state.Functions)
		assert.Empty(t, state.Queues)
	})

	t.Run("stores discovered resources", func(t *testing.T) {
		state := ProjectState{
			ProjectRoot: "/project",
			Functions: map[string]FunctionInfo{
				"api": {
					Name:       "api",
					Runtime:    "go1.x",
					SourcePath: "src/functions/api",
					Handler:    "bootstrap",
					TFResource: "aws_lambda_function.api",
				},
			},
			Queues: map[string]QueueInfo{
				"orders": {
					Name:       "orders",
					TFResource: "module.orders_queue",
				},
			},
		}

		assert.Len(t, state.Functions, 1)
		assert.Equal(t, "api", state.Functions["api"].Name)
		assert.Equal(t, "go1.x", state.Functions["api"].Runtime)

		assert.Len(t, state.Queues, 1)
		assert.Equal(t, "orders", state.Queues["orders"].Name)
	})
}

// TestResourceConfig tests configuration data structure
func TestResourceConfig(t *testing.T) {
	t.Run("creates config without integration", func(t *testing.T) {
		config := ResourceConfig{
			Type:   ResourceS3,
			Name:   "data_bucket",
			Module: true,
			Variables: map[string]interface{}{
				"versioning": true,
				"encryption": "AES256",
			},
			Integration: nil,
		}

		assert.Equal(t, ResourceS3, config.Type)
		assert.True(t, config.Module)
		assert.True(t, config.Variables["versioning"].(bool))
		assert.Nil(t, config.Integration)
	})

	t.Run("creates config with integration", func(t *testing.T) {
		config := ResourceConfig{
			Type:   ResourceSQS,
			Name:   "jobs_queue",
			Module: true,
			Integration: &IntegrationConfig{
				TargetFunction: "worker",
				EventSource: &EventSourceConfig{
					ARNExpression: "module.jobs_queue.queue_arn",
					BatchSize:     10,
				},
				IAMPermissions: []IAMPermission{
					{
						Effect:    "Allow",
						Actions:   []string{"sqs:ReceiveMessage", "sqs:DeleteMessage"},
						Resources: []string{"module.jobs_queue.queue_arn"},
					},
				},
				EnvVars: map[string]string{
					"QUEUE_URL": "module.jobs_queue.queue_url",
				},
			},
		}

		assert.NotNil(t, config.Integration)
		assert.Equal(t, "worker", config.Integration.TargetFunction)
		assert.Equal(t, 10, config.Integration.EventSource.BatchSize)
		assert.Len(t, config.Integration.IAMPermissions, 1)
		assert.Contains(t, config.Integration.EnvVars, "QUEUE_URL")
	})
}

// TestGeneratedCode tests code generation data structure
func TestGeneratedCode(t *testing.T) {
	t.Run("creates generated code with resources", func(t *testing.T) {
		code := GeneratedCode{
			Resources: `resource "aws_sqs_queue" "orders" {
  name = "orders"
}`,
			Variables: `variable "queue_name" {
  type = string
}`,
			Outputs: `output "queue_url" {
  value = aws_sqs_queue.orders.url
}`,
			ModuleCalls: "",
			Files: []FileToWrite{
				{
					Path:    "sqs.tf",
					Content: "# SQS resources",
					Mode:    WriteModeCreate,
				},
			},
		}

		assert.Contains(t, code.Resources, "aws_sqs_queue")
		assert.Contains(t, code.Variables, "variable")
		assert.Contains(t, code.Outputs, "output")
		assert.Len(t, code.Files, 1)
		assert.Equal(t, WriteModeCreate, code.Files[0].Mode)
	})
}

// TestWriteMode tests write mode constants
func TestWriteMode(t *testing.T) {
	t.Run("all write modes are defined", func(t *testing.T) {
		assert.Equal(t, WriteMode("create"), WriteModeCreate)
		assert.Equal(t, WriteMode("append"), WriteModeAppend)
		assert.Equal(t, WriteMode("update"), WriteModeUpdate)
	})
}

// TestRegistry tests generator registry
func TestRegistry(t *testing.T) {
	t.Run("creates empty registry", func(t *testing.T) {
		registry := NewRegistry()

		assert.NotNil(t, registry)
		assert.NotNil(t, registry.generators)
		assert.Empty(t, registry.generators)
	})

	t.Run("registers and retrieves generator", func(t *testing.T) {
		registry := NewRegistry()
		mockGen := &mockGenerator{}

		registry.Register(ResourceSQS, mockGen)

		gen, ok := registry.Get(ResourceSQS)
		assert.True(t, ok)
		assert.Equal(t, mockGen, gen)
	})

	t.Run("returns false for unregistered type", func(t *testing.T) {
		registry := NewRegistry()

		_, ok := registry.Get(ResourceSQS)
		assert.False(t, ok)
	})

	t.Run("registers multiple generators", func(t *testing.T) {
		registry := NewRegistry()
		sqsGen := &mockGenerator{}
		snsGen := &mockGenerator{}

		registry.Register(ResourceSQS, sqsGen)
		registry.Register(ResourceSNS, snsGen)

		gen1, ok1 := registry.Get(ResourceSQS)
		gen2, ok2 := registry.Get(ResourceSNS)

		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.Equal(t, sqsGen, gen1)
		assert.Equal(t, snsGen, gen2)
	})

	t.Run("overwrites generator on re-register", func(t *testing.T) {
		registry := NewRegistry()
		gen1 := &mockGenerator{}
		gen2 := &mockGenerator{}

		registry.Register(ResourceSQS, gen1)
		registry.Register(ResourceSQS, gen2) // Overwrite

		gen, ok := registry.Get(ResourceSQS)
		assert.True(t, ok)
		assert.Equal(t, gen2, gen)
	})
}

// TestFunctionInfo tests function info data structure
func TestFunctionInfo(t *testing.T) {
	t.Run("stores complete function metadata", func(t *testing.T) {
		info := FunctionInfo{
			Name:       "api",
			Runtime:    "python3.13",
			SourcePath: "src/functions/api",
			Handler:    "app.handler",
			TFResource: "aws_lambda_function.api",
		}

		assert.Equal(t, "api", info.Name)
		assert.Equal(t, "python3.13", info.Runtime)
		assert.Equal(t, "app.handler", info.Handler)
	})
}

// TestIntegrationConfig tests integration configuration
func TestIntegrationConfig(t *testing.T) {
	t.Run("creates integration with event source", func(t *testing.T) {
		config := IntegrationConfig{
			TargetFunction: "processor",
			EventSource: &EventSourceConfig{
				ARNExpression:         "aws_sqs_queue.jobs.arn",
				BatchSize:             10,
				MaxBatchingWindowSecs: 5,
				MaxConcurrency:        2,
			},
		}

		assert.Equal(t, "processor", config.TargetFunction)
		assert.NotNil(t, config.EventSource)
		assert.Equal(t, 10, config.EventSource.BatchSize)
		assert.Equal(t, 5, config.EventSource.MaxBatchingWindowSecs)
	})

	t.Run("creates integration with IAM permissions", func(t *testing.T) {
		config := IntegrationConfig{
			TargetFunction: "processor",
			IAMPermissions: []IAMPermission{
				{
					Effect:    "Allow",
					Actions:   []string{"dynamodb:GetItem", "dynamodb:PutItem"},
					Resources: []string{"aws_dynamodb_table.users.arn"},
				},
			},
		}

		assert.Len(t, config.IAMPermissions, 1)
		assert.Equal(t, "Allow", config.IAMPermissions[0].Effect)
		assert.Len(t, config.IAMPermissions[0].Actions, 2)
	})

	t.Run("creates integration with environment variables", func(t *testing.T) {
		config := IntegrationConfig{
			TargetFunction: "api",
			EnvVars: map[string]string{
				"TABLE_NAME": "aws_dynamodb_table.users.name",
				"QUEUE_URL":  "aws_sqs_queue.jobs.url",
			},
		}

		assert.Len(t, config.EnvVars, 2)
		assert.Contains(t, config.EnvVars, "TABLE_NAME")
		assert.Contains(t, config.EnvVars, "QUEUE_URL")
	})
}

// mockGenerator implements Generator for testing
type mockGenerator struct {
	promptFunc   func(context.Context, ResourceIntent, ProjectState) E.Either[error, ResourceConfig]
	generateFunc func(ResourceConfig, ProjectState) E.Either[error, GeneratedCode]
	validateFunc func(ResourceConfig) E.Either[error, ResourceConfig]
}

func (m *mockGenerator) Prompt(ctx context.Context, intent ResourceIntent, state ProjectState) E.Either[error, ResourceConfig] {
	if m.promptFunc != nil {
		return m.promptFunc(ctx, intent, state)
	}
	return E.Right[error, ResourceConfig](ResourceConfig{Type: intent.Type, Name: intent.Name})
}

func (m *mockGenerator) Generate(config ResourceConfig, state ProjectState) E.Either[error, GeneratedCode] {
	if m.generateFunc != nil {
		return m.generateFunc(config, state)
	}
	return E.Right[error, GeneratedCode](GeneratedCode{Resources: "# Generated"})
}

func (m *mockGenerator) Validate(config ResourceConfig) E.Either[error, ResourceConfig] {
	if m.validateFunc != nil {
		return m.validateFunc(config)
	}
	return E.Right[error, ResourceConfig](config)
}
