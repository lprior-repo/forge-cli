import uuid
from typing import Any

from aws_lambda_powertools.utilities.typing import LambdaContext

from service.models.input import RequestInput
from service.models.output import RequestOutput
from service.dal.dynamodb_handler import save_item
from service.handlers.utils.observability import logger, tracer



@tracer.capture_method
def process_request(
    request_input: RequestInput,
    table_name: str | None,
    context: LambdaContext,
) -> RequestOutput:
    """Process the incoming request."""
    logger.info('processing request', request=request_input.model_dump())

    # Generate unique ID
    item_id = str(uuid.uuid4())

    # Business logic here
    result = {
        'id': item_id,
        'name': request_input.name,
        'count': request_input.count,
        'status': 'created',
    }

    # Save to DynamoDB
    if table_name:
        save_item(table_name=table_name, item=result)

    logger.info('request processed successfully', item_id=item_id)

    return RequestOutput(**result)
