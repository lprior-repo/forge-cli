package sqs_test

import (
	"strings"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lewis/forge/internal/generators"
	"github.com/lewis/forge/internal/generators/sqs"
)

// Helper function to extract Right value from Either.
func extractConfig(result E.Either[error, generators.ResourceConfig]) generators.ResourceConfig {
	return E.Fold(
		func(error) generators.ResourceConfig { return generators.ResourceConfig{} },
		func(c generators.ResourceConfig) generators.ResourceConfig { return c },
	)(result)
}

// Helper function to extract error from Either.
func extractError(result E.Either[error, generators.ResourceConfig]) error {
	return E.Fold(
		func(e error) error { return e },
		func(generators.ResourceConfig) error { return nil },
	)(result)
}

// Helper function to extract generated code.
func extractCode(result E.Either[error, generators.GeneratedCode]) generators.GeneratedCode {
	return E.Fold(
		func(error) generators.GeneratedCode { return generators.GeneratedCode{} },
		func(c generators.GeneratedCode) generators.GeneratedCode { return c },
	)(result)
}

// Helper function to extract error from GeneratedCode Either.
func extractCodeError(result E.Either[error, generators.GeneratedCode]) error {
	return E.Fold(
		func(e error) error { return e },
		func(generators.GeneratedCode) error { return nil },
	)(result)
}

// TestNew verifies generator creation.
func TestNew(t *testing.T) {
	gen := sqs.New()
	assert.NotNil(t, gen)
}

// TestPrompt_Standalone tests standalone queue configuration.
func TestPrompt_Standalone(t *testing.T) {
	gen := sqs.New()
	ctx := t.Context()

	intent := generators.ResourceIntent{
		Type:      generators.ResourceSQS,
		Name:      "orders-queue",
		UseModule: true,
	}

	state := generators.ProjectState{
		Functions: make(map[string]generators.FunctionInfo),
	}

	result := gen.Prompt(ctx, intent, state)

	require.True(t, E.IsRight(result), "Prompt should succeed")
	config := extractConfig(result)

	assert.Equal(t, generators.ResourceSQS, config.Type)
	assert.Equal(t, "orders-queue", config.Name)
	assert.True(t, config.Module)
	assert.Nil(t, config.Integration, "Standalone queue should have no integration")

	// Verify default variables
	assert.Equal(t, 30, config.Variables["visibility_timeout_seconds"])
	assert.Equal(t, 345600, config.Variables["message_retention_seconds"])
	assert.Equal(t, true, config.Variables["create_dlq"])
}

// TestPrompt_WithLambdaIntegration tests Lambda integration.
func TestPrompt_WithLambdaIntegration(t *testing.T) {
	gen := sqs.New()
	ctx := t.Context()

	intent := generators.ResourceIntent{
		Type:      generators.ResourceSQS,
		Name:      "orders-queue",
		ToFunc:    "processor",
		UseModule: true,
	}

	state := generators.ProjectState{
		Functions: map[string]generators.FunctionInfo{
			"processor": {
				Name:       "processor",
				Runtime:    "go1.x",
				Handler:    "bootstrap",
				TFResource: "aws_lambda_function.processor",
			},
		},
	}

	result := gen.Prompt(ctx, intent, state)

	require.True(t, E.IsRight(result), "Prompt should succeed")
	config := extractConfig(result)

	assert.NotNil(t, config.Integration, "Should have integration config")
	assert.Equal(t, "processor", config.Integration.TargetFunction)

	// Verify event source config
	assert.NotNil(t, config.Integration.EventSource)
	assert.Equal(t, "module.orders_queue.queue_arn", config.Integration.EventSource.ARNExpression)
	assert.Equal(t, 10, config.Integration.EventSource.BatchSize)
	assert.Equal(t, 5, config.Integration.EventSource.MaxBatchingWindowSecs)
	assert.Equal(t, 10, config.Integration.EventSource.MaxConcurrency)

	// Verify IAM permissions
	require.Len(t, config.Integration.IAMPermissions, 1)
	perm := config.Integration.IAMPermissions[0]
	assert.Equal(t, "Allow", perm.Effect)
	assert.Contains(t, perm.Actions, "sqs:ReceiveMessage")
	assert.Contains(t, perm.Actions, "sqs:DeleteMessage")
	assert.Contains(t, perm.Actions, "sqs:GetQueueAttributes")
	assert.Contains(t, perm.Resources, "module.orders_queue.queue_arn")
}

