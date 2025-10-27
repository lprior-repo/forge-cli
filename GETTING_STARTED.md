# Forge - Getting Started Guide

Forge is a modern serverless infrastructure tool that combines Terraform's power with streamlined Lambda deployment workflows. This guide will help you get started quickly.

## Prerequisites

- Go 1.21+ installed
- AWS credentials configured (`aws configure`)
- Terraform 1.0+ installed (optional for development, Forge can manage this)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/your-org/forge.git
cd forge

# Build the binary
go build -o forge ./cmd/forge

# Add to PATH (optional)
sudo mv forge /usr/local/bin/
```

### Verify Installation

```bash
forge version
```

## Quick Start

### 1. Create a New Project

Create a new serverless project with a single command:

```bash
# Create a new project with Go runtime
forge new my-api

# Create with Python runtime
forge new my-api --runtime python3.11

# Create with Node.js runtime
forge new my-api --runtime nodejs20.x
```

This creates:
```
my-api/
├── forge.hcl           # Main configuration file
├── functions/
│   └── hello/
│       └── main.go     # Function source code
└── README.md
```

### 2. Configure Your Application

Edit `forge.hcl` to define your serverless architecture:

```hcl
service = "my-api"

provider = {
  region = "us-east-1"
}

functions = {
  hello = {
    handler = "main"
    runtime = "go1.x"
    source  = {
      path = "./functions/hello"
    }
    http_routing = {
      path   = "/hello"
      method = "GET"
    }
  }
}

api_gateway = {
  name          = "my-api-gateway"
  protocol_type = "HTTP"
}
```

### 3. Deploy Your Application

```bash
cd my-api

# Initialize Terraform
forge init

# Deploy everything
forge deploy
```

Forge will:
1. Build your Lambda functions
2. Generate Terraform configuration
3. Plan infrastructure changes
4. Apply changes to AWS

### 4. Test Your API

After deployment, Forge outputs the API Gateway URL:

```bash
curl https://your-api-id.execute-api.us-east-1.amazonaws.com/hello
```

## Core Concepts

### Configuration File (`forge.hcl`)

Forge uses HCL (HashiCorp Configuration Language) for infrastructure definitions.

#### Basic Structure

```hcl
service = "my-service"

provider = {
  region  = "us-east-1"
  profile = "default"  # Optional AWS profile
}

functions = {
  # Function definitions
}

api_gateway = {
  # API Gateway configuration
}

