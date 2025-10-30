// Package s3 provides S3 bucket generation for forge add s3 command.
// It follows functional programming principles with pure generation logic.
package s3

import (
	"context"
	"errors"
	"fmt"
	"strings"

	E "github.com/IBM/fp-go/either"

	"github.com/lewis/forge/internal/generators"
)

type (
	// Generator implements generators.Generator for S3 buckets.
	Generator struct{}
)

// New creates a new S3 generator.
func New() *Generator {
	return &Generator{}
}

// Prompt gathers configuration from user (I/O ACTION).
func (g *Generator) Prompt(ctx context.Context, intent generators.ResourceIntent, state generators.ProjectState) E.Either[error, generators.ResourceConfig] {
	// For MVP, use sensible defaults
	// In Phase 3, this will launch interactive TUI

	config := generators.ResourceConfig{
		Type:   generators.ResourceS3,
		Name:   intent.Name,
		Module: intent.UseModule,
		Variables: map[string]interface{}{
			"versioning_enabled":        true,
			"block_public_acls":         true,
			"block_public_policy":       true,
			"ignore_public_acls":        true,
			"restrict_public_buckets":   true,
			"force_destroy":             false, // Safety: require manual deletion
			"lifecycle_rules":           []map[string]interface{}{},
			"cors_rules":                []map[string]interface{}{},
			"server_side_encryption":    "AES256", // Default AWS managed encryption
			"enable_event_notification": false,
		},
	}

	// If integrating with Lambda, add integration config
	if intent.ToFunc != "" {
		// Verify target function exists
		if _, exists := state.Functions[intent.ToFunc]; !exists {
			return E.Left[generators.ResourceConfig](
				fmt.Errorf("target function '%s' not found", intent.ToFunc),
			)
		}

		// Enable event notifications for Lambda integration
		config.Variables["enable_event_notification"] = true

		config.Integration = &generators.IntegrationConfig{
			TargetFunction: intent.ToFunc,
			IAMPermissions: []generators.IAMPermission{
				{
					Effect: "Allow",
					Actions: []string{
						"s3:GetObject",
						"s3:ListBucket",
					},
					Resources: []string{
						fmt.Sprintf("module.%s.s3_bucket_arn", sanitizeName(intent.Name)),
						fmt.Sprintf("module.%s.s3_bucket_arn/*", sanitizeName(intent.Name)),
					},
				},
			},
			EnvVars: map[string]string{
				strings.ToUpper(sanitizeName(intent.Name)) + "_BUCKET_NAME": fmt.Sprintf("module.%s.s3_bucket_id", sanitizeName(intent.Name)),
			},
		}
	}

	return E.Right[error](config)
}

// Generate creates Terraform code from configuration (PURE CALCULATION).
func (g *Generator) Generate(config generators.ResourceConfig, state generators.ProjectState) E.Either[error, generators.GeneratedCode] {
	// Validate first, then chain generation - automatic error short-circuiting
	return E.Chain(func(validConfig generators.ResourceConfig) E.Either[error, generators.GeneratedCode] {
		var files []generators.FileToWrite

		// 1. Generate main S3 resource file
		if validConfig.Module {
			files = append(files, generators.FileToWrite{
				Path:    "s3.tf",
				Content: generateModuleCode(validConfig),
				Mode:    generators.WriteModeAppend,
			})
		} else {
			files = append(files, generators.FileToWrite{
				Path:    "s3.tf",
				Content: generateRawResourceCode(validConfig),
				Mode:    generators.WriteModeAppend,
			})
		}

		// 2. Generate outputs
		files = append(files, generators.FileToWrite{
			Path:    "outputs.tf",
			Content: generateOutputs(validConfig),
			Mode:    generators.WriteModeAppend,
		})

		// 3. If integration, update Lambda function file
		if validConfig.Integration != nil {
			lambdaFile := fmt.Sprintf("lambda_%s.tf", validConfig.Integration.TargetFunction)
			files = append(files, generators.FileToWrite{
				Path:    lambdaFile,
				Content: generateIntegrationCode(validConfig),
				Mode:    generators.WriteModeAppend,
			})
		}

		return E.Right[error](generators.GeneratedCode{
			Files: files,
		})
	})(g.Validate(config))
}

