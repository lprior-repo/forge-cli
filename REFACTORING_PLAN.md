# Forge Refactoring Plan - Martin Fowler Quality Standards

## Executive Summary

After deep dive analysis of the codebase, here's the refactoring plan to achieve Martin Fowler-level quality: highly testable, pure functional, well-designed, and zero redundancy.

## Current State Analysis

### ✅ What's Working Well

1. **Functional Programming Patterns**
   - Either monad for error handling
   - Option monad for optional values
   - Pure functions in build/ and terraform/ packages
   - Function types over interfaces (excellent for testing)

2. **Test Coverage**
   - 226 total tests (189 unit + 37 integration)
   - 100% pass rate
   - Comprehensive test suites for all core functionality

3. **Pipeline Architecture**
   - Composable stages
   - Railway-oriented programming
   - Clean separation of concerns

4. **Build System**
   - Multi-runtime support (Go, Python, Node.js)
   - Higher-order functions (WithCache, WithLogging)
   - Registry pattern with Option monad

### ⚠️ Issues Found

1. **Unused Code**
   - `internal/ci/` package (252 lines) - NOT imported anywhere
   - Should be removed or integrated

2. **Naming Inconsistencies**
   - `build.Config` vs `config.Config` - potential confusion
   - Should rename to `BuildConfig` and `ProjectConfig`

3. **Incomplete Implementation**
   - `internal/lingon/generator.go` - placeholder exportToTerraform()
   - Needs actual Lingon resource generation
   - Missing reference resolution (${table.users.arn})

4. **Missing Integration**
   - Imported files.zip has complete Lingon stack.go
   - Not yet integrated with our generator
   - HCL config parser needed for forge.config.hcl

5. **Test Gaps**
   - No integration tests for Lingon generation
   - Missing property-based tests
   - No end-to-end tests with actual Terraform

## Refactoring Strategy

### Phase 1: Remove Redundancies ✂️

**Actions**:
1. Delete `internal/ci/` package (unused)
2. Rename `build.Config` → `BuildConfig` for clarity
3. Rename `config.Config` → `ProjectConfig` for consistency
4. Remove any dead code found during analysis

**Impact**: -252 lines, improved naming clarity

**Tests Required**: Update existing tests with new names

### Phase 2: Integrate Lingon Stack 🔗

**Actions**:
1. Copy `stack.go` from files.zip to `internal/lingon/`
2. Implement actual Terraform resource generation
3. Replace placeholder `exportToTerraform()` with real implementation
4. Add resource creation functions:
   - `createLambdaFunctionResources()`
   - `createAPIGatewayResources()`
   - `createDynamoDBTableResources()`
   - etc.

**Impact**: +800 lines of production-ready Lingon code

**Tests Required**:
- Unit tests for each resource creation function
- Integration tests with actual Lingon stack export

### Phase 3: Reference Resolution System 🔗

**Actions**:
1. Implement variable reference parser
2. Add dependency graph for correct resource ordering
3. Support `${function.api.arn}`, `${table.users.name}`, etc.
4. Validate references before generation

**Impact**: +200 lines

**Tests Required**:
- Unit tests for reference parsing
- Integration tests for complex dependency chains
- Circular dependency detection tests

### Phase 4: HCL Config Parser 📄

**Actions**:
1. Add HCL parser for `forge.config.hcl` (from files.zip)
2. Support both YAML and HCL formats
3. Unify configuration loading
4. Add validation with detailed error messages

**Impact**: +150 lines

**Tests Required**:
- Unit tests for HCL parsing
- Error handling tests
- Format detection tests

### Phase 5: Pure Function Refactoring 🔧

**Actions**:
1. Audit all functions for side effects
2. Extract I/O to edges (ports & adapters pattern)
3. Make core logic purely functional
4. Apply Fowler's "Refactoring" patterns:
   - Extract Function
   - Replace Temp with Query
   - Introduce Parameter Object
   - Preserve Whole Object

**Impact**: Improved testability, no line count change

**Tests Required**: All existing tests should still pass

### Phase 6: Comprehensive Testing 🧪

**Actions**:
1. Add property-based tests using fp-go
2. Add integration tests for Lingon generation
3. Add end-to-end tests (optional, requires Terraform binary)
4. Achieve >90% coverage on all packages
5. Add mutation testing

**Impact**: +500 lines of tests

**Test Goals**:
- Unit tests: 250+ (currently 189)
- Integration tests: 50+ (currently 37)
- Total: 300+ tests

