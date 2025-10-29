from typing import Annotated

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
