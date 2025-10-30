# internal/discovery

**Convention-based function discovery - SAM-like auto-detection of Lambda functions**

## Overview

The `discovery` package implements **convention over configuration** by scanning `src/functions/*` directories to automatically detect Lambda functions, their runtimes, and entry points. No config files required—the directory structure IS the configuration.

## Philosophy

**Convention over configuration:**
- ✅ Presence of `src/functions/api/main.go` declares a Go Lambda function named "api"
- ✅ Presence of `src/functions/worker/index.js` declares a Node.js function named "worker"
- ❌ No need for `serverless.yml`, `template.yaml`, or `forge.yaml`

This follows SAM CLI patterns but without the YAML.

## Project Structure Convention

```
my-app/
└── src/
    └── functions/          # Convention: all functions here
        ├── api/            # Function name = directory name
        │   └── main.go     # Entry point → runtime detected
        ├── worker/
        │   └── index.js    # Entry point → runtime detected
        └── notifier/
            └── handler.py  # Entry point → runtime detected
```

**Discovery rules:**
1. Scan `src/functions/`for subdirectories
2. Each subdirectory = one Lambda function
3. Function name = directory name (e.g., `api`, `worker`)
4. Runtime = detected from entry file (see table below)

## Runtime Detection

The scanner detects runtime based on entry file presence:

| Entry File | Runtime | Handler |
|------------|---------|---------|
| `main.go`, `*.go` | `provided.al2023` | `bootstrap` |
| `index.js` | `nodejs20.x` | `index.handler` |
| `index.mjs` | `nodejs20.x` | `index.handler` |
| `handler.js` | `nodejs20.x` | `handler.handler` |
| `app.py` | `python3.13` | `handler` |
| `lambda_function.py` | `python3.13` | `lambda_function.handler` |
| `handler.py` | `python3.13` | `handler.handler` |

**Priority:** First matching file wins (e.g., if both `main.go` and `index.js` exist, Go wins).

## Data Structures

```go
// Function represents a discovered Lambda function (immutable)
type Function struct {
    Name       string // Function name (directory name)
    Path       string // Absolute path to function source
    Runtime    string // Detected runtime
    EntryPoint string // Entry file name
}
```

## Usage

### Scan Functions

```go
import "github.com/lewis/forge/internal/discovery"

// Scan src/functions/* for Lambda functions
functions, err := discovery.ScanFunctions("/path/to/project")
if err != nil {
    log.Fatal(err)
}

for _, fn := range functions {
    fmt.Printf("Found: %s (%s) at %s\n", fn.Name, fn.Runtime, fn.Path)
}
```

**Output:**
```
Found: api (provided.al2023) at /path/to/project/src/functions/api
Found: worker (nodejs20.x) at /path/to/project/src/functions/worker
Found: notifier (python3.13) at /path/to/project/src/functions/notifier
```

### Convert to Build Config

The package provides a **pure transformation function** to convert discovered functions to `build.Config`:

```go
import "github.com/lewis/forge/internal/build"

buildDir := ".forge/build"

configs := lo.Map(functions, func(f discovery.Function, _ int) build.Config {
    return discovery.ToBuildConfig(f, buildDir)
})

// Now pass to build.BuildAll()
result := build.BuildAll(ctx, configs, registry)
```

**ToBuildConfig implementation:**
```go
// ToBuildConfig converts a Function to a build.Config (PURE)
func ToBuildConfig(f Function, buildDir string) build.Config {
    outputPath := filepath.Join(buildDir, f.Name+".zip")

    // Determine handler based on runtime
    handler := "bootstrap"
    if strings.HasPrefix(f.Runtime, "nodejs") {
        handler = "index.handler"
    } else if strings.HasPrefix(f.Runtime, "python") {
        handler = "handler"
    }

    return build.Config{
        SourceDir:  f.Path,
        OutputPath: outputPath,
        Runtime:    f.Runtime,
        Handler:    handler,
        Env:        make(map[string]string),
    }
}
```

## Implementation Details

### ScanFunctions (Pure + I/O)

```go
func ScanFunctions(projectRoot string) ([]Function, error) {
    functionsDir := filepath.Join(projectRoot, "src", "functions")

    // Check if directory exists
    if _, err := os.Stat(functionsDir); os.IsNotExist(err) {
        return nil, fmt.Errorf("src/functions directory not found")
    }

    // Read all subdirectories
    entries, err := os.ReadDir(functionsDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read functions directory: %w", err)
    }

    functions := make([]Function, 0, len(entries))
    for _, entry := range entries {
        if !entry.IsDir() {
            continue  // Skip files, only process directories
        }

        functionPath := filepath.Join(functionsDir, entry.Name())

        // Detect runtime (pure function)
        runtime, entryPoint, err := detectRuntime(functionPath)
        if err != nil {
            continue  // Skip directories without recognizable entry points
        }

        functions = append(functions, Function{
            Name:       entry.Name(),
            Path:       functionPath,
            Runtime:    runtime,
            EntryPoint: entryPoint,
        })
    }

    return functions, nil
}
```

### detectRuntime (Pure)