### Phase 7: Documentation & Polish 📚

**Actions**:
1. Update all documentation with new architecture
2. Add architecture decision records (ADRs)
3. Create contribution guidelines
4. Add code examples for all major features
5. Update README with complete getting started guide

**Impact**: Better developer experience

## Martin Fowler Quality Checklist

### Design Patterns
- ✅ **Repository Pattern** - terraform executor abstracts terraform-exec
- ✅ **Strategy Pattern** - build registry with different builders
- ✅ **Decorator Pattern** - WithCache, WithLogging
- ✅ **Pipeline Pattern** - composable stages
- ⏳ **Adapter Pattern** - Need for Lingon resources
- ⏳ **Factory Pattern** - Need for resource creation

### SOLID Principles
- ✅ **Single Responsibility** - Each package has clear purpose
- ✅ **Open/Closed** - Function types allow extension
- ✅ **Liskov Substitution** - N/A (no inheritance)
- ✅ **Interface Segregation** - Small, focused function types
- ✅ **Dependency Inversion** - Function types invert dependencies

### Code Smells to Remove
- ✅ **Dead Code** - Remove ci/ package
- ✅ **Duplicate Code** - No duplicates found
- ⏳ **Long Functions** - Need to check Lingon integration
- ⏳ **Large Classes** - ForgeConfig has 300+ fields (acceptable for config)
- ✅ **Comments** - Well-documented, not excessive

### Testability Metrics
- ✅ **Pure Functions** - Most core logic is pure
- ✅ **No Global State** - All state passed explicitly
- ✅ **Dependency Injection** - Function types enable easy mocking
- ✅ **Fast Tests** - Unit tests < 1s
- ✅ **Isolated Tests** - No test dependencies

## Implementation Timeline

### Week 1: Foundation
- Day 1-2: Phase 1 (Remove redundancies)
- Day 3-4: Phase 2 (Integrate Lingon stack)
- Day 5: Testing and validation

### Week 2: Features
- Day 1-2: Phase 3 (Reference resolution)
- Day 3: Phase 4 (HCL config parser)
- Day 4-5: Phase 5 (Pure function refactoring)

### Week 3: Quality
- Day 1-3: Phase 6 (Comprehensive testing)
- Day 4-5: Phase 7 (Documentation)

## Success Criteria

### Code Quality
- [ ] Zero unused code
- [ ] 100% consistent naming
- [ ] All functions pure or clearly marked as I/O
- [ ] No circular dependencies
- [ ] Clear separation of concerns

### Test Coverage
- [ ] 300+ total tests
- [ ] >90% coverage on all packages
- [ ] All tests passing
- [ ] <1s unit test execution
- [ ] <30s integration test execution

### Documentation
- [ ] Complete API documentation
- [ ] Usage examples for all features
- [ ] Architecture decision records
- [ ] Contribution guidelines
- [ ] Getting started guide

### Performance
- [ ] O(1) cache lookups
- [ ] O(n) build operations
- [ ] <100ms config validation
- [ ] <1s Terraform generation

## Files to Modify

### Delete
- `internal/ci/ci.go` (252 lines)

### Rename
- None (build.Config and config.Config are in different packages - OK)

### Add
- `internal/lingon/stack.go` (from files.zip)
- `internal/lingon/resources.go` (new)
- `internal/lingon/references.go` (new)
- `internal/config/hcl_parser.go` (new)
- `internal/lingon/stack_test.go` (new)
- `internal/lingon/resources_test.go` (new)
- `internal/lingon/references_test.go` (new)

### Modify
- `internal/lingon/generator.go` - Replace placeholder
- `internal/lingon/config_types.go` - Add HCL tags
- `internal/config/config.go` - Add HCL parser
- All test files - Add new test cases
- All documentation - Update

## Risk Assessment

### Low Risk
- Removing unused ci/ package
- Adding new tests
- Documentation updates

### Medium Risk
- Integrating Lingon stack (well-tested pattern)
- Adding HCL parser (standard library)
- Reference resolution (complex but isolated)

### High Risk
- None identified

## Rollback Plan

- All changes in feature branches
- Comprehensive testing before merge
- Tag stable versions
- Keep old code in git history

## Monitoring

Post-refactoring metrics to track:
- Test execution time
- Build times
- Test coverage percentage
- Cyclomatic complexity
- Code churn rate

---

**Next Steps**: Begin Phase 1 - Remove redundancies
