# Resource Configuration Reference

This document lists all configuration options available for each resource type, based on the Terraform modules from [serverless.tf](https://serverless.tf/).

## Table of Contents

- [SQS Queues](#sqs-queues)
- [DynamoDB Tables](#dynamodb-tables)
- [SNS Topics](#sns-topics)
- [S3 Buckets](#s3-buckets)
- [Future Resources](#future-resources)

---

## SQS Queues

Module: `terraform-aws-modules/sqs/aws` v4.0

### Currently Implemented (MVP)

```yaml
visibility_timeout_seconds: 30      # Lambda timeout + buffer
message_retention_seconds:  345600  # 4 days
create_dlq:                 true    # Dead letter queue
dlq_message_retention:      1209600 # 14 days
```

### Lambda Integration (--to flag)

```yaml
batch_size:             10   # Messages per invocation
max_batching_window:    5    # Seconds to wait for full batch
max_concurrency:        10   # Concurrent Lambda invocations
```

### Full Module Configuration (Phase 3)

The module supports 40+ additional parameters:

```yaml
# Queue Configuration
delay_seconds:                    0-900
max_message_size:                 1024-262144
receive_wait_time_seconds:        0-20
kms_master_key_id:                string
kms_data_key_reuse_period_seconds: 60-86400

# FIFO Queue Options
fifo_queue:                       boolean
content_based_deduplication:      boolean
deduplication_scope:              "queue" | "messageGroup"
fifo_throughput_limit:            "perQueue" | "perMessageGroupId"

# Dead Letter Queue
dlq_redrive_allow_policy:         json
dlq_redrive_max_receive_count:    1-1000

# Policies
sqs_managed_sse_enabled:          boolean
create_queue_policy:              boolean
queue_policy_statements:          list

# Tags
tags:                             map
```

**Reference**: https://github.com/terraform-aws-modules/terraform-aws-sqs

---

## DynamoDB Tables

Module: `terraform-aws-modules/dynamodb-table/aws` v4.0

### Currently Implemented (MVP)

```yaml
hash_key:               "id"
billing_mode:           "PAY_PER_REQUEST"  # On-demand pricing
stream_enabled:         false              # true when --to used
stream_view_type:       "NEW_AND_OLD_IMAGES"
point_in_time_recovery: true
attributes:
  - name: "id"
    type: "S"  # String
```

### Lambda Integration (--to flag)

```yaml
# DynamoDB Streams Configuration
stream_enabled:         true
stream_view_type:       "NEW_AND_OLD_IMAGES"
batch_size:             100
starting_position:      "LATEST"
max_concurrency:        10
```

### Full Module Configuration (Phase 3)

The module supports 30+ additional parameters:

```yaml
# Table Configuration
name:                   string
billing_mode:           "PROVISIONED" | "PAY_PER_REQUEST"
read_capacity:          number  # if PROVISIONED
write_capacity:         number  # if PROVISIONED

# Keys
hash_key:               string (required)
range_key:              string (optional)
attributes:             list of {name, type}
  # Types: S (string), N (number), B (binary)

# Global Secondary Indexes (GSI)
global_secondary_indexes:
  - name:               string
    hash_key:           string
    range_key:          string (optional)
    projection_type:    "ALL" | "KEYS_ONLY" | "INCLUDE"
    non_key_attributes: list
    read_capacity:      number
    write_capacity:     number

# Local Secondary Indexes (LSI)
local_secondary_indexes:
  - name:               string
    range_key:          string
    projection_type:    "ALL" | "KEYS_ONLY" | "INCLUDE"
    non_key_attributes: list

# Streams
stream_enabled:         boolean
stream_view_type:       "KEYS_ONLY" | "NEW_IMAGE" | "OLD_IMAGE" | "NEW_AND_OLD_IMAGES"

# TTL
ttl_enabled:            boolean
ttl_attribute_name:     string

# Autoscaling (for PROVISIONED mode)
autoscaling_enabled:    boolean
autoscaling_read:
  scale_in_cooldown:    seconds
  scale_out_cooldown:   seconds
  target_value:         percentage
  max_capacity:         number
autoscaling_write:
  # same as read

# Backup
point_in_time_recovery_enabled: boolean

# Encryption
server_side_encryption_enabled: boolean
server_side_encryption_kms_key_arn: string

# Replica Regions (Global Tables)
replica_regions:
  - region_name:        string
    kms_key_arn:        string
    propagate_tags:     boolean

# Tags
tags:                   map
table_class:            "STANDARD" | "STANDARD_INFREQUENT_ACCESS"
```

**Reference**: https://github.com/terraform-aws-modules/terraform-aws-dynamodb-table

---

## SNS Topics

Module: `terraform-aws-modules/sns/aws` v6.0

### Currently Implemented (MVP)

```yaml
name:         string
display_name: string
fifo_topic:   false
```

### Lambda Integration (--to flag)

```yaml
# Subscription Configuration
protocol:     "lambda"
endpoint:     Lambda function ARN
```

### Full Module Configuration (Phase 3)

The module supports 20+ additional parameters:

```yaml
# Topic Configuration
name:                       string
display_name:               string
fifo_topic:                 boolean
content_based_deduplication: boolean (FIFO only)

# Delivery Policy
delivery_policy:            json
  # Controls retry backoff, max receives, etc.

# Encryption
kms_master_key_id:          string

# Topic Policy
create_topic_policy:        boolean
topic_policy:               json
topic_policy_statements:    list

# Subscriptions
subscriptions:
  protocol:                 "lambda" | "sqs" | "email" | "http" | "https" | "sms"
  endpoint:                 string (ARN, email, URL, phone)
  filter_policy:            json
  filter_policy_scope:      "MessageAttributes" | "MessageBody"
  raw_message_delivery:     boolean
  redrive_policy:           json
  delivery_policy:          json

# Data Protection Policy
data_protection_policy:     json

# Archive Policy
archive_policy:             json

# FIFO Configuration
signature_version:          number
tracing_config:             string

# Tags
tags:                       map
```

**Reference**: https://github.com/terraform-aws-modules/terraform-aws-sns

---

## S3 Buckets

Module: `terraform-aws-modules/s3-bucket/aws` v4.0

### Currently Implemented (MVP)

```yaml
versioning_enabled:      true
block_public_acls:       true
block_public_policy:     true
ignore_public_acls:      true
restrict_public_buckets: true
force_destroy:           false
server_side_encryption:  "AES256"
```

### Lambda Integration (--to flag)

```yaml
# S3 Event Notification
events:                  ["s3:ObjectCreated:*"]
lambda_function_arn:     string
filter_prefix:           string (optional)
filter_suffix:           string (optional)
```

### Full Module Configuration (Phase 3)

The module supports 50+ additional parameters:

```yaml
# Bucket Configuration
bucket:                  string
bucket_prefix:           string (auto-generate name)
force_destroy:           boolean
object_lock_enabled:     boolean
expected_bucket_owner:   string

# Versioning
versioning:
  enabled:               boolean
  mfa_delete:            boolean

# Logging
logging:
  target_bucket:         string
  target_prefix:         string
  target_object_key_format: "SimplePrefix" | "PartitionedPrefix"

# Server-Side Encryption
server_side_encryption_configuration:
  rule:
    apply_server_side_encryption_by_default:
      sse_algorithm:     "AES256" | "aws:kms"
      kms_master_key_id: string
    bucket_key_enabled:  boolean

# Lifecycle Rules
lifecycle_rules:
  - id:                  string
    enabled:             boolean
    prefix:              string
    tags:                map

    transition:
      days:              number
      storage_class:     "STANDARD_IA" | "ONEZONE_IA" | "INTELLIGENT_TIERING" | "GLACIER" | "DEEP_ARCHIVE"

    expiration:
      days:              number
      expired_object_delete_marker: boolean

    noncurrent_version_transition:
      noncurrent_days:   number
      storage_class:     string

    noncurrent_version_expiration:
      noncurrent_days:   number

    abort_incomplete_multipart_upload:
      days_after_initiation: number

# CORS Configuration
cors_rules:
  - allowed_headers:     list
    allowed_methods:     list ("GET", "PUT", "POST", "DELETE", "HEAD")
    allowed_origins:     list
    expose_headers:      list
    max_age_seconds:     number

# Website Configuration
website:
  index_document:        string
  error_document:        string
  redirect_all_requests_to:
    host_name:           string
    protocol:            "http" | "https"
  routing_rules:         json

# Acceleration
acceleration_status:     "Enabled" | "Suspended"

# Request Payer
request_payer:           "Requester" | "BucketOwner"

# Public Access Block
block_public_acls:       boolean
block_public_policy:     boolean
ignore_public_acls:      boolean
restrict_public_buckets: boolean

# Object Lock Configuration
object_lock_configuration:
  rule:
    default_retention:
      mode:              "GOVERNANCE" | "COMPLIANCE"
      days:              number
      years:             number

# Replication Configuration
replication_configuration:
  role:                  string (IAM role ARN)
  rules:
    - id:                string
      status:            "Enabled" | "Disabled"
      priority:          number
      filter:
        prefix:          string
        tags:            map
      destination:
        bucket:          string (destination bucket ARN)
        storage_class:   string
        replica_kms_key_id: string
        replication_time:
          status:        "Enabled"
          time:
            minutes:     number
        metrics:
          status:        "Enabled"
          event_threshold:
            minutes:     number

# Inventory Configuration
inventory_configuration:
  - id:                  string
    enabled:             boolean
    filter:
      prefix:            string
    destination:
      bucket:
        format:          "CSV" | "ORC" | "Parquet"
        bucket_arn:      string
        prefix:          string
        encryption:      map
    schedule:
      frequency:         "Daily" | "Weekly"
    included_object_versions: "All" | "Current"
    optional_fields:     list

# Analytics Configuration
analytics_configuration:
  - id:                  string
    filter:
      prefix:            string
      tags:              map
    storage_class_analysis:
      data_export:
        destination:
          s3_bucket_destination:
            bucket_arn:  string
            prefix:      string

# Event Notifications
event_notifications:
  lambda_functions:
    - lambda_function_arn: string
      events:            list
      filter_prefix:     string
      filter_suffix:     string

  sqs_queues:
    - queue_arn:         string
      events:            list
      filter_prefix:     string
      filter_suffix:     string

  sns_topics:
    - topic_arn:         string
      events:            list
      filter_prefix:     string
      filter_suffix:     string

# Intelligent Tiering
intelligent_tiering:
  - name:                string
    status:              "Enabled" | "Disabled"
    filter:
      prefix:            string
      tags:              map
    tiering:
      access_tier:       "ARCHIVE_ACCESS" | "DEEP_ARCHIVE_ACCESS"
      days:              number

# Ownership Controls
object_ownership:        "BucketOwnerPreferred" | "ObjectWriter" | "BucketOwnerEnforced"

# ACL
acl:                     "private" | "public-read" | "public-read-write" | "aws-exec-read" | "authenticated-read" | "log-delivery-write"

# Bucket Policy
attach_policy:           boolean
policy:                  json

# Tags
tags:                    map
```

**Reference**: https://github.com/terraform-aws-modules/terraform-aws-s3-bucket

---

## Future Resources

### API Gateway (Phase 3)

Module: `terraform-aws-modules/apigateway-v2/aws`

- HTTP APIs with JWT authorizers
- WebSocket APIs
- REST APIs with request validators
- Custom domains with Route53
- CloudWatch logging
- Throttling and rate limiting

### EventBridge (Phase 3)

Module: `terraform-aws-modules/eventbridge/aws`

- Event buses
- Event rules with patterns
- Multiple targets per rule
- Dead letter queues
- Input transformers
- Archive and replay

### Step Functions (Phase 4)

Module: `terraform-aws-modules/step-functions/aws`

- State machines (Express and Standard)
- ASL definition generation
- CloudWatch logging
- X-Ray tracing
- Error handling and retries

### Lambda Functions (Phase 4)

Module: `terraform-aws-modules/lambda/aws`

- Function packaging
- Layers
- Aliases and versions
- Reserved concurrency
- Environment variables
- VPC configuration

---

## Using Advanced Configurations

### Option 1: Edit Generated Terraform (Current)

After running `forge add`, edit the generated `.tf` files to add advanced configuration:

```bash
forge add dynamodb users-table
# Edit infra/dynamodb.tf to add GSI, autoscaling, etc.
```

### Option 2: Interactive TUI (Phase 3)

Future versions will provide interactive configuration:

```bash
forge add dynamodb users-table --interactive
```

This will launch a TUI allowing you to configure all module parameters visually.

### Option 3: Configuration File (Phase 4)

Provide a YAML configuration file:

```bash
forge add dynamodb users-table --config=users-table.yaml
```

Where `users-table.yaml` contains:

```yaml
hash_key: user_id
range_key: created_at
billing_mode: PAY_PER_REQUEST
global_secondary_indexes:
  - name: email-index
    hash_key: email
    projection_type: ALL
attributes:
  - name: user_id
    type: S
  - name: email
    type: S
  - name: created_at
    type: N
```

---

## Contributing

To add support for additional module parameters:

1. Update the `Variables` map in `internal/generators/<resource>/generator.go`
2. Add parameter handling in `generateModuleCode()` function
3. Add validation in `Validate()` method
4. Update tests in `generator_test.go`
5. Document the parameter in this file

See `internal/generators/sqs/generator.go` for reference implementation.
