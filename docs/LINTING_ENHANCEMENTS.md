# Linting Enhancements for Functional Programming & Architectural Quality

## Overview

This document explains the comprehensive linting improvements made to enforce functional programming principles, immutability, and architectural simplicity in the Forge codebase.

## Philosophy

The enhanced linting configuration enforces:

1. **Functional Programming**: Pure functions, immutability, no side effects in core logic
2. **Architectural Simplicity**: Small interfaces (ISP), DRY principle, minimal complexity
3. **Martin Fowler Standards**: Self-documenting code, zero duplication, single responsibility
4. **Railway-Oriented Programming**: Explicit error handling via Either monad

## New Linters Added (52 → 91 linters)

### 🎯 CRITICAL: Immutability Enforcement

#### `reassign` (HIGHEST PRIORITY)
**Purpose**: Detect variable reassignment to enforce immutability

**Why Critical**: The core principle of FP is immutability. This linter catches:
```go
// ❌ BAD: Mutation
result := initial
for _, item := range items {
    result = transform(result, item)  // CAUGHT: reassignment!
}

// ✅ GOOD: Use functional reduce
result := A.Reduce(
    func(acc T, item I) T {
        return transform(acc, item)  // No mutation, returns new value
    },
    initial,
)(items)
```

**Exclusions**: Allowed in specific FP patterns (Either chains, reduce operations) where reassignment is the canonical pattern.

---

#### `revive: modifies-parameter` (HIGHEST PRIORITY)
**Purpose**: Detect functions that mutate their parameters

**Why Critical**: Parameter mutation breaks referential transparency and makes code unpredictable:
```go
// ❌ BAD: Parameter mutation
func AddTag(artifact Artifact, tag string) {
    artifact.Tags = append(artifact.Tags, tag)  // CAUGHT: modifies parameter!
}

// ✅ GOOD: Return new value
func AddTag(artifact Artifact, tag string) Artifact {
    newArtifact := artifact
    newArtifact.Tags = append(artifact.Tags, tag)  // Copy, don't mutate
    return newArtifact
}
```

---

#### `revive: modifies-value-receiver` (HIGHEST PRIORITY)
**Purpose**: Detect methods that mutate value receivers

**Why Critical**: Value receiver mutation is a common source of bugs in Go:
```go
// ❌ BAD: Mutating value receiver
func (c Config) SetRuntime(runtime string) {
    c.Runtime = runtime  // CAUGHT: doesn't work with value receiver!
}

// ✅ GOOD: Pointer receiver (if mutation needed) or return new value
func (c Config) WithRuntime(runtime string) Config {
    newConfig := c
    newConfig.Runtime = runtime
    return newConfig
}
```

---

### 🔥 Code Quality & Simplification

#### `gocritic` (Comprehensive Analysis)
**Purpose**: 200+ checks for correctness, performance, style

**Key Checks**:
- `appendAssign`: `x = append(x, y)` can fail with shared slices
- `assignOp`: Use `x += 1` instead of `x = x + 1`
- `captureLoopVar`: Loop variable captured incorrectly
- `deferUnlambda`: `defer func() { f() }()` → `defer f()`
- `ifElseChain`: Long if-else chains → switch or map
- `nestingReduce`: Reduce nesting depth
- `rangeExprCopy`: Range over large structs by reference
- `sloppyLen`: `len(x) >= 0` is always true
- `unnecessaryBlock`: Remove unnecessary braces

**Tags Enabled**:
- `diagnostic`: Code correctness issues
- `experimental`: Cutting-edge checks
- `opinionated`: Best practice enforcement
- `performance`: Performance optimizations
- `style`: Code style consistency

**Example**:
```go
// ❌ BAD: Unnecessary block
if err != nil {
    {  // CAUGHT: Unnecessary braces
        return err
    }
}

// ✅ GOOD
if err != nil {
    return err
}
```

---

#### `dupl` (DRY Principle)
**Purpose**: Detect code duplication (threshold: 100 tokens)

**Why Important**: Duplicated code violates DRY and increases maintenance burden:
```go
// ❌ BAD: Duplicated validation logic
func ValidateGoBuild(cfg Config) error {
    if cfg.SourceDir == "" { return errors.New("source dir required") }
    if cfg.OutputDir == "" { return errors.New("output dir required") }
    // ...
}

func ValidatePythonBuild(cfg Config) error {
    if cfg.SourceDir == "" { return errors.New("source dir required") }
    if cfg.OutputDir == "" { return errors.New("output dir required") }
    // ... CAUGHT: 80% duplication!
}

// ✅ GOOD: Extract common validation
func ValidateBasicBuild(cfg Config) error { /* shared logic */ }
```

