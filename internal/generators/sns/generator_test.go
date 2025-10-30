package sns_test

import (
	"strings"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lewis/forge/internal/generators"
	"github.com/lewis/forge/internal/generators/sns"
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

// Helper function to find file by path.
func findFile(files []generators.FileToWrite, path string) *generators.FileToWrite {
	for i := range files {
		if files[i].Path == path {
			return &files[i]
		}
	}
	return nil
}

// TestNew verifies generator creation.
func TestNew(t *testing.T) {
	gen := sns.New()
	assert.NotNil(t, gen)
}

// TestPrompt_Standalone tests standalone topic configuration.
func TestPrompt_Standalone(t *testing.T) {
	gen := sns.New()
	ctx := t.Context()

	intent := generators.ResourceIntent{
		Type:      generators.ResourceSNS,
		Name:      "notifications",
		UseModule: true,
	}

	state := generators.ProjectState{
		Functions: make(map[string]generators.FunctionInfo),
	}

	result := gen.Prompt(ctx, intent, state)

	require.True(t, E.IsRight(result), "Prompt should succeed")
	config := extractConfig(result)

	assert.Equal(t, generators.ResourceSNS, config.Type)
	assert.Equal(t, "notifications", config.Name)
	assert.True(t, config.Module)
	assert.Nil(t, config.Integration, "Standalone topic should have no integration")

	// Verify default variables
	assert.Equal(t, "notifications", config.Variables["display_name"])
	assert.Equal(t, false, config.Variables["fifo_topic"])
	assert.Equal(t, false, config.Variables["content_based_deduplication"])
	assert.Equal(t, "", config.Variables["kms_master_key_id"])
}

// TestPrompt_WithLambdaIntegration tests Lambda integration.
func TestPrompt_WithLambdaIntegration(t *testing.T) {
	gen := sns.New()
	ctx := t.Context()

	intent := generators.ResourceIntent{
		Type:      generators.ResourceSNS,
		Name:      "notifications",
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

	// Verify IAM permissions
	require.Len(t, config.Integration.IAMPermissions, 1)
	perm := config.Integration.IAMPermissions[0]
	assert.Equal(t, "Allow", perm.Effect)
	assert.Contains(t, perm.Actions, "sns:Publish")
	assert.Contains(t, perm.Resources, "module.notifications.topic_arn")

	// Verify environment variables
	assert.NotNil(t, config.Integration.EnvVars)
	assert.Contains(t, config.Integration.EnvVars, "NOTIFICATIONS_TOPIC_ARN")
	assert.Equal(t, "module.notifications.topic_arn", config.Integration.EnvVars["NOTIFICATIONS_TOPIC_ARN"])
}

// TestPrompt_FunctionNotFound tests error when target function doesn't exist.
func TestPrompt_FunctionNotFound(t *testing.T) {
	gen := sns.New()
	ctx := t.Context()

	intent := generators.ResourceIntent{
		Type:      generators.ResourceSNS,
		Name:      "notifications",
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
	gen := sns.New()

	tests := []struct {
		name      string
		config    generators.ResourceConfig
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid configuration",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSNS,
				Name:   "notifications",
				Module: true,
			},
			expectErr: false,
		},
		{
			name: "missing name",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSNS,
				Name:   "",
				Module: true,
			},
			expectErr: true,
			errMsg:    "topic name is required",
		},
		{
			name: "invalid name with spaces",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSNS,
				Name:   "invalid name",
				Module: true,
			},
			expectErr: true,
			errMsg:    "alphanumeric",
		},
		{
			name: "invalid name with special chars",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSNS,
				Name:   "topic@123",
				Module: true,
			},
			expectErr: true,
			errMsg:    "alphanumeric",
		},
		{
			name: "valid name with hyphens",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSNS,
				Name:   "notifications-v2",
				Module: true,
			},
			expectErr: false,
		},
		{
			name: "valid name with underscores",
			config: generators.ResourceConfig{
				Type:   generators.ResourceSNS,
				Name:   "notifications_v2",
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
	gen := sns.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSNS,
		Name:   "notifications",
		Module: true,
		Variables: map[string]interface{}{
			"display_name":                "notifications",
			"fifo_topic":                  false,
			"content_based_deduplication": false,
			"kms_master_key_id":           "",
			"delivery_policy":             "",
			"create_topic_policy":         false,
		},
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsRight(result), "Generate should succeed")
	code := extractCode(result)

	// Should generate 2 files
	require.Len(t, code.Files, 2)

	// Check sns.tf
	snsFile := findFile(code.Files, "sns.tf")
	require.NotNil(t, snsFile)
	assert.Equal(t, generators.WriteModeAppend, snsFile.Mode)
	assert.Contains(t, snsFile.Content, `module "notifications"`)
	assert.Contains(t, snsFile.Content, `source  = "terraform-aws-modules/sns/aws"`)
	assert.Contains(t, snsFile.Content, `version = "~> 6.0"`)
	assert.Contains(t, snsFile.Content, `name = "${var.namespace}notifications"`)
	assert.Contains(t, snsFile.Content, `display_name = "notifications"`)

	// Check outputs.tf
	outputsFile := findFile(code.Files, "outputs.tf")
	require.NotNil(t, outputsFile)
	assert.Equal(t, generators.WriteModeAppend, outputsFile.Mode)
	assert.Contains(t, outputsFile.Content, `output "notifications_topic_arn"`)
	assert.Contains(t, outputsFile.Content, `output "notifications_topic_id"`)
	assert.Contains(t, outputsFile.Content, "module.notifications.topic_arn")
	assert.Contains(t, outputsFile.Content, "module.notifications.topic_id")
}

