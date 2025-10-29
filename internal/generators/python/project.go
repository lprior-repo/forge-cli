package python

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectConfig defines the configuration for a Python Lambda project
type ProjectConfig struct {
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

// Generate creates the complete Python Lambda project structure
// Pure function - takes projectRoot and config as parameters
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

// createDirectoryStructure creates all necessary directories
func createDirectoryStructure() error {
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

	for _, dir := range dirs {
		path := filepath.Join(projectRoot, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateProjectFiles generates all project files
func generateProjectFiles() error {
	files := map[string]func() string{
		"pyproject.toml":                          g.generatePyProjectToml,
		"README.md":                               g.generateReadme,
		".gitignore":                              g.generateGitignore,
		"Makefile":                                g.generateMakefile,
		"service/__init__.py":                     g.generateEmptyInit,
		"service/handlers/__init__.py":            g.generateEmptyInit,
		"service/handlers/models/__init__.py":     g.generateEmptyInit,
		"service/handlers/utils/__init__.py":      g.generateEmptyInit,
		"service/logic/__init__.py":               g.generateEmptyInit,
		"service/logic/utils/__init__.py":         g.generateEmptyInit,
		"service/dal/__init__.py":                 g.generateEmptyInit,
		"service/dal/models/__init__.py":          g.generateEmptyInit,
		"service/models/__init__.py":              g.generateEmptyInit,
		"service/handlers/handle_request.py":      g.generateHandler,
		"service/handlers/models/env_vars.py":     g.generateEnvVars,
		"service/handlers/utils/observability.py": g.generateObservability,
		"service/handlers/utils/rest_api.py":      g.generateRestAPI,
		"service/models/input.py":                 g.generateInputModel,
		"service/models/output.py":                g.generateOutputModel,
		"service/logic/business_logic.py":         g.generateBusinessLogic,
		"tests/__init__.py":                       g.generateEmptyInit,
		"tests/unit/__init__.py":                  g.generateEmptyInit,
		"tests/integration/__init__.py":           g.generateEmptyInit,
		"tests/e2e/__init__.py":                   g.generateEmptyInit,
	}

	if config.UseDynamoDB {
		files["service/dal/dynamodb_handler.py"] = g.generateDynamoDBHandler
		files["service/dal/models/db.py"] = g.generateDBModel
	}

	for filePath, generator := range files {
		fullPath := filepath.Join(projectRoot, filePath)
		content := generator()
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filePath, err)
		}
	}

	return nil
}

// generatePyProjectToml generates pyproject.toml with Poetry configuration
func generatePyProjectToml() string {
	pythonConstraint := fmt.Sprintf("^%s", config.PythonVersion)

	content := fmt.Sprintf(`[build-system]
requires = ["poetry>=2.0.1"]
build-backend = "poetry.core.masonry.api"

[tool.poetry]
name = "%s"
version = "1.0.0"
description = "%s"
authors = ["Your Name <you@example.com>"]
readme = "README.md"

[tool.poetry.dependencies]
python = "%s"
pydantic = "^2.0.0"
`, config.ServiceName, config.Description, pythonConstraint)

	if config.UsePowertools {
		content += `aws-lambda-powertools = {extras = ["tracer"], version = "^3.7.0"}
aws-lambda-env-modeler = "*"
`
	}

	content += `boto3 = "^1.26.0"
`

	if config.UseDynamoDB {
		content += `mypy-boto3-dynamodb = "*"
`
	}

	if config.UseIdempotency {
		content += `cachetools = "*"
`
	}

	content += `
[tool.poetry.group.dev.dependencies]
pytest = "*"
pytest-mock = "*"
pytest-cov = "*"
ruff = "*"
mypy = "*"

[tool.ruff]
line-length = 150
target-version = "py` + config.PythonVersion[:2] + `13"

[tool.ruff.lint]
select = ["E", "W", "F", "I", "C", "B"]
ignore = ["E203", "E266", "E501", "W191"]

[tool.ruff.format]
quote-style = "single"
indent-style = "space"
`

	return content
}

// generateReadme generates README.md
func generateReadme() string {
	return fmt.Sprintf(`# %s

%s

## Getting Started

### Prerequisites

- Python %s
- Poetry 2.0+
- AWS CLI configured

### Installation

'''bash
poetry install
'''

### Development

'''bash
# Run tests
poetry run pytest

# Format code
poetry run ruff format .

# Type check
poetry run mypy service/
'''

### Deployment

'''bash
# Deploy to AWS
cdk deploy
'''

## Project Structure

'''
service/
├── handlers/       # Lambda handler entry points
├── logic/          # Business logic layer
├── dal/            # Data access layer
└── models/         # Pydantic models
'''

## API

- **%s %s** - %s
`, config.ServiceName, config.Description, config.PythonVersion,
		config.HTTPMethod, config.APIPath, config.Description)
}

// generateGitignore generates .gitignore
func generateGitignore() string {
	return `__pycache__/
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
.env
.venv
env/
venv/
.vscode/
.idea/
*.swp
*.swo
*~
.DS_Store
cdk.out/
.build/
`
}

// generateMakefile generates Makefile
func generateMakefile() string {
	return fmt.Sprintf(`.PHONY: install test format lint deploy clean

install:
	poetry install

test:
	poetry run pytest tests/ -v

format:
	poetry run ruff format .

lint:
	poetry run ruff check .
	poetry run mypy service/

deploy:
	cdk deploy

clean:
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
	rm -rf .pytest_cache .coverage htmlcov/ dist/ build/
`)
}

// generateEmptyInit generates empty __init__.py
func generateEmptyInit() string {
	return ""
}

// generateHandler generates the Lambda handler
func generateHandler() string {
	if config.UsePowertools {
		return g.generatePowertoolsHandler()
	}
	return g.generateBasicHandler()
}

// generatePowertoolsHandler generates handler with Powertools
func generatePowertoolsHandler() string {
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
@logger.inject_lambda_context(correlation_id_path=correlation_paths.API_GATEWAY_REST)
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

// generateBasicHandler generates basic handler without Powertools
func generateBasicHandler() string {
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
