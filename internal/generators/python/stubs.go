package python

import (
	"fmt"
	"os"
	"path/filepath"
)

// generateTerraformFiles generates Terraform infrastructure using type-safe tfmodules.
func generateTerraformFiles(projectRoot string, config ProjectConfig) error {
	terraformDir := filepath.Join(projectRoot, "terraform")
	if err := os.MkdirAll(terraformDir, 0o755); err != nil {
		return fmt.Errorf("failed to create terraform directory: %w", err)
	}

	// Generate type-safe modules
	lambdaModule := GenerateLambdaModule(config)
	lambdaModuleName := "lambda_function"
	lambdaHCL := GenerateLambdaModuleHCL(lambdaModule)

	apiModule := GenerateAPIGatewayModule(config)
	apiGatewayHCL := GenerateAPIGatewayModuleHCL(apiModule, lambdaModuleName)

	// Conditionally generate DynamoDB
	var dynamoDBHCL string
	if config.UseDynamoDB {
		dynamoDBModule := GenerateDynamoDBModule(config)
		dynamoDBHCL = GenerateDynamoDBModuleHCL(dynamoDBModule, lambdaModuleName)
	}

	// Create Terraform files map
	files := map[string]string{
		filepath.Join(terraformDir, "main.tf"):       generateTerraformMain(config),
		filepath.Join(terraformDir, "variables.tf"):  generateTerraformVariables(config),
		filepath.Join(terraformDir, "outputs.tf"):    generateTerraformOutputs(config),
		filepath.Join(terraformDir, "lambda.tf"):     lambdaHCL,
		filepath.Join(terraformDir, "apigateway.tf"): apiGatewayHCL,
		filepath.Join(projectRoot, "Taskfile.yml"):   generateTaskfile(config),
	}

	// Add DynamoDB file if enabled
	if config.UseDynamoDB {
		files[filepath.Join(terraformDir, "dynamodb.tf")] = dynamoDBHCL
	}

	// Write all files
	for filePath, content := range files {
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filePath, err)
		}
	}

	return nil
}

// generateTerraformMain generates the main Terraform configuration.
func generateTerraformMain(config ProjectConfig) string {
	return `terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      ManagedBy   = "Terraform"
      Generator   = "Forge"
      Service     = var.service_name
      Environment = var.environment
    }
  }
}
`
}

// generateTerraformVariables generates Terraform variables.
func generateTerraformVariables(config ProjectConfig) string {
	return `variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "service_name" {
  description = "Service name"
  type        = string
  default     = "` + config.ServiceName + `"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "dev"
}
`
}

// generateTerraformOutputs generates Terraform outputs.
func generateTerraformOutputs(config ProjectConfig) string {
	outputs := `output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = module.lambda_function.lambda_function_name
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = module.lambda_function.lambda_function_arn
}

output "api_endpoint" {
  description = "API Gateway endpoint URL"
  value       = module.api_gateway.api_endpoint
}

output "api_id" {
  description = "API Gateway ID"
  value       = module.api_gateway.api_id
}
`

	if config.UseDynamoDB {
		outputs += `
output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = module.dynamodb_table.dynamodb_table_id
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table"
  value       = module.dynamodb_table.dynamodb_table_arn
}
`
	}

	return outputs
}

// generateTaskfile generates a Taskfile.yml for development workflow.
func generateTaskfile(config ProjectConfig) string {
	return `version: '3'

tasks:
  install:
    desc: Install dependencies using UV
    cmds:
      - uv pip install -r requirements.txt

  install:dev:
    desc: Install dev dependencies
    cmds:
      - uv pip install -r requirements.txt
      - uv pip install pytest pytest-cov ruff mypy

  format:
    desc: Format code with ruff
    cmds:
      - ruff format .

  lint:
    desc: Lint code with ruff
    cmds:
      - ruff check .

  type-check:
    desc: Run mypy type checking
    cmds:
      - mypy service/

  test:
    desc: Run tests with coverage
    cmds:
      - pytest tests/ -v --cov=service --cov-report=term-missing --cov-report=html

  full-test:
    desc: Run all checks (format, lint, type-check, test)
    cmds:
      - task: format
      - task: lint
      - task: type-check
      - task: test

  build:
    desc: Build Lambda deployment package
    cmds:
      - mkdir -p .build
      - uv pip install -r requirements.txt --target .build/python
      - cp -r service .build/python/
      - cd .build && zip -r lambda.zip python/

  deploy:
    desc: Deploy infrastructure with Terraform
    dir: terraform
    cmds:
      - task: build
      - terraform init
      - terraform plan
      - terraform apply -auto-approve

  status:
    desc: Show Terraform deployment status
    dir: terraform
    cmds:
      - terraform show

  outputs:
    desc: Show Terraform outputs
    dir: terraform
    cmds:
      - terraform output

  destroy:
    desc: Destroy infrastructure
    dir: terraform
    cmds:
      - terraform destroy -auto-approve

  test-api:
    desc: Test deployed API endpoint
    cmds:
      - |
        API_URL=$(cd terraform && terraform output -raw api_endpoint)
        curl -X ` + config.HTTPMethod + ` "${API_URL}` + config.APIPath + `" \
          -H "Content-Type: application/json" \
          -d '{"name": "test", "count": 42}'

  clean:
    desc: Clean build artifacts
    cmds:
      - rm -rf .build/
      - rm -rf __pycache__/
      - rm -rf **/__pycache__/
      - rm -rf .pytest_cache/
      - rm -rf htmlcov/
      - rm -rf .coverage
`
}