// TestPrompt_FunctionNotFound tests error when target function doesn't exist.
func TestPrompt_FunctionNotFound(t *testing.T) {
	gen := sqs.New()
	ctx := t.Context()

	intent := generators.ResourceIntent{
		Type:      generators.ResourceSQS,
		Name:      "orders-queue",
		ToFunc:    "nonexistent",
		UseModule: true,
	}

	state := generators.ProjectState{
		Functions: make(map[string]generators.FunctionInfo),
	}

	result := gen.Prompt(ctx, intent, state)

	require.True(t, E.IsLeft(result), "Should fail when function doesn't exist")
	err := extractError(result)
	assert.Contains(t, err.Error(), "nonexistent")
	assert.Contains(t, err.Error(), "not found")
}

// TestValidate tests configuration validation.
func TestValidate(t *testing.T) {
	gen := sqs.New()

	tests := []struct {
		name      string
		config    generators.ResourceConfig
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid configuration",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSQS,
				Name:   "orders-queue",
				Module: true,
			},
			expectErr: false,
		},
		{
			name: "missing name",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSQS,
				Name:   "",
				Module: true,
			},
			expectErr: true,
			errMsg:    "queue name is required",
		},
		{
			name: "invalid name with spaces",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSQS,
				Name:   "invalid name",
				Module: true,
			},
			expectErr: true,
			errMsg:    "alphanumeric",
		},
		{
			name: "invalid name with special chars",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSQS,
				Name:   "queue@123",
				Module: true,
			},
			expectErr: true,
			errMsg:    "alphanumeric",
		},
		{
			name: "valid name with hyphens",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSQS,
				Name:   "orders-queue-v2",
				Module: true,
			},
			expectErr: false,
		},
		{
			name: "valid name with underscores",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSQS,
				Name:   "orders_queue_v2",
				Module: true,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.Validate(tt.config)

			if tt.expectErr {
				require.True(t, E.IsLeft(result), "Should return error")
				err := extractError(result)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.True(t, E.IsRight(result), "Should succeed")
			}
		})
	}
}

// TestGenerate_ModuleMode tests module-based generation.
func TestGenerate_ModuleMode(t *testing.T) {
	gen := sqs.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSQS,
		Name:   "orders-queue",
		Module: true,
		Variables: map[string]interface{}{
			"visibility_timeout_seconds": 30,
			"message_retention_seconds":  345600,
			"create_dlq":                 true,
		},
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsRight(result), "Generate should succeed")
	code := extractCode(result)

	// Should generate 2 files
	require.Len(t, code.Files, 2)

	// Check sqs.tf
	sqsFile := findFile(code.Files, "sqs.tf")
	require.NotNil(t, sqsFile)
	assert.Equal(t, generators.WriteModeAppend, sqsFile.Mode)
	assert.Contains(t, sqsFile.Content, `module "orders_queue"`)
	assert.Contains(t, sqsFile.Content, `source  = "terraform-aws-modules/sqs/aws"`)
	assert.Contains(t, sqsFile.Content, `version = "~> 4.0"`)
	assert.Contains(t, sqsFile.Content, `name = "${var.namespace}orders-queue"`)
	assert.Contains(t, sqsFile.Content, "visibility_timeout_seconds = 30")
	assert.Contains(t, sqsFile.Content, "message_retention_seconds  = 345600")
	assert.Contains(t, sqsFile.Content, "create_dlq                    = true")
	assert.Contains(t, sqsFile.Content, `dlq_name                      = "${var.namespace}orders-queue-dlq"`)

	// Check outputs.tf
	outputsFile := findFile(code.Files, "outputs.tf")
	require.NotNil(t, outputsFile)
	assert.Equal(t, generators.WriteModeAppend, outputsFile.Mode)
	assert.Contains(t, outputsFile.Content, `output "orders_queue_url"`)
	assert.Contains(t, outputsFile.Content, `output "orders_queue_arn"`)
	assert.Contains(t, outputsFile.Content, "module.orders_queue.queue_url")
	assert.Contains(t, outputsFile.Content, "module.orders_queue.queue_arn")
}