---

#### `mnd` (Magic Number Detector)
**Purpose**: Detect magic numbers that should be named constants

**Why Important**: Magic numbers reduce code readability and maintainability:
```go
// ❌ BAD: Magic numbers
timeout := time.Duration(300) * time.Second  // CAUGHT: What is 300?
if size > 1048576 { /* ... */ }             // CAUGHT: What is 1048576?

// ✅ GOOD: Named constants
const (
    DefaultTimeoutSeconds = 300
    MaxFileSizeBytes      = 1 * 1024 * 1024  // 1 MB
)
timeout := time.Duration(DefaultTimeoutSeconds) * time.Second
if size > MaxFileSizeBytes { /* ... */ }
```

**Exclusions**: Common numbers (0, 1, 2, 100), file permissions, time durations.

---

#### `stylecheck` (Naming & Documentation)
**Purpose**: Enforce Go naming conventions and godoc standards

**Key Checks**:
- `ST1001`: Dot imports (allowed for FP: `E.Either`, `O.Option`, `A.Array`)
- `ST1003`: Underscores in package names
- `ST1005`: Error strings should not be capitalized
- `ST1006`: Receiver names (should be consistent and short)
- `ST1016`: Methods on the same type should have same receiver name
- `ST1017`: Don't use `yoda` conditions (`nil == x` → `x == nil`)

**Initialisms**: Comprehensive list (AWS, API, HTTP, JSON, S3, etc.) for correct capitalization.

**Example**:
```go
// ❌ BAD: Incorrect naming
type HTTPSConnection struct { /* ... */ }  // CAUGHT: should be HTTPSConnection
func (h *HTTPSConnection) getURL() { /* ... */ }  // CAUGHT: should be GetURL

// ✅ GOOD
type HTTPSConnection struct { /* ... */ }
func (c *HTTPSConnection) GetURL() { /* ... */ }
```

---

### 🛡️ Safety & Correctness

#### `interfacebloat` (Interface Segregation)
**Purpose**: Limit interface size to 6 methods (ISP principle)

**Why Important**: Large interfaces violate the Interface Segregation Principle:
```go
// ❌ BAD: God interface (10+ methods)
type Builder interface {
    Build(Config) Artifact
    Validate(Config) error
    Clean(Config) error
    Test(Config) error
    Deploy(Config) error
    Rollback(Config) error
    Monitor(Config) Status
    // ... CAUGHT: Too many responsibilities!
}

// ✅ GOOD: Segregated interfaces
type Builder interface {
    Build(Config) Artifact
}

type Validator interface {
    Validate(Config) error
}

type Cleaner interface {
    Clean(Config) error
}
```

---

#### `forcetypeassert` (Type Assertion Safety)
**Purpose**: Require checked type assertions

**Why Important**: Unchecked type assertions cause panics:
```go
// ❌ BAD: Unchecked type assertion
value := data.(string)  // CAUGHT: Panics if not string!

// ✅ GOOD: Checked type assertion
value, ok := data.(string)
if !ok {
    return errors.New("expected string")
}

// ✅ BEST: Use Option monad
func SafeAsString(data interface{}) O.Option[string] {
    if s, ok := data.(string); ok {
        return O.Some(s)
    }
    return O.None[string]()
}
```

---

#### `contextcheck` (Context Propagation)
**Purpose**: Ensure context.Context is passed correctly through function chains

**Why Important**: Context cancellation and deadlines must propagate:
```go
// ❌ BAD: Context not propagated
func ProcessPipeline(ctx context.Context, stages []Stage) {
    for _, stage := range stages {
        stage(context.Background(), state)  // CAUGHT: Using background ctx!
    }
}

// ✅ GOOD: Propagate context
func ProcessPipeline(ctx context.Context, stages []Stage) {
    for _, stage := range stages {
        stage(ctx, state)  // Parent context propagated
    }
}
```

---

#### `containedctx` (Context in Structs)
**Purpose**: Detect structs containing context.Context fields

