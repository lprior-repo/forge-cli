# Functional Programming Improvements Summary

**Date:** 2025-10-29
**Status:** In Progress

## Completed Work

### ✅ P0: State Mutation in Pipeline (COMPLETE)

**Status:** Already implemented correctly!

Upon audit, discovered that the pipeline stages were already following immutable state patterns:
- All stages in `convention_stages.go` create new `State` instances
- All stages in `terraform_stages.go` create new output maps and return new `State`
- No direct mutation of the `State` struct parameter

**Files Verified:**
- `internal/pipeline/convention_stages.go` - ✅ Immutable
- `internal/pipeline/terraform_stages.go` - ✅ Immutable

### ✅ P0: Build Function Purity (PARTIALLY COMPLETE)

**Status:** Go and Python refactored successfully

#### What Was Fixed

**Go Builder** (`internal/build/go_builder.go`):
- Created `GoBuildSpec` struct for pure build specifications
- Implemented `GenerateGoBuildSpec()` - **PURE** function that generates build commands
- Implemented `ExecuteGoBuildSpec()` - **ACTION** function that performs I/O
- Refactored `GoBuild()` to compose pure + impure: `spec := Generate(cfg); Execute(ctx, spec)`

**Python Builder** (`internal/build/python_builder.go`):
- Created `PythonBuildSpec` struct for pure build specifications
- Implemented `GeneratePythonBuildSpec()` - **PURE** function that generates pip/uv commands
- Implemented `ExecutePythonBuildSpec()` - **ACTION** function that performs I/O
- Refactored `PythonBuild()` to compose pure + impure
- Marked helper functions appropriately:
  - `envSlice()` - **PURE** (calculation)
  - `addDirToZip()` - **ACTION** (I/O)
  - `shouldSkipFile()` - **PURE** (calculation)

**Shared Infrastructure** (`internal/build/builder.go`):
- Added `executeCommand()` helper - **ACTION** that executes commands
- Clearly marked all I/O operations with comments

#### Test Updates

Fixed test expectations to handle new error messages from generic `executeCommand()` helper:
- `go_builder_test.go` - Updated error message assertion
- `python_builder_test.go` - Updated error message assertion

All 60+ build tests passing ✅

#### Remaining Work

**Node Builder** (`internal/build/node_builder.go`):
- TODO: Create `NodeBuildSpec` struct
- TODO: Implement `GenerateNodeBuildSpec()` - PURE
- TODO: Implement `ExecuteNodeBuildSpec()` - ACTION
- TODO: Refactor `NodeBuild()` to compose pure + impure

**Java Builder** (`internal/build/java_builder.go`):
- TODO: Create `JavaBuildSpec` struct
- TODO: Implement `GenerateJavaBuildSpec()` - PURE
- TODO: Implement `ExecuteJavaBuildSpec()` - ACTION
- TODO: Refactor `JavaBuild()` to compose pure + impure

**Estimated Effort:** 2-3 hours (similar pattern to Go/Python)

## Pending Work

### P1: Console I/O in Pipeline Business Logic (HIGH PRIORITY)

**Problem:** 15+ instances of `fmt.Println` / `fmt.Printf` embedded in pipeline stages.

**Files Affected:**
- `internal/pipeline/convention_stages.go` (12 instances)
- `internal/pipeline/terraform_stages.go` (3 instances)

**Solution Approach:**

1. Create event data structures:
```go
type StageEvent struct {
    Level   string // "info", "success", "error"
    Message string
    Data    map[string]interface{}
}

type StageResult struct {
    State  State
    Events []StageEvent
}
```

2. Refactor stages to return events as data instead of printing:
```go
func ConventionScan() func(context.Context, State) E.Either[error, StageResult] {
    return func(ctx context.Context, s State) E.Either[error, StageResult] {
        functions, err := discovery.ScanFunctions(s.ProjectDir)
        if err != nil {
            return E.Left[StageResult](err)
        }

        events := []StageEvent{
            {Level: "info", Message: "Scanning for Lambda functions..."},
            {Level: "info", Message: fmt.Sprintf("Found %d function(s)", len(functions))},
        }

        return E.Right[error](StageResult{
            State:  newState,
            Events: events,
        })
    }
}
```

3. Create imperative shell function to print events:
```go
func PrintEvents(events []StageEvent) {
    for _, e := range events {
        fmt.Println(e.Message)
    }
}
```

**Estimated Effort:** 2-3 days

### P2: Additional Improvements (MEDIUM PRIORITY)

See `docs/FP_AUDIT_CLEANUP.md` for full list:
- Type-safe State.Config (discriminated union pattern)
- Railway-oriented error context (add stage names to errors)
- Reduce repetitive error wrapping
- Fix defer error handling
- Fix map iteration non-determinism

## Testing Status

### Current Test Results
- **Build Package:** All 60+ tests passing ✅
- **Coverage:** ~85% (target: 90%)
- **Linting:** Not yet run

### Next Steps
1. Run full test suite: `task test:all`
2. Check coverage: `task coverage:check`
3. Run linter: `task lint`
4. Fix any issues found

## Functional Programming Score Progress

| Audit Point | Before | After | Target |
|-------------|--------|-------|--------|
| Overall Score | 6.5/10 | **7.5/10** | 9/10 |
| Build Package | 5/10 | **8/10** | 9/10 |
| Pipeline Package | 4/10 | **5/10** | 9/10 |

**Improvements:**
- ✅ State immutability verified (was already good!)
- ✅ Build functions now properly separate pure core from imperative shell
- ✅ Clear labeling of PURE vs ACTION functions
- ⏳ Console I/O still embedded in pipeline (next priority)

## Key Learnings

### What Worked Well
1. **Spec Pattern:** Creating `*BuildSpec` structs for pure specifications worked excellently
2. **Composition:** `Build() = Generate(spec) + Execute(spec)` is clean and testable
3. **Clear Labels:** Comments marking **PURE** vs **ACTION** improve code clarity
4. **Helper Functions:** Generic `executeCommand()` reduces duplication

### Challenges Encountered
1. **Test Updates:** Changing error messages required updating test assertions
2. **I/O Detection:** Some I/O operations (like `exec.LookPath`) are hard to isolate
3. **Trade-offs:** Perfect purity (no I/O in entry points) vs pragmatic separation

### Best Practices Established
- Always separate calculation (pure) from I/O (action)
- Use `{Type}Spec` structs for pure specifications
- Mark functions with `// PURE:` or `// ACTION:` comments
- Compose pure + impure at the top level
- Test pure functions without mocks

## Next Session Plan

1. **Finish P0:** Refactor Node and Java builders (2-3 hours)
2. **Start P1:** Remove console I/O from pipeline (2-3 days)
3. **Run Tests:** Full test suite + coverage check
4. **Document:** Update CLAUDE.md with new patterns

**Total Remaining Effort:** ~4-5 days to reach 9/10 FP score
