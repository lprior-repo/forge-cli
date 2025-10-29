from typing import Annotated, Literal

from pydantic import BaseModel, Field, HttpUrl, PositiveInt


class Observability(BaseModel):
    """Observability configuration."""
    POWERTOOLS_SERVICE_NAME: Annotated[str, Field(min_length=1)]
    LOG_LEVEL: Literal['DEBUG', 'INFO', 'ERROR', 'CRITICAL', 'WARNING']


class Idempotency(BaseModel):
    """Idempotency configuration."""
    IDEMPOTENCY_TABLE_NAME: Annotated[str, Field(min_length=1)]


class HandlerEnvVars(Observability, Idempotency):
    """Handler environment variables."""
    TABLE_NAME: Annotated[str, Field(min_length=1)]
