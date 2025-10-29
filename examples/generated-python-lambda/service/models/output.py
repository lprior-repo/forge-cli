from typing import Annotated

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
