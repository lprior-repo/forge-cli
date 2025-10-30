package s3_test

import (
	"strings"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lewis/forge/internal/generators"
	"github.com/lewis/forge/internal/generators/s3"
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
	gen := s3.New()
	assert.NotNil(t, gen)
}

// TestPrompt_Standalone tests standalone bucket configuration.
func TestPrompt_Standalone(t *testing.T) {
	gen := s3.New()
	ctx := t.Context()

	intent := generators.ResourceIntent{
		Type:      generators.ResourceS3,
		Name:      "uploads-bucket",
		UseModule: true,
	}

	state := generators.ProjectState{
		Functions: make(map[string]generators.FunctionInfo),
	}

	result := gen.Prompt(ctx, intent, state)

	require.True(t, E.IsRight(result), "Prompt should succeed")
	config := extractConfig(result)

	assert.Equal(t, generators.ResourceS3, config.Type)
	assert.Equal(t, "uploads-bucket", config.Name)
	assert.True(t, config.Module)
	assert.Nil(t, config.Integration, "Standalone bucket should have no integration")

	// Verify default variables
	assert.Equal(t, true, config.Variables["versioning_enabled"])
	assert.Equal(t, true, config.Variables["block_public_acls"])
	assert.Equal(t, true, config.Variables["block_public_policy"])
	assert.Equal(t, true, config.Variables["ignore_public_acls"])
	assert.Equal(t, true, config.Variables["restrict_public_buckets"])
	assert.Equal(t, false, config.Variables["force_destroy"])
	assert.Equal(t, "AES256", config.Variables["server_side_encryption"])
	assert.Equal(t, false, config.Variables["enable_event_notification"])
}