// TestGenerate_RawMode tests raw resource generation.
func TestGenerate_RawMode(t *testing.T) {
	gen := sns.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSNS,
		Name:   "notifications",
		Module: false,
		Variables: map[string]interface{}{
			"display_name":                "notifications",
			"fifo_topic":                  false,
			"content_based_deduplication": false,
			"kms_master_key_id":           "",
			"delivery_policy":             "",
			"create_topic_policy":         false,
		},
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsRight(result), "Generate should succeed")
	code := extractCode(result)

	// Check sns.tf
	snsFile := findFile(code.Files, "sns.tf")
	require.NotNil(t, snsFile)
	assert.Contains(t, snsFile.Content, `resource "aws_sns_topic" "notifications"`)
	assert.Contains(t, snsFile.Content, `name = "${var.namespace}notifications"`)
	assert.Contains(t, snsFile.Content, `display_name = "notifications"`)

	// Check outputs.tf
	outputsFile := findFile(code.Files, "outputs.tf")
	require.NotNil(t, outputsFile)
	assert.Contains(t, outputsFile.Content, "aws_sns_topic.notifications.arn")
	assert.Contains(t, outputsFile.Content, "aws_sns_topic.notifications.id")
}

// TestGenerate_FIFOTopic tests FIFO topic generation.
func TestGenerate_FIFOTopic(t *testing.T) {
	gen := sns.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSNS,
		Name:   "orders",
		Module: true,
		Variables: map[string]interface{}{
			"display_name":                "orders",
			"fifo_topic":                  true,
			"content_based_deduplication": true,
			"kms_master_key_id":           "",
			"delivery_policy":             "",
			"create_topic_policy":         false,
		},
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsRight(result), "Generate should succeed")
	code := extractCode(result)

	snsFile := findFile(code.Files, "sns.tf")
	require.NotNil(t, snsFile)
	assert.Contains(t, snsFile.Content, "fifo_topic                  = true")
	assert.Contains(t, snsFile.Content, "content_based_deduplication = true")
}