tables = {
  # DynamoDB tables
}
```

### Functions

Define Lambda functions with comprehensive configuration:

```hcl
functions = {
  api = {
    # Core Configuration
    handler     = "index.handler"
    runtime     = "nodejs20.x"
    timeout     = 30
    memory_size = 256
    description = "API handler"

    # Source Code
    source = {
      path = "./functions/api"
    }

    # Environment Variables
    environment = {
      NODE_ENV = "production"
      API_KEY  = "your-api-key"
    }

    # VPC Configuration
    vpc = {
      subnet_ids         = ["subnet-123", "subnet-456"]
      security_group_ids = ["sg-789"]
    }

    # Lambda Layers
    layers = [
      "arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1"
    ]

    # Dead Letter Queue
    dead_letter_config = {
      target_arn = "arn:aws:sqs:us-east-1:123456789012:dlq"
    }

    # X-Ray Tracing
    tracing_mode = "Active"

    # Reserved Concurrency
    reserved_concurrent_executions = 10

    # Architecture
    architectures = ["arm64"]  # or ["x86_64"]

    # Tags
    tags = {
      Environment = "production"
      Team        = "platform"
    }

    # HTTP Routing (API Gateway)
    http_routing = {
      path          = "/api"
      method        = "POST"
      authorizer_id = "cognito-auth"
    }
  }
}
```

### API Gateway

Configure HTTP and WebSocket APIs:

```hcl
api_gateway = {
  name          = "my-api"
  protocol_type = "HTTP"
  description   = "Main API"

  # CORS Configuration
  cors = {
    allow_origins     = ["https://example.com"]
    allow_methods     = ["GET", "POST", "PUT", "DELETE"]
    allow_headers     = ["Content-Type", "Authorization"]
    allow_credentials = true
    max_age           = 3600
  }

  # Authorizers
  authorizers = {
    cognito = {
      type            = "JWT"
      identity_source = ["$request.header.Authorization"]
      jwt_configuration = {
        issuer   = "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_123456789"
        audience = ["client-id-123"]
      }
    }
  }

  # Throttling
  default_route_settings = {
    throttling_burst_limit = 500
    throttling_rate_limit  = 1000
  }

  # Access Logs
  access_logs = {
    destination_arn = "arn:aws:logs:us-east-1:123456789012:log-group:/aws/apigateway/my-api"
    format          = "$requestId $context.error.message"
  }

  # Tags
  tags = {
    Environment = "production"
  }
}
```

### DynamoDB Tables

```hcl
tables = {
  users = {
    billing_mode = "PAY_PER_REQUEST"
    hash_key     = "userId"

    attributes = [
      { name = "userId", type = "S" },
      { name = "email", type = "S" }
    ]

    global_secondary_indexes = [
      {
        name            = "email-index"
        hash_key        = "email"
        projection_type = "ALL"
      }
    ]

    # Streams for DynamoDB triggers
    stream_enabled   = true
    stream_view_type = "NEW_AND_OLD_IMAGES"

    # Point-in-time recovery
    point_in_time_recovery = true

    # Tags
    tags = {
      Environment = "production"
    }
  }
}
```

### Event Source Mappings

Connect Lambda to event sources:

```hcl
functions = {
  processor = {
    handler = "index.handler"
    runtime = "nodejs20.x"
    source  = {
      path = "./functions/processor"
    }

    # DynamoDB Stream
    event_source_mappings = [
      {
        event_source_arn       = "arn:aws:dynamodb:us-east-1:123456789012:table/users/stream/2021-01-01T00:00:00.000"
        starting_position      = "LATEST"
        batch_size             = 100
        parallelization_factor = 2
        maximum_retry_attempts = 3

        # Event Filtering
        filter_criteria = {
          filters = [
            { pattern = "{\"eventName\": [\"INSERT\", \"MODIFY\"]}" }
          ]
        }

        # Failure Destination
        destination_config = {
          on_failure = {
            destination = "arn:aws:sqs:us-east-1:123456789012:dlq"
          }
        }
      }
    ]
  }
}
```

### Lambda Aliases

Implement blue/green deployments:

```hcl
functions = {
  api = {
    handler = "index.handler"
    runtime = "nodejs20.x"
    source  = { path = "./functions/api" }
    publish = true  # Enable versioning

    aliases = [
      {
        name             = "live"
        function_version = "2"
        description      = "Live traffic"

        # Weighted routing (blue/green)
        routing_config = {
          additional_version_weights = {
            "1" = 0.1  # 10% to version 1
          }
        }
      },
      {
        name             = "dev"
        function_version = "$LATEST"
        description      = "Development alias"
      }
    ]
  }
}
```

## Advanced Features

### Container-Based Functions

```hcl
functions = {
  api = {
    package_type = "Image"
    image_uri    = "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-function:latest"

    image_config = {
      entry_point       = ["/app/bootstrap"]
      command           = ["handler"]
      working_directory = "/app"
    }
  }
}
```

### Lambda@Edge

```hcl
functions = {
  edge = {
    handler       = "index.handler"
    runtime       = "nodejs20.x"
    lambda_at_edge = true
    source        = { path = "./functions/edge" }
  }
}
```

### EFS File Systems

```hcl
functions = {
  processor = {
    handler = "index.handler"
    runtime = "python3.11"
    source  = { path = "./functions/processor" }

    file_system_configs = [
      {
        arn              = "arn:aws:elasticfilesystem:us-east-1:123456789012:access-point/fsap-123"
        local_mount_path = "/mnt/data"
      }
    ]
  }
}
```

## Commands

### Create New Project

```bash
# New project with Go
forge new my-api

# New project with specific runtime
forge new my-api --runtime python3.11

# Add stack to existing project
forge new --stack compute --runtime go1.x
```

### Initialize Terraform

```bash
# Initialize all stacks
forge init

# With verbose output
forge init --verbose
```

### Deploy

```bash
# Deploy all stacks
forge deploy

# Deploy specific stack
forge deploy --stack api

# With auto-approve (skip confirmation)
forge deploy --auto-approve

# Dry run (plan only)
forge deploy --plan-only
```

### Destroy

```bash
# Destroy all infrastructure
forge destroy

# Destroy specific stack
forge destroy --stack api

# With auto-approve
forge destroy --auto-approve
```

### Version

```bash
forge version
```

## Project Structure

### Generated Project

```
my-api/
├── forge.hcl                  # Main configuration
├── functions/
│   ├── hello/
│   │   ├── main.go           # Go function
│   │   └── go.mod
│   ├── api/
│   │   ├── index.js          # Node.js function
│   │   └── package.json
│   └── processor/
│       ├── handler.py        # Python function
│       └── requirements.txt
├── stacks/
│   └── main/                 # Generated Terraform
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
└── .terraform/               # Terraform state
```

### Multi-Stack Projects

```
my-project/
├── forge.hcl                 # Root config
├── stacks/
│   ├── api/
│   │   ├── forge.hcl        # API stack config
│   │   └── functions/
│   ├── worker/
│   │   ├── forge.hcl        # Worker stack config
│   │   └── functions/
│   └── data/
│       └── forge.hcl        # Data stack config
└── shared/
    └── layers/              # Shared Lambda layers
