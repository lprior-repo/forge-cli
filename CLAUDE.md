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

**Performance Note**: The codebase includes `internal/lingon/aws/` with 2,671 generated AWS resource packages (~1M lines of code) providing complete type-safe Terraform support. These are excluded from test runs by default for performance (saves ~2 minutes per test run). The test tasks automatically filter these out using:
```bash
go list ./internal/... | grep -v '/internal/lingon/aws'
```

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
│                   # Commands: new, init, deploy, destroy, version
├── config/         # HCL configuration loading and validation
├── lingon/         # Type-safe Terraform generation (Lingon integration)
│                   # 170+ Lambda params, 80+ API Gateway, 50+ DynamoDB
│                   # Complete serverless.tf specification support
├── pipeline/       # Pipeline orchestration using functional composition
│                   # Railway-oriented programming with Either monad
├── scaffold/       # Project and stack scaffolding
├── stack/          # Stack detection, dependency graph, topological sort
└── terraform/      # Terraform executor wrapper (uses hashicorp/terraform-exec)
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

## Lingon Integration (Type-Safe Terraform)

Forge supports declarative infrastructure configuration via `forge.yaml` with complete serverless.tf specification:

### Configuration Format
```yaml
service: my-app
provider:
  region: us-east-1

functions:
  api:
    handler: index.handler
    runtime: nodejs20.x
    timeout: 30
    memorySize: 1024
    # ... 170+ Lambda parameters supported

apiGateway:
  name: my-api
  # ... 80+ API Gateway v2 parameters

tables:
  users:
    hashKey: userId
    # ... 50+ DynamoDB parameters
```

See `LINGON_SPEC.md` for complete parameter reference (1,500+ lines) and `examples/forge.yaml` for working examples.

### Variable References
Use `${}` syntax to reference other resources:
```yaml
environment:
  TABLE_NAME: ${tables.users.name}
  QUEUE_URL: ${queues.jobs.url}
```

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

### Adding New Lingon Resource Types
1. Add config types in `internal/lingon/config_types.go`
2. Implement generation logic in `internal/lingon/generator.go`
3. Add validation in `internal/lingon/validation.go`
4. Write tests in `internal/lingon/generator_test.go`

### Debugging Failed Builds
- Build logs: check stdout/stderr from builder functions
- Terraform errors: examine `.terraform/` directory in stack folder
- Test failures: run with `-v` flag for detailed output

## Important Notes

- **HCL Configuration**: Project uses `forge.hcl` (project) and `stack.forge.hcl` (per-stack)
- **YAML Configuration**: Lingon integration uses `forge.yaml` for declarative infrastructure
- **Terraform State**: Each stack maintains independent Terraform state
- **CI/CD**: Auto-detects GitHub Actions/GitLab CI and provides native integrations
- **Inspiration**: Patterns from terraform-exec, Terramate, Atlantis, Lingon

## References

- README.md - User-facing documentation
- LINGON_SPEC.md - Complete Lingon parameter reference (1,500+ lines)
- CODEBASE_AUDIT.md - Code quality audit (Martin Fowler standards)
- TDD_PROGRESS.md - Test-driven development journey
- examples/forge.yaml - Complete working example configuration
