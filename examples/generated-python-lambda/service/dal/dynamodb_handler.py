import boto3
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
