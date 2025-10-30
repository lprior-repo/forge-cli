# internal/pipeline

**Functional pipeline orchestration for multi-stage deployment workflows**

## Overview

The `pipeline` package provides a **functional, composable pipeline system** for orchestrating complex deployment workflows. It uses **railway-oriented programming** with Either monads to handle errors elegantly and supports both sequential and parallel stage execution.

## Core Concepts

### State

**Immutable data** that flows through the pipeline:

```go
type State struct {
    ProjectDir string                   // Project root directory
    Artifacts  map[string]Artifact      // Built function artifacts
    Outputs    map[string]interface{}   // Terraform outputs
    Config     interface{}              // Configuration data
}
```

### Stage

**Pure function** that transforms state:

```go
// Stage is a function that transforms state
// Uses Either monad for error handling (railway-oriented programming)
type Stage func(context.Context, State) E.Either[error, State]
```

### Pipeline

**Immutable collection** of stages:

```go
type Pipeline struct {
    stages []Stage  // Ordered list of stages to execute
}
```

## Usage

### Basic Pipeline

```go
// Define stages
buildStage := func(ctx context.Context, s State) E.Either[error, State] {
    // Build functions
    artifacts, err := build.BuildAll(ctx, ...)
    if err != nil {
        return E.Left[State](err)
    }

    s.Artifacts = artifacts
    return E.Right[error](s)
}

initStage := func(ctx context.Context, s State) E.Either[error, State] {
    // Terraform init
    err := terraform.Init(ctx, s.ProjectDir)
    if err != nil {
        return E.Left[State](err)
    }
    return E.Right[error](s)
}

// Create pipeline
deployPipeline := pipeline.New(buildStage, initStage, planStage, applyStage)

// Run pipeline
initialState := pipeline.State{ProjectDir: "."}
result := pipeline.Run(deployPipeline, ctx, initialState)

// Handle result
E.Fold(
    func(err error) { fmt.Printf("Failed: %v\n", err) },
    func(s State) { fmt.Println("Success!") },
)(result)
```

### Railway-Oriented Programming

Pipelines **short-circuit on first error** (like a railway switch):

```
buildStage → initStage → planStage → applyStage → Success
              ↓ error
              Failure (skip remaining stages)
```

```go
func Run(p Pipeline, ctx context.Context, initial State) E.Either[error, State] {
    result := E.Right[error](initial)

    for _, stage := range p.stages {
        if E.IsLeft(result) {
            return result  // Short-circuit on error
        }

        state := extractState(result)
        result = stage(ctx, state)  // Run next stage
    }

    return result
}
```

### Composing Pipelines

```go
// Create sub-pipelines
buildPipeline := pipeline.New(discoveryStage, buildStage)
terraformPipeline := pipeline.New(initStage, planStage, applyStage)

// Chain into single pipeline
deployPipeline := pipeline.Chain(buildPipeline, terraformPipeline)

// Run combined pipeline
result := pipeline.Run(deployPipeline, ctx, initialState)
```

### Parallel Execution (Future)

```go
// Run stages in parallel (NOT YET IMPLEMENTED)
parallelStage := pipeline.Parallel(
    buildGoFunctionsStage,
    buildPythonFunctionsStage,
    buildNodeFunctionsStage,
)
```

## Pre-built Stages

### Convention Stages (`convention_stages.go`)

**Discovery Stage:**
```go
func NewDiscoveryStage() Stage {
    return func(ctx context.Context, s State) E.Either[error, State] {
        functions, err := discovery.ScanFunctions(s.ProjectDir)
        if err != nil {
            return E.Left[State](err)
        }
        s.DiscoveredFunctions = functions
        return E.Right[error](s)
    }
}
```

**Build Stage:**
```go
func NewBuildStage(registry build.Registry) Stage {
    return func(ctx context.Context, s State) E.Either[error, State] {
        configs := convertToConfigs(s.DiscoveredFunctions)
        result := build.BuildAll(ctx, configs, registry)

        return E.Fold(
            func(err error) E.Either[error, State] {
                return E.Left[State](err)
            },
            func(artifacts []build.Artifact) E.Either[error, State] {
                s.Artifacts = artifactsToMap(artifacts)
                return E.Right[error](s)
            },
        )(result)
    }
}
```

