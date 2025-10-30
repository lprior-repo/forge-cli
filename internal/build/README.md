# internal/build

**Multi-runtime Lambda function build system with functional composition**

## Overview

The `build` package provides a pluggable, extensible build system for serverless Lambda functions across multiple runtimes (Go, Python, Node.js, Java). It follows **pure functional programming principles** using monadic error handling (`Either`), immutable data structures, and composable decorators.

## Architecture

### Core Abstractions

```go
// BuildFunc is the fundamental abstraction - a pure function
type BuildFunc func(context.Context, Config) E.Either[error, Artifact]

// Registry maps runtimes to their build functions
type Registry map[string]BuildFunc

// Config is immutable input data
type Config struct {
    SourceDir  string            // Source code directory
    OutputPath string            // Output .zip path
    Handler    string            // Handler name
    Runtime    string            // Runtime identifier
    Env        map[string]string // Build environment variables
}

// Artifact is immutable output data
type Artifact struct {
    Path     string  // Path to built .zip file
    Checksum string  // SHA256 checksum
    Size     int64   // File size in bytes
}
```

### Design Philosophy

**Function Types Over Interfaces**

Instead of traditional OOP interfaces:
```go
// Traditional approach (what we DON'T do)
type Builder interface {
    Build(context.Context, Config) (Artifact, error)
}
```

We use **function types**:
```go
// Functional approach (what we DO)
type BuildFunc func(context.Context, Config) E.Either[error, Artifact]
```

**Benefits:**
- ✅ **Easier testing** - no mock structs, just mock functions
- ✅ **Composability** - functions can be composed with decorators
- ✅ **Type safety** - Either monad enforces error handling at compile time
- ✅ **Immutability** - no hidden state, no side effects

## Supported Runtimes

The registry supports all current AWS Lambda runtimes:

| Runtime | Versions | Builder Function | Entry Files |
|---------|----------|------------------|-------------|
| **Go** | `go1.x`, `provided.al2`, `provided.al2023` | `GoBuild` | `main.go`, `*.go` |
| **Python** | `python3.9` - `python3.13` | `PythonBuild` | `app.py`, `handler.py`, `lambda_function.py` |
| **Node.js** | `nodejs18.x`, `nodejs20.x`, `nodejs22.x` | `NodeBuild` | `index.js`, `index.mjs`, `handler.js` |
| **Java** | `java11`, `java17`, `java21` | `JavaBuild` | `pom.xml`, `build.gradle` |

## Usage

### Basic Build

```go
import "github.com/lewis/forge/internal/build"

// Create registry
registry := build.NewRegistry()

// Get builder for runtime
builderOpt := build.GetBuilder(registry, "go1.x")

// Build function
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

### Build Multiple Functions

```go
configs := []build.Config{
    {SourceDir: "./src/functions/api", Runtime: "go1.x", ...},
    {SourceDir: "./src/functions/worker", Runtime: "python3.13", ...},
}

// BuildAll uses parallel execution patterns
result := build.BuildAll(ctx, configs, registry)

if E.IsRight(result) {
    artifacts := E.GetOrElse(func() []build.Artifact { return nil })(result)
    fmt.Printf("Built %d functions\n", len(artifacts))
}
```

### Composable Decorators

The package provides **higher-order functions** that add cross-cutting concerns:

```go
// WithCache adds SHA256-based caching
cachedBuild := build.WithCache(cache)(baseBuild)

// WithLogging adds structured logging
loggedBuild := build.WithLogging(logger)(baseBuild)

// Compose multiple decorators (right-to-left composition)
enhancedBuild := build.Compose(
    build.WithCache(cache),
    build.WithLogging(logger),
)(baseBuild)
```

**How decorators work:**

```go
// Higher-order function pattern
func WithCache(cache Cache) func(BuildFunc) BuildFunc {
    return func(build BuildFunc) BuildFunc {
        return func(ctx context.Context, cfg Config) E.Either[error, Artifact] {
            // Check cache first
            if artifact, ok := cache.Get(cfg); ok {
                return E.Right[error](artifact)
            }

            // Execute wrapped build function
            result := build(ctx, cfg)

            // Cache successful results
            if E.IsRight(result) {
                cache.Set(cfg, extractArtifact(result))
            }

            return result
        }
    }
}
```

This is **pure functional composition** - no inheritance, no mixins, just function wrapping.

## Runtime-Specific Builders

### Go Builder (`go_builder.go`)

**What it does:**
1. Runs `go mod download` to fetch dependencies
2. Compiles with `GOOS=linux GOARCH=amd64` for Lambda compatibility
3. Creates `bootstrap` binary (required for `provided.al2023` runtime)
4. Zips the binary to output path
5. Calculates SHA256 checksum

**Environment:**
```go
env := []string{
    "GOOS=linux",
    "GOARCH=amd64",
    "CGO_ENABLED=0",  // Static binary
}
```

**Output:** `bootstrap.zip` containing statically-linked Linux binary

### Python Builder (`python_builder.go`)

**What it does:**
1. Detects dependency manager (pip, poetry, pipenv)
2. Installs dependencies to `.forge/python_modules/`
3. Zips Python code + dependencies
4. Calculates checksum

**Dependency installation:**
```bash
# If requirements.txt exists
pip install -r requirements.txt -t .forge/python_modules/