// Validate checks if configuration is valid (PURE CALCULATION).
func (g *Generator) Validate(config generators.ResourceConfig) E.Either[error, generators.ResourceConfig] {
	if config.Name == "" {
		return E.Left[generators.ResourceConfig](
			errors.New("bucket name is required"),
		)
	}

	// S3 bucket names have stricter requirements
	if !isValidS3Name(config.Name) {
		return E.Left[generators.ResourceConfig](
			errors.New("bucket name must be lowercase alphanumeric with hyphens, 3-63 characters"),
		)
	}

	return E.Right[error](config)
}

// generateModuleCode creates Terraform module code (PURE).
func generateModuleCode(config generators.ResourceConfig) string {
	moduleName := sanitizeName(config.Name)
	bucketName := config.Name

	versioningEnabled, ok := config.Variables["versioning_enabled"].(bool)
	_ = ok
	blockPublicACLs, ok := config.Variables["block_public_acls"].(bool)
	_ = ok
	forceDestroy, ok := config.Variables["force_destroy"].(bool)
	_ = ok
	encryption, ok := config.Variables["server_side_encryption"].(string)
	_ = ok

	var parts []string

	parts = append(parts, "# Generated by forge add s3 "+config.Name)
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("module \"%s\" {", moduleName))
	parts = append(parts, "  source  = \"terraform-aws-modules/s3-bucket/aws\"")
	parts = append(parts, "  version = \"~> 4.0\"")
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("  bucket = \"${var.namespace}%s\"", bucketName))
	parts = append(parts, fmt.Sprintf("  force_destroy = %t", forceDestroy))

	// Versioning
	if versioningEnabled {
		parts = append(parts, "")
		parts = append(parts, "  # Versioning configuration")
		parts = append(parts, "  versioning = {")
		parts = append(parts, "    enabled = true")
		parts = append(parts, "  }")
	}

	// Public access block
	if blockPublicACLs {
		parts = append(parts, "")
		parts = append(parts, "  # Block all public access")
		parts = append(parts, "  block_public_acls       = true")
		parts = append(parts, "  block_public_policy     = true")
		parts = append(parts, "  ignore_public_acls      = true")
		parts = append(parts, "  restrict_public_buckets = true")
	}

	// Server-side encryption
	parts = append(parts, "")
	parts = append(parts, "  # Server-side encryption")
	parts = append(parts, "  server_side_encryption_configuration = {")
	parts = append(parts, "    rule = {")
	parts = append(parts, "      apply_server_side_encryption_by_default = {")
	parts = append(parts, fmt.Sprintf("        sse_algorithm = \"%s\"", encryption))
	parts = append(parts, "      }")
	parts = append(parts, "    }")
	parts = append(parts, "  }")

	parts = append(parts, "")
	parts = append(parts, "  tags = {")
	parts = append(parts, "    ManagedBy = \"forge\"")
	parts = append(parts, "    Namespace = var.namespace")
	parts = append(parts, "  }")
	parts = append(parts, "}")
	parts = append(parts, "")

	return strings.Join(parts, "\n")
}

// generateRawResourceCode creates raw Terraform resource code (PURE).
func generateRawResourceCode(config generators.ResourceConfig) string {
	resourceName := sanitizeName(config.Name)
	bucketName := config.Name

	versioningEnabled, ok := config.Variables["versioning_enabled"].(bool)
	_ = ok
	forceDestroy, ok := config.Variables["force_destroy"].(bool)
	_ = ok

	var parts []string

	parts = append(parts, "# Generated by forge add s3 "+config.Name+" --raw")
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("resource \"aws_s3_bucket\" \"%s\" {", resourceName))
	parts = append(parts, fmt.Sprintf("  bucket = \"${var.namespace}%s\"", bucketName))
	parts = append(parts, fmt.Sprintf("  force_destroy = %t", forceDestroy))
	parts = append(parts, "")
	parts = append(parts, "  tags = {")
	parts = append(parts, "    ManagedBy = \"forge\"")
	parts = append(parts, "    Namespace = var.namespace")
	parts = append(parts, "  }")
	parts = append(parts, "}")
	parts = append(parts, "")

	// Versioning (separate resource in raw mode)
	if versioningEnabled {
		parts = append(parts, fmt.Sprintf("resource \"aws_s3_bucket_versioning\" \"%s\" {", resourceName))
		parts = append(parts, fmt.Sprintf("  bucket = aws_s3_bucket.%s.id", resourceName))
		parts = append(parts, "")
		parts = append(parts, "  versioning_configuration {")
		parts = append(parts, "    status = \"Enabled\"")
		parts = append(parts, "  }")
		parts = append(parts, "}")
		parts = append(parts, "")
	}

	// Public access block (separate resource in raw mode)
	parts = append(parts, fmt.Sprintf("resource \"aws_s3_bucket_public_access_block\" \"%s\" {", resourceName))
	parts = append(parts, fmt.Sprintf("  bucket = aws_s3_bucket.%s.id", resourceName))
	parts = append(parts, "")
	parts = append(parts, "  block_public_acls       = true")
	parts = append(parts, "  block_public_policy     = true")
	parts = append(parts, "  ignore_public_acls      = true")
	parts = append(parts, "  restrict_public_buckets = true")
	parts = append(parts, "}")
	parts = append(parts, "")

	return strings.Join(parts, "\n")
}

