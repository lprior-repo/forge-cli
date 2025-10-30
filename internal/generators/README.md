# internal/generators

**Advanced project generators for opinionated, production-ready Lambda architectures**

## Overview

The `generators` package provides **opinionated code generators** that create complete, production-ready Lambda projects with best practices baked in. Unlike the basic scaffolding in `internal/scaffold`, generators create **enterprise-grade** projects with:

- ✅ **Layered architecture** (handlers, logic, DAL)
- ✅ **AWS Powertools** integration (logging, tracing, metrics)
- ✅ **Testing infrastructure** (unit, integration, E2E)
- ✅ **Type safety** (Pydantic models, TypeScript types)
- ✅ **Idempotency patterns** (for critical operations)
- ✅ **DynamoDB integration** (with proper modeling)

## Philosophy

**Approved patterns for teams:**

- Teams customize generators to enforce their architecture standards
- Generated code follows hexagonal/clean architecture principles
- 100% test coverage from day 1 (with passing tests)
- Production-ready observability (CloudWatch, X-Ray)

**Comparison:**

| Package | Complexity | Output | Use Case |
|---------|------------|--------|----------|
| `internal/scaffold` | Simple | Basic hello-world | Getting started, prototypes |
| `internal/generators` | Advanced | Production-grade architecture | Enterprise teams, critical systems |

## Generators

### Python Generator (`generators/python/`)

Creates a complete Python Lambda project with hexagonal architecture.

**Generated structure:**
```
my-service/
├── pyproject.toml              # Poetry config with dependencies
├── Makefile                    # Dev commands (test, lint, deploy)
├── README.md                   # Service documentation
├── service/
│   ├── handlers/               # API layer (AWS Lambda handlers)
│   │   ├── __init__.py
│   │   ├── create_order.py     # Lambda entry point
│   │   └── models/             # Request/response models (Pydantic)
│   │       └── order.py
│   ├── logic/                  # Business logic layer
│   │   ├── __init__.py
│   │   ├── order_service.py    # Core business logic
│   │   └── utils/              # Business utilities
│   ├── dal/                    # Data access layer
│   │   ├── __init__.py
│   │   ├── order_repository.py # DynamoDB operations
│   │   └── models/             # Database models
│   │       └── order_entity.py
│   └── models/                 # Domain models (shared)
│       └── order.py
├── tests/
│   ├── unit/                   # Unit tests (100% coverage)
│   ├── integration/            # Integration tests
│   └── e2e/                    # End-to-end tests
└── infra/                      # Terraform infrastructure
    ├── main.tf                 # Lambda + DynamoDB + API Gateway
    ├── variables.tf
    └── outputs.tf
```

#### Features

**AWS Powertools Integration:**
```python
from aws_lambda_powertools import Logger, Tracer, Metrics
from aws_lambda_powertools.utilities.typing import LambdaContext

logger = Logger()
tracer = Tracer()
metrics = Metrics()

@logger.inject_lambda_context
@tracer.capture_lambda_handler
@metrics.log_metrics
def lambda_handler(event: dict, context: LambdaContext) -> dict:
    logger.info("Processing order", extra={"order_id": order_id})
    # ...
```

**Pydantic Models for Type Safety:**
```python
# service/handlers/models/order.py
from pydantic import BaseModel, Field

class CreateOrderRequest(BaseModel):
    customer_id: str = Field(..., description="Customer ID")
    items: List[OrderItem]
    total_amount: Decimal = Field(..., gt=0)

class CreateOrderResponse(BaseModel):
    order_id: str
    status: str
    created_at: datetime
```

**Idempotency for Critical Operations:**
```python
from aws_lambda_powertools.utilities.idempotency import idempotent

@idempotent(persistence_store=dynamodb_store)
def process_payment(order_id: str) -> PaymentResult:
    # Guaranteed to execute only once, even if Lambda retries
    return charge_credit_card(order_id)
```

**DynamoDB Repository Pattern:**
```python
# service/dal/order_repository.py
class OrderRepository:
    def save(self, order: Order) -> None:
        item = {
            "PK": f"ORDER#{order.id}",
            "SK": f"ORDER#{order.id}",
            "customer_id": order.customer_id,
            "items": [item.dict() for item in order.items],
            # ...
        }
        self.table.put_item(Item=item)

    def get_by_id(self, order_id: str) -> Optional[Order]:
        response = self.table.get_item(
            Key={"PK": f"ORDER#{order_id}", "SK": f"ORDER#{order_id}"}
        )
        return Order.from_dynamodb(response["Item"]) if "Item" in response else None
```

#### Usage

```go
import "github.com/lewis/forge/internal/generators/python"

config := python.ProjectConfig{
    ServiceName:    "order-service",
    FunctionName:   "create-order",
    Description:    "Creates customer orders",
    PythonVersion:  "3.13",
    UsePowertools:  true,
    UseIdempotency: true,
    UseDynamoDB:    true,
    TableName:      "orders",
    APIPath:        "/api/orders",
    HTTPMethod:     "POST",
}

err := python.Generate("./order-service", config)
```