**Why Important**: Contexts should be passed as function parameters, not stored:
```go
// ❌ BAD: Context in struct
type Pipeline struct {
    ctx    context.Context  // CAUGHT: Don't store context!
    stages []Stage
}

// ✅ GOOD: Context as parameter
type Pipeline struct {
    stages []Stage
}

func (p Pipeline) Run(ctx context.Context, state State) Either[error, State] {
    // ctx passed as parameter
}
```

---

#### `exportloopref` (Loop Variable Capture)
**Purpose**: Detect loop variables incorrectly captured in closures

**Why Important**: Classic Go gotcha that causes bugs:
```go
// ❌ BAD: Loop variable captured
for _, item := range items {
    go func() {
        process(item)  // CAUGHT: item reference changes!
    }()
}

// ✅ GOOD: Copy loop variable
for _, item := range items {
    item := item  // Shadow to create new variable
    go func() {
        process(item)  // Safe: uses copied value
    }()
}
```

---

### 📦 Import & Dependency Management

#### `importas` (FP Alias Enforcement)
**Purpose**: Enforce consistent import aliases for fp-go packages

**Required Aliases**:
```go
import (
    E "github.com/IBM/fp-go/either"   // ✅ Required
    O "github.com/IBM/fp-go/option"   // ✅ Required
    A "github.com/IBM/fp-go/array"    // ✅ Required
    C "github.com/IBM/fp-go/context"  // ✅ Required
    T "github.com/IBM/fp-go/tuple"    // ✅ Required
    F "github.com/IBM/fp-go/function" // ✅ Required
)
```

**Why Important**: Consistent aliases improve readability across the codebase:
```go
// ❌ BAD: Inconsistent aliases
import either "github.com/IBM/fp-go/either"  // CAUGHT!
import opt "github.com/IBM/fp-go/option"     // CAUGHT!

// ✅ GOOD: Consistent aliases
import E "github.com/IBM/fp-go/either"
import O "github.com/IBM/fp-go/option"
```

---

#### `gci` (Import Ordering)
**Purpose**: Enforce consistent import grouping and ordering

**Order**:
1. Standard library imports
2. External packages
3. Internal packages (`github.com/lewis/forge`)

**Example**:
```go
// ✅ GOOD: Proper import order
import (
    // Standard library
    "context"
    "fmt"

    // External packages
    E "github.com/IBM/fp-go/either"
    "github.com/spf13/cobra"

    // Internal packages
    "github.com/lewis/forge/internal/build"
    "github.com/lewis/forge/internal/pipeline"
)
```

---

#### `gomodguard` (Dependency Control)
**Purpose**: Block deprecated or unwanted dependencies

**Blocked Dependencies**:
```yaml
github.com/pkg/errors:
  reason: "Use stdlib errors and Either monad instead"
  recommendations:
    - errors
    - fmt

github.com/sirupsen/logrus:
  reason: "Use stdlib slog for structured logging"
  recommendations:
    - log/slog
```

**Why Important**: Prevents technical debt from accumulating via outdated dependencies.

---

### 📝 Documentation & Formatting

#### `godot` (Documentation Quality)
**Purpose**: Ensure all comments end with proper punctuation

**Why Important**: Professional documentation standards:
```go
// ❌ BAD: No punctuation
// BuildFunc is the core abstraction

// ✅ GOOD: Complete sentence
// BuildFunc is the core abstraction for building artifacts.
```

---

#### `gofumpt` (Stricter Formatting)
**Purpose**: Stricter version of gofmt with extra rules

**Key Rules**:
- Remove empty lines at start/end of blocks
- Format struct field alignment
- Group imports by category
- Simplify composite literals

**Why Important**: Consistent formatting reduces review noise and improves readability.

---

#### `grouper` (Declaration Grouping)
**Purpose**: Enforce grouping of similar declarations

**Rules**:
- Group constants together
- Group variables together
- Group type definitions together

**Example**:
```go
// ❌ BAD: Scattered declarations
const MaxRetries = 3
var defaultTimeout = 30 * time.Second
const DefaultRegion = "us-east-1"
var cache = NewCache()

// ✅ GOOD: Grouped declarations
const (
    MaxRetries    = 3
    DefaultRegion = "us-east-1"
)

var (
    defaultTimeout = 30 * time.Second
    cache          = NewCache()
)
```

---

### 🧪 Testing Quality

#### `thelper` (Test Helper Marking)
**Purpose**: Ensure test helpers call `t.Helper()`