### Terraform Stages (`terraform_stages.go`)

**Init Stage:**
```go
func NewInitStage(executor terraform.Executor) Stage {
    return func(ctx context.Context, s State) E.Either[error, State] {
        err := executor.Init(ctx, filepath.Join(s.ProjectDir, "infra"))
        if err != nil {
            return E.Left[State](fmt.Errorf("terraform init failed: %w", err))
        }
        return E.Right[error](s)
    }
}
```

**Plan Stage:**
```go
func NewPlanStage(executor terraform.Executor) Stage {
    return func(ctx context.Context, s State) E.Either[error, State] {
        hasChanges, err := executor.Plan(ctx, filepath.Join(s.ProjectDir, "infra"))
        if err != nil {
            return E.Left[State](fmt.Errorf("terraform plan failed: %w", err))
        }

        s.HasChanges = hasChanges
        return E.Right[error](s)
    }
}
```

**Apply Stage:**
```go
func NewApplyStage(executor terraform.Executor) Stage {
    return func(ctx context.Context, s State) E.Either[error, State] {
        if !s.HasChanges {
            return E.Right[error](s)  // Skip if no changes
        }

        err := executor.Apply(ctx, filepath.Join(s.ProjectDir, "infra"))
        if err != nil {
            return E.Left[State](fmt.Errorf("terraform apply failed: %w", err))
        }

        // Fetch outputs
        outputs, _ := executor.Output(ctx, filepath.Join(s.ProjectDir, "infra"))
        s.Outputs = outputs

        return E.Right[error](s)
    }
}
```

## Complete Deployment Pipeline

```go
func NewDeploymentPipeline(tfExecutor terraform.Executor) Pipeline {
    buildRegistry := build.NewRegistry()

    return pipeline.Chain(
        // Build phase
        pipeline.New(
            NewDiscoveryStage(),
            NewBuildStage(buildRegistry),
        ),
        // Terraform phase
        pipeline.New(
            NewInitStage(tfExecutor),
            NewPlanStage(tfExecutor),
            NewApplyStage(tfExecutor),
        ),
    )
}
```

## Design Principles

1. **Immutable state** - State is never modified in-place, always returns new State
2. **Pure stages** - Same input always produces same output (except I/O)
3. **Railway-oriented** - Errors short-circuit automatically
4. **Composable** - Pipelines can be combined via `Chain()`
5. **Testable** - Each stage can be tested independently

## Testing

```go
func TestBuildStage(t *testing.T) {
    mockRegistry := build.NewMockRegistry()
    stage := NewBuildStage(mockRegistry)

    initialState := pipeline.State{
        ProjectDir: "./testdata",
        DiscoveredFunctions: []discovery.Function{
            {Name: "api", Runtime: "go1.x"},
        },
    }

    result := stage(context.Background(), initialState)

    assert.True(t, E.IsRight(result))
    finalState := extractState(result)
    assert.NotEmpty(t, finalState.Artifacts)
}
```

## Files

- **`pipeline.go`** - Core `Pipeline`, `Stage`, `Run`, `Chain`
- **`convention_stages.go`** - Discovery and build stages
- **`terraform_stages.go`** - Terraform init/plan/apply stages
- **`stages.go`** - Generic stage utilities
- **`*_test.go`** - Unit tests for each stage

## Dependencies

```go
import (
    E "github.com/IBM/fp-go/either"           // Either monad
    "github.com/lewis/forge/internal/build"
    "github.com/lewis/forge/internal/discovery"
    "github.com/lewis/forge/internal/terraform"
)
```

## Future Enhancements

- [ ] Parallel stage execution using goroutines
- [ ] Conditional stages (`If()` combinator)
- [ ] Retry stages with exponential backoff
- [ ] Pipeline visualization (DAG generation)
- [ ] Stage metrics (timing, success rate)
- [ ] Dry-run mode (skip I/O stages)