**Generated dependencies (pyproject.toml):**
```toml
[tool.poetry.dependencies]
python = "^3.13"
aws-lambda-powertools = "^2.30.0"
pydantic = "^2.5.0"
boto3 = "^1.34.0"

[tool.poetry.group.dev.dependencies]
pytest = "^7.4.0"
pytest-cov = "^4.1.0"
moto = "^4.2.0"  # AWS mocking
black = "^23.12.0"
ruff = "^0.1.9"
mypy = "^1.8.0"
```

**Generated Terraform:**
```hcl
# infra/main.tf
resource "aws_dynamodb_table" "orders" {
  name           = "orders"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "PK"
  range_key      = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  tags = {
    Service   = "order-service"
    ManagedBy = "forge"
  }
}

resource "aws_lambda_function" "create_order" {
  function_name = "create-order"
  role          = aws_iam_role.lambda.arn
  handler       = "service.handlers.create_order.lambda_handler"
  runtime       = "python3.13"
  filename      = "../dist/function.zip"
  timeout       = 30
  memory_size   = 256

  environment {
    variables = {
      TABLE_NAME               = aws_dynamodb_table.orders.name
      POWERTOOLS_SERVICE_NAME  = "order-service"
      POWERTOOLS_METRICS_NAMESPACE = "OrderService"
      LOG_LEVEL                = "INFO"
    }
  }

  tracing_config {
    mode = "Active"  # Enable X-Ray tracing
  }
}

resource "aws_iam_role_policy" "lambda_dynamodb" {
  role = aws_iam_role.lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:Query",
        ]
        Resource = aws_dynamodb_table.orders.arn
      },
      {
        Effect = "Allow"
        Action = [
          "xray:PutTraceSegments",
          "xray:PutTelemetryRecords"
        ]
        Resource = "*"
      }
    ]
  })
}
```

## Generator Configuration

Each generator accepts a config struct:

```go
type ProjectConfig struct {
    ServiceName    string  // e.g., "order-service"
    FunctionName   string  // e.g., "create-order"
    Description    string  // Service description
    PythonVersion  string  // "3.13"
    UsePowertools  bool    // Include AWS Powertools
    UseIdempotency bool    // Add idempotency support
    UseDynamoDB    bool    // Generate DynamoDB resources
    TableName      string  // DynamoDB table name
    APIPath        string  // API Gateway path
    HTTPMethod     string  // HTTP method (GET, POST, etc.)
}
```

## Implementation

### Pure Functions

All generators use **pure functions** for code generation:

```go
// generatePyProjectToml creates pyproject.toml content (PURE)
func generatePyProjectToml(config ProjectConfig) string {
    dependencies := []string{
        fmt.Sprintf("python = \"^%s\"", config.PythonVersion),
    }

    if config.UsePowertools {
        dependencies = append(dependencies, "aws-lambda-powertools = \"^2.30.0\"")
    }

    // ... build TOML string
    return tomlContent
}
```

### Testing

```go
func TestPythonGenerator(t *testing.T) {
    tmpDir := t.TempDir()

    config := python.ProjectConfig{
        ServiceName:   "test-service",
        FunctionName:  "test-function",
        PythonVersion: "3.13",
        UsePowertools: true,
    }

    err := python.Generate(tmpDir, config)

    assert.NoError(t, err)
    assert.FileExists(t, filepath.Join(tmpDir, "pyproject.toml"))
    assert.FileExists(t, filepath.Join(tmpDir, "service/handlers/__init__.py"))
    assert.FileExists(t, filepath.Join(tmpDir, "tests/unit/test_handler.py"))

    // Verify dependencies
    content, _ := os.ReadFile(filepath.Join(tmpDir, "pyproject.toml"))
    assert.Contains(t, string(content), "aws-lambda-powertools")
}
```

## Files

- **`python/`** - Python generator
  - `project.go` - Main generator function
  - `models.go` - Pydantic model templates
  - `terraform.go` - Terraform generation
  - `example_test.go` - Generator tests

## Design Principles

1. **Opinionated architecture** - enforces team standards
2. **Production-ready** - observability, error handling, testing from day 1
3. **Type safety** - Pydantic models, mypy configuration
4. **Layered architecture** - handlers, logic, DAL separation
5. **Testable** - 100% test coverage with passing tests

## Future Generators

- [ ] **Node.js/TypeScript Generator** - Hexagonal architecture with Zod validation
- [ ] **Go Generator** - Clean architecture with domain-driven design
- [ ] **Event-Driven Generator** - SQS/SNS/EventBridge integration
- [ ] **GraphQL API Generator** - AppSync + resolvers
- [ ] **Step Functions Generator** - State machines + orchestration
- [ ] **Custom Generator Plugin System** - Teams define their own templates

## Customization

Teams can fork generators and customize:

1. **Architecture patterns** - Add your team's preferred structure
2. **Dependencies** - Include internal libraries, custom SDKs
3. **Testing frameworks** - Use your preferred tools
4. **Observability** - Integrate with Datadog, New Relic, etc.
5. **Security** - Add secrets management, encryption patterns

**Example enterprise customization:**

```go
// Custom generator for Acme Corp
config := python.ProjectConfig{
    ServiceName:     "order-service",
    UsePowertools:   true,
    UseAcmeLogger:   true,    // Internal logging library
    UseAcmeMetrics:  true,    // Internal metrics
    SecretsManager:  "vault", // HashiCorp Vault integration
    DeploymentMode:  "blue-green",
}
```

This enables **approved patterns** across the organization while maintaining flexibility.