**Why Important**: Correct line numbers in test failures:
```go
// ❌ BAD: Missing t.Helper()
func assertNoError(t *testing.T, err error) {
    if err != nil {
        t.Fatal(err)  // Line number points here, not caller
    }
}

// ✅ GOOD: Marked as helper
func assertNoError(t *testing.T, err error) {
    t.Helper()  // Line number points to caller
    if err != nil {
        t.Fatal(err)
    }
}
```

---

#### `tenv` (Environment Variables in Tests)
**Purpose**: Use `t.Setenv()` instead of `os.Setenv()` in tests

**Why Important**: Automatic cleanup and test isolation:
```go
// ❌ BAD: Manual cleanup required
func TestWithEnv(t *testing.T) {
    old := os.Getenv("AWS_REGION")
    os.Setenv("AWS_REGION", "us-east-1")  // CAUGHT!
    defer os.Setenv("AWS_REGION", old)    // Easy to forget
}

// ✅ GOOD: Automatic cleanup
func TestWithEnv(t *testing.T) {
    t.Setenv("AWS_REGION", "us-east-1")  // Auto-restored after test
}
```

---

#### `testpackage` (Blackbox Testing)
**Purpose**: Encourage blackbox tests using `_test` package suffix

**Why Important**: Tests should verify public API, not internal implementation:
```go
// ❌ DISCOURAGED: Whitebox test
package build

func TestInternalFunction(t *testing.T) {
    // Can access unexported functions - fragile!
}

// ✅ ENCOURAGED: Blackbox test
package build_test

import "github.com/lewis/forge/internal/build"

func TestPublicAPI(t *testing.T) {
    // Can only access exported API - robust!
}
```

**Exclusions**: Allowed for `internal_test.go` and `export_test.go` files.

---

### 🎨 Style & Consistency

#### `decorder` (Declaration Order)
**Purpose**: Enforce consistent declaration order in files

**Order**:
1. `const` declarations
2. `var` declarations
3. `type` definitions
4. `func` definitions

**Why Important**: Consistent structure makes code easier to navigate.

---

#### `tagalign` (Struct Tag Alignment)
**Purpose**: Align and sort struct tags for readability

**Order**: `json` → `yaml` → `yml` → `toml` → `mapstructure` → `binding` → `validate`

**Example**:
```go
// ❌ BAD: Unsorted, unaligned
type Config struct {
    Name    string `yaml:"name" json:"name"`
    Region  string `json:"region" yaml:"region"`
}

// ✅ GOOD: Sorted and aligned
type Config struct {
    Name   string `json:"name"   yaml:"name"`
    Region string `json:"region" yaml:"region"`
}
```

---

#### `whitespace` (Blank Line Consistency)
**Purpose**: Enforce blank lines after multi-line constructs

**Rules**:
- Blank line after multi-line `if` statement
- Blank line after multi-line function signature

**Example**:
```go
// ❌ BAD: No blank line
if condition1 &&
    condition2 &&
    condition3 {
    // ...
}
nextStatement()  // CAUGHT: No blank line after multi-line if

// ✅ GOOD: Blank line added
if condition1 &&
    condition2 &&
    condition3 {
    // ...
}

nextStatement()
```

---

### 🔒 Security

#### `nolintlint` (Linter Disable Justification)
**Purpose**: Require explanation for all `//nolint` directives

**Why Important**: Prevents blindly disabling linters:
```go
// ❌ BAD: No justification
func unsafeOperation() {
    //nolint:gosec  // CAUGHT: No explanation!
    exec.Command(userInput).Run()
}

// ✅ GOOD: Justified disable
func unsafeOperation() {
    //nolint:gosec // User input is validated by validateCommand() first
    exec.Command(userInput).Run()
}
```

**Rules**:
- `allow-unused: false` - No unused nolint directives
- `allow-no-explanation: []` - ALL directives require explanation
- `require-explanation: true` - Explanation is mandatory
- `require-specific: true` - Must specify linter name

---

## Exclusions & Pragmatism

### Test Files (`_test.go`)
Tests have relaxed rules for pragmatism:
- `errcheck`: Skip error checks in setup code
- `gosec`: Less strict security (test data is controlled)
- `unparam`: Test helpers can have unused params
- `goconst`: Test data can have repeated strings
- `dupl`: Tests can have similar patterns
- `mnd`: Magic numbers acceptable in test data
- `forcetypeassert`: Type assertions ok (controlled environment)

