# orders-service

Orders service API - Create order endpoint

## Getting Started

### Prerequisites

- Python 3.13
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

- **POST /api/orders** - Orders service API - Create order endpoint
