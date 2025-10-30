# internal/lingon

**Type-safe Terraform generation using Lingon with YAML configuration support.**

## Overview

The `lingon` package provides declarative infrastructure-as-code capabilities by converting YAML configurations (`forge.yaml`) into type-safe Terraform code using the Lingon library. It supports the complete AWS Lambda serverless stack with 170+ Lambda parameters, 80+ API Gateway parameters, and 50+ DynamoDB parameters.

## Architecture

```
┌──────────────────────────────────────────────────┐
│          Lingon Generator (Pure Core)            │
│   - Reads forge.yaml configuration               │
│   - Validates resource specifications            │
│   - Generates type-safe Terraform via Lingon     │
└──────────────────────────────────────────────────┘
                      ↓
    ┌─────────────────┴─────────────────┐
    ↓                                   ↓
┌─────────────────┐           ┌─────────────────┐
│  Config Types   │           │   AWS Resources │
│  (Pure Data)    │           │   (Lingon/Terra)│
└─────────────────┘           └─────────────────┘
```

## Key Types

```go
// ForgeConfig represents the complete forge.yaml structure (PURE DATA)
type ForgeConfig struct {
    Service   string                        // Service name
    Provider  ProviderConfig                // AWS provider config
    Functions map[string]FunctionConfig     // Lambda functions
    Tables    map[string]DynamoDBConfig     // DynamoDB tables
    Queues    map[string]SQSConfig          // SQS queues
    Topics    map[string]SNSConfig          // SNS topics
    Buckets   map[string]S3Config           // S3 buckets
}

// FunctionConfig supports 170+ Lambda parameters
type FunctionConfig struct {
    Handler      string            // Handler function
    Runtime      string            // Lambda runtime
    Timeout      int               // Timeout in seconds
    MemorySize   int               // Memory in MB
    Environment  map[string]string // Environment variables
    // ... 165+ more parameters
}

// Generator provides Terraform generation
type Generator struct {
    Generate GeneratorFunc
}

type GeneratorFunc func(ctx context.Context, config ForgeConfig) E.Either[error, []byte]
```

## Core Functions

### Generate Terraform from Config

```go
// Create generator
gen := lingon.NewGenerator()

// Load configuration
config, err := lingon.LoadForgeConfig("forge.yaml")
if err != nil {
    log.Fatal(err)
}

// Generate Terraform (returns Either)
result := gen.Generate(context.Background(), config)

// Handle result
E.Fold(
    func(err error) {
        log.Fatalf("Generation failed: %v", err)
    },
    func(terraform []byte) {
        os.WriteFile("main.tf", terraform, 0644)
    },
)(result)
```

### Variable References

Use `${}` syntax to reference other resources:

```yaml
functions:
  api:
    handler: index.handler
    runtime: nodejs20.x
    environment:
      TABLE_NAME: ${tables.users.name}    # Reference DynamoDB table
      QUEUE_URL: ${queues.jobs.url}       # Reference SQS queue
      BUCKET: ${buckets.assets.bucket}    # Reference S3 bucket

tables:
  users:
    hashKey: userId
    name: ${service}-users-${var.environment}
```

## Supported AWS Resources

### Lambda Functions
- **170+ parameters** - Complete AWS Lambda configuration
- Handler, Runtime, Timeout, Memory, Layers
- VPC, IAM, Environment variables
- Tracing, Reserved concurrency, Dead letter queues

### API Gateway v2 (HTTP)
- **80+ parameters** - Complete API Gateway HTTP configuration
- Routes, Integrations, Authorizers
- CORS, Custom domains, Stages
- Throttling, Logging, Access logs

### DynamoDB Tables
- **50+ parameters** - Complete DynamoDB configuration
- Hash/Range keys, Global/Local indexes
- Billing mode, Capacity, Auto-scaling
- Streams, TTL, Point-in-time recovery

### Additional Resources
- **SQS Queues** - Standard and FIFO queues
- **SNS Topics** - Pub/sub messaging
- **S3 Buckets** - Object storage with versioning

## Functional Design

### Pure Calculation Functions

All generation logic is **pure** - same input → same output:

```go
// Pure function - no I/O, no mutation
func generateLambdaResource(fn FunctionConfig) *terra.Resource {
    return &terra.Resource{
        Type: "aws_lambda_function",
        Name: fn.Name,
        // ...
    }
}

// Pure validation
func validateConfig(config ForgeConfig) E.Either[error, ForgeConfig] {
    if config.Service == "" {
        return E.Left[ForgeConfig](fmt.Errorf("service name required"))
    }
    return E.Right[error](config)
}
```

### I/O at Boundaries

File operations are isolated to action functions:

```go
// ACTION - reads file from disk
func LoadForgeConfig(path string) (ForgeConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return ForgeConfig{}, err
    }
    // Parse YAML...
}

// ACTION - writes Terraform to disk
func WriteGeneratedTerraform(terraform []byte, path string) error {
    return os.WriteFile(path, terraform, 0644)
}
```

## Complete Specification

See `LINGON_SPEC.md` for:
- Complete parameter reference (1,500+ lines)
- All supported AWS resource types
- Variable reference syntax
- Example configurations
- Best practices

## Example Configuration

```yaml
service: my-app

provider:
  region: us-east-1
  profile: default

functions:
  api:
    handler: index.handler
    runtime: nodejs20.x
    timeout: 30
    memorySize: 1024
    environment:
      TABLE_NAME: ${tables.users.name}
      NODE_ENV: production
    vpc:
      securityGroupIds:
        - ${var.security_group_id}
      subnetIds:
        - ${var.subnet_a}
        - ${var.subnet_b}

tables:
  users:
    hashKey: userId
    rangeKey: createdAt
    billingMode: PAY_PER_REQUEST
    streamEnabled: true
    streamViewType: NEW_AND_OLD_IMAGES
    attributes:
      - name: userId
        type: S
      - name: createdAt
        type: N
      - name: email
        type: S
    globalSecondaryIndexes:
      - name: EmailIndex
        hashKey: email
        projectionType: ALL
```

## Related Packages

- **internal/config** - HCL configuration loading (alternative to YAML)
- **internal/terraform** - Terraform executor wrapper
- **internal/scaffold** - Project scaffolding with forge.yaml generation

## Design Principles

1. **Type Safety** - Lingon provides compile-time type checking for Terraform
2. **Declarative** - Infrastructure defined in YAML, not imperative code
3. **Pure Core** - All generation logic is pure functions
4. **Complete Coverage** - 300+ AWS parameters across all resource types
5. **Composable** - Resources can reference each other via variables