```

## Best Practices

### 1. Use Environment Variables

Store sensitive data in environment variables, not in `forge.hcl`:

```hcl
functions = {
  api = {
    environment = {
      DB_HOST     = "db.example.com"
      DB_PASSWORD = env("DB_PASSWORD")  # From environment
    }
  }
}
```

### 2. Leverage Lambda Layers

Share code across functions:

```hcl
functions = {
  api = {
    layers = [
      "arn:aws:lambda:us-east-1:123456789012:layer:shared-utils:1"
    ]
  }
}
```

### 3. Use VPC for Private Resources

```hcl
functions = {
  api = {
    vpc = {
      subnet_ids         = ["subnet-private-1", "subnet-private-2"]
      security_group_ids = ["sg-lambda"]
    }
  }
}
```

### 4. Enable Tracing

```hcl
functions = {
  api = {
    tracing_mode = "Active"
  }
}
```

### 5. Set Appropriate Timeouts and Memory

```hcl
functions = {
  fast_api = {
    timeout     = 5
    memory_size = 256
  }

  batch_processor = {
    timeout     = 900   # 15 minutes
    memory_size = 3008  # More memory = more CPU
  }
}
```

## Troubleshooting

### Build Failures

```bash
# Check build logs
forge deploy --verbose

# Verify function source
ls -la functions/my-function/
```

### Deployment Errors

```bash
# Check Terraform state
cd stacks/main
terraform state list

# View detailed errors
forge deploy --verbose
```

### Permission Issues

Ensure your AWS credentials have necessary permissions:
- Lambda: Full access
- IAM: Create roles and policies
- API Gateway: Full access
- DynamoDB: Full access
- CloudWatch Logs: Create log groups

### Function Invocation Errors

```bash
# Check CloudWatch Logs
aws logs tail /aws/lambda/my-service-my-function --follow

# Test function locally
cd functions/my-function
go run main.go
```

## Examples

### Simple REST API

```hcl
service = "todo-api"

provider = {
  region = "us-east-1"
}

functions = {
  create = {
    handler = "index.handler"
    runtime = "nodejs20.x"
    source  = { path = "./functions/create" }
    http_routing = {
      path   = "/todos"
      method = "POST"
    }
  }

  list = {
    handler = "index.handler"
    runtime = "nodejs20.x"
    source  = { path = "./functions/list" }
    http_routing = {
      path   = "/todos"
      method = "GET"
    }
  }
}

api_gateway = {
  name          = "todo-api"
  protocol_type = "HTTP"
}

tables = {
  todos = {
    billing_mode = "PAY_PER_REQUEST"
    hash_key     = "id"
    attributes   = [{ name = "id", type = "S" }]
  }
}
```

### Event-Driven Processing

```hcl
service = "data-pipeline"

provider = {
  region = "us-east-1"
}

functions = {
  ingest = {
    handler = "main"
    runtime = "go1.x"
    source  = { path = "./functions/ingest" }
    http_routing = {
      path   = "/ingest"
      method = "POST"
    }
  }

  processor = {
    handler = "main"
    runtime = "go1.x"
    source  = { path = "./functions/processor" }
    event_source_mappings = [
      {
        event_source_arn  = "arn:aws:dynamodb:us-east-1:123456789012:table/events/stream/..."
        starting_position = "LATEST"
        batch_size        = 100
      }
    ]
  }
}

tables = {
  events = {
    billing_mode       = "PAY_PER_REQUEST"
    hash_key           = "id"
    stream_enabled     = true
    stream_view_type   = "NEW_AND_OLD_IMAGES"
    attributes         = [{ name = "id", type = "S" }]
  }
}
```

## Next Steps

1. **Explore Examples**: Check the `examples/` directory for more use cases
2. **Read Documentation**: See `docs/` for detailed API reference
3. **Join Community**: Star the repo and open issues for questions
4. **Contribute**: PRs welcome!

## Resources

- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/)
- [API Gateway Documentation](https://docs.aws.amazon.com/apigateway/)
- [DynamoDB Documentation](https://docs.aws.amazon.com/dynamodb/)
- [Lingon Documentation](https://github.com/golingon/lingon)

## Support

- **Issues**: [GitHub Issues](https://github.com/your-org/forge/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/forge/discussions)
- **Email**: support@forge.dev

---

Happy Building! 🚀