### Functional Programming Patterns
Special allowances for FP idioms:
- `reassign`: Allowed in Either chains and reduce operations (canonical pattern)
- `dot-imports`: Allowed for FP monads (`E.Either`, `O.Option`, `A.Array`)
- `shadow`: Variable shadowing intentional in FP (loop variable copies)

### I/O Boundary (CLI Commands)
CLI code has pragmatic allowances:
- `deep-exit`: `os.Exit` allowed in CLI entry points
- `modifies-parameter`: Cobra command setup mutates cmd objects

### Generated Code
All linters disabled for generated files:
- `*_gen.go`: Code generation output
- `*.pb.go`: Protobuf generated code
- `internal/lingon/aws/`: 2,671 AWS resource packages (1M+ LOC)

---

## Impact Summary

### Before: 35 linters
- Basic error checking and formatting
- Some security and style rules
- Manual enforcement of FP principles

### After: 91 linters (+56 new linters)
- **Immutability enforcement** via `reassign`, `modifies-parameter`, `modifies-value-receiver`
- **Interface quality** via `interfacebloat` (ISP principle)
- **Code duplication detection** via `dupl` (DRY principle)
- **Magic number detection** via `mnd`
- **FP alias consistency** via `importas`
- **Documentation quality** via `godot`, `stylecheck`
- **Safety guarantees** via `forcetypeassert`, `contextcheck`, `exportloopref`
- **Test quality** via `thelper`, `tenv`, `testpackage`
- **Dependency control** via `gomodguard`
- **Comprehensive analysis** via `gocritic` (200+ checks)

---

## Integration with CI/CD

The enhanced linting is enforced via:

```bash
# Run linting (must pass with zero issues)
task lint

# Full CI pipeline
task ci        # Unit tests + lint
task ci:full   # Integration tests + lint

# Coverage enforcement (90% minimum)
task coverage:check
```

**CI/CD Requirements**:
1. ✅ Zero linting errors or warnings
2. ✅ Zero test failures
3. ✅ 90% minimum test coverage
4. ✅ 80% minimum mutation score (critical packages)

---

## Adopting in Your Team

### Phase 1: Gradual Rollout
Enable new linters incrementally:
```yaml
# Start with low-noise linters
- dupl
- godot
- gofumpt
- gci
- importas
```

### Phase 2: Safety & Correctness
Add safety-focused linters:
```yaml
- forcetypeassert
- contextcheck
- exportloopref
- containedctx
```

### Phase 3: Immutability (Most Disruptive)
Enable immutability enforcement:
```yaml
- reassign
- revive: modifies-parameter
- revive: modifies-value-receiver
```

### Phase 4: Comprehensive Analysis
Enable comprehensive checks:
```yaml
- gocritic
- interfacebloat
- mnd
```

---

## Benefits

### Code Quality
- **Zero duplication**: DRY principle enforced automatically
- **Consistent naming**: Go conventions enforced everywhere
- **No magic numbers**: All constants are named and documented

### Safety
- **No unchecked type assertions**: Panics prevented
- **Context propagation**: Cancellation works correctly
- **Loop variable capture**: Classic Go gotchas caught

### Functional Programming
- **Immutability enforced**: Variable reassignment caught
- **No parameter mutation**: Referential transparency maintained
- **Small interfaces**: ISP principle enforced (max 6 methods)

### Developer Experience
- **Fewer bugs**: Catch issues before code review
- **Faster reviews**: Automated checks reduce human review burden
- **Better documentation**: godoc standards enforced

### Maintainability
- **Consistent imports**: FP aliases standardized
- **Proper grouping**: Declarations organized logically
- **No technical debt**: Deprecated dependencies blocked

---

## References

- [golangci-lint Documentation](https://golangci-lint.run/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [Martin Fowler: Refactoring](https://refactoring.com/)
- [fp-go Documentation](https://github.com/IBM/fp-go)

---

## Conclusion

This ultra-strict linting configuration transforms Go into a **functional programming environment** with **architectural guardrails**. It enforces:

1. ✅ **Immutability** (no reassignment, no mutation)
2. ✅ **Small interfaces** (ISP principle)
3. ✅ **Zero duplication** (DRY principle)
4. ✅ **Safe operations** (checked type assertions, context propagation)
5. ✅ **Professional documentation** (godoc standards)
6. ✅ **Consistent style** (FP aliases, import ordering, declaration grouping)

The result is a codebase that **prevents bugs at compile time**, **enforces best practices automatically**, and **maintains Martin Fowler-level quality standards**.
