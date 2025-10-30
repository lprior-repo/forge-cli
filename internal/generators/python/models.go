package python

import (
	"fmt"
)

// generateEnvVars generates environment variables model.
func generateEnvVars(config ProjectConfig) string {
	content := `from typing import Annotated, Literal

from pydantic import BaseModel, Field, HttpUrl, PositiveInt


class Observability(BaseModel):
    """Observability configuration."""
    POWERTOOLS_SERVICE_NAME: Annotated[str, Field(min_length=1)]
    LOG_LEVEL: Literal['DEBUG', 'INFO', 'ERROR', 'CRITICAL', 'WARNING']


`

	if config.UseIdempotency {
		content += `class Idempotency(BaseModel):
    """Idempotency configuration."""
    IDEMPOTENCY_TABLE_NAME: Annotated[str, Field(min_length=1)]


`
	}

	content += `class HandlerEnvVars(`
	if config.UsePowertools {
		content += `Observability`
		if config.UseIdempotency {
			content += `, Idempotency`
		}
	} else {
		content += `BaseModel`
	}
	content += `):`
	content += `
    """Handler environment variables."""
`

	if config.UseDynamoDB {
		content += `    TABLE_NAME: Annotated[str, Field(min_length=1)]
`
	}

	return content
}

// generateInputModel generates Pydantic input model.
func generateInputModel(config ProjectConfig) string {
	return `from typing import Annotated

from pydantic import BaseModel, Field, field_validator


class RequestInput(BaseModel):
    """Request input model."""
    name: Annotated[str, Field(min_length=1, max_length=100, description='Name field')]
    count: Annotated[int, Field(strict=True, description='Count field')]

    @field_validator('count')
    @classmethod
    def check_count(cls, v):
        """Validate count is positive."""
        if v <= 0:
            raise ValueError('count must be larger than 0')
        return v

    model_config = {
        'json_schema_extra': {
            'examples': [
                {
                    'name': 'example',
                    'count': 5,
                }
            ]
        }
    }
`
}

// generateOutputModel generates Pydantic output model.
func generateOutputModel(config ProjectConfig) string {
	return `from typing import Annotated

from pydantic import BaseModel, Field


class RequestOutput(BaseModel):
    """Request output model."""
    id: Annotated[str, Field(description='Unique identifier')]
    name: Annotated[str, Field(description='Name field')]
    count: Annotated[int, Field(description='Count field')]
    status: Annotated[str, Field(description='Status')]

    model_config = {
        'json_schema_extra': {
            'examples': [
                {
                    'id': '123e4567-e89b-12d3-a456-426614174000',
                    'name': 'example',
                    'count': 5,
                    'status': 'created',
                }
            ]
        }
    }


class ErrorOutput(BaseModel):
    """Error output model."""
    error: Annotated[str, Field(description='Error message')]
    details: Annotated[str | None, Field(default=None, description='Error details')]
`
}

// generateObservability generates observability utilities.
func generateObservability(config ProjectConfig) string {
	if config.UsePowertools {
		return fmt.Sprintf(`from aws_lambda_powertools import Logger, Metrics, Tracer

# Initialize Powertools
logger = Logger(service='%s')
metrics = Metrics(namespace='%s', service='%s')
tracer = Tracer(service='%s')
`, config.ServiceName, config.ServiceName, config.ServiceName, config.ServiceName)
	}

	return `import logging

# Basic logging setup
logger = logging.getLogger()
logger.setLevel(logging.INFO)


def log_info(message: str, **kwargs):
    """Log info message."""
    logger.info(f"{message} - {kwargs}")


def log_error(message: str, **kwargs):
    """Log error message."""
    logger.error(f"{message} - {kwargs}")
`
}

// generateRestAPI generates REST API resolver.
func generateRestAPI(config ProjectConfig) string {
	if config.UsePowertools {
		return fmt.Sprintf(`from aws_lambda_powertools.event_handler import APIGatewayHttpResolver

# API configuration
API_PATH = '%s'

# Initialize API resolver (HTTP API v2)
app = APIGatewayHttpResolver()
`, config.APIPath)
	}

	return `# API configuration (basic setup)
API_PATH = '` + config.APIPath + `'
`
}

// generateBusinessLogic generates business logic layer.
func generateBusinessLogic(config ProjectConfig) string {
	content := `import uuid
from typing import Any

`
	if config.UsePowertools {
		content += `from aws_lambda_powertools.utilities.typing import LambdaContext

`
	}

	content += `from service.models.input import RequestInput
from service.models.output import RequestOutput
`

	if config.UseDynamoDB {
		content += `from service.dal.dynamodb_handler import save_item
`
	}

	if config.UsePowertools {
		content += `from service.handlers.utils.observability import logger, tracer

`
	}

	content += `

`

	if config.UsePowertools {
		content += `@tracer.capture_method
`
	}

	content += `def process_request(
    request_input: RequestInput,
    table_name: str | None,`

	if config.UsePowertools {
		content += `
    context: LambdaContext,`
	} else {
		content += `
    context: Any,`
	}

	content += `
) -> RequestOutput:
    """Process the incoming request."""
`

	if config.UsePowertools {
		content += `    logger.info('processing request', request=request_input.model_dump())
`
	}

	content += `
    # Generate unique ID
    item_id = str(uuid.uuid4())

    # Business logic here
    result = {
        'id': item_id,
        'name': request_input.name,
        'count': request_input.count,
        'status': 'created',
    }
`

	if config.UseDynamoDB {
		content += `
    # Save to DynamoDB
    if table_name:
        save_item(table_name=table_name, item=result)
`
	}

	if config.UsePowertools {
		content += `
    logger.info('request processed successfully', item_id=item_id)
`
	}

	content += `
    return RequestOutput(**result)
`

	return content
}

// generateDynamoDBHandler generates DynamoDB data access layer.
func generateDynamoDBHandler(config ProjectConfig) string {
	return `import boto3
from typing import Any

# Initialize DynamoDB client
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

// generateDBModel generates database model.
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
