# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## ⚠️ CRITICAL: Code Quality Standards (HIGHEST PRIORITY)

**These standards are MANDATORY and of the HIGHEST MAGNITUDE:**

### 1. Test Coverage Requirement: 90% MINIMUM
- **Aggregate coverage across all packages must be ≥90%**
- **Includes unit tests, integration tests, and E2E tests**
- **NO exceptions for I/O boundary code** - even CLI commands must have comprehensive tests
- Run `task coverage:check` to verify compliance
- Run `task test:all` to execute all test suites

### 2. Linting Requirement: ZERO Issues
- **Zero linting errors or warnings allowed**
- Uses `golangci-lint` with comprehensive linter configuration
- Run `task lint` before every commit
- Configuration in `.golangci.yml`

### 3. Test Requirements: ALL Tests Must Pass
- **100% test pass rate required**
- **NO FAILURES allowed** in any test suite:
  - Unit tests: `task test:unit`
  - Integration tests: `task test:integration`
  - E2E tests: `task test:e2e`
  - All tests: `task test:all`

### 4. Mutation Testing
- Mutation score should be ≥80% for critical packages
- Run `task mutation` to verify test suite quality
- Ensures tests actually catch bugs, not just pass

**ENFORCEMENT**: CI/CD pipeline will reject any PR that violates these standards.

## Project Overview

**Forge** is a developer-friendly CLI for building and deploying serverless applications on AWS Lambda. It combines Terraform with streamlined Lambda deployment workflows, featuring multi-runtime support (Go, Python, Node.js), dependency management, and incremental deploys.

The codebase follows **functional programming principles** using monadic error handling (Either/Option), pure functions, and immutable data structures from `github.com/IBM/fp-go`.

## Project Vision & Philosophy

### The Problem Forge Solves

Serverless deployment tools fall into two camps:
1. **Too opinionated** (Serverless Framework, SAM): Lock you into YAML configs, proprietary patterns, vendor lock-in
2. **Too low-level** (raw Terraform): Verbose, repetitive, steep learning curve

**Forge bridges this gap**: Convention over configuration + raw Terraform power + zero lock-in.

### Core Principles

1. **Convention Over Configuration (Omakase)**
   - Inspired by Ruby on Rails and DHH's philosophy
   - Zero config files (`forge.yaml`, `serverless.yml`, etc.)
   - Smart defaults that just work
   - Exit ramp: customize Terraform directly when needed

2. **Pure Functional Programming**
   - Data/Actions/Calculations separation
   - Pure core, imperative shell
   - Railway-oriented programming (Either monad)
   - Immutable data structures
   - No hidden state, no surprises

3. **Minimal Magic, Maximum Control**
   - No black box abstractions
   - Generated Terraform is readable and editable
   - Conventions are discoverable (scan `src/functions/*`)
   - Developer owns the infrastructure code

4. **Production-Ready from Day 1**
   - Ephemeral PR environments built-in
   - Terraform state management handled
   - Multi-region support
   - Cost tracking via namespace tags

### The Forge Workflow

```bash
# 1. Scaffold new project (convention-based)
forge new my-app --runtime=go --auto-state

# 2. Build functions (auto-discovery)
forge build

# 3. Deploy to production
forge deploy

# 4. Preview PR changes (ephemeral env)
forge deploy --namespace=pr-123

# 5. Cleanup
forge destroy --namespace=pr-123
```

### What Makes Forge Different

| Feature | Forge | Serverless Framework | SAM | Raw Terraform |
|---------|-------|---------------------|-----|---------------|
| Config files | **0** (convention) | serverless.yml | template.yaml | *.tf |
| Lock-in | **None** | High | Medium | None |
| Terraform control | **Full** | Hidden | None | Full |
| PR previews | **Built-in** | Plugin | Manual | Manual |
| State management | **Auto** | N/A | N/A | Manual |
| Learning curve | **Low** | Medium | Medium | High |
| Exit strategy | **Edit .tf** | Eject | Switch tools | N/A |

## Core Philosophy: Convention Over Configuration

