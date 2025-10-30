# Functional Programming Audit - Cleanup Roadmap

**Date:** 2025-10-29
**Overall Score:** 6.5/10
**Target Score:** 9/10
**Estimated Effort:** 12-18 developer days

---

## Executive Summary

The Forge codebase demonstrates a **genuine commitment to functional programming principles** with good use of the fp-go library (Either/Option monads), but contains **critical violations** that compromise its purity claims. The code is in a **transitional state** - aspiring toward functional purity but not fully achieving it.

---

## Critical Violations (MUST FIX)

### P0: State Mutation in Pipeline Stages
**Files:**
- `internal/pipeline/convention_stages.go` (lines 36, 77-78, 121, 215-218)
- `internal/pipeline/terraform_stages.go` (lines 73-74, 81)

**Issue:** Pipeline stages directly mutate the `State` struct instead of creating new immutable instances.

**Current Code:**
```go
// VIOLATION - Direct mutation of state
s.Config = functions           // Line 36
s.Artifacts = make(map[string]Artifact)  // Line 78
s.Outputs = outputs            // Line 218
```

**Correct Approach:**
```go
// PURE: Create new state instead of mutating
func ConventionScan() Stage {
    return func(ctx context.Context, s State) E.Either[error, State] {
        functions, err := discovery.ScanFunctions(s.ProjectDir)
        if err != nil {
            return E.Left[State](fmt.Errorf("failed to scan functions: %w", err))
        }

        // Create NEW state with updated config
        newState := State{
            ProjectDir: s.ProjectDir,
            Artifacts:  copyArtifacts(s.Artifacts), // Deep copy
            Outputs:    copyOutputs(s.Outputs),     // Deep copy
            Config:     functions,                   // New value
        }
        return E.Right[error](newState)
    }
}

// Helper for immutable updates
func copyArtifacts(m map[string]Artifact) map[string]Artifact {
    result := make(map[string]Artifact, len(m))
    for k, v := range m {
        result[k] = v
    }
    return result
}
```

**Effort:** Medium (2-3 days)
**Impact:** High
**Status:** ❌ TODO

---

### P0: Build Functions Mislabeled as Pure
**Files:**
- `internal/build/go_builder.go` (lines 14-81)
- `internal/build/python_builder.go` (lines 14-81)
- `internal/build/node_builder.go` (lines 14-81)
- `internal/build/java_builder.go` (lines 14-81)

**Issue:** Build functions labeled as "pure" but perform extensive I/O operations directly in the function body:
- File system operations (`os.MkdirAll`, `os.Create`, `os.WriteFile`)
- Process execution (`exec.CommandContext`)
- Temp directory creation (`os.MkdirTemp`)

**Current Code:**
```go
// CLAIMED: "GoBuild is a pure function"
// REALITY: Performs I/O throughout the function body
func GoBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
    artifact, err := func() (Artifact, error) {
        // IMPURE: File system I/O
        if err := os.MkdirAll(filepath.Dir(outputPath), 0754); err != nil {
            return Artifact{}, fmt.Errorf("...")
        }

        // IMPURE: Process execution
        cmd := exec.CommandContext(ctx, "go", "build", ...)
        output, err := cmd.CombinedOutput()  // I/O!

        // IMPURE: File system reads
        checksum, err := calculateChecksum(outputPath)  // I/O!
    }()
}
```

**Correct Approach (Imperative Shell, Functional Core):**
```go
// PURE: Build specification (calculation)
type BuildSpec struct {
    Command    []string
    Env        map[string]string
    WorkDir    string
    OutputPath string
}

// PURE: Generate build specification
func GenerateBuildSpec(cfg Config) BuildSpec {
    return BuildSpec{
        Command: []string{"go", "build", "-tags", "lambda.norpc",
                        "-ldflags", "-s -w", "-o", cfg.OutputPath},
        Env: []string{
            "GOOS=linux",
            "GOARCH=amd64",
            "CGO_ENABLED=0",
        },
        WorkDir:    cfg.SourceDir,
        OutputPath: cfg.OutputPath,
    }
}

// ACTION: Execute build spec (imperative shell)
func ExecuteBuildSpec(ctx context.Context, spec BuildSpec) E.Either[error, Artifact] {
    // All I/O happens here, clearly marked as ACTION
    cmd := exec.CommandContext(ctx, spec.Command[0], spec.Command[1:]...)
    cmd.Env = append(os.Environ(), spec.Env...)
    cmd.Dir = spec.WorkDir

    if err := cmd.Run(); err != nil {
        return E.Left[Artifact](err)
    }

    // More I/O...
    return E.Right[error](artifact)
}

// COMPOSITION: Compose pure + impure
func GoBuild(ctx context.Context, cfg Config) E.Either[error, Artifact] {
    spec := GenerateBuildSpec(cfg)  // PURE
    return ExecuteBuildSpec(ctx, spec)  // ACTION (I/O)
}
```

