package python

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateTerraformFiles generates Terraform infrastructure code
func generateTerraformFiles() error {
	terraformDir := filepath.Join(projectRoot, "terraform")
	if err := os.MkdirAll(terraformDir, 0755); err != nil {
		return fmt.Errorf("failed to create terraform directory: %w", err)
	}

	files := map[string]func() string{
		"terraform/main.tf":      g.generateTerraformMain,
		"terraform/variables.tf": g.generateTerraformVariables,
		"terraform/outputs.tf":   g.generateTerraformOutputs,
		"terraform/lambda.tf":    g.generateTerraformLambda,
		"terraform/iam.tf":       g.generateTerraformIAM,
	}

	if config.UseDynamoDB {
		files["terraform/dynamodb.tf"] = g.generateTerraformDynamoDB
	}

	// Always include API Gateway for REST APIs
	files["terraform/apigateway.tf"] = g.generateTerraformAPIGateway

	for filePath, generator := range files {
		fullPath := filepath.Join(projectRoot, filePath)
		content := generator()
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filePath, err)
		}
	}

	return nil
}

// generateTerraformMain generates main.tf
func generateTerraformMain() string {
	return `terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = var.tags
  }
}
`
}

// generateTerraformVariables generates variables.tf
func generateTerraformVariables() string {
	serviceName := strings.ReplaceAll(config.ServiceName, "_", "-")

	content := fmt.Sprintf(`variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "service_name" {
  description = "Service name"
  type        = string
  default     = "%s"
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default = {
    Service     = "%s"
    ManagedBy   = "Terraform"
    Generator   = "Forge"
  }
}

variable "lambda_runtime" {
  description = "Lambda runtime"
  type        = string
  default     = "python%s"
}

variable "lambda_timeout" {
  description = "Lambda timeout in seconds"
  type        = number
  default     = 30
}

variable "lambda_memory_size" {
  description = "Lambda memory size in MB"
  type        = number
  default     = 512
}
`, serviceName, serviceName, config.PythonVersion)

	if config.UseDynamoDB {
		content += `
variable "dynamodb_billing_mode" {
  description = "DynamoDB billing mode"
  type        = string
  default     = "PAY_PER_REQUEST"
}

variable "dynamodb_read_capacity" {
  description = "DynamoDB read capacity units (only used if billing_mode is PROVISIONED)"
  type        = number
  default     = 5
}

variable "dynamodb_write_capacity" {
  description = "DynamoDB write capacity units (only used if billing_mode is PROVISIONED)"
  type        = number
  default     = 5
}
`
	}

	return content
}

// generateTerraformOutputs generates outputs.tf
func generateTerraformOutputs() string {
	content := `output "lambda_function_name" {
  description = "Lambda function name"
  value       = aws_lambda_function.main.function_name
}

output "lambda_function_arn" {
  description = "Lambda function ARN"
  value       = aws_lambda_function.main.arn
}

output "lambda_invoke_arn" {
  description = "Lambda invoke ARN"
  value       = aws_lambda_function.main.invoke_arn
}

output "api_gateway_url" {
  description = "API Gateway URL"
  value       = aws_apigatewayv2_stage.default.invoke_url
}

output "api_endpoint" {
  description = "Full API endpoint URL"
  value       = "${aws_apigatewayv2_stage.default.invoke_url}${var.service_name}"
}
`

	if config.UseDynamoDB {
		content += `
output "dynamodb_table_name" {
  description = "DynamoDB table name"
  value       = aws_dynamodb_table.main.name
}

output "dynamodb_table_arn" {
  description = "DynamoDB table ARN"
  value       = aws_dynamodb_table.main.arn
}
`
	}

	return content
}

// generateTerraformLambda generates lambda.tf
func generateTerraformLambda() string {
	return fmt.Sprintf(`# Lambda deployment package
data "archive_file" "lambda" {
  type        = "zip"
  source_dir  = "${path.module}/../.build/lambda"
  output_path = "${path.module}/../.build/lambda.zip"
}

# Lambda function
resource "aws_lambda_function" "main" {
  filename         = data.archive_file.lambda.output_path
  function_name    = "${var.service_name}-${var.environment}"
  role            = aws_iam_role.lambda.arn
  handler         = "service.handlers.handle_request.lambda_handler"
  source_code_hash = data.archive_file.lambda.output_base64sha256
  runtime         = var.lambda_runtime
  timeout         = var.lambda_timeout
  memory_size     = var.lambda_memory_size

  environment {
    variables = {
      POWERTOOLS_SERVICE_NAME      = var.service_name
      LOG_LEVEL                    = "INFO"
      ENVIRONMENT                  = var.environment%s
    }
  }

  tracing_config {
    mode = "Active"
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    aws_cloudwatch_log_group.lambda,
  ]
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.service_name}-${var.environment}"
  retention_in_days = 7
}

# Lambda permission for API Gateway
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.main.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.main.execution_arn}/*/*"
}
`, func() string {
		if config.UseDynamoDB {
			return `
      TABLE_NAME                   = aws_dynamodb_table.main.name`
		}
		return ""
	}())
}

// generateTerraformIAM generates iam.tf
func generateTerraformIAM() string {
	content := `# Lambda execution role
resource "aws_iam_role" "lambda" {
  name = "${var.service_name}-lambda-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# CloudWatch Logs policy
resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# X-Ray tracing policy
resource "aws_iam_role_policy_attachment" "lambda_xray" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/AWSXRayDaemonWriteAccess"
}
`

	if config.UseDynamoDB {
		content += `
# DynamoDB access policy
resource "aws_iam_role_policy" "dynamodb" {
  name = "${var.service_name}-dynamodb-policy-${var.environment}"
  role = aws_iam_role.lambda.id

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
        Resource = aws_dynamodb_table.main.arn
      }
    ]
  })
}
`
	}

	return content
}

// generateTerraformDynamoDB generates dynamodb.tf
func generateTerraformDynamoDB() string {
	tableName := config.TableName
	if tableName == "" {
		tableName = "${var.service_name}-${var.environment}"
	}

	return fmt.Sprintf(`# DynamoDB table
resource "aws_dynamodb_table" "main" {
  name         = "%s"
  billing_mode = var.dynamodb_billing_mode
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  ttl {
    attribute_name = "ttl"
    enabled        = false
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled = true
  }

  tags = {
    Name = "%s"
  }
}
`, tableName, tableName)
}

// generateTerraformAPIGateway generates apigateway.tf
func generateTerraformAPIGateway() string {
	apiPath := strings.TrimPrefix(config.APIPath, "/")

	return fmt.Sprintf(`# API Gateway HTTP API (v2)
resource "aws_apigatewayv2_api" "main" {
  name          = "${var.service_name}-${var.environment}"
  protocol_type = "HTTP"
  description   = "%s"

  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers = ["content-type", "x-amz-date", "authorization", "x-api-key"]
    max_age       = 300
  }
}

# API Gateway stage
resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = "$default"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
    })
  }
}

# CloudWatch Log Group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${var.service_name}-${var.environment}"
  retention_in_days = 7
}

# API Gateway integration with Lambda
resource "aws_apigatewayv2_integration" "lambda" {
  api_id             = aws_apigatewayv2_api.main.id
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.main.invoke_arn
  payload_format_version = "2.0"
}

# API Gateway route
resource "aws_apigatewayv2_route" "main" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "%s /%s"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}
`, config.Description, config.HTTPMethod, apiPath)
}