**Forge follows Omakase principles** (inspired by Ruby on Rails and DHH's philosophy):

### Zero Configuration Files
- **NO `forge.yaml`** - Everything is convention-based
- **NO template files** - Code generation is imperative
- **Exit ramp provided** - Users can customize Terraform directly

### Required Project Structure
```
my-app/
├── infra/              # REQUIRED: Terraform infrastructure
│   ├── main.tf         # Define AWS resources explicitly
│   ├── variables.tf    # namespace variable for ephemeral envs
│   └── outputs.tf
└── src/                # OPTIONAL: Application code (any structure)
    └── functions/      # Convention: Lambda functions here
        ├── api/        # Function name = directory name
        │   └── main.go # Runtime detected from entry file
        └── worker/
            └── index.js
```

### Convention-Based Discovery
Forge **scans `src/functions/*`** to automatically detect:
- **Function names**: Directory name (e.g., `api`, `worker`)
- **Runtimes**: Detected from entry files:
  - `main.go` or `*.go` → Go (provided.al2023)
  - `index.js`, `index.mjs`, `handler.js` → Node.js (nodejs20.x)
  - `app.py`, `lambda_function.py`, `handler.py` → Python (python3.13)
- **Build targets**: Automatically builds to `.forge/build/{name}.zip`

### SAM-Inspired Workflow
```bash
# 1. Create new project
forge new my-app --runtime=go
  → Generates infra/ with example Lambda
  → Generates src/functions/api/ with hello-world

# 2. Build all functions
forge build
  → Scans src/functions/*
  → Detects runtimes automatically
  → Builds each to .forge/build/*.zip
  → Creates stub zips first (for terraform init)

# 3. Deploy (build + terraform apply)
forge deploy
  → Runs forge build
  → Runs terraform init/plan/apply
  → All in one command!

# 4. Ephemeral PR environments
forge deploy --namespace=pr-123
  → Sets TF_VAR_namespace=pr-123
  → All resources prefixed: my-app-pr-123-api
  → Isolated preview environment

# 5. Cleanup
forge destroy --namespace=pr-123
  → Tears down ephemeral environment
```

### Developer Responsibilities (NOT Forge's Job)
Forge is **minimal by design**. The developer handles:
- **Dependencies**: go.mod, requirements.txt, package.json (per function)
- **Shared code**: Organize as needed, ensure it compiles
- **Secrets**: .env files, AWS Secrets Manager, SSM Parameter Store
- **IAM permissions**: Define in Terraform
- **API Gateway routing**: Define in Terraform
- **VPC configuration**: Define in Terraform
- **Environment variables**: Define in Terraform Lambda resources
- **Local testing**: Use AWS SAM, LocalStack, or similar tools
- **Logs**: Use AWS CloudWatch directly (or `forge logs` when added)
- **Cost management**: Tag resources in Terraform

### Terraform Integration
Terraform is the **source of truth** for infrastructure:

```hcl
# infra/main.tf
variable "namespace" {
  type    = string
  default = ""
}

resource "aws_lambda_function" "api" {
  function_name    = "${var.namespace}my-app-api"
  runtime          = "go1.x"
  handler          = "bootstrap"
  filename         = "../.forge/build/api.zip"
  source_code_hash = filebase64sha256("../.forge/build/api.zip")

  environment {
    variables = {
      TABLE_NAME = aws_dynamodb_table.users.name  # Explicit wiring
    }
  }
}
```

Forge reads this to know what to build, then runs `terraform apply`.

### Ephemeral Pipeline Pattern
When PR is opened:
```yaml
# .github/workflows/pr-preview.yml (user creates this)
name: PR Preview
on: pull_request
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: forge deploy --namespace=pr-${{ github.event.number }}
```

When PR is closed:
```yaml
# .github/workflows/pr-cleanup.yml
on:
  pull_request:
    types: [closed]
jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: forge destroy --namespace=pr-${{ github.event.number }}
```

This creates isolated AWS environments per PR with zero configuration.

## Feature Roadmap & Implementation Status

### ✅ Phase 1: Core Foundation (COMPLETE)
- [x] Convention-based function discovery (`src/functions/*`)
- [x] Runtime auto-detection (Go, Python, Node.js)
- [x] Functional pipeline architecture (pure core + imperative shell)
- [x] `forge build` - Build all functions with caching
- [x] `forge deploy` - Deploy with Terraform
- [x] Namespace support for ephemeral environments
- [x] `forge new` - Scaffold convention-based projects

### 🚧 Phase 2: Production Readiness (IN PROGRESS)
- [x] Terraform state management design
- [ ] **`forge new --auto-state`** - Auto-provision S3 bucket + DynamoDB for state
- [ ] Backend.tf generation with namespace-aware state keys
- [ ] Complete Terraform templates (Lambda + IAM + Function URLs)
- [ ] AWS credential validation
- [ ] Multi-account support (AWS profiles)

### 📋 Phase 3: Developer Experience (PLANNED)
- [ ] **Interactive TUI** for project setup (bubbletea)
  - Runtime selection
  - AWS account/region picker
  - State bucket configuration
  - Visual feedback during deployment
- [ ] `forge logs` - Tail CloudWatch logs by namespace
- [ ] `forge list` - Show all deployed namespaces
- [ ] `forge destroy` - Enhanced with namespace discovery
- [ ] Hot reload / watch mode (`forge watch`)

### 📋 Phase 4: CI/CD Integration (PLANNED)
- [ ] GitHub Actions workflow generation
- [ ] GitLab CI pipeline generation
- [ ] Automatic PR environment provisioning
- [ ] Automatic cleanup on PR close
- [ ] Cost estimation per PR
- [ ] Deployment status comments on PRs

### 📋 Phase 5: Advanced Features (PLANNED)
- [ ] Lambda Layers support for shared dependencies
- [ ] Multi-function projects (API Gateway routing)
- [ ] DynamoDB table auto-provisioning
- [ ] SQS/SNS auto-wiring
- [ ] Custom domains with Route53
- [ ] VPC configuration helpers
- [ ] Secrets management integration (AWS Secrets Manager)

### 📋 Phase 6: Observability (PLANNED)
- [ ] Cost tracking dashboard per namespace
- [ ] Performance metrics
- [ ] Error rate monitoring
- [ ] Deployment rollback support
- [ ] State drift detection

## Terraform State Management Design

### Problem
Terraform requires remote state for team collaboration and PR environments. Manual S3 bucket setup is tedious and error-prone.

### Solution: `forge new --auto-state`

**What it does:**
1. Detects AWS credentials (profile, env vars, or prompt)
2. Creates S3 bucket: `{project-name}-terraform-state`
3. Enables versioning and encryption
4. Creates DynamoDB table: `{project-name}-state-lock`
5. Generates `infra/backend.tf` with dynamic state keys

**Generated backend.tf:**
```hcl
terraform {
  backend "s3" {
    bucket         = "my-app-terraform-state"
    key            = "forge/${var.namespace}terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "my-app-state-lock"
  }
}
```

**Namespace-aware state:**
- Default deployment: `forge/terraform.tfstate`
- PR environment: `forge/pr-123-terraform.tfstate`
- Each namespace has isolated state

**Pure functional design:**
```go
// PURE: Generate backend config (calculation)
func GenerateBackendTF(opts StateConfig) string

// ACTION: Provision S3 bucket (I/O)
func ProvisionStateBucket(cfg StateConfig) Either[error, BucketInfo]

// ACTION: Provision DynamoDB table (I/O)
func ProvisionLockTable(cfg StateConfig) Either[error, TableInfo]
```

### Interactive TUI Design (Phase 3)

**Bubbletea-based TUI** following The Elm Architecture (functional!):
```go
type Model struct {
    // Immutable state
    Step       SetupStep
    Runtime    string
    Region     string
    Profile    string
    AutoState  bool
    BucketName string
}

type Msg interface{} // Messages (events)

// Pure: Update model based on message
func (m Model) Update(msg Msg) (Model, tea.Cmd)

// Pure: Render view from model
func (m Model) View() string
```

**User experience:**
```
┌─────────────────────────────────────────┐
│  🔨 Forge Project Setup                 │
├─────────────────────────────────────────┤
│                                         │
│  Project: my-app                        │
│                                         │
│  Select Runtime:                        │
│    ○ Go                                 │
│    ● Python         ← (selected)       │
│    ○ Node.js                            │
│                                         │
│  ⚙️  AWS Configuration                  │
│    Region:  [us-east-1        ▼]       │
│    Profile: [default          ▼]       │
│                                         │
│  🗄️  Terraform State                    │
│    ☑ Auto-create S3 bucket             │
│    Bucket: my-app-terraform-state       │
│    ☑ Enable state locking (DynamoDB)   │
│                                         │
│  ✓ AWS credentials validated            │
│                                         │
│  [Continue ⏎] [Cancel ^C]              │
└─────────────────────────────────────────┘
```

This creates isolated AWS environments per PR with zero configuration.

## Development Commands

**IMPORTANT**: This project uses **Taskfile** (not Make) for all development tasks. Run `task` or `task --list` to see available commands.

### Testing
```bash
# Unit tests (fast, ~5s) - Excludes AWS resources for speed
task test
task test:unit

# Integration tests (requires terraform binary, ~10s)
task test:integration

# All tests (unit + integration)
task test:all

# Coverage reports
task coverage              # Text coverage report
task coverage:html         # HTML coverage report in browser
task coverage:check        # Verify 90% coverage threshold

# Benchmarks
task bench                 # Run benchmarks

# Mutation Testing
task mutation              # Run mutation testing on all non-generated code
task mutation:verbose      # Verbose mutation testing output
task mutation:package PKG=internal/build  # Test specific package
```

**Performance Note**: Tests run quickly (~5s for unit tests) with no external dependencies required for the core test suite.

**Mutation Testing**: Uses [go-mutesting](https://github.com/avito-tech/go-mutesting) to verify test suite quality by introducing mutations (bugs) and checking if tests catch them. The mutation score indicates the percentage of mutations killed by tests (higher is better). Requires Nushell (`nu`).

### Building
```bash
# Build binary to bin/forge
task build

# Install dependencies
task install
```

### Code Quality
```bash
# Format code
task fmt

# Run go vet
task vet

# Run linter (requires golangci-lint)
task lint

# Full CI checks (fmt, vet, test, coverage)
task ci

# Full CI with integration tests
task ci:full

# Clean artifacts
task clean
```

## Architecture

### Package Structure
```
internal/
├── build/          # Build system with runtime-specific builders (Go, Python, Node.js)
│                   # Pure functions, functional decorators (WithCache, WithLogging)
├── cli/            # Cobra commands (I/O boundary)
│                   # Commands: new, add, build, deploy, destroy, version
├── config/         # HCL configuration loading and validation
├── generators/     # Code generators for Python Lambda projects
│   └── python/     # Python-specific generators with tfmodules integration
├── pipeline/       # Pipeline orchestration using functional composition
│                   # Railway-oriented programming with Either monad
├── scaffold/       # Project and stack scaffolding
├── stack/          # Stack detection, dependency graph, topological sort
├── terraform/      # Terraform executor wrapper (uses hashicorp/terraform-exec)
└── tfmodules/      # Type-safe Terraform module wrappers
    ├── apigateway/ # API Gateway v2 module
    ├── dynamodb/   # DynamoDB table module
    ├── hclgen/     # HCL generation engine (reflection-based)
    ├── lambda/     # Lambda function module
    ├── s3/         # S3 bucket module
    ├── sns/        # SNS topic module
    └── sqs/        # SQS queue module
```

### Functional Programming Patterns

**Either Monad** - All fallible operations return `E.Either[error, T]`:
```go
func Build(ctx context.Context, cfg Config) E.Either[error, Artifact] {
    // Railway-oriented programming - automatic error short-circuiting
}
```

**Option Monad** - Optional values use `O.Option[T]` (no nil checks):
```go
func (r *Registry) Get(runtime string) O.Option[BuildFunc] {
    // Type-safe optional handling
}
```

**Pure Functions** - Core logic has no side effects:
```go
// Same inputs always produce same outputs
func generateStack(config ForgeConfig) (*Stack, error)
```

**Function Composition** - Decorators and higher-order functions:
```go
// Composable build decorators
cachedBuild := WithCache(cache)(WithLogging(logger)(baseBuild))
```

### Key Design Patterns
- **Repository Pattern**: terraform executor abstraction
- **Strategy Pattern**: build registry for different runtimes
- **Decorator Pattern**: WithCache, WithLogging build wrappers
- **Pipeline Pattern**: composable deployment stages
- **Registry Pattern**: runtime builder registry

### Testing Strategy (TDD)
- **226 total tests** (189 unit, 37 integration) with 100% pass rate
- **~85% coverage** on functional code
- **Fast unit tests** (<1s) with no external dependencies
- **Integration tests** verify Terraform integration
- **Property-based thinking** (formal property tests coming in Phase 6)

## Type-Safe Terraform Generation (tfmodules)

Forge uses **type-safe Go structs** to generate Terraform HCL instead of string concatenation or YAML configuration files. This provides compile-time safety, IDE autocomplete, and refactoring support.

### tfmodules Pattern
All Terraform modules follow this pattern:

```go
// Define module configuration using Go structs
module := &lambda.Module{
    FunctionName: "my-api",
    Runtime:      "python3.13",
    Handler:      "service.handlers.handle_request.lambda_handler",
    Timeout:      30,
    MemorySize:   1024,
}

// Generate HCL using reflection-based engine
hcl, err := module.Configuration()
```

### HCL Generation Engine (hclgen)

The `internal/tfmodules/hclgen` package provides reflection-based HCL generation:

```go
// ToHCL converts any Go struct to Terraform HCL
func ToHCL(localName, source, version string, config interface{}) (string, error)
```

**Features:**
- Automatic struct field → HCL attribute conversion
- Terraform reference detection (`${var.namespace}`)
- Nested block support (maps, slices, structs)
- Zero-value omission (only set fields are rendered)

### Supported Modules
- **Lambda**: `internal/tfmodules/lambda` - 170+ parameters
- **API Gateway v2**: `internal/tfmodules/apigateway` - 80+ parameters
- **DynamoDB**: `internal/tfmodules/dynamodb` - 50+ parameters
- **S3**: `internal/tfmodules/s3` - Bucket configuration
- **SNS**: `internal/tfmodules/sns` - Topic configuration
- **SQS**: `internal/tfmodules/sqs` - Queue configuration

## Key Implementation Details

### Build System
- **Pluggable builders** via Registry pattern
- **Runtime support**: Go (GOOS=linux GOARCH=amd64), Python (pip/poetry), Node.js (npm/yarn/pnpm)
- **Caching**: SHA256-based build artifact caching
- **Functional decorators**: composable build enhancements

### Dependency Management
- **Topological sorting** of stack dependencies via `after` relationships
- **Parallel deployment** when stacks are independent
- **Cycle detection** prevents circular dependencies

### Terraform Integration
- Uses `hashicorp/terraform-exec` for reliable Terraform operations
- **Function types** over interfaces for easy testing
- **Functional wrapper**: `InitFunc`, `PlanFunc`, `ApplyFunc`, etc.

### Error Handling Philosophy
- **Either monad** forces explicit error handling at compile time
- **No panics** in library code (only CLI entry point)
- **Railway-oriented programming**: automatic short-circuiting on errors
- **Descriptive errors** with context wrapping

## Code Quality Standards

This codebase follows Martin Fowler's principles:
- **Pure functions**: no side effects in core logic
- **Immutable data**: all config/artifact structs are immutable
- **Single Responsibility**: each package has one clear purpose
- **Zero code duplication**: DRY principle strictly applied
- **Minimal dependencies**: 8 direct dependencies, all stable and well-maintained
- **Self-documenting code**: clear naming eliminates most comments

Audit rating: **9/10** (see `CODEBASE_AUDIT.md`)

## Common Development Workflows

### Adding a New Runtime
1. Implement `BuildFunc` in `internal/build/`
2. Register in `internal/build/functional.go` registry
3. Add tests in `internal/build/*_test.go`
4. Update scaffold templates in `internal/scaffold/`

### Adding New Terraform Modules
1. Create new package in `internal/tfmodules/{service}/`
2. Define module struct with all parameters in `types.go`
3. Implement `Configuration()` method using `hclgen.ToHCL()`
4. Write comprehensive tests with 90%+ coverage
5. Update generators to use the new module

### Debugging Failed Builds
- Build logs: check stdout/stderr from builder functions
- Terraform errors: examine `.terraform/` directory in stack folder
- Test failures: run with `-v` flag for detailed output

## Important Notes

- **Convention-Based**: Zero configuration files - project structure determines behavior
- **Type-Safe Generation**: Go structs → HCL via reflection-based `hclgen` engine
- **Terraform State**: Namespace-aware state management for ephemeral environments
- **CI/CD**: Auto-detects GitHub Actions/GitLab CI and provides native integrations
- **Inspiration**: Patterns from terraform-exec, Terramate, Atlantis, AWS SAM

## References

- README.md - User-facing documentation
- TDD_PROGRESS.md - Test-driven development journey
- TECHNICAL_DECISIONS.md - Architecture and design decisions
- VISION.md - Product vision and philosophy
- internal/tfmodules/ - Type-safe Terraform module implementations
