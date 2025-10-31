package python

import (
	"fmt"

	"github.com/lewis/forge/internal/tfmodules/dynamodb"
)

// GenerateDynamoDBModule creates a type-safe DynamoDB module configuration.
// PURE: Calculation - same inputs always produce same outputs.
func GenerateDynamoDBModule(config ProjectConfig) *dynamodb.Module {
	tableName := config.TableName
	if tableName == "" {
		tableName = fmt.Sprintf("%s-%s", config.ServiceName, "${var.environment}")
	}

	// Create base DynamoDB module
	table := dynamodb.NewModule(tableName)

	// Configure billing mode - PAY_PER_REQUEST for serverless
	billingMode := "PAY_PER_REQUEST"
	table.BillingMode = &billingMode

	// Define primary key (id: string)
	hashKey := "id"
	table.HashKey = &hashKey

	// Attributes
	table.Attributes = []dynamodb.Attribute{
		{
			Name: "id",
			Type: "S", // String
		},
	}

	// Enable point-in-time recovery for data protection
	pitr := true
	table.PointInTimeRecoveryEnabled = &pitr

	// Enable server-side encryption
	encryption := true
	table.ServerSideEncryptionEnabled = &encryption

	// TTL configuration (optional but commonly used)
	ttlEnabled := false
	table.TTLEnabled = &ttlEnabled
	ttlAttribute := "ttl"
	table.TTLAttributeName = &ttlAttribute

	return table
}

// GenerateDynamoDBModuleHCL converts DynamoDB module to HCL string.
// PURE: Calculation - deterministic output from module configuration.
func GenerateDynamoDBModuleHCL(module *dynamodb.Module, lambdaModuleName string) string {
	// Use a static module name (variable interpolation not allowed in module names)
	moduleName := "dynamodb_table"

	hcl := fmt.Sprintf(`# DynamoDB table module
module "%s" {
  source  = "%s"
  version = "%s"

  name         = "%s"
  billing_mode = "%s"
  hash_key     = "%s"

`, moduleName, module.Source, module.Version,
		*module.Name, *module.BillingMode, *module.HashKey)

	// Attributes
	if len(module.Attributes) > 0 {
		hcl += "  attributes = [\n"
		for _, attr := range module.Attributes {
			hcl += fmt.Sprintf("    {\n      name = \"%s\"\n      type = \"%s\"\n    },\n", attr.Name, attr.Type)
		}
		hcl += "  ]\n\n"
	}

	// TTL configuration
	if module.TTLEnabled != nil {
		hcl += fmt.Sprintf("  ttl_enabled        = %t\n", *module.TTLEnabled)
		if module.TTLAttributeName != nil {
			hcl += fmt.Sprintf("  ttl_attribute_name = \"%s\"\n\n", *module.TTLAttributeName)
		}
	}

	// Point-in-time recovery
	if module.PointInTimeRecoveryEnabled != nil && *module.PointInTimeRecoveryEnabled {
		hcl += "  point_in_time_recovery_enabled = true\n\n"
	}

	// Server-side encryption
	if module.ServerSideEncryptionEnabled != nil && *module.ServerSideEncryptionEnabled {
		hcl += "  server_side_encryption_enabled = true\n\n"
	}

	// Tags
	hcl += "  tags = {\n"
	hcl += "    ManagedBy   = \"Terraform\"\n"
	hcl += "    Generator   = \"Forge\"\n"
	hcl += "    Service     = var.service_name\n"
	hcl += "    Environment = var.environment\n"
	hcl += fmt.Sprintf("    Name        = \"%s\"\n", *module.Name)
	hcl += "  }\n"

	hcl += "}\n\n"

	// Add IAM policy for Lambda to access DynamoDB
	hcl += fmt.Sprintf(`# IAM policy attachment for Lambda to access DynamoDB
resource "aws_iam_policy" "dynamodb_access" {
  name        = "${var.service_name}-${var.environment}-dynamodb-access"
  description = "Allow Lambda to access DynamoDB table"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [
          module.%s.dynamodb_table_arn,
          "${module.%s.dynamodb_table_arn}/index/*"
        ]
      }
    ]
  })

  tags = {
    ManagedBy   = "Terraform"
    Generator   = "Forge"
    Service     = var.service_name
    Environment = var.environment
  }
}

resource "aws_iam_role_policy_attachment" "lambda_dynamodb" {
  role       = module.%s.lambda_role_name
  policy_arn = aws_iam_policy.dynamodb_access.arn
}
`, moduleName, moduleName, lambdaModuleName)

	return hcl
}