**Effort:** Large (5-7 days)
**Impact:** High
**Status:** ❌ TODO

---

### P1: Console I/O in Business Logic
**Files:**
- `internal/pipeline/convention_stages.go` (15 instances at lines 17, 29-33, 58, 68, 85, 127-128, 139, 153, 163, 172, 182-189, 192, 211)

**Issue:** Business logic functions directly print to stdout, coupling computation to I/O.

**Current Code:**
```go
func ConventionScan() Stage {
    return func(ctx context.Context, s State) E.Either[error, State] {
        fmt.Println("==> Scanning for Lambda functions...")  // IMPURE!

        functions, err := discovery.ScanFunctions(s.ProjectDir)

        fmt.Printf("Found %d function(s):\n", len(functions))  // IMPURE!
        for _, fn := range functions {
            fmt.Printf("  - %s (%s)\n", fn.Name, fn.Runtime)  // IMPURE!
        }
    }
}
```

**Correct Approach:**
```go
// PURE: Stage returns both state and events (data)
type StageEvent struct {
    Level   string // "info", "success", "error"
    Message string
    Data    map[string]interface{}
}

type StageResult struct {
    State  State
    Events []StageEvent
}

func ConventionScan() func(context.Context, State) E.Either[error, StageResult] {
    return func(ctx context.Context, s State) E.Either[error, StageResult] {
        functions, err := discovery.ScanFunctions(s.ProjectDir)
        if err != nil {
            return E.Left[StageResult](err)
        }

        // Build events (pure data)
        events := []StageEvent{
            {Level: "info", Message: "Scanning for Lambda functions..."},
            {Level: "info", Message: fmt.Sprintf("Found %d function(s)", len(functions))},
        }
        for _, fn := range functions {
            events = append(events, StageEvent{
                Level: "info",
                Message: fmt.Sprintf("- %s (%s)", fn.Name, fn.Runtime),
            })
        }

        newState := State{
            ProjectDir: s.ProjectDir,
            Config:     functions,
            // ... immutable copy
        }

        return E.Right[error](StageResult{
            State:  newState,
            Events: events,
        })
    }
}

// IMPERATIVE SHELL: Print events (happens at edges)
func PrintEvents(events []StageEvent) {
    for _, e := range events {
        fmt.Println(e.Message)
    }
}
```

**Effort:** Medium (2-3 days)
**Impact:** Medium
**Status:** ❌ TODO

---

## Medium Priority Issues (SHOULD FIX)

### P1: Type-Unsafe State.Config
**File:** `internal/pipeline/pipeline.go` (lines 11-17)

**Issue:** Using `interface{}` for `Config` defeats Go's type system and forces type assertions everywhere.

**Current Code:**
```go
type State struct {
    ProjectDir string
    Artifacts  map[string]Artifact
    Outputs    map[string]interface{}
    Config     interface{}  // ← TYPE ERASURE! Defeats type system
}
```

**Recommended Refactoring:**
```go
// Use sum type (discriminated union) pattern
type StateConfig interface {
    isStateConfig()
}

type FunctionListConfig struct {
    Functions []discovery.Function
}
func (FunctionListConfig) isStateConfig() {}

type TerraformConfig struct {
    WorkspaceVars map[string]string
}
func (TerraformConfig) isStateConfig() {}

type State struct {
    ProjectDir string
    Artifacts  map[string]Artifact
    Outputs    map[string]interface{}
    Config     StateConfig  // Type-safe!
}
```