// TestGenerate_RawMode tests raw resource generation.
func TestGenerate_RawMode(t *testing.T) {
	gen := sqs.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSQS,
		Name:   "orders-queue",
		Module: false,
		Variables: map[string]interface{}{
			"visibility_timeout_seconds": 30,
			"message_retention_seconds":  345600,
			"create_dlq":                 true,
		},
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsRight(result), "Generate should succeed")
	code := extractCode(result)

	// Check sqs.tf
	sqsFile := findFile(code.Files, "sqs.tf")
	require.NotNil(t, sqsFile)
	assert.Contains(t, sqsFile.Content, `resource "aws_sqs_queue" "orders_queue"`)
	assert.Contains(t, sqsFile.Content, `name = "${var.namespace}orders-queue"`)
	assert.Contains(t, sqsFile.Content, "visibility_timeout_seconds = 30")
	assert.Contains(t, sqsFile.Content, "message_retention_seconds  = 345600")

	// Check outputs.tf
	outputsFile := findFile(code.Files, "outputs.tf")
	require.NotNil(t, outputsFile)
	assert.Contains(t, outputsFile.Content, "aws_sqs_queue.orders_queue.url")
	assert.Contains(t, outputsFile.Content, "aws_sqs_queue.orders_queue.arn")
}

// TestGenerate_WithIntegration tests Lambda integration generation.
func TestGenerate_WithIntegration(t *testing.T) {
	gen := sqs.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSQS,
		Name:   "orders-queue",
		Module: true,
		Variables: map[string]interface{}{
			"visibility_timeout_seconds": 30,
			"message_retention_seconds":  345600,
			"create_dlq":                 true,
		},
		Integration: &generators.IntegrationConfig{
			TargetFunction: "processor",
			EventSource: &generators.EventSourceConfig{
				ARNExpression:         "module.orders_queue.queue_arn",
				BatchSize:             10,
				MaxBatchingWindowSecs: 5,
				MaxConcurrency:        10,
			},
			IAMPermissions: []generators.IAMPermission{
				{
					Effect: "Allow",
					Actions: []string{
						"sqs:ReceiveMessage",
						"sqs:DeleteMessage",
						"sqs:GetQueueAttributes",
					},
					Resources: []string{"module.orders_queue.queue_arn"},
				},
			},
		},
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsRight(result), "Generate should succeed")
	code := extractCode(result)

	// Should generate 3 files
	require.Len(t, code.Files, 3)

	// Check lambda_processor.tf
	lambdaFile := findFile(code.Files, "lambda_processor.tf")
	require.NotNil(t, lambdaFile)
	assert.Contains(t, lambdaFile.Content, `resource "aws_lambda_event_source_mapping" "processor_orders_queue"`)
	assert.Contains(t, lambdaFile.Content, "event_source_arn = module.orders_queue.queue_arn")
	assert.Contains(t, lambdaFile.Content, "function_name    = aws_lambda_function.processor.arn")
	assert.Contains(t, lambdaFile.Content, "batch_size                         = 10")
	assert.Contains(t, lambdaFile.Content, "maximum_batching_window_in_seconds = 5")
	assert.Contains(t, lambdaFile.Content, "maximum_concurrency = 10")

	// Check IAM policy
	assert.Contains(t, lambdaFile.Content, `resource "aws_iam_role_policy" "processor_sqs_orders_queue"`)
	assert.Contains(t, lambdaFile.Content, `role = aws_iam_role.processor.id`)
	assert.Contains(t, lambdaFile.Content, "sqs:ReceiveMessage")
	assert.Contains(t, lambdaFile.Content, "sqs:DeleteMessage")
	assert.Contains(t, lambdaFile.Content, "sqs:GetQueueAttributes")
}

// TestGenerate_InvalidConfig tests generation with invalid config.
func TestGenerate_InvalidConfig(t *testing.T) {
	gen := sqs.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSQS,
		Name:   "", // Invalid: empty name
		Module: true,
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsLeft(result), "Should fail with invalid config")
	err := extractCodeError(result)
	assert.Contains(t, err.Error(), "queue name is required")
}

