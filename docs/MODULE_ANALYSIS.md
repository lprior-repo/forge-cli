# Terraform Module Analysis

This document summarizes the 15 official Terraform modules we've cloned from terraform-aws-modules.

## Module Inventory

| Module | Variables | Lines | Priority | Status |
|--------|-----------|-------|----------|--------|
| **RDS Aurora** | 150+ | 883 | Phase 4 | Not Started |
| **Lambda** | 140+ | 849 | Phase 3 | Not Started |
| **EventBridge** | 80+ | 526 | Phase 3 | Not Started |
| **S3 Bucket** | 80+ | 426 | **Phase 2** | ✅ Basic generator |
| **API Gateway V2** | 60+ | 387 | Phase 3 | Not Started |
| **AppSync** | 50+ | 358 | Phase 4 | Not Started |
| **SQS** | 60+ | 319 | **Phase 1** | ✅ Complete |
| **RDS Proxy** | 40+ | 296 | Phase 4 | Not Started |
| **Step Functions** | 40+ | 276 | Phase 3 | Not Started |
| **AppConfig** | 35+ | 264 | Phase 4 | Not Started |
| **SNS** | 40+ | 235 | **Phase 2** | ✅ Basic generator |
| **DynamoDB** | 40+ | 235 | **Phase 2** | ✅ Basic generator |
| **CloudFront** | 30+ | 212 | Phase 4 | Not Started |
| **Secrets Manager** | 30+ | 211 | Phase 4 | Not Started |
| **SSM Parameter** | 15+ | 87 | Phase 4 | Not Started |

**Total: 870+ variables across 5,564 lines**

## Location

All modules are cloned to:
```
.forge/modules/
├── lambda/
├── appsync/
├── eventbridge/
├── step-functions/
├── cloudfront/
├── apigateway-v2/
├── dynamodb/
├── rds-aurora/
├── rds-proxy/
├── s3/
├── sqs/
├── appconfig/
├── ssm-parameter/
├── secrets-manager/
└── sns/
```

## Type Safety Strategy

### Current Approach (MVP - Phase 1 & 2)

Using `map[string]interface{}` for Variables field:

```go
Variables: map[string]interface{}{
    "visibility_timeout_seconds": 30,
    "message_retention_seconds":  345600,
    "create_dlq":                 true,
}
```

**Pros:**
- ✅ Quick to implement
- ✅ Easy to extend
- ✅ Flexible for MVP

**Cons:**
- ❌ No compile-time type safety
- ❌ Runtime type assertions needed
- ❌ No IDE autocomplete
- ❌ Easy to make typos

### Proposed Approach (Phase 3)

Create strongly-typed configuration structs by parsing module variables:

```go
type SQSConfig struct {
    // Core fields (required)
    Name string `validate:"required"`
    
    // Common fields
    VisibilityTimeoutSeconds *int    `terraform:"visibility_timeout_seconds"`
    MessageRetentionSeconds  *int    `terraform:"message_retention_seconds"`
    FifoQueue                *bool   `terraform:"fifo_queue"`
    DelaySeconds             *int    `terraform:"delay_seconds"`
    
    // DLQ configuration
    CreateDLQ               *bool   `terraform:"create_dlq"`
    DLQMessageRetention     *int    `terraform:"dlq_message_retention_seconds"`
    
    // Encryption
    KmsMasterKeyID          *string `terraform:"kms_master_key_id"`
    SQSManagedSSEEnabled    *bool   `terraform:"sqs_managed_sse_enabled"`
    
    // Advanced (map for rarely-used options)
    Advanced map[string]interface{} `terraform:"-"`
}
```

**Benefits:**
- ✅ Compile-time type safety
- ✅ IDE autocomplete
- ✅ Self-documenting
- ✅ Validation at struct level
- ✅ Hybrid approach for advanced options

## Implementation Plan

### Phase 3: Type-Safe Structs

1. **Generate struct definitions from modules**
   - Parse variables.tf files
   - Extract types, defaults, descriptions
   - Generate Go structs with proper tags

2. **Create struct-to-HCL converter**
   - Convert Go structs to Terraform HCL
   - Handle nil vs zero values (pointers)
   - Support nested structures

3. **Update generators to use structs**
   - Replace `map[string]interface{}`
   - Add validation methods
   - Maintain backwards compatibility

4. **Add Interactive TUI**
   - Use bubbletea for interactive config
   - Show only relevant fields based on context
   - Provide inline help from descriptions

### Example: SQS Variables Analysis

From `.forge/modules/sqs/variables.tf`:

```hcl
variable "visibility_timeout_seconds" {
  description = "The visibility timeout for the queue. An integer from 0 to 43200 (12 hours)"
  type        = number
  default     = null
}

variable "message_retention_seconds" {
  description = "The number of seconds Amazon SQS retains a message. Integer representing seconds, from 60 (1 minute) to 1209600 (14 days)"
  type        = number
  default     = null
}

variable "fifo_queue" {
  description = "Boolean designating a FIFO queue"
  type        = bool
  default     = false
}
```

Generated Go struct:

```go
type SQSConfig struct {
    // The visibility timeout for the queue. An integer from 0 to 43200 (12 hours)
    VisibilityTimeoutSeconds *int `json:"visibility_timeout_seconds,omitempty" terraform:"visibility_timeout_seconds" validate:"min=0,max=43200"`
    
    // The number of seconds Amazon SQS retains a message. Integer representing seconds, from 60 (1 minute) to 1209600 (14 days)
    MessageRetentionSeconds *int `json:"message_retention_seconds,omitempty" terraform:"message_retention_seconds" validate:"min=60,max=1209600"`
    
    // Boolean designating a FIFO queue
    FifoQueue *bool `json:"fifo_queue,omitempty" terraform:"fifo_queue"`
}
```

## Validation Rules from Modules

Many modules include validation in descriptions:

| Module | Variable | Validation |
|--------|----------|------------|
| SQS | visibility_timeout_seconds | 0 to 43200 |
| SQS | message_retention_seconds | 60 to 1209600 |
| SQS | delay_seconds | 0 to 900 |
| SQS | max_message_size | 1024 to 262144 |
| SQS | receive_wait_time_seconds | 0 to 20 |
| DynamoDB | billing_mode | "PROVISIONED" or "PAY_PER_REQUEST" |
| DynamoDB | stream_view_type | "KEYS_ONLY", "NEW_IMAGE", "OLD_IMAGE", "NEW_AND_OLD_IMAGES" |
| S3 | bucket | 3-63 chars, lowercase, no consecutive hyphens |

These can be enforced at compile-time with struct tags and Go validator library.

## Code Generator Tool (Future)

Create a tool to auto-generate structs:

```bash
task generate:structs
```

This would:
1. Parse all modules' `variables.tf` files
2. Extract type information, defaults, descriptions
3. Generate Go struct definitions
4. Add validation tags
5. Create conversion functions

## References

- All modules: `.forge/modules/*/`
- Variable files: `.forge/modules/*/variables.tf`
- Examples: `.forge/modules/*/examples/`