**Effort:** Medium (2-3 days)
**Impact:** Medium
**Status:** ❌ TODO

---

### P1: Missing Railway-Oriented Error Context
**File:** `internal/pipeline/pipeline.go` (lines 42-62)

**Issue:** Error messages don't indicate which stage failed.

**Current Code:**
```go
// Current: errors lack context
for _, stage := range p.stages {
    if E.IsLeft(result) {
        return result  // Which stage failed? Unknown!
    }
    result = stage(ctx, state)
}
```

**Better Approach:**
```go
// Add stage names for debugging
type NamedStage struct {
    Name  string
    Stage Stage
}

func RunWithStageNames(p Pipeline, ctx context.Context, initial State) E.Either[error, State] {
    result := E.Right[error](initial)

    for _, namedStage := range p.namedStages {
        if E.IsLeft(result) {
            return result
        }

        opt := E.ToOption(result)
        state := O.GetOrElse(func() State { return State{} })(opt)

        stageResult := namedStage.Stage(ctx, state)

        // Wrap errors with stage context
        if E.IsLeft(stageResult) {
            return E.MapLeft(func(err error) error {
                return fmt.Errorf("[%s] %w", namedStage.Name, err)
            })(stageResult)
        }

        result = stageResult
    }

    return result
}
```

**Effort:** Small (4-6 hours)
**Impact:** Low
**Status:** ❌ TODO

---

## Low Priority Issues (NICE TO HAVE)

### P2: Repetitive Error Wrapping
**Location:** All builder files (50+ instances)

**Current Code:**
```go
// REPEATED 50+ times across codebase
if err != nil {
    return E.Left[Artifact](err)
}
return E.Right[error](artifact)
```

**Reduction Opportunity:**
```go
// Extract helper (already available in fp-go but underutilized)
func ToEither[T any](value T, err error) E.Either[error, T] {
    if err != nil {
        return E.Left[T](err)
    }
    return E.Right[error](value)
}

// Usage:
artifact, err := buildArtifact()
return ToEither(artifact, err)
```

**Effort:** Trivial (1-2 hours)
**Impact:** Low
**Status:** ❌ TODO

---

### P3: Overly Verbose Option Folding
**Location:** Multiple files (build.go line 132-135, deploy.go line 132-135)

**Current Code:**
```go
// VERBOSE (9 lines)
builder := O.Fold(
    func() build.BuildFunc { return nil },
    func(b build.BuildFunc) build.BuildFunc { return b },
)(builderOpt)
```

**Reduction:**
```go
// CONCISE (1 line) - use O.GetOrElse
builder := O.GetOrElse(func() build.BuildFunc { return nil })(builderOpt)
```

**Effort:** Trivial (1 hour)
**Impact:** Low
**Status:** ❌ TODO

---

### P3: Unused Return Values in Defers
**Location:** Multiple files (python_builder.go lines 76, 81; node_builder.go lines 56, 61)

**Current Code:**
```go
defer func() {
    _ = zipFile.Close() // Best effort close in defer
}()
```

**Better Approach:**
```go
// Named return values + explicit error handling in defer
func PythonBuild(ctx context.Context, cfg Config) (result E.Either[error, Artifact]) {
    zipFile, err := os.Create(outputPath)
    if err != nil {
        return E.Left[Artifact](err)
    }
    defer func() {
        if closeErr := zipFile.Close(); closeErr != nil && E.IsRight(result) {
            result = E.Left[Artifact](closeErr)
        }
    }()
    // ... rest of function
}
```

**Effort:** Small (2-3 hours)
**Impact:** Low
**Status:** ❌ TODO

---

### P3: Map Iteration Non-Determinism
**Location:** `internal/generators/python/project.go` (lines 74-112)

**Current Code:**
```go
// Map iteration order is non-deterministic in Go!
for filePath, generator := range files {
    content := generator()
    if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
        return fmt.Errorf("failed to write %s: %w", filePath, err)
    }
}
```

**Fix:**
```go
// Use slice of ordered pairs for deterministic iteration
type FileGen struct {
    Path string
    Gen  func() string
}

files := []FileGen{
    {"pyproject.toml", func() string { return generatePyProjectToml(config) }},
    {"README.md", func() string { return generateReadme(config) }},
    // ... in order
}

for _, fg := range files {
    // Deterministic iteration!
}
```

