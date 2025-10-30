# Forge API Reference

Go package documentation for building tools on top of Forge or extending Forge's functionality.

## Table of Contents

- [Overview](#overview)
- [Core Packages](#core-packages)
  - [internal/build](#internalbuild)
  - [internal/discovery](#internaldiscovery)
  - [internal/pipeline](#internalpipeline)
  - [internal/terraform](#internalterraform)
  - [internal/agent](#internalagent)
- [Usage Examples](#usage-examples)
- [Testing](#testing)

---

## Overview

Forge's internal packages follow **functional programming principles** with strict separation of concerns:

- **Data** - Immutable structs with no methods
- **Calculations** - Pure functions (same input → same output)
- **Actions** - I/O functions at system edges

All packages are designed to be **composable**, **testable**, and **predictable**.

---

## Core Packages

### internal/build

**Multi-runtime Lambda function build system with functional composition.**

#### Key Types

```go
// Config holds build configuration (immutable)
type Config struct {
    SourceDir  string            // Source code directory
    OutputPath string            // Output .zip path
    Handler    string            // Handler name
    Runtime    string            // Runtime identifier
    Env        map[string]string // Build environment variables
}

// Artifact represents a built artifact (immutable)
type Artifact struct {
    Path     string  // Path to .zip file
    Checksum string  // SHA256 checksum
    Size     int64   // File size in bytes
}

// BuildFunc is the core abstraction - a pure function
type BuildFunc func(context.Context, Config) E.Either[error, Artifact]

// Registry maps runtimes to their build functions
type Registry map[string]BuildFunc
```

#### Core Functions

**Create Registry:**
```go
registry := build.NewRegistry()
// Supports: go1.x, python3.9-3.13, nodejs18.x-22.x, java11-21
```

**Get Builder:**
```go
builderOpt := build.GetBuilder(registry, "go1.x")

result := O.Fold(
    func() E.Either[error, build.Artifact] {
        return E.Left[build.Artifact](errors.New("runtime not found"))
    },
    func(builder build.BuildFunc) E.Either[error, build.Artifact] {
        cfg := build.Config{
            SourceDir:  "./src/functions/api",
            OutputPath: ".forge/build/api.zip",
            Runtime:    "go1.x",
            Handler:    "bootstrap",
        }
        return builder(ctx, cfg)
    },
)(builderOpt)
```

**Build Multiple Functions:**
```go
configs := []build.Config{
    {SourceDir: "./src/functions/api", Runtime: "go1.x", ...},
    {SourceDir: "./src/functions/worker", Runtime: "python3.13", ...},
}

result := build.BuildAll(ctx, configs, registry)

if E.IsRight(result) {
    artifacts := E.GetOrElse(func() []build.Artifact { return nil })(result)
    fmt.Printf("Built %d functions\n", len(artifacts))
}
```

**Composable Decorators:**
```go
// Add caching
cachedBuild := build.WithCache(cache)(baseBuild)

// Add logging
loggedBuild := build.WithLogging(logger)(baseBuild)

// Compose multiple decorators
enhancedBuild := build.Compose(
    build.WithCache(cache),
    build.WithLogging(logger),
)(baseBuild)
```

#### Example: Custom Builder

```go
// Create custom builder for Rust
rustBuild := func(ctx context.Context, cfg build.Config) E.Either[error, build.Artifact] {
    // 1. Build Rust binary
    cmd := exec.CommandContext(ctx, "cargo", "build", "--release", "--target", "x86_64-unknown-linux-musl")
    cmd.Dir = cfg.SourceDir
    if err := cmd.Run(); err != nil {
        return E.Left[build.Artifact](err)
    }

    // 2. Create zip
    binary := filepath.Join(cfg.SourceDir, "target/x86_64-unknown-linux-musl/release/bootstrap")
    if err := zipFile(binary, cfg.OutputPath); err != nil {
        return E.Left[build.Artifact](err)
    }

    // 3. Calculate checksum
    checksum, _ := calculateChecksum(cfg.OutputPath)
    size, _ := getFileSize(cfg.OutputPath)

    return E.Right[error](build.Artifact{
        Path:     cfg.OutputPath,
        Checksum: checksum,
        Size:     size,
    })
}

// Register custom builder
registry := build.NewRegistry()
registry["rust"] = rustBuild
```

---

### internal/discovery

**Convention-based function discovery - SAM-like auto-detection.**

#### Key Types

```go
// Function represents a discovered Lambda function (immutable)
type Function struct {
    Name       string // Function name (directory name)
    Path       string // Absolute path to function source
    Runtime    string // Detected runtime
    EntryPoint string // Entry file name
}
```

#### Core Functions

**Scan Functions:**
```go
functions, err := discovery.ScanFunctions("/path/to/project")
if err != nil {
    log.Fatal(err)
}

for _, fn := range functions {
    fmt.Printf("Found: %s (%s) at %s\n", fn.Name, fn.Runtime, fn.Path)
}
```

**Convert to Build Config:**
```go
buildConfigs := lo.Map(functions, func(f discovery.Function, _ int) build.Config {
    return discovery.ToBuildConfig(f, ".forge/build")
})

// Now pass to build.BuildAll()
result := build.BuildAll(ctx, buildConfigs, registry)
```

#### Example: Custom Runtime Detection

```go
// Add custom detection for Rust
func detectRustRuntime(functionPath string) (string, string, error) {
    if fileExists(functionPath, "Cargo.toml") {
        return "rust", "Cargo.toml", nil
    }
    return "", "", fmt.Errorf("not a Rust function")
}

// Use in discovery pipeline
functions, _ := discovery.ScanFunctions("/project")
rustFunctions := lo.Filter(functions, func(f discovery.Function, _ int) bool {
    runtime, _, _ := detectRustRuntime(f.Path)
    return runtime == "rust"
})
```

---

### internal/pipeline

**Functional pipeline orchestration for multi-stage workflows.**

#### Key Types

```go
// State carries data through the pipeline (immutable)
type State struct {
    ProjectDir string
    Artifacts  map[string]Artifact
    Outputs    map[string]interface{}
    Config     interface{}
}

// Stage is a function that transforms state
type Stage func(context.Context, State) E.Either[error, State]

// Pipeline composes stages functionally
type Pipeline struct {
    stages []Stage
}
```

#### Core Functions

**Create Pipeline:**
```go
buildStage := func(ctx context.Context, s State) E.Either[error, State] {
    // Build functions
    artifacts, err := build.BuildAll(ctx, ...)
    if err != nil {
        return E.Left[State](err)
    }
    s.Artifacts = artifacts
    return E.Right[error](s)
}

deployPipeline := pipeline.New(
    discoveryStage,
    buildStage,
    initStage,
    planStage,
    applyStage,
)
```

**Run Pipeline:**
```go
initialState := pipeline.State{ProjectDir: "."}
result := pipeline.Run(deployPipeline, ctx, initialState)

E.Fold(
    func(err error) { fmt.Printf("Failed: %v\n", err) },
    func(s State) { fmt.Println("Success!") },
)(result)
```

**Compose Pipelines:**
```go
buildPipeline := pipeline.New(discoveryStage, buildStage)
terraformPipeline := pipeline.New(initStage, planStage, applyStage)

// Chain into single pipeline
deployPipeline := pipeline.Chain(buildPipeline, terraformPipeline)
```

#### Example: Custom Stage

```go
// Create validation stage
validationStage := func(ctx context.Context, s State) E.Either[error, State] {
    // Validate project structure
    if _, err := os.Stat(filepath.Join(s.ProjectDir, "forge.hcl")); err != nil {
        return E.Left[State](fmt.Errorf("forge.hcl not found"))
    }

    if _, err := os.Stat(filepath.Join(s.ProjectDir, "infra/main.tf")); err != nil {
        return E.Left[State](fmt.Errorf("infra/main.tf not found"))
    }

    // State passes through unchanged
    return E.Right[error](s)
}

// Add to pipeline
pipeline := pipeline.New(
    validationStage,    // ← Custom stage
    discoveryStage,
    buildStage,
    deployStage,
)
```

---

### internal/terraform

**Terraform executor wrapper - function types over interfaces.**

#### Key Types

```go
// Function types for terraform operations
type InitFunc func(ctx context.Context, dir string, opts ...InitOption) error
type PlanFunc func(ctx context.Context, dir string, opts ...PlanOption) (bool, error)
type ApplyFunc func(ctx context.Context, dir string, opts ...ApplyOption) error
type DestroyFunc func(ctx context.Context, dir string, opts ...DestroyOption) error

// Executor is a collection of terraform operation functions
type Executor struct {
    Init     InitFunc
    Plan     PlanFunc
    Apply    ApplyFunc
    Destroy  DestroyFunc
    Output   OutputFunc
    Validate ValidateFunc
}
```

#### Core Functions

**Create Executor:**
```go
executor := terraform.NewExecutor("terraform")
```

**Run Operations:**
```go
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

**With Options:**
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
```

#### Example: Mock Executor for Testing

```go
func TestDeployPipeline(t *testing.T) {
    // Create mock executor
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

    // Use in pipeline
    pipeline := createDeployPipeline(mockExecutor)
    result := pipeline.Run(ctx, initialState)

    assert.True(t, E.IsRight(result))
}
```

---

### internal/agent

**Go functional writer agent - AI-assisted code generation.**

#### Key Types

```go
// CodeSpec defines what code to generate (PURE DATA)
type CodeSpec struct {
    Type         CodeType          // Function, Struct, Interface
    Name         string            // Symbol name
    Package      string            // Package name
    Inputs       []Parameter       // Parameters or fields
    Outputs      []Parameter       // Return values
    Paradigm     Paradigm          // Data, Calculation, or Action
    Validation   ValidationRules   // Guard clauses
    Assertions   []Assertion       // Internal invariants
    Doc          string            // Documentation
    Visibility   Visibility        // Exported or unexported
}

// Paradigm categorizes code
type Paradigm string
const (
    ParadigmData        Paradigm = "data"        // Inert data
    ParadigmCalculation Paradigm = "calculation" // Pure function
    ParadigmAction      Paradigm = "action"      // I/O function
)
```

#### Core Functions

**Generate Pure Function:**
```go
spec := agent.CodeSpec{
    Type:     agent.TypeFunction,
    Name:     "CalculateDiscount",
    Package:  "pricing",
    Paradigm: agent.ParadigmCalculation,
    Inputs: []agent.Parameter{
        {Name: "price", Type: "decimal.Decimal"},
        {Name: "tier", Type: "string"},
    },
    Outputs: []agent.Parameter{
        {Name: "discounted", Type: "decimal.Decimal"},
        {Name: "err", Type: "error"},
    },
    Validation: agent.ValidationRules{
        Custom: []agent.GuardClause{
            {
                Condition: "price.GreaterThan(decimal.Zero)",
                ErrorMsg:  "price must be positive",
            },
        },
    },
    Doc: "CalculateDiscount applies tier-based discounts",
}

code, err := agent.GenerateFunction(spec)
```

**Generate Data Structure:**
```go
spec := agent.CodeSpec{
    Type:    agent.TypeStruct,
    Name:    "Order",
    Package: "domain",
    Paradigm: agent.ParadigmData,
    Inputs: []agent.Parameter{
        {Name: "ID", Type: "string"},
        {Name: "Total", Type: "decimal.Decimal"},
        {Name: "Status", Type: "OrderStatus"},
    },
    Doc: "Order represents a customer order",
}

code, err := agent.GenerateStruct(spec)
```

**Generate I/O Action:**
```go
spec := agent.CodeSpec{
    Type:     agent.TypeFunction,
    Name:     "SaveOrder",
    Package:  "repository",
    Paradigm: agent.ParadigmAction,
    Inputs: []agent.Parameter{
        {Name: "ctx", Type: "context.Context"},
        {Name: "order", Type: "*Order"},
    },
    Outputs: []agent.Parameter{
        {Name: "", Type: "error"},
    },
    Validation: agent.ValidationRules{
        RequireNonNil: []string{"order"},
    },
    Doc: "SaveOrder persists an order to the database",
}

code, err := agent.GenerateFunction(spec)
```

---

## Usage Examples

### Example 1: Custom Build Tool

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/lewis/forge/internal/build"
    "github.com/lewis/forge/internal/discovery"
    E "github.com/IBM/fp-go/either"
)

func main() {
    ctx := context.Background()

    // 1. Discover functions
    functions, err := discovery.ScanFunctions(".")
    if err != nil {
        log.Fatal(err)
    }

    // 2. Convert to build configs
    buildDir := ".forge/build"
    configs := make([]build.Config, len(functions))
    for i, fn := range functions {
        configs[i] = discovery.ToBuildConfig(fn, buildDir)
    }

    // 3. Build all
    registry := build.NewRegistry()
    result := build.BuildAll(ctx, configs, registry)

    // 4. Handle result
    E.Fold(
        func(err error) {
            log.Fatalf("Build failed: %v", err)
        },
        func(artifacts []build.Artifact) {
            fmt.Printf("✓ Built %d functions\n", len(artifacts))
            for _, a := range artifacts {
                fmt.Printf("  %s (%d bytes)\n", a.Path, a.Size)
            }
        },
    )(result)
}
```

### Example 2: Custom Deployment Pipeline

```go
package main

import (
    "context"
    "fmt"

    "github.com/lewis/forge/internal/pipeline"
    "github.com/lewis/forge/internal/terraform"
    E "github.com/IBM/fp-go/either"
)

func main() {
    ctx := context.Background()

    // Create custom pipeline
    deployPipeline := pipeline.New(
        validationStage,
        buildStage,
        testStage,        // Custom: run tests before deploy
        terraformStage,
        notificationStage, // Custom: send Slack notification
    )

    // Run pipeline
    initialState := pipeline.State{ProjectDir: "."}
    result := pipeline.Run(deployPipeline, ctx, initialState)

    E.Fold(
        func(err error) {
            fmt.Printf("❌ Deployment failed: %v\n", err)
        },
        func(s pipeline.State) {
            fmt.Println("✅ Deployment successful")
        },
    )(result)
}

// Custom stage: Run tests
func testStage(ctx context.Context, s pipeline.State) E.Either[error, pipeline.State] {
    cmd := exec.CommandContext(ctx, "go", "test", "./...")
    cmd.Dir = s.ProjectDir

    if err := cmd.Run(); err != nil {
        return E.Left[pipeline.State](fmt.Errorf("tests failed: %w", err))
    }

    return E.Right[error](s)
}

// Custom stage: Send notification
func notificationStage(ctx context.Context, s pipeline.State) E.Either[error, pipeline.State] {
    // Send Slack notification
    sendSlackMessage("Deployment successful!")
    return E.Right[error](s)
}
```

### Example 3: Code Generation Tool

```go
package main

import (
    "fmt"
    "log"

    "github.com/lewis/forge/internal/agent"
)

func main() {
    // Generate domain model
    orderSpec := agent.CodeSpec{
        Type:    agent.TypeStruct,
        Name:    "Order",
        Package: "domain",
        Paradigm: agent.ParadigmData,
        Inputs: []agent.Parameter{
            {Name: "ID", Type: "string"},
            {Name: "CustomerID", Type: "string"},
            {Name: "Items", Type: "[]OrderItem"},
            {Name: "Total", Type: "decimal.Decimal"},
        },
        Doc: "Order represents a customer order",
    }

    orderCode, _ := agent.GenerateStruct(orderSpec)

    // Generate business logic
    calcSpec := agent.CodeSpec{
        Type:     agent.TypeFunction,
        Name:     "CalculateTotal",
        Package:  "domain",
        Paradigm: agent.ParadigmCalculation,
        Inputs: []agent.Parameter{
            {Name: "items", Type: "[]OrderItem"},
        },
        Outputs: []agent.Parameter{
            {Name: "total", Type: "decimal.Decimal"},
        },
        Doc: "CalculateTotal sums item prices",
    }

    calcCode, _ := agent.GenerateFunction(calcSpec)

    // Generate repository
    repoSpec := agent.CodeSpec{
        Type:     agent.TypeFunction,
        Name:     "SaveOrder",
        Package:  "repository",
        Paradigm: agent.ParadigmAction,
        Inputs: []agent.Parameter{
            {Name: "ctx", Type: "context.Context"},
            {Name: "order", Type: "*Order"},
        },
        Outputs: []agent.Parameter{
            {Name: "", Type: "error"},
        },
        Doc: "SaveOrder persists an order",
    }

    repoCode, _ := agent.GenerateFunction(repoSpec)

    // Write to files
    writeFile("domain/order.go", orderCode)
    writeFile("domain/calculate.go", calcCode)
    writeFile("repository/order.go", repoCode)

    fmt.Println("✓ Generated 3 files")
}
```

---

## Testing

### Unit Testing with Mock Executors

```go
func TestDeployment(t *testing.T) {
    mockTF := terraform.Executor{
        Init: func(ctx context.Context, dir string, opts ...terraform.InitOption) error {
            return nil
        },
        Plan: func(ctx context.Context, dir string, opts ...terraform.PlanOption) (bool, error) {
            return true, nil
        },
        Apply: func(ctx context.Context, dir string, opts ...terraform.ApplyOption) error {
            return nil
        },
    }

    result := deploy(mockTF)

    assert.NoError(t, result)
}
```

### Integration Testing

```go
func TestBuildIntegration(t *testing.T) {
    // Create temp project
    tmpDir := t.TempDir()
    createTestProject(tmpDir)

    // Discover functions
    functions, err := discovery.ScanFunctions(tmpDir)
    assert.NoError(t, err)
    assert.Len(t, functions, 1)

    // Build
    registry := build.NewRegistry()
    configs := []build.Config{
        discovery.ToBuildConfig(functions[0], tmpDir),
    }

    result := build.BuildAll(context.Background(), configs, registry)

    assert.True(t, E.IsRight(result))
}
```

---

## See Also

- [CLI_REFERENCE.md](CLI_REFERENCE.md) - Command-line usage
- [Package READMEs](../internal/) - Detailed package documentation
- [Examples](../examples/) - Sample projects and code
