# Forge CLI Specification - Summary Document

## Overview

This document provides a high-level summary of the Forge CLI OpenAPI specification located in `CLI_OPENAPI_SPEC.yaml`.

## Specification Statistics

- **Format**: OpenAPI 3.1.0
- **Total Commands**: 8 (6 root commands + 2 subcommands)
- **Total Flags**: 26 (2 global + 24 command-specific)
- **Supported Runtimes**: 3 (Go, Python, Node.js)
- **Supported Resource Types**: 4 (SQS, DynamoDB, SNS, S3)

## Command Hierarchy

```
forge (root)
├── build                    # Build Lambda functions
├── deploy                   # Build + deploy with Terraform
├── destroy                  # Tear down AWS resources
├── version                  # Show version info
├── add <type> <name>        # Add AWS resources
├── new [project-name]       # Create project/stack
│   └── lambda [name]        # Create Lambda project
```

## Global Flags

Available to ALL commands via flag persistence:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--verbose` | `-v` | boolean | false | Enable verbose output |
| `--region` | `-r` | string | "" | AWS region override |

## Command Reference

### 1. `forge build`
**Purpose**: Build Lambda functions with convention-based discovery

**Flags**:
- `--stub-only` (boolean): Create stub zips without building

**Auto-Discovery**:
- Scans `src/functions/*`
- Detects runtime from entry files
- Outputs to `.forge/build/{name}.zip`

**Examples**:
```bash
forge build
forge build --stub-only
forge -v build
```

---

### 2. `forge deploy`
**Purpose**: Build and deploy with Terraform

**Flags**:
- `--auto-approve` (boolean): Skip interactive approval
- `--namespace` (string): Namespace for ephemeral environments

**Process**:
1. Build functions
2. Run `terraform init/plan/apply`
3. Output deployed resources

**Examples**:
```bash
forge deploy
forge deploy --namespace=pr-123
forge deploy --auto-approve
forge deploy --region=us-west-2 --namespace=staging --auto-approve
```

---

### 3. `forge destroy`
**Purpose**: Tear down AWS resources

**Flags**:
- `--auto-approve` (boolean): Skip confirmation

**Safety**:
- Interactive confirmation by default
- Shows resource plan before destruction

**Examples**:
```bash
forge destroy
forge destroy --auto-approve
forge destroy --namespace=pr-123 --auto-approve
```

---

### 4. `forge version`
**Purpose**: Display version information

**Flags**: None

**Output**: Version, license, repository URL

**Examples**:
```bash
forge version
```

---

### 5. `forge add <resource-type> <name>`
**Purpose**: Add AWS resources with generated Terraform

**Arguments**:
- `<resource-type>`: sqs, dynamodb, sns, s3
- `<name>`: Resource name

**Flags**:
- `--to` (string): Target Lambda function
- `--raw` (boolean): Generate raw Terraform (not modules)
- `--no-module` (boolean): Alias for `--raw`

**Examples**:
```bash
forge add sqs orders-queue
forge add sqs orders-queue --to=processor
forge add dynamodb users-table
forge add sns notifications --raw
forge add s3 uploads-bucket
```

---

### 6. `forge new [project-name]`
**Purpose**: Create new project or stack

**Arguments**:
- `[project-name]` (optional): Project name

**Flags**:
- `--stack` (string): Create stack in existing project
- `--runtime` (string): Runtime (go1.x, python3.11, nodejs20.x)
- `--description` (string): Stack description
- `--auto-state` (boolean): Auto-provision Terraform state

**Examples**:
```bash
forge new my-app
forge new my-app --auto-state
forge new --stack=my-stack --runtime=python3.11
```

---

### 7. `forge new lambda [project-name]`
**Purpose**: Create production-ready Lambda project

**Arguments**:
- `[project-name]` (required): Project name

**Flags**:
- `--runtime` (string): python, go, nodejs (default: python)
- `--service` (string): Service name
- `--function` (string): Function name (default: handler)
- `--description` (string): Project description
- `--powertools` (boolean): Use AWS Lambda Powertools (default: true)
- `--idempotency` (boolean): Enable idempotency (default: true)
- `--dynamodb` (boolean): Include DynamoDB table (default: true)
- `--table` (string): DynamoDB table name
- `--api-path` (string): API Gateway path (default: /api/orders)
- `--method` (string): HTTP method (default: POST)

**Generated Structure**:
```
my-service/
├── src/
│   └── {function}/
│       ├── main.py
│       ├── requirements.txt
│       └── tests/
└── terraform/
    ├── main.tf
    ├── apigateway.tf
    ├── lambda.tf
    ├── dynamodb.tf
    ├── variables.tf
    └── outputs.tf
```

**Examples**:
```bash
forge new lambda my-service
forge new lambda my-service --dynamodb=false
forge new lambda my-service --runtime=go
forge new lambda my-service --api-path=/api/users --method=GET
```

---

## Common Workflows

### 1. New Project (Basic)
```bash
forge new my-app
cd my-app
forge build
forge deploy
```

### 2. New Project (With Auto State)
```bash
forge new my-app --auto-state
cd my-app
forge deploy
```

### 3. PR Preview Deployment
```bash
# Deploy PR environment
forge deploy --namespace=pr-123 --auto-approve

# Cleanup after PR close
forge destroy --namespace=pr-123 --auto-approve
```

### 4. Add Resources
```bash
forge add sqs orders-queue
forge add dynamodb orders-table
forge add sqs orders-queue --to=processor
forge deploy
```

### 5. Production Lambda
```bash
forge new lambda order-service --runtime=python
cd order-service
# Edit src/handler/main.py
forge build
forge deploy
```

---

## Convention-Based Discovery

### Function Discovery
Forge automatically discovers Lambda functions by scanning `src/functions/*`:

| Entry File | Detected Runtime |
|-----------|------------------|
| `main.go`, `*.go` | Go (provided.al2023) |
| `index.js`, `handler.js`, `index.mjs` | Node.js (nodejs20.x) |
| `app.py`, `lambda_function.py`, `handler.py` | Python (python3.13) |

### Project Structure
```
my-app/
├── infra/              # REQUIRED: Terraform infrastructure
│   ├── main.tf
│   ├── variables.tf
│   └── outputs.tf
└── src/                # OPTIONAL: Application code
    └── functions/      # Convention: Lambda functions here
        ├── api/
        │   └── main.go
        └── worker/
            └── index.js
```

### Namespace Pattern
```bash
# Deploy with namespace
forge deploy --namespace=pr-123

# Results:
# - TF_VAR_namespace=pr-123
# - Resources: my-app-pr-123-api, my-app-pr-123-worker
# - State: forge/pr-123-terraform.tfstate
```

---

## Environment Variables

| Variable | Type | Description |
|----------|------|-------------|
| `AWS_REGION` | string | Default AWS region |
| `AWS_PROFILE` | string | AWS CLI profile |
| `TF_VAR_namespace` | string | Namespace (set by --namespace) |
| `FORGE_VERBOSE` | boolean | Enable verbose output |

---

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Error (invalid args, AWS creds, build failure, etc.) |

---

## Dependencies

### Required
- **Terraform** (>= 1.0.0): Infrastructure provisioning
- **AWS CLI** (>= 2.0.0): AWS authentication

### Optional (Runtime-Specific)
- **Go** (>= 1.21): Building Go Lambda functions
- **Python** (>= 3.11): Building Python Lambda functions
- **Node.js** (>= 20): Building Node.js Lambda functions

---

## Testing

| Test Type | Command | Coverage | Count |
|-----------|---------|----------|-------|
| Unit | `task test:unit` | 85% | 189 |
| Integration | `task test:integration` | 90% | 37 |
| E2E | `task test:e2e` | 95% | 15 |
| Mutation | `task mutation` | 80% | - |

---

## Architecture

**Framework**: Cobra (spf13/cobra)
**Language**: Go
**Paradigm**: Functional Programming
**Error Handling**: Either Monad (Railway-Oriented Programming)

**Design Patterns**:
- Repository Pattern
- Strategy Pattern
- Decorator Pattern
- Pipeline Pattern
- Registry Pattern

**Package Structure**:
- `internal/build`: Runtime-specific builders
- `internal/cli`: Cobra commands (I/O boundary)
- `internal/config`: HCL configuration
- `internal/lingon`: Type-safe Terraform generation
- `internal/pipeline`: Pipeline orchestration
- `internal/scaffold`: Project scaffolding
- `internal/stack`: Stack detection and dependencies
- `internal/terraform`: Terraform executor wrapper

---

## Key Features

### Convention Over Configuration
- Zero config files required
- Auto-discovery of Lambda functions
- Smart runtime detection
- Sensible defaults

### Ephemeral Environments
- Namespace support for PR previews
- Isolated AWS resources per namespace
- State isolation per environment
- Perfect for CI/CD integration

### Type-Safe Terraform
- 170+ Lambda parameters supported
- 80+ API Gateway parameters
- 50+ DynamoDB parameters
- Complete serverless.tf specification

### Production-Ready
- AWS Lambda Powertools integration
- Idempotency support
- DynamoDB tables with GSI
- API Gateway with CORS
- CloudWatch logging
- IAM roles and policies

### Pure Functional Design
- Either monad for error handling
- Option monad for optional values
- Pure functions (no side effects)
- Immutable data structures
- Railway-oriented programming

---

## Specification Files

1. **CLI_OPENAPI_SPEC.yaml** (4,000+ lines)
   - Complete OpenAPI 3.1.0 specification
   - All commands, flags, parameters
   - Response schemas
   - Error codes
   - Workflows
   - Examples

2. **CLI_SPEC_SUMMARY.md** (this file)
   - High-level overview
   - Quick reference
   - Common workflows
   - Architecture summary

---

## Usage Tips

### Pro Tips
1. Use `--verbose` for debugging build/deploy issues
2. Use `--namespace` for isolated PR environments
3. Use `--auto-approve` in CI/CD pipelines
4. Use `--stub-only` for fast Terraform init
5. Use `forge add` to generate boilerplate Terraform

### Best Practices
1. Always use namespaces for non-production deployments
2. Enable `--auto-state` for team projects
3. Version-control generated Terraform (infra/*.tf)
4. Use Lambda Powertools for production applications
5. Test locally before deploying to AWS

### Common Pitfalls
1. Forgetting to set AWS credentials
2. Not installing Terraform binary
3. Modifying `.forge/build/` directly (regenerated on build)
4. Using `--auto-approve` without reviewing plan
5. Not cleaning up namespaced environments

---

## CI/CD Integration

### GitHub Actions (PR Preview)
```yaml
name: PR Preview
on: pull_request
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: forge deploy --namespace=pr-${{ github.event.number }} --auto-approve
```

### GitHub Actions (PR Cleanup)
```yaml
name: PR Cleanup
on:
  pull_request:
    types: [closed]
jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: forge destroy --namespace=pr-${{ github.event.number }} --auto-approve
```

---

## Future Enhancements

**Planned Features**:
- Interactive TUI (bubbletea)
- `forge logs` command
- `forge list` command (show all namespaces)
- Hot reload / watch mode
- Lambda Layers support
- Custom domains with Route53
- VPC configuration helpers
- Cost tracking dashboard
- Deployment rollback support

---

## References

- **Main README**: `README.md`
- **OpenAPI Spec**: `CLI_OPENAPI_SPEC.yaml`
- **Lingon Spec**: `LINGON_SPEC.md`
- **Code Audit**: `CODEBASE_AUDIT.md`
- **TDD Progress**: `TDD_PROGRESS.md`
- **Project Instructions**: `CLAUDE.md`
