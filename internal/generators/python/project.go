package python

import (
	"fmt"
	"os"
	"path/filepath"
)

type (
	// ProjectConfig defines the configuration for a Python Lambda project.
	ProjectConfig struct {
		ServiceName    string
		FunctionName   string
		Description    string
		PythonVersion  string // e.g., "3.13"
		UsePowertools  bool
		UseIdempotency bool
		UseDynamoDB    bool
		TableName      string
		APIPath        string // e.g., "/api/orders"
		HTTPMethod     string // e.g., "POST"
	}
)

// Generate creates a complete Python Lambda project with Terraform infrastructure.
// This is a pure function that takes projectRoot and config as parameters.
func Generate(projectRoot string, config ProjectConfig) error {
	// Create directory structure
	if err := createDirectoryStructure(projectRoot, config); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Generate project files
	if err := generateProjectFiles(projectRoot, config); err != nil {
		return fmt.Errorf("failed to generate project files: %w", err)
	}

	// Generate Terraform infrastructure
	if err := generateTerraformFiles(projectRoot, config); err != nil {
		return fmt.Errorf("failed to generate terraform files: %w", err)
	}

	return nil
}

// createDirectoryStructure creates all necessary directories.
func createDirectoryStructure(projectRoot string, _ ProjectConfig) error {
	dirs := []string{
		"service",
		"service/handlers",
		"service/handlers/models",
		"service/handlers/utils",
		"service/logic",
		"service/logic/utils",
		"service/dal",
		"service/dal/models",
		"service/models",
		"tests",
		"tests/unit",
		"tests/integration",
		"tests/e2e",
	}

	const dirPerms = 0o750 // rwxr-x---
	for _, dir := range dirs {
		path := filepath.Join(projectRoot, dir)
		if err := os.MkdirAll(path, dirPerms); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateProjectFiles generates all project files.
func generateProjectFiles(projectRoot string, config ProjectConfig) error {
	files := map[string]func() string{
		"requirements.txt":                        func() string { return generateRequirementsTxt(config) },
		"README.md":                               func() string { return generateReadme(config) },
		".gitignore":                              func() string { return generateGitignore(config) },
		"service/__init__.py":                     generateEmptyInit,
		"service/handlers/__init__.py":            generateEmptyInit,
		"service/handlers/models/__init__.py":     generateEmptyInit,
		"service/handlers/utils/__init__.py":      generateEmptyInit,
		"service/logic/__init__.py":               generateEmptyInit,
		"service/logic/utils/__init__.py":         generateEmptyInit,
		"service/dal/__init__.py":                 generateEmptyInit,
		"service/dal/models/__init__.py":          generateEmptyInit,
		"service/models/__init__.py":              generateEmptyInit,
		"service/handlers/handle_request.py":      func() string { return generateHandler(config) },
		"service/handlers/models/env_vars.py":     func() string { return generateEnvVars(config) },
		"service/handlers/utils/observability.py": func() string { return generateObservability(config) },
		"service/handlers/utils/rest_api.py":      func() string { return generateRestAPI(config) },
		"service/models/input.py":                 func() string { return generateInputModel(config) },
		"service/models/output.py":                func() string { return generateOutputModel(config) },
		"service/logic/business_logic.py":         func() string { return generateBusinessLogic(config) },
		"tests/__init__.py":                       generateEmptyInit,
		"tests/unit/__init__.py":                  generateEmptyInit,
		"tests/integration/__init__.py":           generateEmptyInit,
		"tests/e2e/__init__.py":                   generateEmptyInit,
	}

	if config.UseDynamoDB {
		files["service/dal/dynamodb_handler.py"] = func() string { return generateDynamoDBHandler(config) }
		files["service/dal/models/db.py"] = func() string { return generateDBModel(config) }
	}

	const filePerms = 0o600 // rw-------
	for filePath, generator := range files {
		fullPath := filepath.Join(projectRoot, filePath)
		content := generator()
		if err := os.WriteFile(fullPath, []byte(content), filePerms); err != nil {
			return fmt.Errorf("failed to write %s: %w", filePath, err)
		}
	}

	return nil
}

// generateRequirementsTxt generates requirements.txt with production dependencies.
// UV will install these during the build step.
func generateRequirementsTxt(config ProjectConfig) string {
	deps := []string{
		"pydantic>=2.0.0",
		"boto3>=1.26.0",
	}

	if config.UsePowertools {
		deps = append(deps,
			"aws-lambda-powertools[tracer]>=3.7.0",
			"aws-lambda-env-modeler",
		)
	}

	if config.UseDynamoDB {
		deps = append(deps, "mypy-boto3-dynamodb")
	}

	if config.UseIdempotency {
		deps = append(deps, "cachetools")
	}

	content := "# Production dependencies\n"
	content += "# Install with: uv pip install -r requirements.txt\n\n"

	for _, dep := range deps {
		content += dep + "\n"
	}

	return content
}

// generateReadme generates README.md.
func generateReadme(config ProjectConfig) string {
	return fmt.Sprintf(`# %s

%s

## Getting Started

### Prerequisites

- Python %s
- [UV](https://github.com/astral-sh/uv) (fast Python package installer)
- [Task](https://taskfile.dev/) (task runner)
- AWS CLI configured
- Terraform >= 1.0

### Quick Start

'''bash
# Install UV (if not already installed)
curl -LsSf https://astral.sh/uv/install.sh | sh

# Run full test cycle
task full-test

# Build Lambda package
task build

# Deploy to AWS
task deploy
'''

### Development

'''bash
# Install dev dependencies
task install

# Run tests with coverage
task test

# Format code
task format

# Lint code
task lint

# Type check
task type-check
'''

### Deployment

'''bash
# Deploy infrastructure
task deploy

# Show deployment status
task status

# View outputs
task outputs

# Test API
task test-api

# Destroy infrastructure
task destroy
'''

## Project Structure

'''
.
├── service/
│   ├── handlers/       # Lambda handler entry points
│   ├── logic/          # Business logic layer
│   ├── dal/            # Data access layer (if using DynamoDB)
│   └── models/         # Pydantic models
├── tests/
│   ├── unit/          # Unit tests
│   ├── integration/   # Integration tests
│   └── e2e/           # End-to-end tests
├── terraform/         # Infrastructure as Code
│   ├── main.tf        # Provider configuration
│   ├── variables.tf   # Input variables
│   ├── lambda.tf      # Lambda module
│   ├── apigateway.tf  # API Gateway module
│   └── outputs.tf     # Stack outputs
└── Taskfile.yml       # Task definitions
'''

## API

- **%s %s** - %s

## Architecture

This project uses:
- **terraform-aws-modules/lambda** for type-safe Lambda configuration
- **terraform-aws-modules/apigateway-v2** for HTTP API
%s
- **UV** for fast Python dependency management
- **Task** for streamlined development workflow

## Testing

'''bash
# Run all tests
task test

# Run specific test file
pytest tests/unit/test_handler.py -v

# Run with coverage
task test
'''
`, config.ServiceName, config.Description, config.PythonVersion,
		config.HTTPMethod, config.APIPath, config.Description,
		func() string {
			if config.UseDynamoDB {
				return "- **terraform-aws-modules/dynamodb-table** for DynamoDB\n"
			}
			return ""
		}())
}

// generateGitignore generates .gitignore.
func generateGitignore(_ ProjectConfig) string {
	return `# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg
.pytest_cache/
.coverage
htmlcov/
.tox/

# Virtual environments
.env
.venv
env/
venv/

# IDEs
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store

# Build artifacts
.build/

# Terraform
terraform/.terraform/
terraform/.terraform.lock.hcl
terraform/terraform.tfstate
terraform/terraform.tfstate.backup
terraform/*.tfplan
terraform/crash.log

# AWS
response.json
`
}

// generateEmptyInit generates empty __init__.py.
func generateEmptyInit() string {
	return ""
}

// generateHandler generates the Lambda handler.
func generateHandler(config ProjectConfig) string {
	if config.UsePowertools {
		return generatePowertoolsHandler(config)
	}
	return generateBasicHandler(config)
}

// generatePowertoolsHandler generates handler with Powertools.
func generatePowertoolsHandler(config ProjectConfig) string {
	return fmt.Sprintf(`from typing import Annotated, Any

from aws_lambda_env_modeler import get_environment_variables, init_environment_variables
from aws_lambda_powertools.event_handler.openapi.params import Body
from aws_lambda_powertools.logging import correlation_paths
from aws_lambda_powertools.metrics import MetricUnit
from aws_lambda_powertools.utilities.typing import LambdaContext

from service.handlers.models.env_vars import HandlerEnvVars
from service.handlers.utils.observability import logger, metrics, tracer
from service.handlers.utils.rest_api import app, API_PATH
from service.logic.business_logic import process_request
from service.models.input import RequestInput
from service.models.output import RequestOutput, ErrorOutput


@app.%s(
    API_PATH,
    summary='%s',
    description='%s',
    response_description='Successful response',
    responses={
        200: {
            'description': 'Success',
            'content': {'application/json': {'model': RequestOutput}},
        },
        500: {
            'description': 'Internal server error',
            'content': {'application/json': {'model': ErrorOutput}},
        },
    },
    tags=['API'],
)
def handle_request(request_input: Annotated[RequestInput, Body(embed=False, media_type='application/json')]) -> RequestOutput:
    """Handle incoming API request."""
    env_vars: HandlerEnvVars = get_environment_variables(model=HandlerEnvVars)
    logger.debug('environment variables', env_vars=env_vars.model_dump())
    logger.info('received request', request=request_input.model_dump())

    metrics.add_metric(name='ValidRequests', unit=MetricUnit.Count, value=1)

    response: RequestOutput = process_request(
        request_input=request_input,
        table_name=env_vars.TABLE_NAME if hasattr(env_vars, 'TABLE_NAME') else None,
        context=app.lambda_context,
    )

    logger.info('finished processing request')
    return response


@init_environment_variables(model=HandlerEnvVars)
@logger.inject_lambda_context(correlation_id_path=correlation_paths.API_GATEWAY_HTTP)
@metrics.log_metrics
@tracer.capture_lambda_handler(capture_response=False)
def lambda_handler(event: dict[str, Any], context: LambdaContext) -> dict[str, Any]:
    """Lambda handler entry point."""
    return app.resolve(event, context)
`, func() string {
		switch config.HTTPMethod {
		case "GET":
			return "get"
		case "POST":
			return "post"
		case "PUT":
			return "put"
		case "DELETE":
			return "delete"
		default:
			return "post"
		}
	}(), config.Description, config.Description)
}

// generateBasicHandler generates basic handler without Powertools.
func generateBasicHandler(_ ProjectConfig) string {
	return `import json
from typing import Any

from service.logic.business_logic import process_request
from service.models.input import RequestInput
from service.models.output import RequestOutput


def lambda_handler(event: dict[str, Any], context: Any) -> dict[str, Any]:
    """Lambda handler entry point."""
    try:
        # Parse input
        body = json.loads(event.get('body', '{}'))
        request_input = RequestInput(**body)

        # Process request
        response = process_request(request_input=request_input, table_name=None, context=context)

        return {
            'statusCode': 200,
            'body': json.dumps(response.model_dump()),
            'headers': {
                'Content-Type': 'application/json'
            }
        }
    except Exception as e:
        return {
            'statusCode': 500,
            'body': json.dumps({'error': str(e)}),
            'headers': {
                'Content-Type': 'application/json'
            }
        }
`
}

// More generator functions will follow...