// generateOutputs creates Terraform outputs (PURE).
func generateOutputs(config generators.ResourceConfig) string {
	moduleName := sanitizeName(config.Name)

	var parts []string

	parts = append(parts, "# Outputs for "+config.Name)

	if config.Module {
		parts = append(parts, fmt.Sprintf("output \"%s_bucket_id\" {", moduleName))
		parts = append(parts, fmt.Sprintf("  description = \"ID of %s\"", config.Name))
		parts = append(parts, fmt.Sprintf("  value       = module.%s.s3_bucket_id", moduleName))
		parts = append(parts, "}")
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("output \"%s_bucket_arn\" {", moduleName))
		parts = append(parts, fmt.Sprintf("  description = \"ARN of %s\"", config.Name))
		parts = append(parts, fmt.Sprintf("  value       = module.%s.s3_bucket_arn", moduleName))
		parts = append(parts, "}")
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("output \"%s_bucket_domain_name\" {", moduleName))
		parts = append(parts, fmt.Sprintf("  description = \"Domain name of %s\"", config.Name))
		parts = append(parts, fmt.Sprintf("  value       = module.%s.s3_bucket_bucket_domain_name", moduleName))
		parts = append(parts, "}")
	} else {
		parts = append(parts, fmt.Sprintf("output \"%s_bucket_id\" {", moduleName))
		parts = append(parts, fmt.Sprintf("  description = \"ID of %s\"", config.Name))
		parts = append(parts, fmt.Sprintf("  value       = aws_s3_bucket.%s.id", moduleName))
		parts = append(parts, "}")
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("output \"%s_bucket_arn\" {", moduleName))
		parts = append(parts, fmt.Sprintf("  description = \"ARN of %s\"", config.Name))
		parts = append(parts, fmt.Sprintf("  value       = aws_s3_bucket.%s.arn", moduleName))
		parts = append(parts, "}")
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("output \"%s_bucket_domain_name\" {", moduleName))
		parts = append(parts, fmt.Sprintf("  description = \"Domain name of %s\"", config.Name))
		parts = append(parts, fmt.Sprintf("  value       = aws_s3_bucket.%s.bucket_domain_name", moduleName))
		parts = append(parts, "}")
	}

	parts = append(parts, "")
	return strings.Join(parts, "\n")
}