**Effort:** Trivial (1 hour)
**Impact:** Low
**Status:** ❌ TODO

---

## Patterns Done Well ✅

- **Either Monad Usage:** Excellent use of `E.Either[error, T]` for error handling throughout build system and pipeline
- **Option Monad Usage:** Proper use of `O.Option[T]` in build registry lookups
- **Function Composition:** Good decorator pattern in `internal/build/functional.go`:
  - `WithCache(cache Cache) func(BuildFunc) BuildFunc`
  - `WithLogging(log Logger) func(BuildFunc) BuildFunc`
  - `Compose(decorators ...func(BuildFunc) BuildFunc) func(BuildFunc) BuildFunc`
- **Immutable Data Structures:** All config and artifact structs are immutable (no pointer receivers, no mutation methods)
- **Pure Agent Code:** `internal/agent/generator.go` is exemplary - 100% pure functions for code generation

---

## Gold Standard Reference

**File:** `internal/agent/generator.go`

This file demonstrates **perfect** functional programming:
- 100% pure functions
- Zero side effects
- Excellent separation of data/calculations/actions
- Well-tested
- Use as template for refactoring other packages

---

## Effort Summary

| Violation | Effort | Impact | Priority | Status |
|-----------|--------|--------|----------|--------|
| State Mutation | Medium (2-3 days) | High | P0 | ❌ TODO |
| Build Function Purity | Large (5-7 days) | High | P0 | ❌ TODO |
| Console I/O | Medium (2-3 days) | Medium | P1 | ❌ TODO |
| State Type Safety | Medium (2-3 days) | Medium | P1 | ❌ TODO |
| Error Context | Small (4-6 hours) | Low | P1 | ❌ TODO |
| Code Duplication | Trivial (1-2 hours) | Low | P2 | ❌ TODO |
| Verbose Option Folding | Trivial (1 hour) | Low | P3 | ❌ TODO |
| Defer Error Handling | Small (2-3 hours) | Low | P3 | ❌ TODO |
| Map Non-Determinism | Trivial (1 hour) | Low | P3 | ❌ TODO |

**Total Estimated Effort:** 12-18 developer days to achieve 9/10 FP score

---

## Package-Level Scores

| Package | FP Score | Notes |
|---------|----------|-------|
| `internal/agent/` | **10/10** | Exemplary - pure generators |
| `internal/build/` | **5/10** | Good structure, but I/O not isolated |
| `internal/pipeline/` | **4/10** | State mutation, console I/O embedded |
| `internal/discovery/` | **8/10** | Mostly pure, minimal I/O |
| `internal/terraform/` | **7/10** | Good functional wrappers |
| `internal/cli/` | **8/10** | Acceptable (imperative shell) |
| `internal/generators/` | **6/10** | Mix of pure and impure |
| `internal/state/` | **7/10** | Good separation marked |

---

## Recommended Implementation Order

1. **Week 1: P0 Issues**
   - Day 1-3: Fix state mutation in pipeline
   - Day 4-7: Refactor build functions (pure core + imperative shell)

2. **Week 2: P1 Issues**
   - Day 1-3: Remove console I/O from business logic
   - Day 4-5: Add type safety to State.Config
   - Day 5: Add error context to pipeline

3. **Week 3: Polish**
   - Day 1: Fix all P2/P3 issues (code duplication, defer handling, etc.)
   - Day 2-3: Review and testing
   - Day 4-5: Documentation updates

---

## Success Criteria

- [ ] All pipeline stages return new State instances (no mutation)
- [ ] Build functions clearly separated into pure calculations + I/O actions
- [ ] Zero console I/O in business logic (events returned as data)
- [ ] State.Config uses discriminated union pattern
- [ ] Pipeline errors include stage names
- [ ] All tests pass at 90%+ coverage
- [ ] Zero linting issues
- [ ] Code review passes with 9/10 FP score

---

## Notes

- **Original Audit Date:** 2025-10-29
- **Target Completion:** 3 weeks from start
- **Blocker:** None (all work is independent)
- **Risk:** Medium (requires careful refactoring of core abstractions)
