// Package dynamodb provides DynamoDB table generation for forge add dynamodb command.
// It follows functional programming principles with pure generation logic.
package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	E "github.com/IBM/fp-go/either"

	"github.com/lewis/forge/internal/generators"
)

const (
	// Default batch size for DynamoDB stream processing.
	defaultBatchSize = 100
	// Default max concurrency for Lambda event source mapping.
	defaultMaxConcurrency = 10
)

type (
	// Generator implements generators.Generator for DynamoDB tables.
	Generator struct{}
)

// New creates a new DynamoDB generator.
func New() *Generator {
	return &Generator{}
}

// Prompt gathers configuration from user (I/O ACTION).
func (*Generator) Prompt(_ context.Context, intent generators.ResourceIntent, state generators.ProjectState) E.Either[error, generators.ResourceConfig] {
	// For MVP, use sensible defaults
	// In Phase 3, this will launch interactive TUI

	config := generators.ResourceConfig{
		Type:   generators.ResourceDynamoDB,
		Name:   intent.Name,
		Module: intent.UseModule,
		Variables: map[string]interface{}{
			"hash_key":               "id",
			"range_key":              "",
			"billing_mode":           "PAY_PER_REQUEST", // On-demand pricing
			"stream_enabled":         false,
			"stream_view_type":       "",
			"ttl_enabled":            false,
			"ttl_attribute":          "",
			"point_in_time_recovery": true,
			"attributes": []map[string]string{
				{"name": "id", "type": "S"}, // String type by default
			},
			"global_secondary_indexes": []map[string]interface{}{},
			"local_secondary_indexes":  []map[string]interface{}{},
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

		// Enable streams for Lambda integration
		config.Variables["stream_enabled"] = true
		config.Variables["stream_view_type"] = "NEW_AND_OLD_IMAGES"

		config.Integration = &generators.IntegrationConfig{
			TargetFunction: intent.ToFunc,
			EventSource: &generators.EventSourceConfig{
				ARNExpression:         fmt.Sprintf("module.%s.stream_arn", sanitizeName(intent.Name)),
				BatchSize:             defaultBatchSize,
				MaxBatchingWindowSecs: 0,
				MaxConcurrency:        defaultMaxConcurrency,
			},
			IAMPermissions: []generators.IAMPermission{
				{
					Effect: "Allow",
					Actions: []string{
						"dynamodb:GetRecords",
						"dynamodb:GetShardIterator",
						"dynamodb:DescribeStream",
						"dynamodb:ListStreams",
					},
					Resources: []string{
						fmt.Sprintf("module.%s.stream_arn", sanitizeName(intent.Name)),
					},
				},
				{
					Effect: "Allow",
					Actions: []string{
						"dynamodb:Query",
						"dynamodb:Scan",
						"dynamodb:GetItem",
						"dynamodb:PutItem",
						"dynamodb:UpdateItem",
						"dynamodb:DeleteItem",
					},
					Resources: []string{
						fmt.Sprintf("module.%s.table_arn", sanitizeName(intent.Name)),
					},
				},
			},
		}
	}

	return E.Right[error](config)
}

// Generate creates Terraform code from configuration (PURE CALCULATION).
func (gen *Generator) Generate(config generators.ResourceConfig, _ generators.ProjectState) E.Either[error, generators.GeneratedCode] {
	// Validate first, then chain generation - automatic error short-circuiting
	return E.Chain(func(validConfig generators.ResourceConfig) E.Either[error, generators.GeneratedCode] {
		var files []generators.FileToWrite

		// 1. Generate main DynamoDB resource file
		if validConfig.Module {
			files = append(files, generators.FileToWrite{
				Path:    "dynamodb.tf",
				Content: generateModuleCode(validConfig),
				Mode:    generators.WriteModeAppend,
			})
		} else {
			files = append(files, generators.FileToWrite{
				Path:    "dynamodb.tf",
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
	})(gen.Validate(config))
}

// Validate checks if configuration is valid (PURE CALCULATION).
func (*Generator) Validate(config generators.ResourceConfig) E.Either[error, generators.ResourceConfig] {
	if config.Name == "" {
		return E.Left[generators.ResourceConfig](
			errors.New("table name is required"),
		)
	}

	if !isValidName(config.Name) {
		return E.Left[generators.ResourceConfig](
			errors.New("table name must be alphanumeric with hyphens/underscores"),
		)
	}

	// Validate hash key is provided
	hashKey, ok := config.Variables["hash_key"].(string)
	if !ok || hashKey == "" {
		return E.Left[generators.ResourceConfig](
			errors.New("hash_key is required"),
		)
	}

	return E.Right[error](config)
}

// generateModuleCode creates Terraform module code (PURE).
func generateModuleCode(config generators.ResourceConfig) string {
	moduleName := sanitizeName(config.Name)
	tableName := config.Name

	hashKey, ok := config.Variables["hash_key"].(string)
	_ = ok
	rangeKey, ok := config.Variables["range_key"].(string)
	_ = ok
	billingMode, ok := config.Variables["billing_mode"].(string)
	_ = ok
	streamEnabled, ok := config.Variables["stream_enabled"].(bool)
	_ = ok
	streamViewType, ok := config.Variables["stream_view_type"].(string)
	_ = ok
	ttlEnabled, ok := config.Variables["ttl_enabled"].(bool)
	_ = ok
	ttlAttribute, ok := config.Variables["ttl_attribute"].(string)
	_ = ok
	pointInTimeRecovery, ok := config.Variables["point_in_time_recovery"].(bool)
	_ = ok
	attributes, ok := config.Variables["attributes"].([]map[string]string)
	_ = ok

	var parts []string

	parts = append(parts, "# Generated by forge add dynamodb "+config.Name)
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("module \"%s\" {", moduleName))
	parts = append(parts, "  source  = \"terraform-aws-modules/dynamodb-table/aws\"")
	parts = append(parts, "  version = \"~> 4.0\"")
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("  name = \"${var.namespace}%s\"", tableName))
	parts = append(parts, fmt.Sprintf("  hash_key  = \"%s\"", hashKey))

	if rangeKey != "" {
		parts = append(parts, fmt.Sprintf("  range_key = \"%s\"", rangeKey))
	}

	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("  billing_mode = \"%s\"", billingMode))

	// Attributes
	parts = append(parts, "")
	parts = append(parts, "  attributes = [")
	for _, attr := range attributes {
		parts = append(parts, "    {")
		parts = append(parts, fmt.Sprintf("      name = \"%s\"", attr["name"]))
		parts = append(parts, fmt.Sprintf("      type = \"%s\"", attr["type"]))
		parts = append(parts, "    },")
	}
	parts = append(parts, "  ]")

	// Streams
	if streamEnabled {
		parts = append(parts, "")
		parts = append(parts, "  # DynamoDB Streams for Lambda triggers")
		parts = append(parts, "  stream_enabled   = true")
		parts = append(parts, fmt.Sprintf("  stream_view_type = \"%s\"", streamViewType))
	}

	// TTL
	if ttlEnabled && ttlAttribute != "" {
		parts = append(parts, "")
		parts = append(parts, "  # Time-to-Live configuration")
		parts = append(parts, "  ttl_enabled        = true")
		parts = append(parts, fmt.Sprintf("  ttl_attribute_name = \"%s\"", ttlAttribute))
	}

	// Point-in-time recovery
	parts = append(parts, "")
	parts = append(parts, "  # Backup configuration")
	parts = append(parts, fmt.Sprintf("  point_in_time_recovery_enabled = %t", pointInTimeRecovery))

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
	tableName := config.Name

	hashKey, ok := config.Variables["hash_key"].(string)
	_ = ok
	rangeKey, ok := config.Variables["range_key"].(string)
	_ = ok
	billingMode, ok := config.Variables["billing_mode"].(string)
	_ = ok
	streamEnabled, ok := config.Variables["stream_enabled"].(bool)
	_ = ok
	streamViewType, ok := config.Variables["stream_view_type"].(string)
	_ = ok
	attributes, ok := config.Variables["attributes"].([]map[string]string)
	_ = ok

	var parts []string

	parts = append(parts, "# Generated by forge add dynamodb "+config.Name+" --raw")
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("resource \"aws_dynamodb_table\" \"%s\" {", resourceName))
	parts = append(parts, fmt.Sprintf("  name = \"${var.namespace}%s\"", tableName))
	parts = append(parts, fmt.Sprintf("  hash_key  = \"%s\"", hashKey))

	if rangeKey != "" {
		parts = append(parts, fmt.Sprintf("  range_key = \"%s\"", rangeKey))
	}

	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("  billing_mode = \"%s\"", billingMode))

	// Attributes
	parts = append(parts, "")
	for _, attr := range attributes {
		parts = append(parts, "  attribute {")
		parts = append(parts, fmt.Sprintf("    name = \"%s\"", attr["name"]))
		parts = append(parts, fmt.Sprintf("    type = \"%s\"", attr["type"]))
		parts = append(parts, "  }")
	}

	// Streams
	if streamEnabled {
		parts = append(parts, "")
		parts = append(parts, "  stream_enabled   = true")
		parts = append(parts, fmt.Sprintf("  stream_view_type = \"%s\"", streamViewType))
	}

	parts = append(parts, "")
	parts = append(parts, "  tags = {")
	parts = append(parts, "    ManagedBy = \"forge\"")
	parts = append(parts, "    Namespace = var.namespace")
	parts = append(parts, "  }")
	parts = append(parts, "}")
	parts = append(parts, "")

	return strings.Join(parts, "\n")
}