// TestPrompt_WithLambdaIntegration tests Lambda integration.
func TestPrompt_WithLambdaIntegration(t *testing.T) {
	gen := s3.New()
	ctx := t.Context()

	intent := generators.ResourceIntent{
		Type:      generators.ResourceS3,
		Name:      "uploads-bucket",
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

	// Event notifications should be enabled
	assert.Equal(t, true, config.Variables["enable_event_notification"])

	// Verify IAM permissions
	require.Len(t, config.Integration.IAMPermissions, 1)
	perm := config.Integration.IAMPermissions[0]
	assert.Equal(t, "Allow", perm.Effect)
	assert.Contains(t, perm.Actions, "s3:GetObject")
	assert.Contains(t, perm.Actions, "s3:ListBucket")
	assert.Contains(t, perm.Resources, "module.uploads_bucket.s3_bucket_arn")
	assert.Contains(t, perm.Resources, "module.uploads_bucket.s3_bucket_arn/*")

	// Verify environment variables
	assert.NotNil(t, config.Integration.EnvVars)
	assert.Contains(t, config.Integration.EnvVars, "UPLOADS_BUCKET_BUCKET_NAME")
	assert.Equal(t, "module.uploads_bucket.s3_bucket_id", config.Integration.EnvVars["UPLOADS_BUCKET_BUCKET_NAME"])
}

// TestPrompt_FunctionNotFound tests error when target function doesn't exist.
func TestPrompt_FunctionNotFound(t *testing.T) {
	gen := s3.New()
	ctx := t.Context()

	intent := generators.ResourceIntent{
		Type:      generators.ResourceS3,
		Name:      "uploads-bucket",
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
	gen := s3.New()

	tests := []struct {
		name      string
		config    generators.ResourceConfig
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid configuration",
			config: generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "uploads-bucket",
				Module: true,
			},
			expectErr: false,
		},
		{
			name: "missing name",
			config: generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "",
				Module: true,
			},
			expectErr: true,
			errMsg:    "bucket name is required",
		},
		{
			name: "invalid name - uppercase",
			config: generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "MyBucket",
				Module: true,
			},
			expectErr: true,
			errMsg:    "lowercase",
		},
		{
			name: "invalid name - too short",
			config: generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "ab",
				Module: true,
			},
			expectErr: true,
			errMsg:    "3-63 characters",
		},
		{
			name: "invalid name - too long",
			config: generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   strings.Repeat("a", 64),
				Module: true,
			},
			expectErr: true,
			errMsg:    "3-63 characters",
		},
		{
			name: "invalid name - starts with hyphen",
			config: generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "-bucket",
				Module: true,
			},
			expectErr: true,
			errMsg:    "lowercase",
		},
		{
			name: "invalid name - ends with hyphen",
			config: generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "bucket-",
				Module: true,
			},
			expectErr: true,
			errMsg:    "lowercase",
		},
		{
			name: "invalid name - consecutive hyphens",
			config: generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "my--bucket",
				Module: true,
			},
			expectErr: true,
			errMsg:    "lowercase",
		},
		{
			name: "valid name with hyphens",
			config: generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "uploads-bucket-v2",
				Module: true,
			},
			expectErr: false,
		},
		{
			name: "valid name with numbers",
			config: generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "bucket123",
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
	gen := s3.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceS3,
		Name:   "uploads-bucket",
		Module: true,
		Variables: map[string]interface{}{
			"versioning_enabled":        true,
			"block_public_acls":         true,
			"block_public_policy":       true,
			"ignore_public_acls":        true,
			"restrict_public_buckets":   true,
			"force_destroy":             false,
			"lifecycle_rules":           []map[string]interface{}{},
			"cors_rules":                []map[string]interface{}{},
			"server_side_encryption":    "AES256",
			"enable_event_notification": false,
		},
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsRight(result), "Generate should succeed")
	code := extractCode(result)

	// Should generate 2 files
	require.Len(t, code.Files, 2)

	// Check s3.tf
	s3File := findFile(code.Files, "s3.tf")
	require.NotNil(t, s3File)
	assert.Equal(t, generators.WriteModeAppend, s3File.Mode)
	assert.Contains(t, s3File.Content, `module "uploads_bucket"`)
	assert.Contains(t, s3File.Content, `source  = "terraform-aws-modules/s3-bucket/aws"`)
	assert.Contains(t, s3File.Content, `version = "~> 4.0"`)
	assert.Contains(t, s3File.Content, `bucket = "${var.namespace}uploads-bucket"`)
	assert.Contains(t, s3File.Content, "force_destroy = false")
	assert.Contains(t, s3File.Content, "enabled = true") // versioning
	assert.Contains(t, s3File.Content, "block_public_acls")
	assert.Contains(t, s3File.Content, "sse_algorithm = \"AES256\"")

	// Check outputs.tf
	outputsFile := findFile(code.Files, "outputs.tf")
	require.NotNil(t, outputsFile)
	assert.Equal(t, generators.WriteModeAppend, outputsFile.Mode)
	assert.Contains(t, outputsFile.Content, `output "uploads_bucket_bucket_id"`)
	assert.Contains(t, outputsFile.Content, `output "uploads_bucket_bucket_arn"`)
	assert.Contains(t, outputsFile.Content, `output "uploads_bucket_bucket_domain_name"`)
	assert.Contains(t, outputsFile.Content, "module.uploads_bucket.s3_bucket_id")
	assert.Contains(t, outputsFile.Content, "module.uploads_bucket.s3_bucket_arn")
}