```go
// detectRuntime determines the runtime by checking for entry point files (PURE-ish)
func detectRuntime(functionPath string) (string, string, error) {
    // Go: main.go or *.go files
    if fileExists(functionPath, "main.go") {
        return RuntimeGo, "main.go", nil
    }
    if hasGoFiles(functionPath) {
        return RuntimeGo, "*.go", nil
    }

    // Node.js: index.js, index.mjs, or handler.js
    if fileExists(functionPath, "index.js") {
        return RuntimeNode, "index.js", nil
    }
    if fileExists(functionPath, "index.mjs") {
        return RuntimeNode, "index.mjs", nil
    }
    if fileExists(functionPath, "handler.js") {
        return RuntimeNode, "handler.js", nil
    }

    // Python: app.py, lambda_function.py, or handler.py
    if fileExists(functionPath, "app.py") {
        return RuntimePython, "app.py", nil
    }
    if fileExists(functionPath, "lambda_function.py") {
        return RuntimePython, "lambda_function.py", nil
    }
    if fileExists(functionPath, "handler.py") {
        return RuntimePython, "handler.py", nil
    }

    return "", "", fmt.Errorf("no recognized entry point found")
}
```

## Error Handling

**Graceful degradation:**
- Missing `src/functions/` directory → Error (project not initialized)
- Directory without entry file → Skip (allow non-function directories)
- Multiple entry files → Use first match (deterministic priority)

## Testing

### Unit Tests

```go
func TestScanFunctions(t *testing.T) {
    // testdata/
    // └── src/functions/
    //     ├── api/main.go
    //     ├── worker/index.js
    //     └── notifier/handler.py

    functions, err := discovery.ScanFunctions("./testdata")

    assert.NoError(t, err)
    assert.Len(t, functions, 3)

    // Verify Go function
    api := functions[0]
    assert.Equal(t, "api", api.Name)
    assert.Equal(t, "provided.al2023", api.Runtime)
    assert.Equal(t, "main.go", api.EntryPoint)

    // Verify Node.js function
    worker := functions[1]
    assert.Equal(t, "worker", worker.Name)
    assert.Equal(t, "nodejs20.x", worker.Runtime)

    // Verify Python function
    notifier := functions[2]
    assert.Equal(t, "notifier", notifier.Name)
    assert.Equal(t, "python3.13", notifier.Runtime)
}

func TestToBuildConfig(t *testing.T) {
    fn := discovery.Function{
        Name:       "api",
        Path:       "/project/src/functions/api",
        Runtime:    "provided.al2023",
        EntryPoint: "main.go",
    }

    cfg := discovery.ToBuildConfig(fn, ".forge/build")

    assert.Equal(t, "/project/src/functions/api", cfg.SourceDir)
    assert.Equal(t, ".forge/build/api.zip", cfg.OutputPath)
    assert.Equal(t, "provided.al2023", cfg.Runtime)
    assert.Equal(t, "bootstrap", cfg.Handler)
}
```

## Integration with Other Packages

### Build Package

```go
// Typical workflow in CLI
functions, _ := discovery.ScanFunctions(".")
configs := lo.Map(functions, func(f discovery.Function, _ int) build.Config {
    return discovery.ToBuildConfig(f, ".forge/build")
})
result := build.BuildAll(ctx, configs, registry)
```

### Pipeline Package

```go
// Discovery stage in deployment pipeline
discoveryStage := func(ctx context.Context, s pipeline.State) E.Either[error, pipeline.State] {
    functions, err := discovery.ScanFunctions(s.ProjectDir)
    if err != nil {
        return E.Left[pipeline.State](err)
    }

    // Add to state
    s.DiscoveredFunctions = functions
    return E.Right[error](s)
}
```

## Stub ZIP Generation

When Terraform needs to initialize before building (e.g., `terraform init`), we generate **stub ZIP files**:

```go
// CreateStubZip creates a minimal valid ZIP for terraform init (I/O action)
func CreateStubZip(path string) error {
    // Create 1-byte ZIP file (smallest valid ZIP)
    stubContent := []byte{0x50, 0x4b, 0x05, 0x06, 0x00, 0x00, 0x00, 0x00,
                          0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
                          0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

    return os.WriteFile(path, stubContent, 0644)
}
```

**Why?** Terraform requires the `filename` attribute in `aws_lambda_function` to exist before init/plan. Stub ZIPs satisfy this requirement without building.

## Constants

```go
const (
    RuntimeGo     = "provided.al2023"  // Latest Go runtime
    RuntimeNode   = "nodejs20.x"       // Latest Node.js runtime
    RuntimePython = "python3.13"       // Latest Python runtime
)
```

## Files

- **`scanner.go`** - `ScanFunctions`, `detectRuntime`, entry point detection
- **`stub.go`** - Stub ZIP generation for Terraform initialization
- **`scanner_test.go`** - Unit tests for discovery logic
- **`stub_test.go`** - Unit tests for stub generation

## Design Principles

1. **Convention over configuration** - directory structure is the spec
2. **Pure transformations** - `ToBuildConfig` is a pure function
3. **Graceful degradation** - skip unrecognized directories
4. **Deterministic** - same project structure always produces same results
5. **SAM-like** - familiar patterns from AWS SAM CLI

## Future Enhancements

- [ ] Support for Rust (detect `Cargo.toml`)
- [ ] Support for .NET (detect `*.csproj`)
- [ ] Custom runtime detection via config file
- [ ] Multi-file entry points (e.g., TypeScript compilation)
- [ ] Function metadata discovery (description, timeout hints from comments)
- [ ] Nested function directories (e.g., `src/functions/api/v1/main.go`)