// generateOutputBlock creates a single Terraform output block (PURE).
func generateOutputBlock(name, description, valueExpr string) string {
	return fmt.Sprintf(`output "%s" {
  description = "%s"
  value       = %s
}`, name, description, valueExpr)
}

// generateOutputs creates Terraform outputs (PURE).
func generateOutputs(config generators.ResourceConfig) string {
	moduleName := sanitizeName(config.Name)
	var parts []string

	parts = append(parts, "# Outputs for "+config.Name)

	// Determine resource reference based on module vs raw resource
	var tableIDRef, tableARNRef, streamARNRef string
	if config.Module {
		tableIDRef = fmt.Sprintf("module.%s.dynamodb_table_id", moduleName)
		tableARNRef = fmt.Sprintf("module.%s.dynamodb_table_arn", moduleName)
		streamARNRef = fmt.Sprintf("module.%s.dynamodb_table_stream_arn", moduleName)
	} else {
		tableIDRef = fmt.Sprintf("aws_dynamodb_table.%s.id", moduleName)
		tableARNRef = fmt.Sprintf("aws_dynamodb_table.%s.arn", moduleName)
		streamARNRef = fmt.Sprintf("aws_dynamodb_table.%s.stream_arn", moduleName)
	}

	// Generate table ID output
	parts = append(parts, generateOutputBlock(
		moduleName+"_table_id",
		"ID of "+config.Name,
		tableIDRef,
	))
	parts = append(parts, "")

	// Generate table ARN output
	parts = append(parts, generateOutputBlock(
		moduleName+"_table_arn",
		"ARN of "+config.Name,
		tableARNRef,
	))

	// Stream ARN if enabled
	streamEnabled, _ := config.Variables["stream_enabled"].(bool)
	if streamEnabled {
		parts = append(parts, "")
		parts = append(parts, generateOutputBlock(
			moduleName+"_stream_arn",
			"Stream ARN of "+config.Name,
			streamARNRef,
		))
	}

	parts = append(parts, "")
	return strings.Join(parts, "\n")
}

