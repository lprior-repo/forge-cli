# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Forge** is a developer-friendly CLI for building and deploying serverless applications on AWS Lambda. It combines Terraform with streamlined Lambda deployment workflows, featuring multi-runtime support (Go, Python, Node.js), dependency management, and incremental deploys.

The codebase follows **functional programming principles** using monadic error handling (Either/Option), pure functions, and immutable data structures from `github.com/IBM/fp-go`.

## Development Commands

### Testing
```bash
# Unit tests (fast, <1s)
make test
go test -v -short ./...

# Integration tests (requires terraform binary, ~10s)
make test-integration
go test -v -tags=integration ./...

# E2E tests (requires AWS credentials, ~30m)
make test-e2e
go test -v -tags=e2e -timeout=30m ./...

# All tests (unit + integration)
make test-all

# Test specific package
go test -v ./internal/build/...
```

### Building
```bash
# Build binary to bin/forge
make build

# Install to $GOPATH/bin
make install
go install ./cmd/forge
```

### Code Quality
```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Tests + linting
make verify

# Tidy dependencies
make tidy
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