// TestGenerate_RawMode tests raw resource generation.
func TestGenerate_RawMode(t *testing.T) {
	gen := s3.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceS3,
		Name:   "uploads-bucket",
		Module: false,
		Variables: map[string]interface{}{
			"versioning_enabled":        true,
			"block_public_acls":         true,
			"block_public_policy":       true,
			"ignore_public_acls":        true,
			"restrict_public_buckets":   true,
			"force_destroy":             false,
			"lifecycle_rules":           []map[string]interface{}{},
			"cors_rules":                []map[string]interface{}{},
			"server_side_encryption":    "AES256",
			"enable_event_notification": false,
		},
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsRight(result), "Generate should succeed")
	code := extractCode(result)

	// Check s3.tf
	s3File := findFile(code.Files, "s3.tf")
	require.NotNil(t, s3File)
	assert.Contains(t, s3File.Content, `resource "aws_s3_bucket" "uploads_bucket"`)
	assert.Contains(t, s3File.Content, `bucket = "${var.namespace}uploads-bucket"`)
	assert.Contains(t, s3File.Content, "force_destroy = false")

	// Raw mode should have separate versioning resource
	assert.Contains(t, s3File.Content, `resource "aws_s3_bucket_versioning" "uploads_bucket"`)
	assert.Contains(t, s3File.Content, "status = \"Enabled\"")

	// Raw mode should have separate public access block resource
	assert.Contains(t, s3File.Content, `resource "aws_s3_bucket_public_access_block" "uploads_bucket"`)

	// Check outputs.tf
	outputsFile := findFile(code.Files, "outputs.tf")
	require.NotNil(t, outputsFile)
	assert.Contains(t, outputsFile.Content, "aws_s3_bucket.uploads_bucket.id")
	assert.Contains(t, outputsFile.Content, "aws_s3_bucket.uploads_bucket.arn")
}

// TestGenerate_WithIntegration tests Lambda integration generation.
func TestGenerate_WithIntegration(t *testing.T) {
	gen := s3.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceS3,
		Name:   "uploads-bucket",
		Module: true,
		Variables: map[string]interface{}{
			"versioning_enabled":        true,
			"block_public_acls":         true,
			"block_public_policy":       true,
			"ignore_public_acls":        true,
			"restrict_public_buckets":   true,
			"force_destroy":             false,
			"lifecycle_rules":           []map[string]interface{}{},
			"cors_rules":                []map[string]interface{}{},
			"server_side_encryption":    "AES256",
			"enable_event_notification": true,
		},
		Integration: &generators.IntegrationConfig{
			TargetFunction: "processor",
			IAMPermissions: []generators.IAMPermission{
				{
					Effect: "Allow",
					Actions: []string{
						"s3:GetObject",
						"s3:ListBucket",
					},
					Resources: []string{
						"module.uploads_bucket.s3_bucket_arn",
						"module.uploads_bucket.s3_bucket_arn/*",
					},
				},
			},
			EnvVars: map[string]string{
				"UPLOADS_BUCKET_BUCKET_NAME": "module.uploads_bucket.s3_bucket_id",
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

	// Check Lambda permission
	assert.Contains(t, lambdaFile.Content, `resource "aws_lambda_permission" "processor_s3_uploads_bucket"`)
	assert.Contains(t, lambdaFile.Content, "AllowExecutionFromS3Bucket")
	assert.Contains(t, lambdaFile.Content, "principal     = \"s3.amazonaws.com\"")
	assert.Contains(t, lambdaFile.Content, "source_arn    = module.uploads_bucket.s3_bucket_arn")

	// Check S3 bucket notification
	assert.Contains(t, lambdaFile.Content, `resource "aws_s3_bucket_notification" "uploads_bucket"`)
	assert.Contains(t, lambdaFile.Content, "bucket = module.uploads_bucket.s3_bucket_id")
	assert.Contains(t, lambdaFile.Content, "lambda_function_arn = aws_lambda_function.processor.arn")
	assert.Contains(t, lambdaFile.Content, "events              = [\"s3:ObjectCreated:*\"]")
	assert.Contains(t, lambdaFile.Content, "depends_on = [aws_lambda_permission.processor_s3_uploads_bucket]")

	// Check IAM policy
	assert.Contains(t, lambdaFile.Content, `resource "aws_iam_role_policy" "processor_s3_uploads_bucket"`)
	assert.Contains(t, lambdaFile.Content, "s3:GetObject")
	assert.Contains(t, lambdaFile.Content, "s3:ListBucket")

	// Check env vars note
	assert.Contains(t, lambdaFile.Content, "UPLOADS_BUCKET_BUCKET_NAME")
}

// TestGenerate_InvalidConfig tests generation with invalid config.
func TestGenerate_InvalidConfig(t *testing.T) {
	gen := s3.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceS3,
		Name:   "", // Invalid: empty name
		Module: true,
	}

	state := generators.ProjectState{}

	result := gen.Generate(config, state)

	require.True(t, E.IsLeft(result), "Should fail with invalid config")
	err := extractCodeError(result)
	assert.Contains(t, err.Error(), "bucket name is required")
}

// TestSanitizeName tests name sanitization.
func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"uploads-bucket", "uploads_bucket"},
		{"my-bucket-v2", "my_bucket_v2"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Sanitization is tested indirectly through module names
			gen := s3.New()
			config := generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   tt.input,
				Module: true,
				Variables: map[string]interface{}{
					"versioning_enabled":        true,
					"block_public_acls":         true,
					"block_public_policy":       true,
					"ignore_public_acls":        true,
					"restrict_public_buckets":   true,
					"force_destroy":             false,
					"lifecycle_rules":           []map[string]interface{}{},
					"cors_rules":                []map[string]interface{}{},
					"server_side_encryption":    "AES256",
					"enable_event_notification": false,
				},
			}

			result := gen.Generate(config, generators.ProjectState{})
			require.True(t, E.IsRight(result))
			code := extractCode(result)

			s3File := findFile(code.Files, "s3.tf")
			require.NotNil(t, s3File)

			// Module name should use underscores
			expectedModule := `module "` + tt.expected + `"`
			assert.Contains(t, s3File.Content, expectedModule)
		})
	}
}