// generateIntegrationCode creates S3 bucket notification and IAM policy (PURE).
func generateIntegrationCode(config generators.ResourceConfig) string {
	if config.Integration == nil {
		return ""
	}

	bucketName := sanitizeName(config.Name)
	functionName := config.Integration.TargetFunction

	var parts []string

	// Lambda permission for S3 to invoke function
	parts = append(parts, "# Permission for S3 to invoke "+functionName)
	parts = append(parts, fmt.Sprintf("resource \"aws_lambda_permission\" \"%s_s3_%s\" {",
		functionName, bucketName))
	parts = append(parts, "  statement_id  = \"AllowExecutionFromS3Bucket\"")
	parts = append(parts, "  action        = \"lambda:InvokeFunction\"")
	parts = append(parts, fmt.Sprintf("  function_name = aws_lambda_function.%s.arn", functionName))
	parts = append(parts, "  principal     = \"s3.amazonaws.com\"")

	if config.Module {
		parts = append(parts, fmt.Sprintf("  source_arn    = module.%s.s3_bucket_arn", bucketName))
	} else {
		parts = append(parts, fmt.Sprintf("  source_arn    = aws_s3_bucket.%s.arn", bucketName))
	}

	parts = append(parts, "}")
	parts = append(parts, "")

	// S3 bucket notification
	parts = append(parts, "# S3 bucket notification for "+config.Name)
	parts = append(parts, fmt.Sprintf("resource \"aws_s3_bucket_notification\" \"%s\" {", bucketName))

	if config.Module {
		parts = append(parts, fmt.Sprintf("  bucket = module.%s.s3_bucket_id", bucketName))
	} else {
		parts = append(parts, fmt.Sprintf("  bucket = aws_s3_bucket.%s.id", bucketName))
	}

	parts = append(parts, "")
	parts = append(parts, "  lambda_function {")
	parts = append(parts, fmt.Sprintf("    lambda_function_arn = aws_lambda_function.%s.arn", functionName))
	parts = append(parts, "    events              = [\"s3:ObjectCreated:*\"]")
	parts = append(parts, "  }")
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("  depends_on = [aws_lambda_permission.%s_s3_%s]", functionName, bucketName))
	parts = append(parts, "}")
	parts = append(parts, "")

	// IAM policy for Lambda to access S3
	if len(config.Integration.IAMPermissions) > 0 {
		perm := config.Integration.IAMPermissions[0]

		parts = append(parts, fmt.Sprintf("# IAM policy for %s to access %s", functionName, config.Name))
		parts = append(parts, fmt.Sprintf("resource \"aws_iam_role_policy\" \"%s_s3_%s\" {",
			functionName, bucketName))
		parts = append(parts, fmt.Sprintf("  name = \"${var.namespace}%s-s3-%s\"", functionName, bucketName))
		parts = append(parts, fmt.Sprintf("  role = aws_iam_role.%s.id", functionName))
		parts = append(parts, "")
		parts = append(parts, "  policy = jsonencode({")
		parts = append(parts, "    Version = \"2012-10-17\"")
		parts = append(parts, "    Statement = [")
		parts = append(parts, "      {")
		parts = append(parts, fmt.Sprintf("        Effect = \"%s\"", perm.Effect))
		parts = append(parts, "        Action = [")

		for i, action := range perm.Actions {
			comma := ","
			if i == len(perm.Actions)-1 {
				comma = ""
			}
			parts = append(parts, fmt.Sprintf("          \"%s\"%s", action, comma))
		}

		parts = append(parts, "        ]")
		parts = append(parts, "        Resource = [")

		for i, resource := range perm.Resources {
			comma := ","
			if i == len(perm.Resources)-1 {
				comma = ""
			}
			parts = append(parts, fmt.Sprintf("          %s%s", resource, comma))
		}

		parts = append(parts, "        ]")
		parts = append(parts, "      }")
		parts = append(parts, "    ]")
		parts = append(parts, "  })}")
		parts = append(parts, "")
	}

	// Environment variables
	if len(config.Integration.EnvVars) > 0 {
		parts = append(parts, fmt.Sprintf("# Note: Add these environment variables to lambda_%s.tf:", functionName))
		for key, value := range config.Integration.EnvVars {
			parts = append(parts, fmt.Sprintf("#   %s = %s", key, value))
		}
		parts = append(parts, "")
	}

	return strings.Join(parts, "\n")
}

// sanitizeName converts a name to a valid Terraform identifier (PURE).
func sanitizeName(name string) string {
	// Replace hyphens with underscores for Terraform identifiers
	return strings.ReplaceAll(name, "-", "_")
}

// isValidS3Name checks if a name is valid for S3 buckets (PURE).
func isValidS3Name(name string) bool {
	// S3 bucket naming rules:
	// - 3-63 characters
	// - lowercase letters, numbers, hyphens
	// - must start with letter or number
	// - cannot end with hyphen
	// - cannot have consecutive hyphens

	if len(name) < 3 || len(name) > 63 {
		return false
	}

	// Must start with letter or number
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= '0' && first <= '9')) {
		return false
	}

	// Cannot end with hyphen
	if name[len(name)-1] == '-' {
		return false
	}

	// Check each character and consecutive hyphens
	prevHyphen := false
	for _, r := range name {
		if r == '-' {
			if prevHyphen {
				return false // consecutive hyphens
			}
			prevHyphen = true
		} else {
			prevHyphen = false
		}

		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}

	return true
}