// TestSanitizeName tests name sanitization.
func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"orders-queue", "orders_queue"},
		{"my-queue-v2", "my_queue_v2"},
		{"simple", "simple"},
		{"already_underscored", "already_underscored"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Sanitization is tested indirectly through module names
			gen := sqs.New()
			config := generators.ResourceConfig{
				Type:   generators.ResourceSQS,
				Name:   tt.input,
				Module: true,
				Variables: map[string]interface{}{
					"visibility_timeout_seconds": 30,
					"message_retention_seconds":  345600,
					"create_dlq":                 true,
				},
			}

			result := gen.Generate(config, generators.ProjectState{})
			require.True(t, E.IsRight(result))
			code := extractCode(result)

			sqsFile := findFile(code.Files, "sqs.tf")
			require.NotNil(t, sqsFile)

			// Module name should use underscores
			expectedModule := `module "` + tt.expected + `"`
			assert.Contains(t, sqsFile.Content, expectedModule)
		})
	}
}

// TestNamespaceSupport tests namespace variable usage.
func TestNamespaceSupport(t *testing.T) {
	gen := sqs.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSQS,
		Name:   "orders-queue",
		Module: true,
		Variables: map[string]interface{}{
			"visibility_timeout_seconds": 30,
			"message_retention_seconds":  345600,
			"create_dlq":                 true,
		},
	}

	result := gen.Generate(config, generators.ProjectState{})
	require.True(t, E.IsRight(result))
	code := extractCode(result)

	sqsFile := findFile(code.Files, "sqs.tf")
	require.NotNil(t, sqsFile)

	// All resources should use ${var.namespace} prefix
	assert.Contains(t, sqsFile.Content, `name = "${var.namespace}orders-queue"`)
	assert.Contains(t, sqsFile.Content, `dlq_name                      = "${var.namespace}orders-queue-dlq"`)
}

// TestDLQConfiguration tests DLQ settings.
func TestDLQConfiguration(t *testing.T) {
	gen := sqs.New()

	tests := []struct {
		name      string
		createDLQ bool
	}{
		{"with DLQ", true},
		{"without DLQ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := generators.ResourceConfig{
				Type:   generators.ResourceSQS,
				Name:   "orders-queue",
				Module: true,
				Variables: map[string]interface{}{
					"visibility_timeout_seconds": 30,
					"message_retention_seconds":  345600,
					"create_dlq":                 tt.createDLQ,
				},
			}

			result := gen.Generate(config, generators.ProjectState{})
			require.True(t, E.IsRight(result))
			code := extractCode(result)

			sqsFile := findFile(code.Files, "sqs.tf")
			require.NotNil(t, sqsFile)

			if tt.createDLQ {
				assert.Contains(t, sqsFile.Content, "create_dlq                    = true")
				assert.Contains(t, sqsFile.Content, "dlq_name")
				assert.Contains(t, sqsFile.Content, "dlq_message_retention_seconds = 1209600")
			} else {
				assert.NotContains(t, sqsFile.Content, "create_dlq")
				assert.NotContains(t, sqsFile.Content, "dlq_name")
			}
		})
	}
}

// Helper function to find file by path.
func findFile(files []generators.FileToWrite, path string) *generators.FileToWrite {
	for i := range files {
		if files[i].Path == path {
			return &files[i]
		}
	}
	return nil
}

// TestGeneratedCodeFormat tests that generated code is well-formatted.
func TestGeneratedCodeFormat(t *testing.T) {
	gen := sqs.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSQS,
		Name:   "orders-queue",
		Module: true,
		Variables: map[string]interface{}{
			"visibility_timeout_seconds": 30,
			"message_retention_seconds":  345600,
			"create_dlq":                 true,
		},
	}

	result := gen.Generate(config, generators.ProjectState{})
	require.True(t, E.IsRight(result))
	code := extractCode(result)

	for _, file := range code.Files {
		// Check that code starts with comment
		assert.True(t, strings.HasPrefix(file.Content, "#"), "Should start with comment")

		// Check proper indentation (2 spaces)
		lines := strings.Split(file.Content, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "  ") {
				// Lines with indentation should use 2 spaces
				trimmed := strings.TrimLeft(line, " ")
				indent := len(line) - len(trimmed)
				assert.Equal(t, 0, indent%2, "Indentation should be multiple of 2")
			}
		}

		// Check no trailing whitespace
		for i, line := range lines {
			assert.Equal(t, strings.TrimRight(line, " \t"), line,
				"Line %d should not have trailing whitespace", i+1)
		}
	}
}