// TestNamespaceSupport tests namespace variable usage.
func TestNamespaceSupport(t *testing.T) {
	gen := s3.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceS3,
		Name:   "uploads-bucket",
		Module: true,
		Variables: map[string]interface{}{
			"versioning_enabled":        true,
			"block_public_acls":         true,
			"block_public_policy":       true,
			"ignore_public_acls":        true,
			"restrict_public_buckets":   true,
			"force_destroy":             false,
			"lifecycle_rules":           []map[string]interface{}{},
			"cors_rules":                []map[string]interface{}{},
			"server_side_encryption":    "AES256",
			"enable_event_notification": false,
		},
	}

	result := gen.Generate(config, generators.ProjectState{})
	require.True(t, E.IsRight(result))
	code := extractCode(result)

	s3File := findFile(code.Files, "s3.tf")
	require.NotNil(t, s3File)

	// All resources should use ${var.namespace} prefix
	assert.Contains(t, s3File.Content, `bucket = "${var.namespace}uploads-bucket"`)
}

// TestGeneratedCodeFormat tests that generated code is well-formatted.
func TestGeneratedCodeFormat(t *testing.T) {
	gen := s3.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceS3,
		Name:   "uploads-bucket",
		Module: true,
		Variables: map[string]interface{}{
			"versioning_enabled":        true,
			"block_public_acls":         true,
			"block_public_policy":       true,
			"ignore_public_acls":        true,
			"restrict_public_buckets":   true,
			"force_destroy":             false,
			"lifecycle_rules":           []map[string]interface{}{},
			"cors_rules":                []map[string]interface{}{},
			"server_side_encryption":    "AES256",
			"enable_event_notification": false,
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
	gen := s3.New()

	config := generators.ResourceConfig{
		Type:   generators.ResourceS3,
		Name:   "uploads-bucket",
		Module: false,
		Variables: map[string]interface{}{
			"versioning_enabled":        true,
			"block_public_acls":         true,
			"block_public_policy":       true,
			"ignore_public_acls":        true,
			"restrict_public_buckets":   true,
			"force_destroy":             false,
			"lifecycle_rules":           []map[string]interface{}{},
			"cors_rules":                []map[string]interface{}{},
			"server_side_encryption":    "AES256",
			"enable_event_notification": true,
		},
		Integration: &generators.IntegrationConfig{
			TargetFunction: "processor",
			IAMPermissions: []generators.IAMPermission{
				{
					Effect: "Allow",
					Actions: []string{
						"s3:GetObject",
						"s3:ListBucket",
					},
					Resources: []string{
						"aws_s3_bucket.uploads_bucket.arn",
						"aws_s3_bucket.uploads_bucket.arn/*",
					},
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
	assert.Contains(t, lambdaFile.Content, "source_arn    = aws_s3_bucket.uploads_bucket.arn")
	assert.Contains(t, lambdaFile.Content, "bucket = aws_s3_bucket.uploads_bucket.id")
}

// TestForceDestroyConfiguration tests force_destroy settings.
func TestForceDestroyConfiguration(t *testing.T) {
	gen := s3.New()

	tests := []struct {
		name         string
		forceDestroy bool
	}{
		{"with force_destroy", true},
		{"without force_destroy", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "test-bucket",
				Module: true,
				Variables: map[string]interface{}{
					"versioning_enabled":        true,
					"block_public_acls":         true,
					"block_public_policy":       true,
					"ignore_public_acls":        true,
					"restrict_public_buckets":   true,
					"force_destroy":             tt.forceDestroy,
					"lifecycle_rules":           []map[string]interface{}{},
					"cors_rules":                []map[string]interface{}{},
					"server_side_encryption":    "AES256",
					"enable_event_notification": false,
				},
			}

			result := gen.Generate(config, generators.ProjectState{})
			require.True(t, E.IsRight(result))
			code := extractCode(result)

			s3File := findFile(code.Files, "s3.tf")
			require.NotNil(t, s3File)

			if tt.forceDestroy {
				assert.Contains(t, s3File.Content, "force_destroy = true")
			} else {
				assert.Contains(t, s3File.Content, "force_destroy = false")
			}
		})
	}
}

// TestVersioningConfiguration tests versioning settings.
func TestVersioningConfiguration(t *testing.T) {
	gen := s3.New()

	tests := []struct {
		name              string
		versioningEnabled bool
		expectedInModule  string
		expectedInRaw     bool
	}{
		{
			name:              "with versioning",
			versioningEnabled: true,
			expectedInModule:  "enabled = true",
			expectedInRaw:     true,
		},
		{
			name:              "without versioning",
			versioningEnabled: false,
			expectedInModule:  "",
			expectedInRaw:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" (module)", func(t *testing.T) {
			config := generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "test-bucket",
				Module: true,
				Variables: map[string]interface{}{
					"versioning_enabled":        tt.versioningEnabled,
					"block_public_acls":         true,
					"block_public_policy":       true,
					"ignore_public_acls":        true,
					"restrict_public_buckets":   true,
					"force_destroy":             false,
					"lifecycle_rules":           []map[string]interface{}{},
					"cors_rules":                []map[string]interface{}{},
					"server_side_encryption":    "AES256",
					"enable_event_notification": false,
				},
			}

			result := gen.Generate(config, generators.ProjectState{})
			require.True(t, E.IsRight(result))
			code := extractCode(result)

			s3File := findFile(code.Files, "s3.tf")
			require.NotNil(t, s3File)

			if tt.expectedInModule != "" {
				assert.Contains(t, s3File.Content, tt.expectedInModule)
			}
		})

		t.Run(tt.name+" (raw)", func(t *testing.T) {
			config := generators.ResourceConfig{
				Type:   generators.ResourceS3,
				Name:   "test-bucket",
				Module: false,
				Variables: map[string]interface{}{
					"versioning_enabled":        tt.versioningEnabled,
					"block_public_acls":         true,
					"block_public_policy":       true,
					"ignore_public_acls":        true,
					"restrict_public_buckets":   true,
					"force_destroy":             false,
					"lifecycle_rules":           []map[string]interface{}{},
					"cors_rules":                []map[string]interface{}{},
					"server_side_encryption":    "AES256",
					"enable_event_notification": false,
				},
			}

			result := gen.Generate(config, generators.ProjectState{})
			require.True(t, E.IsRight(result))
			code := extractCode(result)

			s3File := findFile(code.Files, "s3.tf")
			require.NotNil(t, s3File)

			if tt.expectedInRaw {
				assert.Contains(t, s3File.Content, "aws_s3_bucket_versioning")
			}
		})
	}
}
