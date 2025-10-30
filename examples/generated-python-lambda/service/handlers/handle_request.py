from typing import Annotated, Any

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


@app.post(
    API_PATH,
    summary='Orders service API - Create order endpoint',
    description='Orders service API - Create order endpoint',
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
