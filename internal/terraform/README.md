# internal/terraform

**Terraform executor wrapper - function types over interfaces for easy testing**

## Overview

The `terraform` package provides a **functional wrapper** around HashiCorp's [terraform-exec](https://github.com/hashicorp/terraform-exec) library. It uses **function types instead of interfaces** for easier testing and better composability.

## Design Philosophy

### Function Types Over Interfaces

**Traditional approach (what we DON'T do):**
```go
type TerraformExecutor interface {
    Init(ctx context.Context, dir string) error
    Plan(ctx context.Context, dir string) (bool, error)
    Apply(ctx context.Context, dir string) error
}

// Testing requires creating mock structs:
type mockExecutor struct{}
func (m *mockExecutor) Init(ctx context.Context, dir string) error { return nil }
func (m *mockExecutor) Plan(ctx context.Context, dir string) (bool, error) { return true, nil }
func (m *mockExecutor) Apply(ctx context.Context, dir string) error { return nil }
```

**Functional approach (what we DO):**
```go
// Function types
type InitFunc func(ctx context.Context, dir string, opts ...InitOption) error
type PlanFunc func(ctx context.Context, dir string, opts ...PlanOption) (bool, error)
type ApplyFunc func(ctx context.Context, dir string, opts ...ApplyOption) error

// Executor is just a struct of functions
type Executor struct {
    Init     InitFunc
    Plan     PlanFunc
    Apply    ApplyFunc
    Destroy  DestroyFunc
    Output   OutputFunc
    Validate ValidateFunc
}

// Testing is trivial - just assign functions:
mockExecutor := Executor{
    Init: func(ctx context.Context, dir string, opts ...InitOption) error {
        return nil
    },
    Plan: func(ctx context.Context, dir string, opts ...PlanOption) (bool, error) {
        return true, nil
    },
}
```

**Benefits:**
- ✅ **Simpler testing** - no mock structs, just functions
- ✅ **Easier composition** - can wrap functions with decorators
- ✅ **Type safety** - same signature enforcement as interfaces
- ✅ **Clear intent** - `Executor.Init` is clearly a function, not a method

## Core Types

### Function Types

```go
type InitFunc func(ctx context.Context, dir string, opts ...InitOption) error
type PlanFunc func(ctx context.Context, dir string, opts ...PlanOption) (bool, error)
type ApplyFunc func(ctx context.Context, dir string, opts ...ApplyOption) error
type DestroyFunc func(ctx context.Context, dir string, opts ...DestroyOption) error
type OutputFunc func(ctx context.Context, dir string) (map[string]interface{}, error)
type ValidateFunc func(ctx context.Context, dir string) error
```

### Executor

Collection of Terraform operation functions:

```go
type Executor struct {
    Init     InitFunc     // terraform init
    Plan     PlanFunc     // terraform plan (returns hasChanges)
    Apply    ApplyFunc    // terraform apply
    Destroy  DestroyFunc  // terraform destroy
    Output   OutputFunc   // terraform output
    Validate ValidateFunc // terraform validate
}
```

## Usage

### Real Executor

```go
import "github.com/lewis/forge/internal/terraform"

// Create real executor (uses terraform binary)
executor := terraform.NewExecutor("terraform")

// terraform init
err := executor.Init(ctx, "./infra")

// terraform plan
hasChanges, err := executor.Plan(ctx, "./infra")

// terraform apply (if changes)
if hasChanges {
    err = executor.Apply(ctx, "./infra")
}

// terraform output
outputs, err := executor.Output(ctx, "./infra")
fmt.Printf("Function URL: %s\n", outputs["function_url"])
```

### With Options

Options use the **functional options pattern**:

```go
// terraform init with upgrade
err := executor.Init(ctx, "./infra",
    terraform.WithUpgrade(true),
)

// terraform plan with variables
hasChanges, err := executor.Plan(ctx, "./infra",
    terraform.WithVar("namespace", "pr-123"),
    terraform.WithVar("region", "us-west-2"),
)

// terraform apply with auto-approve
err := executor.Apply(ctx, "./infra",
    terraform.WithAutoApprove(true),
    terraform.WithVar("namespace", "pr-123"),
)

// terraform destroy with auto-approve
err := executor.Destroy(ctx, "./infra",
    terraform.WithAutoApprove(true),
    terraform.WithVar("namespace", "pr-123"),
)
```

### Mock Executor (Testing)

```go
// Create mock executor for tests
mockExecutor := terraform.NewMockExecutor()

// All operations succeed by default
err := mockExecutor.Init(ctx, "./infra")  // nil
hasChanges, err := mockExecutor.Plan(ctx, "./infra")  // true, nil
err = mockExecutor.Apply(ctx, "./infra")  // nil
```

**Custom mock behavior:**
```go
mockExecutor := terraform.Executor{
    Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
        // Custom test behavior
        if dir == "./bad-dir" {
            return errors.New("directory not found")
        }
        return nil
    },
    Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
        // Simulate no changes
        return false, nil
    },
}
```

## Options Pattern

### InitOptions

```go
type InitOptions struct {
    Upgrade bool    // -upgrade flag
}

type InitOption func(*InitOptions)

func WithUpgrade(upgrade bool) InitOption {
    return func(opts *InitOptions) {
        opts.Upgrade = upgrade
    }
}
```

**Usage:**
```go
err := executor.Init(ctx, "./infra", terraform.WithUpgrade(true))
```

### PlanOptions

```go
type PlanOptions struct {
    Vars     map[string]string  // -var flags
    Detailed bool               // -detailed-exitcode
}

type PlanOption func(*PlanOptions)

func WithVar(key, value string) PlanOption {
    return func(opts *PlanOptions) {
        if opts.Vars == nil {
            opts.Vars = make(map[string]string)
        }
        opts.Vars[key] = value
    }
}
```

**Usage:**
```go
hasChanges, err := executor.Plan(ctx, "./infra",
    terraform.WithVar("namespace", "pr-123"),
    terraform.WithVar("region", "us-west-2"),
)
```

### ApplyOptions

```go
type ApplyOptions struct {
    Vars        map[string]string
    AutoApprove bool  // -auto-approve flag
}

type ApplyOption func(*ApplyOptions)

func WithAutoApprove(autoApprove bool) ApplyOption {
    return func(opts *ApplyOptions) {
        opts.AutoApprove = autoApprove
    }
}
```

**Usage:**
```go
err := executor.Apply(ctx, "./infra",
    terraform.WithAutoApprove(true),
    terraform.WithVar("namespace", "pr-123"),
)
```

## Real Implementation

### terraform init

```go
func makeInitFunc(tfPath string) InitFunc {
    return func(ctx context.Context, dir string, opts ...InitOption) error {
        tf, err := tfexec.NewTerraform(dir, tfPath)
        if err != nil {
            return fmt.Errorf("failed to create terraform instance: %w", err)
        }

        // Apply options
        options := &InitOptions{}
        for _, opt := range opts {
            opt(options)
        }

        // Build terraform init options
        initOpts := []tfexec.InitOption{}
        if options.Upgrade {
            initOpts = append(initOpts, tfexec.Upgrade(true))
        }

        // Execute
        return tf.Init(ctx, initOpts...)
    }
}
```

### terraform plan

```go
func makePlanFunc(tfPath string) PlanFunc {
    return func(ctx context.Context, dir string, opts ...PlanOption) (bool, error) {
        tf, err := tfexec.NewTerraform(dir, tfPath)
        if err != nil {
            return false, fmt.Errorf("failed to create terraform instance: %w", err)
        }

        // Apply options
        options := &PlanOptions{Detailed: true}
        for _, opt := range opts {
            opt(options)
        }

        // Build terraform plan options
        planOpts := []tfexec.PlanOption{tfexec.Out("tfplan")}
        for key, value := range options.Vars {
            planOpts = append(planOpts, tfexec.Var(key+"="+value))
        }

        // Execute
        hasChanges, err := tf.Plan(ctx, planOpts...)
        return hasChanges, err
    }
}
```

### terraform apply

```go
func makeApplyFunc(tfPath string) ApplyFunc {
    return func(ctx context.Context, dir string, opts ...ApplyOption) error {
        tf, err := tfexec.NewTerraform(dir, tfPath)
        if err != nil {
            return fmt.Errorf("failed to create terraform instance: %w", err)
        }

        // Apply options
        options := &ApplyOptions{}
        for _, opt := range opts {
            opt(options)
        }

        // Build terraform apply options
        applyOpts := []tfexec.ApplyOption{}
        if options.AutoApprove {
            applyOpts = append(applyOpts, tfexec.AutoApprove(true))
        }
        for key, value := range options.Vars {
            applyOpts = append(applyOpts, tfexec.Var(key+"="+value))
        }

        // Execute
        return tf.Apply(ctx, applyOpts...)
    }
}
```

## Error Handling

Errors are wrapped with context:

```go
if err := executor.Init(ctx, "./infra"); err != nil {
    return fmt.Errorf("terraform init failed: %w", err)
}

if err := executor.Apply(ctx, "./infra"); err != nil {
    return fmt.Errorf("terraform apply failed: %w", err)
}
```

## Testing

### Unit Tests with Mock Executor

```go
func TestDeployWithChanges(t *testing.T) {
    mockExecutor := terraform.Executor{
        Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
            return nil
        },
        Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
            return true, nil  // Has changes
        },
        Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
            return nil
        },
    }

    // Use mockExecutor in pipeline
    result := deployPipeline(mockExecutor)

    assert.NoError(t, result)
}

func TestDeployNoChanges(t *testing.T) {
    mockExecutor := terraform.Executor{
        Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
            return nil
        },
        Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
            return false, nil  // No changes
        },
        Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
            t.Fatal("Apply should not be called when no changes")
            return nil
        },
    }

    result := deployPipeline(mockExecutor)

    assert.NoError(t, result)
}
```

## Files

- **`executor.go`** - `Executor` struct, `NewExecutor`, `NewMockExecutor`
- **`real.go`** - Real implementation using `terraform-exec`
- **`options.go`** - Functional options (`WithVar`, `WithAutoApprove`, etc.)
- **`errors.go`** - Error types and wrapping

## Dependencies

```go
import "github.com/hashicorp/terraform-exec/tfexec"  // Terraform executor library
```

## Design Principles

1. **Function types over interfaces** - easier testing, better composability
2. **Functional options** - clean API for optional parameters
3. **Context propagation** - all operations accept `context.Context`
4. **Error wrapping** - clear error messages with context
5. **Zero magic** - thin wrapper, delegates to `terraform-exec`

## Future Enhancements

- [ ] Terraform version detection and validation
- [ ] Plan file parsing (detailed diff output)
- [ ] State manipulation functions (import, mv, rm)
- [ ] Workspace management (create, switch, delete)
- [ ] Terraform Cloud/Enterprise integration
- [ ] Parallel execution of independent stacks
