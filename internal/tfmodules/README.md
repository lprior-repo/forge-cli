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
| DynamoDB | `tfmodules/dynamodb` | 🚧 In Progress | 40+ |
| SNS | `tfmodules/sns` | 🚧 In Progress | 40+ |
| S3 | `tfmodules/s3` | 🚧 In Progress | 80+ |
| Lambda | `tfmodules/lambda` | 📋 Planned | 140+ |
| EventBridge | `tfmodules/eventbridge` | 📋 Planned | 80+ |
| API Gateway V2 | `tfmodules/apigatewayv2` | 📋 Planned | 60+ |
| Step Functions | `tfmodules/stepfunctions` | 📋 Planned | 40+ |

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

### Example 4: Stack of Multiple Modules

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