// generateIntegrationCode creates Lambda event source mapping (PURE).
func generateIntegrationCode(config generators.ResourceConfig) string {
	if config.Integration == nil {
		return ""
	}

	tableName := sanitizeName(config.Name)
	functionName := config.Integration.TargetFunction
	eventSource := config.Integration.EventSource

	var parts []string

	// Event source mapping for DynamoDB Streams
	parts = append(parts, "# DynamoDB Streams event source mapping for "+config.Name)
	parts = append(parts, fmt.Sprintf("resource \"aws_lambda_event_source_mapping\" \"%s_%s\" {",
		functionName, tableName))
	parts = append(parts, "  event_source_arn = "+eventSource.ARNExpression)
	parts = append(parts, fmt.Sprintf("  function_name    = aws_lambda_function.%s.arn", functionName))
	parts = append(parts, "")
	parts = append(parts, "  starting_position = \"LATEST\"")
	parts = append(parts, fmt.Sprintf("  batch_size        = %d", eventSource.BatchSize))
	parts = append(parts, "")
	parts = append(parts, "  scaling_config {")
	parts = append(parts, fmt.Sprintf("    maximum_concurrency = %d", eventSource.MaxConcurrency))
	parts = append(parts, "  }")
	parts = append(parts, "}")
	parts = append(parts, "")

	// IAM policies
	for i, perm := range config.Integration.IAMPermissions {
		policyName := fmt.Sprintf("%s_dynamodb_%s_%d", functionName, tableName, i)

		parts = append(parts, fmt.Sprintf("# IAM policy for %s to access %s", functionName, config.Name))
		parts = append(parts, fmt.Sprintf("resource \"aws_iam_role_policy\" \"%s\" {", policyName))
		parts = append(parts, fmt.Sprintf("  name = \"${var.namespace}%s\"", policyName))
		parts = append(parts, fmt.Sprintf("  role = aws_iam_role.%s.id", functionName))
		parts = append(parts, "")
		parts = append(parts, "  policy = jsonencode({")
		parts = append(parts, "    Version = \"2012-10-17\"")
		parts = append(parts, "    Statement = [")
		parts = append(parts, "      {")
		parts = append(parts, fmt.Sprintf("        Effect = \"%s\"", perm.Effect))
		parts = append(parts, "        Action = [")

		for j, action := range perm.Actions {
			comma := ","
			if j == len(perm.Actions)-1 {
				comma = ""
			}
			parts = append(parts, fmt.Sprintf("          \"%s\"%s", action, comma))
		}

		parts = append(parts, "        ]")

		// Handle resource references
		resourceRef := perm.Resources[0]
		parts = append(parts, "        Resource = "+resourceRef)

		parts = append(parts, "      }")
		parts = append(parts, "    ]")
		parts = append(parts, "  })}")
		parts = append(parts, "")
	}

	return strings.Join(parts, "\n")
}

// sanitizeName converts a name to a valid Terraform identifier (PURE).
func sanitizeName(name string) string {
	// Replace hyphens with underscores for Terraform identifiers
	return strings.ReplaceAll(name, "-", "_")
}

// isValidName checks if a name is valid (PURE).
func isValidName(name string) bool {
	if len(name) == 0 {
		return false
	}

	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_') {

			return false
		}
	}

	return true
}
