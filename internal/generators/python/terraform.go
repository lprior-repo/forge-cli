package python

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateTerraformFiles generates Terraform infrastructure code
func generateTerraformFiles(projectRoot string, config ProjectConfig) error {
	terraformDir := filepath.Join(projectRoot, "terraform")
	if err := os.MkdirAll(terraformDir, 0755); err != nil {
		return fmt.Errorf("failed to create terraform directory: %w", err)
	}

	files := map[string]func() string{
		"terraform/main.tf":      func() string { return generateTerraformMain(config) },
		"terraform/variables.tf": func() string { return generateTerraformVariables(config) },
		"terraform/outputs.tf":   func() string { return generateTerraformOutputs(config) },
		"terraform/lambda.tf":    func() string { return generateTerraformLambda(config) },
		"terraform/iam.tf":       func() string { return generateTerraformIAM(config) },
		"Taskfile.yml":           func() string { return generateTaskfile(config) },
	}

	if config.UseDynamoDB {
		files["terraform/dynamodb.tf"] = func() string { return generateTerraformDynamoDB(config) }
	}

	// Always include API Gateway for REST APIs
	files["terraform/apigateway.tf"] = func() string { return generateTerraformAPIGateway(config) }

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
func generateTerraformMain(config ProjectConfig) string {
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
func generateTerraformVariables(config ProjectConfig) string {
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
func generateTerraformOutputs(config ProjectConfig) string {
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
func generateTerraformLambda(config ProjectConfig) string {
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
		envVars := ""
		if config.UseDynamoDB {
			envVars += `
      TABLE_NAME                   = aws_dynamodb_table.main.name`
		}
		if config.UseIdempotency && config.UseDynamoDB {
			envVars += `
      IDEMPOTENCY_TABLE_NAME       = aws_dynamodb_table.main.name`
		}
		return envVars
	}())
}

// generateTerraformIAM generates iam.tf
func generateTerraformIAM(config ProjectConfig) string {
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
func generateTerraformDynamoDB(config ProjectConfig) string {
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
func generateTerraformAPIGateway(config ProjectConfig) string {
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

// generateTaskfile generates Taskfile.yml with uv-based build
func generateTaskfile(config ProjectConfig) string {
	serviceName := strings.ReplaceAll(config.ServiceName, "_", "-")

	// Build list of Python dependencies
	deps := []string{
		"pydantic",
		"boto3",
	}

	if config.UsePowertools {
		deps = append(deps, `"aws-lambda-powertools[tracer]"`, "aws-lambda-env-modeler")
	}

	if config.UseDynamoDB {
		deps = append(deps, "mypy-boto3-dynamodb")
	}

	if config.UseIdempotency {
		deps = append(deps, "cachetools")
	}

	depsStr := strings.Join(deps, " ")

	return fmt.Sprintf(`version: '3'

vars:
  BUILD_DIR: .build
  LAMBDA_DIR: .build/lambda
  SERVICE_NAME: %s
  AWS_REGION: us-east-1

tasks:
  clean:
    desc: Clean build artifacts
    cmds:
      - rm -rf {{.BUILD_DIR}}
      - rm -rf terraform/.terraform
      - rm -rf terraform/terraform.tfstate*

  install:
    desc: Install Python dependencies with uv (dev dependencies)
    cmds:
      - uv pip install --python %s %s pytest pytest-mock pytest-cov ruff mypy

  build:
    desc: Build Lambda deployment package using uv (fast!)
    deps: [clean]
    cmds:
      - mkdir -p {{.LAMBDA_DIR}}
      - echo "📦 Copying service code..."
      - cp -r service {{.LAMBDA_DIR}}/
      - echo "📦 Installing production dependencies with uv..."
      - uv pip install --python %s --target {{.LAMBDA_DIR}} %s
      - echo "✅ Lambda package built in {{.LAMBDA_DIR}}"
      - du -sh {{.LAMBDA_DIR}}

  test:
    desc: Run Python tests
    cmds:
      - pytest tests/ -v

  lint:
    desc: Lint Python code
    cmds:
      - ruff check .

  format:
    desc: Format Python code
    cmds:
      - ruff format .

  tf-init:
    desc: Initialize Terraform
    dir: terraform
    cmds:
      - terraform init

  tf-validate:
    desc: Validate Terraform configuration
    dir: terraform
    deps: [tf-init]
    cmds:
      - terraform validate

  tf-plan:
    desc: Terraform plan
    dir: terraform
    deps: [build, tf-init]
    cmds:
      - terraform plan -out=tfplan

  tf-apply:
    desc: Terraform apply
    dir: terraform
    deps: [tf-plan]
    cmds:
      - terraform apply tfplan
      - rm -f tfplan

  deploy:
    desc: Build and deploy to AWS
    cmds:
      - task: build
      - task: tf-apply

  destroy:
    desc: Destroy infrastructure
    dir: terraform
    cmds:
      - terraform destroy -auto-approve

  outputs:
    desc: Show Terraform outputs
    dir: terraform
    cmds:
      - terraform output -json | jq '.'

  test-api:
    desc: Test the deployed API
    dir: terraform
    cmds:
      - |
        API_URL=$(terraform output -raw api_gateway_url)%s
        echo "Testing API at: $API_URL"
        curl -X %s "$API_URL" \
          -H "Content-Type: application/json" \
          -d '{
            "name": "test-order",
            "count": 5
          }' | jq '.'

  logs:
    desc: Tail Lambda logs
    cmds:
      - |
        FUNCTION_NAME=$(cd terraform && terraform output -raw lambda_function_name)
        aws logs tail "/aws/lambda/$FUNCTION_NAME" --follow --region {{.AWS_REGION}}

  invoke:
    desc: Invoke Lambda function directly
    cmds:
      - |
        FUNCTION_NAME=$(cd terraform && terraform output -raw lambda_function_name)
        aws lambda invoke \
          --function-name "$FUNCTION_NAME" \
          --payload '{"body": "{\"name\":\"test\",\"count\":3}"}' \
          --region {{.AWS_REGION}} \
          response.json
        cat response.json | jq '.'
        rm -f response.json

  status:
    desc: Show deployment status
    cmds:
      - |
        cd terraform
        if [ -f terraform.tfstate ]; then
          echo "✅ Infrastructure deployed"
          echo ""
          echo "📊 Resources:"
          terraform show -json | jq -r '.values.root_module.resources[] | "  - \(.type): \(.name)"'
          echo ""
          echo "🌐 Endpoints:"
          terraform output
        else
          echo "❌ Infrastructure not deployed yet"
          echo "Run: task deploy"
        fi

  full-test:
    desc: Full test cycle (lint, test, build, validate)
    cmds:
      - task: format
      - task: lint
      - task: test
      - task: build
      - task: tf-validate
      - echo "✅ All checks passed!"

  help:
    desc: Show available tasks
    cmds:
      - task --list
`, serviceName, config.PythonVersion, depsStr, config.PythonVersion, depsStr, config.APIPath, config.HTTPMethod)
}