// TestGenerate_WithIntegration tests Lambda integration generation.
func TestGenerate_WithIntegration(t *testing.T) {
	gen := sns.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSNS,
		Name:   "notifications",
		Module: true,
		Variables: map[string]interface{}{
			"display_name":                "notifications",
			"fifo_topic":                  false,
			"content_based_deduplication": false,
			"kms_master_key_id":           "",
			"delivery_policy":             "",
			"create_topic_policy":         false,
		},
		Integration: &generators.IntegrationConfig{
			TargetFunction: "processor",
			IAMPermissions: []generators.IAMPermission{
				{
					Effect: "Allow",
					Actions: []string{
						"sns:Publish",
					},
					Resources: []string{"module.notifications.topic_arn"},
				},
			},
			EnvVars: map[string]string{
				"NOTIFICATIONS_TOPIC_ARN": "module.notifications.topic_arn",
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
	assert.Contains(t, lambdaFile.Content, `resource "aws_sns_topic_subscription" "notifications_processor"`)
	assert.Contains(t, lambdaFile.Content, "topic_arn = module.notifications.topic_arn")
	assert.Contains(t, lambdaFile.Content, "protocol  = \"lambda\"")
	assert.Contains(t, lambdaFile.Content, "endpoint  = aws_lambda_function.processor.arn")

	// Check Lambda permission
	assert.Contains(t, lambdaFile.Content, `resource "aws_lambda_permission" "processor_sns_notifications"`)
	assert.Contains(t, lambdaFile.Content, "AllowExecutionFromSNS")
	assert.Contains(t, lambdaFile.Content, "principal     = \"sns.amazonaws.com\"")

	// Check IAM policy
	assert.Contains(t, lambdaFile.Content, `resource "aws_iam_role_policy" "processor_sns_notifications_publish"`)
	assert.Contains(t, lambdaFile.Content, "sns:Publish")

	// Check env vars note
	assert.Contains(t, lambdaFile.Content, "NOTIFICATIONS_TOPIC_ARN")
}

// TestGenerate_InvalidConfig tests generation with invalid config.
func TestGenerate_InvalidConfig(t *testing.T) {
	gen := sns.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSNS,
		Name:   "", // Invalid: empty name
		Module: true,
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsLeft(result), "Should fail with invalid config")
	err := extractCodeError(result)
	assert.Contains(t, err.Error(), "topic name is required")
}

// TestSanitizeName tests name sanitization.
func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"notifications", "notifications"},
		{"my-topic-v2", "my_topic_v2"},
		{"simple", "simple"},
		{"already_underscored", "already_underscored"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Sanitization is tested indirectly through module names
			gen := sns.New()
			config := generators.ResourceConfig{
				Type:   generators.ResourceSNS,
				Name:   tt.input,
				Module: true,
				Variables: map[string]interface{}{
					"display_name":                tt.input,
					"fifo_topic":                  false,
					"content_based_deduplication": false,
					"kms_master_key_id":           "",
					"delivery_policy":             "",
					"create_topic_policy":         false,
				},
			}

			result := gen.Generate(config, generators.ProjectState{})
			require.True(t, E.IsRight(result))
			code := extractCode(result)

			snsFile := findFile(code.Files, "sns.tf")
			require.NotNil(t, snsFile)

			// Module name should use underscores
			expectedModule := `module "` + tt.expected + `"`
			assert.Contains(t, snsFile.Content, expectedModule)
		})
	}
}

// TestNamespaceSupport tests namespace variable usage.
func TestNamespaceSupport(t *testing.T) {
	gen := sns.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSNS,
		Name:   "notifications",
		Module: true,
		Variables: map[string]interface{}{
			"display_name":                "notifications",
			"fifo_topic":                  false,
			"content_based_deduplication": false,
			"kms_master_key_id":           "",
			"delivery_policy":             "",
			"create_topic_policy":         false,
		},
	}

	result := gen.Generate(config, generators.ProjectState{})
	require.True(t, E.IsRight(result))
	code := extractCode(result)

	snsFile := findFile(code.Files, "sns.tf")
	require.NotNil(t, snsFile)

	// All resources should use ${var.namespace} prefix
	assert.Contains(t, snsFile.Content, `name = "${var.namespace}notifications"`)
}

// TestGeneratedCodeFormat tests that generated code is well-formatted.
func TestGeneratedCodeFormat(t *testing.T) {
	gen := sns.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSNS,
		Name:   "notifications",
		Module: true,
		Variables: map[string]interface{}{
			"display_name":                "notifications",
			"fifo_topic":                  false,
			"content_based_deduplication": false,
			"kms_master_key_id":           "",
			"delivery_policy":             "",
			"create_topic_policy":         false,
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

// TestRawResourceWithIntegration tests raw resource mode with Lambda integration.
func TestRawResourceWithIntegration(t *testing.T) {
	gen := sns.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceSNS,
		Name:   "notifications",
		Module: false,
		Variables: map[string]interface{}{
			"display_name":                "notifications",
			"fifo_topic":                  false,
			"content_based_deduplication": false,
			"kms_master_key_id":           "",
			"delivery_policy":             "",
			"create_topic_policy":         false,
		},
		Integration: &generators.IntegrationConfig{
			TargetFunction: "processor",
			IAMPermissions: []generators.IAMPermission{
				{
					Effect: "Allow",
					Actions: []string{
						"sns:Publish",
					},
					Resources: []string{"aws_sns_topic.notifications.arn"},
				},
			},
		},
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsRight(result), "Generate should succeed")
	code := extractCode(result)

	lambdaFile := findFile(code.Files, "lambda_processor.tf")
	require.NotNil(t, lambdaFile)

	// In raw mode, should reference raw resource instead of module
	assert.Contains(t, lambdaFile.Content, "topic_arn = aws_sns_topic.notifications.arn")
	assert.Contains(t, lambdaFile.Content, "source_arn    = aws_sns_topic.notifications.arn")
}