// Python code generation stubs below

func generateEnvVars(config ProjectConfig) string {
	return `from typing import Annotated, Literal

from pydantic import BaseModel, Field


class Observability(BaseModel):
    """Observability configuration."""
    POWERTOOLS_SERVICE_NAME: Annotated[str, Field(min_length=1)]
    LOG_LEVEL: Literal['DEBUG', 'INFO', 'ERROR', 'CRITICAL', 'WARNING']


class HandlerEnvVars(Observability):
    """Handler environment variables."""
    TABLE_NAME: Annotated[str, Field(min_length=1)]
`
}

func generateInputModel(config ProjectConfig) string {
	return `from typing import Annotated

from pydantic import BaseModel, Field, field_validator


class RequestInput(BaseModel):
    """Request input model."""
    name: Annotated[str, Field(min_length=1, max_length=100)]
    count: Annotated[int, Field(strict=True)]

    @field_validator('count')
    @classmethod
    def check_count(cls, v):
        if v <= 0:
            raise ValueError('count must be larger than 0')
        return v
`
}

func generateOutputModel(config ProjectConfig) string {
	return `from typing import Annotated

from pydantic import BaseModel, Field


class RequestOutput(BaseModel):
    """Request output model."""
    id: Annotated[str, Field(description='Unique identifier')]
    name: Annotated[str, Field(description='Name field')]
    count: Annotated[int, Field(description='Count field')]
    status: Annotated[str, Field(description='Status')]


class ErrorOutput(BaseModel):
    """Error output model."""
    error: Annotated[str, Field(description='Error message')]
    details: Annotated[str | None, Field(default=None)]
`
}

func generateObservability(config ProjectConfig) string {
	return `from aws_lambda_powertools import Logger, Metrics, Tracer

logger = Logger(service='` + config.ServiceName + `')
metrics = Metrics(namespace='` + config.ServiceName + `')
tracer = Tracer(service='` + config.ServiceName + `')
`
}

func generateRestAPI(config ProjectConfig) string {
	return `from aws_lambda_powertools.event_handler import APIGatewayHttpResolver

API_PATH = '` + config.APIPath + `'
app = APIGatewayHttpResolver()
`
}

func generateBusinessLogic(config ProjectConfig) string {
	return `import uuid
from typing import Any

from aws_lambda_powertools.utilities.typing import LambdaContext

from service.models.input import RequestInput
from service.models.output import RequestOutput
from service.handlers.utils.observability import logger, tracer


@tracer.capture_method
def process_request(
    request_input: RequestInput,
    table_name: str | None,
    context: LambdaContext,
) -> RequestOutput:
    """Process the incoming request."""
    logger.info('processing request', request=request_input.model_dump())

    item_id = str(uuid.uuid4())
    result = {
        'id': item_id,
        'name': request_input.name,
        'count': request_input.count,
        'status': 'created',
    }

    logger.info('request processed successfully', item_id=item_id)
    return RequestOutput(**result)
`
}

func generateDynamoDBHandler(config ProjectConfig) string {
	return `import boto3
from typing import Any

dynamodb = boto3.resource('dynamodb')


def save_item(table_name: str, item: dict[str, Any]) -> None:
    """Save item to DynamoDB table."""
    table = dynamodb.Table(table_name)
    table.put_item(Item=item)


def get_item(table_name: str, key: dict[str, Any]) -> dict[str, Any] | None:
    """Get item from DynamoDB table."""
    table = dynamodb.Table(table_name)
    response = table.get_item(Key=key)
    return response.get('Item')


def delete_item(table_name: str, key: dict[str, Any]) -> None:
    """Delete item from DynamoDB table."""
    table = dynamodb.Table(table_name)
    table.delete_item(Key=key)
`
}

func generateDBModel(config ProjectConfig) string {
	return `from typing import Annotated

from pydantic import BaseModel, Field


class DBItem(BaseModel):
    """Database item model."""
    id: Annotated[str, Field(description='Primary key')]
    name: Annotated[str, Field(description='Name field')]
    count: Annotated[int, Field(description='Count field')]
    status: Annotated[str, Field(description='Status field')]
`
}
