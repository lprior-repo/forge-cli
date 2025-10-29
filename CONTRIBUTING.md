# Contributing to Forge

Thank you for your interest in contributing to Forge! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.22 or later
- Terraform 1.5+ (for integration tests)
- AWS CLI configured (for E2E tests)
- Make

### Getting Started

1. Fork and clone the repository:
```bash
git clone https://github.com/YOUR_USERNAME/forge
cd forge
```

2. Install dependencies:
```bash
task install
```

3. Run tests:
```bash
task test
```

4. Build the binary:
```bash
task build
```

## Project Structure

```
forge/
├── cmd/forge/              # Main entry point
├── internal/
│   ├── terraform/          # Terraform wrapper (uses terraform-exec)
│   ├── stack/              # Stack detection and dependency graph
│   ├── build/              # Build system for different runtimes
│   ├── scaffold/           # Project/stack generation
│   ├── config/             # HCL configuration loading
│   ├── cli/                # Cobra command implementations
│   └── ci/                 # CI/CD environment detection
├── testdata/               # Test fixtures
├── Makefile                # Build tasks
└── README.md
```

## Testing

Forge uses a three-level testing strategy:

### 1. Unit Tests (Default)

Fast tests with no external dependencies. Use mocks for Terraform operations.

```bash
task test
# or
go test -short ./...
```

**When to write:**
- Testing business logic
- Testing options pattern
- Testing dependency graph
- Testing configuration parsing

**Example:**
```go
func TestStackValidate(t *testing.T) {
    stack := &Stack{Name: "test", Runtime: "go1.x"}
    err := stack.Validate()
    assert.NoError(t, err)
}
```

### 2. Integration Tests

Require Terraform binary but no AWS credentials.

```bash
task test:integration
# or
go test -tags=integration ./...
```

**When to write:**
- Testing actual Terraform execution
- Testing build system with real compilers

**Example:**
```go
//go:build integration
// +build integration

func TestRealTerraformInit(t *testing.T) {
    tmpDir := t.TempDir()
    tf, _ := terraform.New(tmpDir, "terraform")
    err := tf.Init(context.Background())
    assert.NoError(t, err)
}
```

### 3. E2E Tests

Require AWS credentials and actually deploy resources.

```bash
task test:e2e
# or
go test -tags=e2e -timeout=30m ./...
```

**When to write:**
- Testing full deployment workflow
- Validating Lambda function creation

## Code Style

### Formatting

```bash
task fmt
```

### Linting

```bash
task lint
```

We use `golangci-lint`. Install from: https://golangci-lint.run/usage/install/

### Naming Conventions

- **Interfaces**: Descriptive names (e.g., `Executor`, `Builder`)
- **Concrete types**: Noun phrases (e.g., `RealExecutor`, `GoBuilder`)
- **Functions**: Verb phrases (e.g., `NewExecutor`, `BuildStack`)
- **Test functions**: `Test<FunctionName>` or `Test<Type>_<Method>`

### Error Handling

Always wrap errors with context:

```go
if err != nil {
    return fmt.Errorf("failed to build stack: %w", err)
}
```

Use custom error types for user-facing errors:

```go
type ValidationError struct {
    Message string
    File    string
    Line    int
}
```

## Adding New Features

### Adding a New Runtime

1. Create a builder in `internal/build/`:
```go
// internal/build/rust_builder.go
type RustBuilder struct{}

func (b *RustBuilder) Build(ctx context.Context, cfg *Config) (*Artifact, error) {
    // Implementation
}
```

2. Register in `init()`:
```go
func init() {
    Register("rust", &RustBuilder{})
}
```

3. Add templates in `internal/scaffold/templates/`:
- `rust_main.rs.tmpl`
- `rust_cargo.toml.tmpl`
- `rust_main.tf.tmpl`

4. Update generator in `internal/scaffold/generator.go`

5. Add tests:
```go
func TestRustBuilder(t *testing.T) {
    // Test implementation
}
```

### Adding a New CLI Command

1. Create command file in `internal/cli/`:
```go
// internal/cli/status.go
func NewStatusCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "status",
        Short: "Show deployment status",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runStatus()
        },
    }
}
```

2. Register in `internal/cli/root.go`:
```go
cmd.AddCommand(
    NewStatusCmd(),
)
```

3. Add tests:
```go
func TestStatusCommand(t *testing.T) {
    // Test implementation
}
```

## Testing Guidelines

### Use t.TempDir()

Always use `t.TempDir()` for temporary files:

```go
func TestScaffold(t *testing.T) {
    tmpDir := t.TempDir()  // Automatically cleaned up
    // Test code
}
```

### Table-Driven Tests

Prefer table-driven tests for multiple scenarios:

```go
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   *Config
        wantErr bool
    }{
        {
            name:    "valid config",
            input:   &Config{Name: "test"},
            wantErr: false,
        },
        {
            name:    "missing name",
            input:   &Config{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.input.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Use testdata/

Place test fixtures in `testdata/`:

```
testdata/
├── basic/
│   ├── forge.hcl
│   └── api/
│       └── stack.forge.hcl
└── multi-function/
    ├── forge.hcl
    ├── shared/
    └── api/
```

## Pull Request Process

1. **Create an issue** describing the feature/bug
2. **Fork and create a branch**: `git checkout -b feature/my-feature`
3. **Write tests** for new functionality
4. **Ensure tests pass**: `task test:all`
5. **Run linter**: `task lint`
6. **Update documentation** (README, CONTRIBUTING, etc.)
7. **Commit with clear messages**:
   ```
   feat: add Rust runtime support

   - Implement RustBuilder
   - Add Rust templates
   - Add integration tests
   ```
8. **Push and create PR**

### Commit Message Format

Follow conventional commits:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `chore:` Build process, dependencies

## Architecture Principles

### 1. Use Proven Patterns

Forge is based on battle-tested patterns from:
- **terraform-exec**: Terraform execution
- **Terramate**: Stack management
- **Standard Go**: Functional options, interfaces

### 2. Keep It Simple

- Prefer composition over inheritance
- Use interfaces for testability
- Avoid premature abstraction

### 3. Make It Testable

- Use dependency injection
- Provide mock implementations
- Keep functions pure when possible

### 4. User Experience First

- Clear error messages with context
- Helpful CLI output
- Sensible defaults

## Questions?

- Open an issue for questions
- Join discussions in GitHub Discussions
- Check existing issues and PRs

## Code of Conduct

Be respectful and constructive in all interactions.

---

Thank you for contributing to Forge! 🚀
