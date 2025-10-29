from typing import Annotated

from pydantic import BaseModel, Field


class DBItem(BaseModel):
    """Database item model."""
    id: Annotated[str, Field(description='Primary key')]
    name: Annotated[str, Field(description='Name field')]
    count: Annotated[int, Field(description='Count field')]
    status: Annotated[str, Field(description='Status field')]
