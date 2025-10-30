# Forge CLI Reference

Complete command-line reference for Forge - the developer-friendly serverless deployment tool.

## Table of Contents

- [Installation](#installation)
- [Global Flags](#global-flags)
- [Commands](#commands)
  - [forge new](#forge-new)
  - [forge build](#forge-build)
  - [forge deploy](#forge-deploy)
  - [forge destroy](#forge-destroy)
  - [forge version](#forge-version)
- [Workflows](#workflows)
- [Environment Variables](#environment-variables)
- [Exit Codes](#exit-codes)

---

## Installation

### From Source

```bash
git clone https://github.com/lprior-repo/forge-cli.git
cd forge-cli
task build
sudo mv bin/forge /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/lprior-repo/forge-cli/cmd/forge@latest
```

### Verify Installation

```bash
forge version
```

---

## Global Flags

These flags work with all commands:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--verbose` | `-v` | boolean | Enable verbose logging and debug output |
| `--region` | `-r` | string | Override AWS region from forge.hcl |
| `--help` | `-h` | boolean | Show help for command |

**Examples:**

```bash
# Verbose output
forge build --verbose

# Override region
forge deploy --region us-west-2

# Get help
forge new --help
```

---

## Commands

### forge new

**Create a new Forge project with convention-based structure.**

#### Syntax

```bash
forge new <project-name> [flags]
```

#### Arguments

- `<project-name>` - **Required**. Name of the project to create. Must be lowercase alphanumeric with hyphens.

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--runtime` | string | `go` | Initial function runtime (`go`, `python`, `nodejs`, `java`) |
| `--region` | string | `us-east-1` | AWS region for deployment |
| `--auto-state` | boolean | `false` | Auto-provision S3 + DynamoDB for Terraform state |

#### What It Creates

**Without `--auto-state`:**
```
my-app/
├── forge.hcl              # Project configuration
├── .gitignore             # Terraform artifacts
├── README.md              # Quick start guide
├── infra/                 # Terraform infrastructure
│   ├── main.tf           # Lambda + IAM resources
│   ├── variables.tf      # namespace variable
│   └── outputs.tf        # Function URLs, ARNs
└── src/functions/         # Lambda functions
    └── api/              # Hello-world function
        └── main.go       # (or handler.py, index.js)
```

**With `--auto-state`:**
```
my-app/
├── (all files above)
└── infra/
    ├── backend.tf        # ← Auto-generated S3 backend config
    ├── main.tf
    ├── variables.tf
    └── outputs.tf
```

**Generated backend.tf:**
```hcl
terraform {
  backend "s3" {
    bucket         = "forge-state-my-app"
    key            = "${var.namespace}/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "forge-locks-my-app"
  }
}
```

**AWS Resources Created (with `--auto-state`):**
- S3 bucket: `forge-state-my-app` (versioning + encryption enabled)
- DynamoDB table: `forge-locks-my-app` (state locking)

#### Examples

**Basic Go project:**
```bash
forge new my-api --runtime=go
cd my-api
forge build
```

**Python project with auto-state:**
```bash
forge new order-service --runtime=python --auto-state
```

**Node.js project in different region:**
```bash
forge new notification-service --runtime=nodejs --region=eu-west-1
```

#### Generated forge.hcl

```hcl
project {
  name   = "my-app"
  region = "us-east-1"
}
```

#### Generated Hello-World Code

**Go (`src/functions/api/main.go`):**
```go
package main

import (
    "context"
    "encoding/json"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    response := map[string]interface{}{
        "message": "Hello from Forge!",
        "path":    event.Path,
        "method":  event.HTTPMethod,
    }

    body, _ := json.Marshal(response)

    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Headers:    map[string]string{"Content-Type": "application/json"},
        Body:       string(body),
    }, nil
}

func main() {
    lambda.Start(handler)
}
```

**Python (`src/functions/api/handler.py`):**
```python
import json
import logging

logger = logging.getLogger()
logger.setLevel(logging.INFO)

def handler(event, context):
    """Lambda handler function"""
    logger.info(f"Received event: {json.dumps(event)}")

    return {
        "statusCode": 200,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps({
            "message": "Hello from Forge!",
            "path": event.get("path", "/"),
            "method": event.get("httpMethod", "GET"),
        })
    }
```

#### Validation Rules

Project name must:
- Be lowercase
- Contain only alphanumeric characters and hyphens
- Not start or end with hyphen
- Be between 3-50 characters

**Valid:**
- `my-app`
- `order-service`
- `api-v2`

**Invalid:**
- `My-App` (uppercase)
- `my_app` (underscore)
- `-my-app` (starts with hyphen)
- `my-app-` (ends with hyphen)

---

### forge build

**Build all Lambda functions using convention-based discovery.**

#### Syntax

```bash
forge build [flags]
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--verbose` | boolean | `false` | Show detailed build logs |
| `--clean` | boolean | `false` | Clean build cache before building |

#### How It Works

1. **Discovery** - Scans `src/functions/*` directories
2. **Runtime Detection** - Identifies runtime from entry files:
   - `main.go` → Go (`provided.al2023`)
   - `index.js` → Node.js (`nodejs20.x`)
   - `handler.py` → Python (`python3.13`)
3. **Build** - Compiles/packages each function
4. **Cache** - Stores artifacts in `.forge/build/`
5. **Output** - Creates `.zip` files for Lambda deployment

#### Build Process by Runtime

**Go:**
```bash
# For src/functions/api/
cd src/functions/api
go mod download
GOOS=linux GOARCH=amd64 go build -o bootstrap
zip ../../.forge/build/api.zip bootstrap
```

**Python:**
```bash
# For src/functions/worker/
cd src/functions/worker
pip install -r requirements.txt -t .forge/python_modules/
zip -r ../../.forge/build/worker.zip . .forge/python_modules/
```

**Node.js:**
```bash
# For src/functions/notifier/
cd src/functions/notifier
npm install --production
zip -r ../../.forge/build/notifier.zip . node_modules/
```

#### Output Example

```bash
$ forge build

Building 3 functions...
✓ api (go1.x) - 2.1s
✓ worker (python3.13) - 1.8s
✓ notifier (nodejs20.x) - 1.2s

Built 3 functions in 5.1s
Artifacts in .forge/build/
  api.zip (2.3 MB)
  worker.zip (8.1 MB)
  notifier.zip (1.4 MB)
```

#### Caching Strategy

Forge uses **SHA256-based caching** to skip unchanged builds:

```bash
$ forge build
Building 3 functions...
✓ api (cached) - 0.1s
✓ worker (go1.x) - 1.8s  # Changed
✓ notifier (cached) - 0.1s

Built 3 functions in 2.0s (2 cached)
```

**Cache invalidation triggers:**
- Source code changes
- Dependency file changes (`go.mod`, `requirements.txt`, `package.json`)
- `--clean` flag

#### Examples

**Standard build:**
```bash
forge build
```

**Clean build (ignore cache):**
```bash
forge build --clean
```

**Verbose build logs:**
```bash
forge build --verbose
```

**Output:**
```
[DEBUG] Scanning src/functions/
[DEBUG] Found: api (runtime: go1.x)
[DEBUG] Found: worker (runtime: python3.13)
[DEBUG] Building api...
[DEBUG]   Running: go mod download
[DEBUG]   Running: GOOS=linux GOARCH=amd64 go build -o bootstrap
[DEBUG]   Creating zip: .forge/build/api.zip
[DEBUG]   Checksum: a1b2c3d4...
✓ api (go1.x) - 2.1s
...
```

#### Troubleshooting

**Build fails with "go.mod not found":**
```bash
cd src/functions/api
go mod init api
go mod tidy
cd ../../..
forge build
```

**Build fails with "requirements.txt not found":**
```bash
# Create empty requirements.txt
touch src/functions/worker/requirements.txt
forge build
```

**Clear build cache:**
```bash
rm -rf .forge/build
forge build --clean
```

---

### forge deploy

**Deploy infrastructure to AWS via Terraform (pipeline-first).**

#### Syntax

```bash
forge deploy [flags]
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--namespace` | string | `""` | Resource namespace for isolated environments |
| `--auto-approve` | boolean | `false` | Skip interactive plan approval |
| `--var` | string[] | `[]` | Pass Terraform variables (`key=value`) |

#### What It Does

1. **Build** - Runs `forge build` automatically
2. **Init** - `terraform init` (downloads providers)
3. **Plan** - `terraform plan` (shows changes)
4. **Approval** - Prompts user (unless `--auto-approve`)
5. **Apply** - `terraform apply` (creates/updates resources)
6. **Outputs** - Displays function URLs, ARNs

#### Deployment Flow

```
┌─────────────┐
│ forge build │  ← Build all functions
└──────┬──────┘
       ↓
┌─────────────┐
│ terraform   │  ← Initialize providers
│    init     │
└──────┬──────┘
       ↓
┌─────────────┐
│ terraform   │  ← Preview changes
│    plan     │
└──────┬──────┘
       ↓
┌─────────────┐
│  User       │  ← Approve changes
│  Approval   │    (unless --auto-approve)
└──────┬──────┘
       ↓
┌─────────────┐
│ terraform   │  ← Create/update resources
│   apply     │
└──────┬──────┘
       ↓
┌─────────────┐
│  Display    │  ← Show outputs
│  Outputs    │
└─────────────┘
```

#### Examples

**Production deployment:**
```bash
forge deploy
```

**Output:**
```
Building 3 functions...
✓ Built 3 functions in 5.1s

Initializing Terraform...
✓ Terraform initialized

Planning changes...

Terraform will perform the following actions:

  # aws_lambda_function.api will be created
  + resource "aws_lambda_function" "api" {
      + function_name = "my-app-api"
      + runtime       = "provided.al2023"
      + handler       = "bootstrap"
      ...
    }

Plan: 5 to add, 0 to change, 0 to destroy.

Do you want to perform these actions? (yes/no): yes

Applying changes...
✓ Deployment successful

Outputs:
  api_function_url = "https://abc123.lambda-url.us-east-1.on.aws/"
  api_function_arn = "arn:aws:lambda:us-east-1:123456789:function:my-app-api"
```

**PR preview environment:**
```bash
forge deploy --namespace=pr-123
```

**Namespace behavior:**
- Sets `TF_VAR_namespace=pr-123`
- Prefixes all resources: `my-app-pr-123-api`
- Uses separate state file: `forge/pr-123-terraform.tfstate`
- Completely isolated from production

**CI/CD deployment (auto-approve):**
```bash
forge deploy --auto-approve
```

**Custom Terraform variables:**
```bash
forge deploy --var="region=us-west-2" --var="memory=512"
```

#### Namespace Patterns

**Production (no namespace):**
```bash
forge deploy
```
- Resources: `my-app-api`, `my-app-worker`
- State: `forge/terraform.tfstate`

**PR Preview:**
```bash
forge deploy --namespace=pr-456
```
- Resources: `my-app-pr-456-api`, `my-app-pr-456-worker`
- State: `forge/pr-456-terraform.tfstate`

**Staging:**
```bash
forge deploy --namespace=staging
```
- Resources: `my-app-staging-api`, `my-app-staging-worker`
- State: `forge/staging-terraform.tfstate`

#### Pipeline Integration

**GitHub Actions (`.github/workflows/deploy.yml`):**
```yaml
name: Deploy to Production
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Deploy
        run: forge deploy --auto-approve
```

**PR Preview (`.github/workflows/pr-preview.yml`):**
```yaml
name: PR Preview Environment
on:
  pull_request:

jobs:
  preview:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Deploy Preview
        run: forge deploy --namespace=pr-${{ github.event.number }} --auto-approve

      - name: Comment PR
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '🚀 Preview environment deployed!\n\nNamespace: `pr-${{ github.event.number }}`'
            })
```

#### Why Deploy in CI/CD (Not Locally)?

**Problem with local deploys:**
1. Developer A runs `forge deploy` locally at 2:00 PM
2. Developer B pushes code at 2:02 PM → pipeline starts
3. **Concurrent Terraform operations** on same state → **lock conflict**

**Solution: Pipeline-only deploys**
- Only CI/CD runs `forge deploy`
- Developers run `forge plan` locally (preview only)
- State locking prevents pipeline conflicts
- All changes auditable in CI logs

---

### forge destroy

**Tear down infrastructure for a specific namespace.**

#### Syntax

```bash
forge destroy --namespace=<name> [flags]
```

#### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--namespace` | string | **Yes** | Namespace to destroy (prevents accidental production deletion) |
| `--auto-approve` | boolean | No | Skip confirmation prompt |

#### Safety Mechanisms

1. **Namespace required** - Cannot destroy without `--namespace` (prevents production accidents)
2. **Confirmation prompt** - Shows plan before destroying
3. **Preview** - Runs `terraform plan -destroy` first

#### What It Does

1. **Validation** - Ensures `--namespace` is provided
2. **Plan** - Shows what will be destroyed
3. **Confirmation** - Prompts user (unless `--auto-approve`)
4. **Destroy** - Runs `terraform destroy`
5. **Cleanup** - Deletes state file for namespace

#### Examples

**Destroy PR environment:**
```bash
forge destroy --namespace=pr-123
```

**Output:**
```
WARNING: This will destroy all resources in namespace: pr-123

Planning destruction...

Terraform will perform the following actions:

  # aws_lambda_function.api will be destroyed
  - resource "aws_lambda_function" "api" {
      - function_name = "my-app-pr-123-api"
      ...
    }

  # aws_iam_role.lambda will be destroyed
  - resource "aws_iam_role" "lambda" {
      - name = "my-app-pr-123-lambda-role"
      ...
    }

Plan: 0 to add, 0 to change, 5 to destroy.

Are you sure you want to destroy these resources? (yes/no): yes

Destroying resources...
✓ Destroyed namespace: pr-123
```

**Auto-approve (CI/CD):**
```bash
forge destroy --namespace=pr-456 --auto-approve
```

#### PR Cleanup Workflow

**GitHub Actions (`.github/workflows/pr-cleanup.yml`):**
```yaml
name: Cleanup PR Environment
on:
  pull_request:
    types: [closed]

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Destroy Preview Environment
        run: forge destroy --namespace=pr-${{ github.event.number }} --auto-approve

      - name: Comment PR
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '🗑️ Preview environment destroyed.\n\nNamespace: `pr-${{ github.event.number }}`'
            })
```

#### Error Cases

**Missing namespace:**
```bash
$ forge destroy

Error: --namespace is required for destroy (safety check)

This prevents accidental destruction of production resources.
To destroy a specific environment, use:
  forge destroy --namespace=pr-123
```

**Namespace not found:**
```bash
$ forge destroy --namespace=pr-999

Error: No resources found for namespace 'pr-999'
State file does not exist: forge/pr-999-terraform.tfstate
```

---

### forge version

**Show version information for debugging and support.**

#### Syntax

```bash
forge version
```

#### Output

```
Forge v0.2.0
Go version: go1.21.5
Commit: 77be70a
Built: 2025-01-15T14:30:00Z
```

#### Build Information

Version info is set at compile time via `-ldflags`:

```bash
go build -ldflags="\
  -X 'main.version=v0.2.0' \
  -X 'main.commit=$(git rev-parse HEAD)' \
  -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' \
  -X 'main.goVersion=$(go version | awk '{print $3}')' \
" -o bin/forge cmd/forge/main.go
```

---

## Workflows

### Local Development Workflow

```bash
# 1. Create project
forge new my-api --runtime=go

# 2. Develop locally
cd my-api
# ... edit code ...

# 3. Build and test
forge build
go test ./...

# 4. Preview infrastructure changes
forge plan  # (not implemented yet)

# 5. Push to Git
git add .
git commit -m "Add new feature"
git push
# → Pipeline deploys automatically
```

### PR Preview Workflow

```bash
# 1. Create feature branch
git checkout -b feature/new-endpoint

# 2. Make changes
# ... edit code ...

# 3. Push branch
git push origin feature/new-endpoint

# 4. Open PR
# → GitHub Actions deploys preview environment automatically
# → Namespace: pr-123

# 5. Test preview
curl https://pr-123.lambda-url.us-east-1.on.aws/

# 6. Merge PR
# → GitHub Actions destroys preview environment automatically
```

### Multi-Environment Workflow

**Environments:**
- **Production** - `forge deploy` (no namespace)
- **Staging** - `forge deploy --namespace=staging`
- **Development** - `forge deploy --namespace=dev`

**Pipeline structure:**
```yaml
# Deploy to dev on every commit to develop branch
on:
  push:
    branches: [develop]
jobs:
  deploy-dev:
    runs-on: ubuntu-latest
    steps:
      - run: forge deploy --namespace=dev --auto-approve

# Deploy to staging on release branches
on:
  push:
    branches: [release/*]
jobs:
  deploy-staging:
    runs-on: ubuntu-latest
    steps:
      - run: forge deploy --namespace=staging --auto-approve

# Deploy to production on tags
on:
  push:
    tags: [v*]
jobs:
  deploy-production:
    runs-on: ubuntu-latest
    steps:
      - run: forge deploy --auto-approve  # No namespace = production
```

---

## Environment Variables

Forge respects these environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `FORGE_REGION` | Override AWS region | `us-west-2` |
| `AWS_PROFILE` | AWS credentials profile | `my-profile` |
| `AWS_ACCESS_KEY_ID` | AWS access key | `AKIA...` |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | `***` |
| `TF_VAR_namespace` | Terraform namespace variable | `pr-123` |
| `FORGE_VERBOSE` | Enable verbose logging | `true` |

**Example:**
```bash
export FORGE_REGION=eu-west-1
export FORGE_VERBOSE=true
forge deploy
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Generic error (validation, build failure, etc.) |
| `2` | User cancelled operation |
| `3` | AWS authentication error |
| `4` | Terraform error |

**Example usage in scripts:**
```bash
if ! forge build; then
    echo "Build failed"
    exit 1
fi

if ! forge deploy --auto-approve; then
    echo "Deployment failed"
    # Rollback logic here
    exit 1
fi
```

---

## Best Practices

### 1. Always Use Namespaces in CI/CD

**Bad:**
```yaml
# Don't do this - concurrent deploys will conflict
on: pull_request
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - run: forge deploy  # ❌ All PRs deploy to same resources
```

**Good:**
```yaml
on: pull_request
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - run: forge deploy --namespace=pr-${{ github.event.number }}  # ✅ Isolated
```

### 2. Never Deploy from Local Machine

**Why?**
- Concurrent operations cause state conflicts
- Non-reproducible deployments
- No audit trail

**Solution:**
- Use `forge plan` locally (when implemented)
- Let CI/CD handle all `forge deploy` calls

### 3. Always Clean Up Ephemeral Environments

**Bad:**
```yaml
# PR preview workflow without cleanup
# → Orphaned AWS resources → Cost creep
```

**Good:**
```yaml
# PR preview with automatic cleanup
on:
  pull_request:
    types: [closed]
jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - run: forge destroy --namespace=pr-${{ github.event.number }} --auto-approve
```

### 4. Use Auto-State for Team Projects

**Individual:**
```bash
forge new my-experiment --runtime=go
# Manual Terraform state setup OK
```

**Team:**
```bash
forge new team-api --runtime=go --auto-state
# Shared S3 backend for collaboration
```

---

## Troubleshooting

### Build Failures

**Error: "go.mod not found"**
```bash
cd src/functions/api
go mod init api
go mod tidy
```

**Error: "requirements.txt not found"**
```bash
touch src/functions/worker/requirements.txt
forge build
```

### Deployment Failures

**Error: "State lock conflict"**
```
Error: Error acquiring the state lock

Another Terraform operation is in progress.
```

**Solution:**
- Wait for other deployment to finish
- Or manually release lock (dangerous):
  ```bash
  terraform force-unlock <LOCK_ID>
  ```

**Error: "AWS credentials not found"**
```bash
export AWS_ACCESS_KEY_ID=AKIA...
export AWS_SECRET_ACCESS_KEY=...
forge deploy
```

### State Issues

**Error: "State file not found"**
```
Error: Failed to read state
```

**Solution:**
```bash
# Ensure backend.tf exists
ls infra/backend.tf

# Initialize Terraform
cd infra
terraform init
cd ..
```

---

## See Also

- [VISION.md](../VISION.md) - Project philosophy and design
- [TECHNICAL_DECISIONS.md](../TECHNICAL_DECISIONS.md) - Why we made specific choices
- [README.md](../README.md) - Project overview
- [Examples](../examples/) - Sample projects
