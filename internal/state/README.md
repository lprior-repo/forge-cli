# internal/state

**Terraform state backend management - S3 + DynamoDB provisioning and configuration**

## Overview

The `state` package handles **Terraform remote state** setup, including S3 bucket provisioning, DynamoDB table creation for state locking, and namespace-aware state key generation.

## Problem It Solves

**Manual Terraform state setup is tedious:**

```bash
# Manual approach (what users DON'T want to do)
aws s3 mb s3://my-app-terraform-state
aws s3api put-bucket-versioning --bucket my-app-terraform-state --versioning-configuration Status=Enabled
aws s3api put-bucket-encryption --bucket my-app-terraform-state --server-side-encryption-configuration ...
aws dynamodb create-table --table-name my-app-state-lock --attribute-definitions AttributeName=LockID,AttributeType=S ...
# ... then write backend.tf manually
```

**Forge automates this with `--auto-state`:**

```bash
forge new my-app --auto-state
```

This provisions S3 bucket + DynamoDB table + generates `backend.tf` automatically.

## Data Structures

All structs are **immutable data** (pure calculations):

```go
// BackendConfig represents Terraform backend configuration
type BackendConfig struct {
    Bucket         string
    Key            string
    Region         string
    DynamoDBTable  string
    Encrypt        bool
    EnableLocking  bool
}

// S3BucketSpec represents S3 bucket configuration
type S3BucketSpec struct {
    Name          string
    Region        string
    EnableLogging bool
    Tags          map[string]string
}

// DynamoDBTableSpec represents DynamoDB table configuration
type DynamoDBTableSpec struct {
    Name        string
    Region      string
    BillingMode string  // "PAY_PER_REQUEST"
    HashKey     string  // "LockID"
    Tags        map[string]string
}

// StateResources represents complete state backend resources
type StateResources struct {
    S3Bucket      S3BucketSpec
    DynamoDBTable DynamoDBTableSpec
    BackendConfig BackendConfig
}
```

## Pure Functions (Calculations)

### Naming Conventions

```go
// GenerateStateBucketName creates S3 bucket name from project name (PURE)
// Pattern: forge-state-{project}
func GenerateStateBucketName(projectName string) string {
    normalized := strings.ToLower(projectName)
    normalized = strings.ReplaceAll(normalized, "_", "-")
    return fmt.Sprintf("forge-state-%s", normalized)
}

// GenerateLockTableName creates DynamoDB table name from project name (PURE)
// Pattern: forge_locks_{project}
func GenerateLockTableName(projectName string) string {
    normalized := strings.ToLower(projectName)
    normalized = strings.ReplaceAll(normalized, "-", "_")
    return fmt.Sprintf("forge_locks_%s", normalized)
}
```

**Examples:**
```go
GenerateStateBucketName("my-app")      // → "forge-state-my-app"
GenerateStateBucketName("my_app")      // → "forge-state-my-app" (normalized)
GenerateLockTableName("my-app")        // → "forge_locks_my_app"
```

### Namespace-Aware State Keys

```go
// GenerateStateKey creates namespace-aware state key (PURE)
// With namespace: {namespace}/terraform.tfstate
// Without namespace: terraform.tfstate
func GenerateStateKey(namespace string) string {
    if namespace != "" {
        return fmt.Sprintf("%s/terraform.tfstate", namespace)
    }
    return "terraform.tfstate"
}
```

**Examples:**
```go
GenerateStateKey("")         // → "terraform.tfstate" (production)
GenerateStateKey("pr-123")   // → "pr-123/terraform.tfstate" (PR environment)
GenerateStateKey("staging")  // → "staging/terraform.tfstate"
```

**Why namespaces?**
- ✅ Isolate PR environments in separate state files
- ✅ Prevent state conflicts between environments
- ✅ Enable parallel deployments without locking collisions
- ✅ Cost tracking per namespace

### Resource Specifications

```go
// GenerateS3BucketSpec creates S3 bucket specification (PURE)
func GenerateS3BucketSpec(projectName, region string) S3BucketSpec {
    bucketName := GenerateStateBucketName(projectName)

    return S3BucketSpec{
        Name:          bucketName,
        Region:        region,
        EnableLogging: false,  // Can be enabled in future
        Tags: map[string]string{
            "Project":   projectName,
            "ManagedBy": "forge",
            "Purpose":   "terraform-state",
        },
    }
}

// GenerateDynamoDBTableSpec creates DynamoDB table specification (PURE)
func GenerateDynamoDBTableSpec(projectName, region string) DynamoDBTableSpec {
    tableName := GenerateLockTableName(projectName)

    return DynamoDBTableSpec{
        Name:        tableName,
        Region:      region,
        BillingMode: "PAY_PER_REQUEST",  // On-demand pricing
        HashKey:     "LockID",           // Required by Terraform
        Tags: map[string]string{
            "Project":   projectName,
            "ManagedBy": "forge",
            "Purpose":   "terraform-state-locking",
        },
    }
}

// GenerateBackendConfig creates Terraform backend configuration (PURE)
func GenerateBackendConfig(bucket, key, region, table string) BackendConfig {
    return BackendConfig{
        Bucket:         bucket,
        Key:            key,
        Region:         region,
        DynamoDBTable:  table,
        Encrypt:        true,   // Always encrypt state files
        EnableLocking:  true,   // Always use locking
    }
}
```

