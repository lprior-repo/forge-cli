# tfmodules - Type-Safe Terraform Modules for Lingon

This package extends [Lingon](https://github.com/golingon/lingon) to support terraform-aws-modules with full type safety and compile-time validation.

## Overview

Lingon provides type-safe Terraform resource definitions for AWS providers. This package extends that concept to **Terraform modules**, giving you the same compile-time safety and IDE support for terraform-aws-modules.

## Problem

Using Terraform modules traditionally requires string-based configuration:

```hcl
module "my_queue" {
  source = "terraform-aws-modules/sqs/aws"
  version = "~> 4.0"

  name = "orders-queue"
  visibility_timeout_seconds = 30
  message_retention_seconds = 345600
  create_dlq = true
}
```

When generating this from Go, you typically use `map[string]interface{}`:

```go
// ❌ No type safety
Variables: map[string]interface{}{
    "visibility_timeout_seconds": 30,
    "message_retention_seconds":  345600,
    "create_dlq":                 true,
}
```

**Problems:**
- ❌ No compile-time type checking
- ❌ No IDE autocomplete
- ❌ Easy to make typos in variable names
- ❌ No validation until Terraform runs
- ❌ Not self-documenting

## Solution: Lingon-Style Type-Safe Modules

```go
import "github.com/lewis/forge/internal/tfmodules/sqs"

// ✅ Fully type-safe
queue := sqs.NewModule("orders_queue")

visibility := 30
retention := 345600
createDLQ := true

queue.VisibilityTimeoutSeconds = &visibility
queue.MessageRetentionSeconds = &retention
queue.CreateDLQ = &createDLQ
```

**Benefits:**
- ✅ Compile-time type checking
- ✅ Full IDE autocomplete
- ✅ Impossible to mistype variable names
- ✅ Validation rules in struct tags
- ✅ Self-documenting with comments from module docs
- ✅ Fluent builder API for common patterns

## Features

### 1. Strongly-Typed Configuration

All 60+ variables from the SQS module are defined as struct fields with proper Go types:

```go
type Module struct {
    // The visibility timeout for the queue. An integer from 0 to 43200 (12 hours)
    VisibilityTimeoutSeconds *int `validate:"min=0,max=43200"`

    // The number of seconds Amazon SQS retains a message
    MessageRetentionSeconds *int `validate:"min=60,max=1209600"`

    // Boolean designating a FIFO queue
    FifoQueue *bool

    // ... 57 more fields, all strongly typed
}
```

### 2. Fluent Builder API

Common configuration patterns have dedicated methods:

```go
queue := sqs.NewModule("orders_queue").
    WithFIFO(true).
    WithEncryption("arn:aws:kms:us-east-1:123456789012:key/12345").
    WithTags(map[string]string{
        "Environment": "production",
        "ManagedBy":   "forge",
    })
```

### 3. Type-Safe Output References

Module outputs are type-safe references that can be used in other resources:

```go
queue := sqs.NewModule("orders_queue")

// Type-safe output reference
queueARN := tfmodules.NewOutput(queue, "queue_arn")

// Use in Lambda environment variables, IAM policies, etc.
envVars := map[string]string{
    "QUEUE_ARN": queueARN.Ref().String(),
}

// Generates: "module.orders_queue.queue_arn"
```

### 4. Sensible Defaults

NewModule() provides production-ready defaults:

```go
queue := sqs.NewModule("my_queue")
// Automatically configured with:
// - 30s visibility timeout
// - 4 days message retention
// - DLQ enabled (14 days retention)
// - SQS-managed encryption
```

### 5. Compile-Time Validation

Struct tags enable validation:

```go
type Module struct {
    // Valid range enforced at compile-time with tags
    VisibilityTimeoutSeconds *int `validate:"min=0,max=43200"`
    DelaySeconds *int `validate:"min=0,max=900"`
}
```

## Supported Modules

| Module | Package | Status | Variables |
|--------|---------|--------|-----------|
| SQS | `tfmodules/sqs` | ✅ Complete | 60+ |
| DynamoDB | `tfmodules/dynamodb` | ✅ Complete | 48 |
| SNS | `tfmodules/sns` | ✅ Complete | 45 |
| S3 | `tfmodules/s3` | ✅ Complete | 90+ |
| Lambda | `tfmodules/lambda` | ✅ Complete | 100+ |
| API Gateway V2 | `tfmodules/apigatewayv2` | ✅ Complete | 70+ |
| EventBridge | `tfmodules/eventbridge` | ✅ Complete | 80+ |
| Step Functions | `tfmodules/stepfunctions` | ✅ Complete | 40+ |
| Secrets Manager | `tfmodules/secretsmanager` | ✅ Complete | 35+ |
| SSM Parameter | `tfmodules/ssm` | ✅ Complete | 15 |
| AppConfig | `tfmodules/appconfig` | ✅ Complete | 50+ |
| **CloudFront** | `tfmodules/cloudfront` | ✅ Complete | 30+ |
| **AppSync** | `tfmodules/appsync` | ✅ Complete | 50+ |

## Usage Examples

### Example 1: Standalone SQS Queue

```go
package main

import "github.com/lewis/forge/internal/tfmodules/sqs"

func main() {
    queue := sqs.NewModule("orders_queue")

    // Customize as needed
    visibility := 60
    queue.VisibilityTimeoutSeconds = &visibility

    // Configure DLQ retention
    dlqRetention := 1209600 // 14 days
    queue.DLQMessageRetentionSeconds = &dlqRetention
}
```

### Example 2: FIFO Queue with Encryption

```go
queue := sqs.NewModule("transactions_fifo").
    WithFIFO(true).
    WithEncryption("alias/aws/sqs")
```

### Example 3: Queue for Lambda Integration

```go
import (
    "github.com/lewis/forge/internal/tfmodules"
    "github.com/lewis/forge/internal/tfmodules/sqs"
)

func main() {
    // Create queue
    queue := sqs.NewModule("events_queue")

    // Get output references for Lambda
    queueARN := tfmodules.NewOutput(queue, "queue_arn")
    queueURL := tfmodules.NewOutput(queue, "queue_url")

    // Use in Lambda event source mapping
    eventSource := map[string]interface{}{
        "event_source_arn": queueARN.Ref().String(),
        "batch_size":       10,
    }
}
```

### Example 4: DynamoDB Table with Streams

```go
import "github.com/lewis/forge/internal/tfmodules/dynamodb"

func main() {
    table := dynamodb.NewModule("users")

    // Configure primary key
    table.WithHashKey("userId", "S").
        WithRangeKey("timestamp", "N")

    // Enable streams for Lambda triggers
    table.WithStreams("NEW_AND_OLD_IMAGES")

    // Add Global Secondary Index
    gsi := dynamodb.GlobalSecondaryIndex{
        Name:           "email-index",
        HashKey:        "email",
        ProjectionType: "ALL",
    }
    table.WithGSI(gsi)

    // Enable TTL
    table.WithTTL("expiresAt")
}
```

### Example 5: SNS Topic with Subscriptions

```go
import "github.com/lewis/forge/internal/tfmodules/sns"

func main() {
    topic := sns.NewModule("notifications")

    // Configure as FIFO topic
    topic.WithFIFO(true).WithEncryption("alias/aws/sns")

    // Add Lambda subscription
    topic.WithLambdaSubscription(
        "processor",
        "arn:aws:lambda:us-east-1:123456789012:function:processor",
    )

    // Add SQS subscription with raw message delivery
    topic.WithSQSSubscription(
        "queue_sub",
        "arn:aws:sqs:us-east-1:123456789012:queue",
        true, // raw message delivery
    )
}
```

### Example 6: S3 Bucket with Security Features

```go
import "github.com/lewis/forge/internal/tfmodules/s3"

func main() {
    bucket := s3.NewModule("secure-data")

    // Enable versioning and encryption
    bucket.WithVersioning(true).
        WithEncryption("arn:aws:kms:us-east-1:123456789012:key/12345")

    // Configure access logging
    bucket.WithLogging("logs-bucket", "secure-data/")

    // Add CORS rules
    bucket.WithCORS(
        []string{"https://example.com"},
        []string{"GET", "PUT"},
        []string{"*"},
    )

    // Public access is blocked by default
}
```

### Example 7: Lambda Function with VPC

```go
import "github.com/lewis/forge/internal/tfmodules/lambda"

func main() {
    fn := lambda.NewModule("api_handler")

    // Configure runtime and resources
    fn.WithRuntime("python3.13", "app.handler").
        WithMemoryAndTimeout(1024, 30)

    // VPC configuration
    fn.WithVPC(
        []string{"subnet-123", "subnet-456"},
        []string{"sg-789"},
    )

    // Environment variables
    fn.WithEnvironment(map[string]string{
        "TABLE_NAME": "users",
        "REGION":     "us-east-1",
    })

    // Enable tracing
    fn.WithTracing("Active")

    // Add layers
    fn.WithLayers(
        "arn:aws:lambda:us-east-1:123456789012:layer:common:1",
    )
}
```

### Example 8: API Gateway V2 with JWT Auth

```go
import "github.com/lewis/forge/internal/tfmodules/apigatewayv2"

func main() {
    api := apigatewayv2.NewModule("my_api")

    // Configure CORS
    api.WithCORS(
        []string{"https://example.com"},
        []string{"GET", "POST", "PUT", "DELETE"},
        []string{"Content-Type", "Authorization"},
    )

    // Add JWT authorizer
    api.WithJWTAuthorizer(
        "cognito",
        "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_ABC123",
        []string{"client-id-123"},
    )

    // Custom domain
    api.WithDomainName(
        "api.example.com",
        "arn:aws:acm:us-east-1:123456789012:certificate/abc123",
    )
}
```

### Example 9: EventBridge Integration Patterns

```go
import "github.com/lewis/forge/internal/tfmodules/eventbridge"

func main() {
    bus := eventbridge.NewModule("orders")

    // Pattern 1: Event-driven Lambda integration
    bus.WithEventPatternRule("order_created", "Process new orders", `{
        "source": ["order.service"],
        "detail-type": ["Order Created"]
    }`, true).
        WithLambdaTarget("order_created", "arn:aws:lambda:us-east-1:123:function:process-order")

    // Pattern 2: Schedule-driven Step Functions integration
    bus.WithScheduleRule("daily_report", "Generate daily report", "cron(0 9 * * ? *)", true).
        WithStepFunctionsTarget("daily_report", "arn:aws:states:us-east-1:123:stateMachine:report")

    // Pattern 3: Fan-out to SQS and SNS
    bus.WithEventPatternRule("inventory_low", "Low inventory alert", `{
        "source": ["inventory.service"],
        "detail-type": ["Stock Low"]
    }`, true).
        WithSQSTarget("inventory_low", "arn:aws:sqs:us-east-1:123:queue/restock").
        WithSNSTarget("inventory_low", "arn:aws:sns:us-east-1:123:topic/alerts")

    // Pattern 4: Kinesis stream integration for real-time analytics
    bus.WithEventPatternRule("user_activity", "Track user events", `{
        "source": ["app.frontend"],
        "detail-type": ["User Action"]
    }`, true).
        WithKinesisTarget("user_activity", "arn:aws:kinesis:us-east-1:123:stream/analytics")

    // Pattern 5: ECS task integration for batch processing
    bus.WithEventPatternRule("batch_job", "Trigger batch processing", `{
        "source": ["batch.scheduler"],
        "detail-type": ["Job Ready"]
    }`, true).
        WithECSTarget("batch_job",
            "arn:aws:ecs:us-east-1:123:cluster/batch-cluster",
            "arn:aws:ecs:us-east-1:123:task-definition/processor:1",
            []string{"subnet-abc123", "subnet-def456"},
        )

    // Pattern 6: API Destination for HTTP webhooks
    bus.WithEventPatternRule("webhook_event", "Send to external API", `{
        "source": ["webhook.trigger"]
    }`, true).
        WithAPIDestinationTarget("webhook_event", "arn:aws:events:us-east-1:123:destination/webhook")
}
```

### Example 10: Step Functions Workflow

```go
import "github.com/lewis/forge/internal/tfmodules/stepfunctions"

func main() {
    workflow := stepfunctions.NewModule("order_processor")

    // State machine definition
    definition := `{
        "StartAt": "ProcessOrder",
        "States": {
            "ProcessOrder": {
                "Type": "Task",
                "Resource": "arn:aws:lambda:us-east-1:123456789012:function:process",
                "End": true
            }
        }
    }`

    workflow.WithDefinition(definition).
        WithLogging("ALL", true).
        WithTracing().
        WithLambdaIntegration(
            "arn:aws:lambda:us-east-1:123456789012:function:process",
        )
}
```

### Example 11: Secrets Manager with Rotation

```go
import "github.com/lewis/forge/internal/tfmodules/secretsmanager"

func main() {
    secret := secretsmanager.NewModule("db_credentials")

    // Store database credentials as JSON
    credentials := `{
        "username": "admin",
        "password": "changeme",
        "host": "db.example.com"
    }`

    secret.WithSecretJSON(credentials).
        WithKMSKey("arn:aws:kms:us-east-1:123456789012:key/abc123").
        WithRotation(
            "arn:aws:lambda:us-east-1:123456789012:function:rotate",
            30, // days
        ).
        WithReplication(
            "us-west-2",
            "arn:aws:kms:us-west-2:123456789012:key/def456",
        )
}
```

### Example 12: SSM Parameter Store

```go
import "github.com/lewis/forge/internal/tfmodules/ssm"

func main() {
    // Simple string parameter
    apiKey := ssm.NewModule("/myapp/api_key").
        WithSecureString(
            "secret-api-key-value",
            "alias/aws/ssm",
        )

    // StringList parameter
    endpoints := ssm.NewModule("/myapp/endpoints").
        WithStringList([]string{
            "https://api1.example.com",
            "https://api2.example.com",
        })

    // Advanced tier for large values
    config := ssm.NewModule("/myapp/config").
        WithValue(largeJSONConfig).
        WithAdvancedTier().
        WithValidation("^\\{.*\\}$") // Must be valid JSON
}
```

### Example 13: AppConfig Feature Flags

```go
import "github.com/lewis/forge/internal/tfmodules/appconfig"

func main() {
    app := appconfig.NewModule("myapp")

    // Add production environment
    app.WithEnvironment("production", appconfig.Environment{
        Name: "production",
        Monitors: []appconfig.Monitor{
            {AlarmARN: "arn:aws:cloudwatch:us-east-1:123:alarm:api-errors"},
        },
    })

    // Feature flags configuration
    featureFlags := `{
        "flags": {
            "new_checkout": {
                "name": "new_checkout",
                "description": "Enable new checkout flow",
                "_deprecation": {"status": "planned"},
                "attributes": {
                    "enabled": {"constraints": {"type": "boolean"}}
                }
            }
        },
        "values": {
            "new_checkout": {"enabled": true}
        },
        "version": "1"
    }`

    app.WithFeatureFlags(featureFlags).
        WithDeploymentStrategy(
            10,   // 10 min duration
            20.0, // 20% growth factor
            5,    // 5 min bake time
        )
}
```

### Example 14: CloudFront Distribution with S3 Origin

```go
import "github.com/lewis/forge/internal/tfmodules/cloudfront"

func main() {
    cdn := cloudfront.NewModule("My Static Website")

    // Configure S3 origin with Origin Access Control
    cdn.WithOriginAccessControl("s3_oac", "S3 OAC for static website").
        WithS3Origin(
            "s3_origin",
            "my-bucket.s3.amazonaws.com",
            "origin-access-identity/cloudfront/ABCDEFG1234567",
        )

    // Default cache behavior
    cdn.WithDefaultCacheBehavior("s3_origin", "redirect-to-https").
        WithAliases("example.com", "www.example.com").
        WithCertificate(
            "arn:aws:acm:us-east-1:123456789012:certificate/abc123",
            "TLSv1.2_2021",
        )

    // Enable logging and WAF
    cdn.WithLogging("logs-bucket.s3.amazonaws.com", "cdn/", false).
        WithWAF("arn:aws:wafv2:us-east-1:123456789012:global/webacl/my-waf/abc123").
        WithPriceClass("PriceClass_100")
}
```

### Example 15: CloudFront with Lambda@Edge

```go
import "github.com/lewis/forge/internal/tfmodules/cloudfront"

func main() {
    cdn := cloudfront.NewModule("Dynamic Content CDN")

    // Custom origin (ALB or API Gateway)
    cdn.WithCustomOrigin("api_origin", "api.example.com", true)

    // Default cache behavior with Lambda@Edge
    cdn.WithDefaultCacheBehavior("api_origin", "https-only").
        WithLambdaEdge(
            "viewer-request",
            "arn:aws:lambda:us-east-1:123456789012:function:auth-checker:1",
        ).
        WithLambdaEdge(
            "origin-response",
            "arn:aws:lambda:us-east-1:123456789012:function:header-injector:1",
        )

    // Geographic restrictions
    cdn.WithGeoRestriction("whitelist", []string{"US", "CA", "GB"})
}
```

### Example 16: AppSync GraphQL API with DynamoDB

```go
import "github.com/lewis/forge/internal/tfmodules/appsync"

func main() {
    api := appsync.NewModule("my_api")

    // GraphQL schema
    schema := `
        type Query {
            getUser(id: ID!): User
            listUsers: [User]
        }
        type Mutation {
            createUser(name: String!, email: String!): User
        }
        type User {
            id: ID!
            name: String!
            email: String!
        }
    `

    api.WithSchema(schema).
        WithCognitoAuth("us-east-1_ABC123", "us-east-1").
        WithLogging("ALL", false).
        WithXRayTracing()

    // Add DynamoDB data source
    api.WithDynamoDBDataSource("users_table", "users")

    // Add resolvers
    api.WithResolver("get_user", appsync.Resolver{
        Type:       "Query",
        Field:      "getUser",
        DataSource: strPtr("users_table"),
        RequestTemplate: strPtr(`{
            "version": "2017-02-28",
            "operation": "GetItem",
            "key": {
                "id": $util.dynamodb.toDynamoDBJson($ctx.args.id)
            }
        }`),
        ResponseTemplate: strPtr("$util.toJson($ctx.result)"),
    })
}
```

### Example 17: AppSync with Lambda Resolvers

```go
import "github.com/lewis/forge/internal/tfmodules/appsync"

func main() {
    api := appsync.NewModule("serverless_api")

    schema := `
        type Query {
            processOrder(orderId: ID!): Order
        }
        type Order {
            id: ID!
            status: String!
            total: Float!
        }
    `

    api.WithSchema(schema).
        WithIAMAuth().
        WithCaching("SMALL", 3600, true, true)

    // Lambda data source with direct integration
    api.WithLambdaDataSource(
        "order_processor",
        "arn:aws:lambda:us-east-1:123456789012:function:process-order",
    )

    // Pipeline resolver with multiple functions
    api.WithFunction("validate_order", appsync.Function{
        DataSource: "order_processor",
        RequestTemplate: strPtr(`{
            "version": "2018-05-29",
            "operation": "Invoke",
            "payload": {
                "action": "validate",
                "orderId": $util.toJson($ctx.args.orderId)
            }
        }`),
        ResponseTemplate: strPtr("$util.toJson($ctx.result)"),
    })

    api.WithResolver("process_order", appsync.Resolver{
        Type:  "Query",
        Field: "processOrder",
        Kind:  strPtr("PIPELINE"),
        PipelineConfig: &appsync.PipelineConfig{
            Functions: []string{"validate_order"},
        },
    })
}

func strPtr(s string) *string { return &s }
```

### Example 18: Stack of Multiple Modules

```go
import "github.com/lewis/forge/internal/tfmodules"

func main() {
    stack := tfmodules.NewStack("my-app")

    // Add multiple queues
    ordersQueue := sqs.NewModule("orders")
    eventsQueue := sqs.NewModule("events")
    dlqQueue := sqs.NewModule("dead_letters").WithoutDLQ()

    stack.AddModule(ordersQueue)
    stack.AddModule(eventsQueue)
    stack.AddModule(dlqQueue)

    // Validate all modules
    if err := stack.Validate(); err != nil {
        log.Fatal(err)
    }

    // Generate HCL
    hcl, _ := stack.ToHCL()
    fmt.Println(hcl)
}
```

## Architecture

### Type System

All modules follow this pattern:

```go
type Module struct {
    // Required fields (no pointer)
    Source  string
    Version string
    Name    *string

    // Optional fields (pointer, nil = not set)
    VisibilityTimeoutSeconds *int
    FifoQueue                *bool
    Tags                     map[string]string
}
```

**Why pointers?**
- Distinguish between "not set" (nil) and "set to zero value" (pointer to 0)
- Terraform modules treat these differently
- nil means "use module default"
- pointer to zero means "explicitly set to zero"

### Interface Compatibility

All modules implement the `Module` interface:

```go
type Module interface {
    LocalName() string
    Configuration() (string, error)
}
```

This makes them compatible with Lingon's resource management.

### Validation

Modules can implement the `Validator` interface:

```go
type Validator interface {
    Validate() error
}
```

Validation runs before HCL generation to catch errors early.

## Code Generation

Module structs are generated from terraform-aws-modules variable definitions:

```bash
# Generate all module types from cloned repos
task generate:modules

# Or generate a specific module
task generate:module MODULE=sqs
```

This parses `.forge/modules/*/variables.tf` and generates Go structs with:
- Proper Go types from Terraform types
- Struct tags for validation
- Comments from variable descriptions
- Validation rules extracted from descriptions

## Comparison

### Before (Phase 1 & 2): map[string]interface{}

```go
config := generators.ResourceConfig{
    Variables: map[string]interface{}{
        "visibility_timeout_seconds": 30,
        "message_retention_seconds":  345600,
        "create_dlq":                 true,
        "dlq_message_retention_seconds": 1209600,
    },
}

// ❌ Typo not caught until runtime
config.Variables["visibilty_timeout_seconds"] = 60

// ❌ Wrong type not caught until Terraform runs
config.Variables["create_dlq"] = "true"

// ❌ No IDE help
```

### After (Phase 3): Lingon-style types

```go
queue := sqs.NewModule("orders_queue")

visibility := 30
retention := 345600
createDLQ := true
dlqRetention := 1209600

queue.VisibilityTimeoutSeconds = &visibility
queue.MessageRetentionSeconds = &retention
queue.CreateDLQ = &createDLQ
queue.DLQMessageRetentionSeconds = &dlqRetention

// ✅ Typo caught at compile time
// queue.VisibiltyTimeoutSeconds = &visibility  // COMPILE ERROR

// ✅ Wrong type caught at compile time
// queue.CreateDLQ = "true"  // COMPILE ERROR: cannot use "true" (string) as *bool

// ✅ Full IDE autocomplete and docs
```

## Integration with Generators

The generators package now uses these type-safe modules:

```go
// internal/generators/sqs/generator.go

func (g *Generator) Generate(config ResourceConfig) Either[error, GeneratedCode] {
    // Create type-safe module
    module := sqs.NewModule(config.Name)

    // Configure from ResourceConfig
    if timeout, ok := config.Variables["visibility_timeout_seconds"].(int); ok {
        module.VisibilityTimeoutSeconds = &timeout
    }

    // Generate HCL
    hcl, err := module.Configuration()
    if err != nil {
        return E.Left[GeneratedCode](err)
    }

    return E.Right[error](GeneratedCode{
        Files: []FileToWrite{
            {Path: "sqs.tf", Content: hcl},
        },
    })
}
```

## Testing

All modules include comprehensive tests:

```bash
# Run module tests
go test ./internal/tfmodules/...

# Run with coverage
go test -cover ./internal/tfmodules/...

# Run benchmarks
go test -bench=. ./internal/tfmodules/...
```

## References

- [Lingon](https://github.com/golingon/lingon) - Type-safe Terraform from Go
- [terraform-aws-modules](https://github.com/terraform-aws-modules) - Official AWS modules
- Module sources: `.forge/modules/*/`
- Variable definitions: `.forge/modules/*/variables.tf`

## Future Work

### Phase 4: HCL Marshaling

Use Lingon's HCL marshaling or hclwrite for proper HCL generation:

```go
func (m *Module) Configuration() (string, error) {
    // Use Lingon's marshaling
    return lingon.MarshalHCL(m)
}
```

### Phase 5: Full Module Coverage

Generate structs for all 15 terraform-aws-modules:
- Lambda (140+ variables)
- EventBridge (80+ variables)
- RDS Aurora (150+ variables)
- API Gateway V2 (60+ variables)
- And 11 more...

### Phase 6: Interactive TUI

Use these structs to power an interactive configuration UI:

```bash
forge add dynamodb users-table --interactive
```

Shows a Bubbletea TUI with all available options, inline help from struct comments, and real-time validation.