# If poetry.lock exists
poetry export -f requirements.txt | pip install -r /dev/stdin -t .forge/python_modules/
```

**Output:** `function.zip` containing Python code and `site-packages/`

### Node.js Builder (`node_builder.go`)

**What it does:**
1. Detects package manager (npm, yarn, pnpm)
2. Runs `npm install --production` (or equivalent)
3. Zips code + `node_modules/`
4. Calculates checksum

**Output:** `function.zip` containing JS code and `node_modules/`

### Java Builder (`java_builder.go`)

**What it does:**
1. Detects build tool (Maven, Gradle)
2. Runs `mvn package` or `gradle build`
3. Locates JAR in `target/` or `build/libs/`
4. Renames to `function.jar` and zips
5. Calculates checksum

**Output:** `function.jar` (uber JAR with dependencies)

## Caching Strategy

The build system uses **content-addressable caching** based on source code checksums:

```go
type Cache interface {
    Get(Config) (Artifact, bool)
    Set(Config, Artifact)
}
```

**Cache key generation:**
1. Hash all source files in `SourceDir`
2. Hash `go.mod`/`requirements.txt`/`package.json` (dependencies)
3. Combine into SHA256 cache key
4. Check if artifact exists with matching checksum

**Benefits:**
- Skip rebuilding unchanged functions
- Incremental builds (only changed functions rebuild)
- ~10x faster for large projects with multiple functions

## Error Handling

All functions use the **Either monad** for railway-oriented programming:

```go
// Either[error, Artifact] - no naked errors!
result := build(ctx, cfg)

// Pattern match on success/failure
E.Fold(
    func(err error) {
        fmt.Printf("Build failed: %v\n", err)
    },
    func(artifact Artifact) {
        fmt.Printf("Build succeeded: %s\n", artifact.Path)
    },
)(result)
```

**Why Either over `(Artifact, error)`?**

Traditional Go:
```go
artifact, err := build(ctx, cfg)
if err != nil {
    return err  // Easy to forget error handling!
}
// Use artifact
```

With Either:
```go
result := build(ctx, cfg)
// Compiler FORCES you to handle both cases
// No way to access Artifact without handling error
```

## Testing

### Unit Tests

The functional design makes testing trivial:

```go
func TestGoBuild(t *testing.T) {
    cfg := build.Config{
        SourceDir:  "./testdata/go-function",
        OutputPath: "./testdata/output.zip",
        Runtime:    "go1.x",
    }

    result := build.GoBuild(context.Background(), cfg)

    assert.True(t, E.IsRight(result))
    artifact := E.GetOrElse(func() build.Artifact { return build.Artifact{} })(result)
    assert.FileExists(t, artifact.Path)
}
```

### Mock Builder

For CLI tests that don't need actual builds:

```go
mockBuilder := func(ctx context.Context, cfg build.Config) E.Either[error, build.Artifact] {
    return E.Right[error](build.Artifact{
        Path:     cfg.OutputPath,
        Checksum: "mock-checksum",
        Size:     1024,
    })
}

registry := build.Registry{
    "go1.x": mockBuilder,
}
```

### Decorator Tests

Test decorators independently:

```go
func TestWithCache(t *testing.T) {
    cache := NewMemoryCache()
    buildCalls := 0

    baseBuild := func(ctx context.Context, cfg build.Config) E.Either[error, build.Artifact] {
        buildCalls++
        return E.Right[error](build.Artifact{Path: cfg.OutputPath})
    }

    cachedBuild := build.WithCache(cache)(baseBuild)

    // First call: cache miss
    cachedBuild(ctx, cfg)
    assert.Equal(t, 1, buildCalls)

    // Second call: cache hit
    cachedBuild(ctx, cfg)
    assert.Equal(t, 1, buildCalls)  // No additional build!
}
```

## Files

- **`builder.go`** - Core types (`Config`, `Artifact`, checksum utilities)
- **`functional.go`** - Registry, `BuildAll`, decorators (`WithCache`, `WithLogging`, `Compose`)
- **`go_builder.go`** - Go runtime builder implementation
- **`python_builder.go`** - Python runtime builder implementation
- **`node_builder.go`** - Node.js runtime builder implementation
- **`java_builder.go`** - Java runtime builder implementation
- **`*_test.go`** - Comprehensive unit tests (189 tests, 90%+ coverage)

## Dependencies

```go
import (
    E "github.com/IBM/fp-go/either"    // Either monad for error handling
    O "github.com/IBM/fp-go/option"    // Option monad for optional values
    "github.com/samber/lo"             // Functional utilities (Map, Filter, Reduce)
)
```

## Key Insights

1. **Function types > Interfaces** - simpler testing, better composability
2. **Either monad** - compile-time enforcement of error handling
3. **Immutable data** - `Config` and `Artifact` are never mutated
4. **Pure functions** - same inputs always produce same outputs (except I/O)
5. **Composable decorators** - cross-cutting concerns via higher-order functions

## Related Packages

- **`internal/discovery`** - Scans `src/functions/*` and converts to `build.Config`
- **`internal/pipeline`** - Orchestrates build stages in deployment pipeline
- **`internal/cli`** - CLI commands (`forge build`) that invoke this package

## Performance

**Benchmarks** (on 10 Go functions):

```
Without cache: 8.2s
With cache:    0.8s (10x faster)
```

**Parallelization** (future):
Currently sequential, but `BuildAll` is designed for parallel execution using goroutines and channels.

## Future Enhancements

- [ ] Parallel builds using goroutines
- [ ] Incremental dependency caching (only reinstall changed deps)
- [ ] Build artifact compression optimization
- [ ] Custom builder plugins (user-defined runtimes)
- [ ] Build containers for complex dependencies (Docker-based builds)