## Actions (I/O)

### Provision State Backend

```go
// ProvisionStateBackend provisions S3 bucket and DynamoDB table (I/O ACTION)
func ProvisionStateBackend(projectName, region string) (*StateResources, error) {
    // 1. Generate specifications (pure)
    s3Spec := GenerateS3BucketSpec(projectName, region)
    dynamoSpec := GenerateDynamoDBTableSpec(projectName, region)

    // 2. Provision S3 bucket (I/O)
    err := createS3Bucket(s3Spec)
    if err != nil {
        return nil, fmt.Errorf("failed to create S3 bucket: %w", err)
    }

    // 3. Enable versioning (I/O)
    err = enableS3Versioning(s3Spec.Name, region)
    if err != nil {
        return nil, fmt.Errorf("failed to enable versioning: %w", err)
    }

    // 4. Enable encryption (I/O)
    err = enableS3Encryption(s3Spec.Name, region)
    if err != nil {
        return nil, fmt.Errorf("failed to enable encryption: %w", err)
    }

    // 5. Create DynamoDB table (I/O)
    err = createDynamoDBTable(dynamoSpec)
    if err != nil {
        return nil, fmt.Errorf("failed to create DynamoDB table: %w", err)
    }

    // 6. Generate backend config (pure)
    backendConfig := GenerateBackendConfig(
        s3Spec.Name,
        "terraform.tfstate",
        region,
        dynamoSpec.Name,
    )

    return &StateResources{
        S3Bucket:      s3Spec,
        DynamoDBTable: dynamoSpec,
        BackendConfig: backendConfig,
    }, nil
}
```

### Generate backend.tf

```go
// GenerateBackendTF generates backend.tf content (PURE)
func GenerateBackendTF(config BackendConfig) string {
    return fmt.Sprintf(`terraform {
  backend "s3" {
    bucket         = "%s"
    key            = "${var.namespace}/%s"
    region         = "%s"
    encrypt        = %t
    dynamodb_table = "%s"
  }
}
`, config.Bucket, config.Key, config.Region, config.Encrypt, config.DynamoDBTable)
}
```

**Generated backend.tf:**
```hcl
terraform {
  backend "s3" {
    bucket         = "forge-state-my-app"
    key            = "${var.namespace}/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "forge_locks_my_app"
  }
}
```

**Key feature:** `${var.namespace}` makes state keys dynamic:
- Production deploy (no namespace): `terraform.tfstate`
- PR deploy (`--namespace=pr-123`): `pr-123/terraform.tfstate`

## Usage

### CLI Integration

```go
// In forge new --auto-state
if autoState {
    resources, err := state.ProvisionStateBackend(projectName, region)
    if err != nil {
        return err
    }

    // Write backend.tf
    backendTF := state.GenerateBackendTF(resources.BackendConfig)
    err = os.WriteFile("infra/backend.tf", []byte(backendTF), 0644)
    if err != nil {
        return err
    }

    fmt.Printf("✓ Provisioned state backend\n")
    fmt.Printf("  S3 bucket: %s\n", resources.S3Bucket.Name)
    fmt.Printf("  DynamoDB table: %s\n", resources.DynamoDBTable.Name)
}
```

## Testing

```go
func TestGenerateStateBucketName(t *testing.T) {
    assert.Equal(t, "forge-state-my-app", state.GenerateStateBucketName("my-app"))
    assert.Equal(t, "forge-state-my-app", state.GenerateStateBucketName("my_app"))
}

func TestGenerateStateKey(t *testing.T) {
    assert.Equal(t, "terraform.tfstate", state.GenerateStateKey(""))
    assert.Equal(t, "pr-123/terraform.tfstate", state.GenerateStateKey("pr-123"))
}

func TestGenerateBackendTF(t *testing.T) {
    config := state.BackendConfig{
        Bucket:        "test-bucket",
        Key:           "terraform.tfstate",
        Region:        "us-east-1",
        DynamoDBTable: "test-locks",
        Encrypt:       true,
    }

    tf := state.GenerateBackendTF(config)

    assert.Contains(t, tf, `bucket         = "test-bucket"`)
    assert.Contains(t, tf, `encrypt        = true`)
}
```

## Files

- **`backend.go`** - Pure functions for naming, specs, and backend.tf generation
- **`provisioner.go`** - I/O actions for AWS resource provisioning
- **`backend_test.go`** - Unit tests for pure functions
- **`provisioner_test.go`** - Integration tests for AWS provisioning

## Design Principles

1. **Pure calculations** - Naming and spec generation are pure functions
2. **Explicit I/O** - Actions that touch AWS are clearly separated
3. **Namespace support** - State keys adapt to deployment context
4. **Security defaults** - Encryption and versioning always enabled
5. **Cost-efficient** - DynamoDB uses on-demand pricing (PAY_PER_REQUEST)

## Future Enhancements

- [ ] State bucket replication (multi-region)
- [ ] S3 access logging
- [ ] Lifecycle policies for old state versions
- [ ] State file import/export utilities
- [ ] Automatic state migration (local → remote)
- [ ] State integrity checks (checksum validation)
