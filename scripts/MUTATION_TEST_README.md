# Mutation Testing Script Documentation

## Overview

The `mutation-test.nu` script provides production-ready mutation testing for the Forge codebase using Nushell. It tests the quality of test suites by introducing mutations (bugs) and verifying that tests catch them.

## Features

### Production-Ready Patterns

- **Pure Functional Design**: Data/Actions/Calculations separation
- **Explicit Type Signatures**: All functions documented with input/output types
- **Railway-Oriented Error Handling**: Comprehensive error propagation
- **Parallel Execution**: Uses `par-each` for concurrent package testing
- **Timeout Support**: Per-package timeout with graceful handling
- **Streaming Results**: Real-time progress as packages complete
- **Robust Parsing**: Error-safe parsing with fallback values
- **Immutable Pipelines**: No mutable state, pure data transformations

### Key Capabilities

1. **Parallel Execution** (default: 4 workers)
   - Configurable with `--parallel` flag
   - Independent package testing
   - Streaming results as they complete

2. **Timeout Management** (default: 8 hours per package)
   - Prevents hanging on complex packages
   - Graceful timeout handling
   - Reports timeout status clearly

3. **Comprehensive Error Handling**
   - Timeout errors (exit code 124)
   - Execution errors (non-zero exit codes)
   - Parse errors (malformed output)
   - No mutation cases (empty packages)

4. **Rich Reporting**
   - Real-time progress updates
   - Per-package breakdown with progress bars
   - Overall statistics summary
   - Exit codes based on quality thresholds

## Usage

### Basic Commands

```bash
# Run all packages in parallel (default: 4 workers)
nu scripts/mutation-test.nu

# Test a specific package
nu scripts/mutation-test.nu --package internal/build

# Increase parallelism (8 workers)
nu scripts/mutation-test.nu --parallel 8

# Set timeout to 1 hour per package
nu scripts/mutation-test.nu --timeout 3600

# Verbose output (show detailed errors)
nu scripts/mutation-test.nu --verbose

# Combination
nu scripts/mutation-test.nu --package internal/cli --verbose --timeout 1800
```

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--package` | `-p` | string | "" | Specific package to test (e.g., internal/build) |
| `--verbose` | `-v` | flag | false | Show detailed mutation results and errors |
| `--parallel` | `-j` | int | 4 | Number of parallel jobs |
| `--timeout` | `-t` | int | 28800 | Timeout per package in seconds (8 hours) |

### Package List

The script tests 35 packages by default:

**Core** (8 packages):
- internal/build
- internal/cli
- internal/config
- internal/discovery
- internal/scaffold
- internal/pipeline
- internal/terraform
- internal/state

**Generators** (6 packages):
- internal/generators
- internal/generators/dynamodb
- internal/generators/python
- internal/generators/s3
- internal/generators/sns
- internal/generators/sqs

**Lingon** (1 package):
- internal/lingon

**TF Modules** (15 packages):
- internal/tfmodules
- internal/tfmodules/apigatewayv2
- internal/tfmodules/appconfig
- internal/tfmodules/appsync
- internal/tfmodules/cloudfront
- internal/tfmodules/dynamodb
- internal/tfmodules/eventbridge
- internal/tfmodules/lambda
- internal/tfmodules/s3
- internal/tfmodules/secretsmanager
- internal/tfmodules/sns
- internal/tfmodules/sqs
- internal/tfmodules/ssm
- internal/tfmodules/stepfunctions

**UI** (1 package):
- internal/ui

## Output Format

### Real-Time Progress

```
[internal/build] ✅ Score: 85.7% (12/14 mutations killed)
[internal/cli] ⚠️  Score: 72.3% (8/11 mutations killed)
[internal/config] ⏱️  TIMEOUT after 28800s (8.0h)
[internal/discovery] ⚠️  No mutations generated
```

### Final Summary

```
==================================
📊 Overall Mutation Test Results
==================================

Total Mutations:  143
Killed Mutations: 121
Overall Score:    84.6%

Per-Package Breakdown:

  ✅ internal/build                    ████████████████████████████████████████ 90.5%
  ✅ internal/cli                      ████████████████████████████████████░░░░ 88.2%
  ⚠️  internal/config                  ████████████████████████████░░░░░░░░░░░░ 72.3%
  ❌ internal/discovery                ████████████████░░░░░░░░░░░░░░░░░░░░░░░░ 45.0%
  ⚠️  internal/terraform               [TIMEOUT]

✅ EXCELLENT: Test suite catches 85%+ of mutations
```

## Exit Codes

The script exits with codes based on mutation score thresholds:

| Exit Code | Threshold | Meaning |
|-----------|-----------|---------|
| 0 | ≥85% | EXCELLENT - Test suite is high quality |
| 0 | 75-85% | GOOD - Test suite is acceptable (aim for 85%) |
| 1 | 65-75% | FAIR - Test suite needs improvement |
| 1 | <65% | NEEDS IMPROVEMENT - Critical gaps in testing |

## Architecture

### Pure Functions (Calculations)

All data transformations are pure functions with explicit type signatures:

```nushell
# Type: string -> string -> MutationResult
def parse_mutation_output [pkg: string, stdout: string]: nothing -> record

# Type: list<MutationResult> -> float
def calculate_overall_score []: list -> float

# Type: float -> string
def get_status_emoji []: float -> string

# Type: float -> string
def format_score []: float -> string

# Type: float -> int -> string
def generate_progress_bar [width: int]: float -> string
```

### Actions (I/O Operations)

Side-effecting operations are isolated:

```nushell
# Type: string -> int -> bool -> MutationResult
def run_package_mutation_test [
    pkg: string
    timeout_secs: int
    verbose: bool
]: nothing -> record
```

### Data Types

**MutationResult** record type:
```nushell
{
  package: string      # Package path (e.g., internal/build)
  score: float         # Mutation score (0.0 - 1.0)
  passed: int          # Number of mutations killed
  failed: int          # Number of mutations that survived
  total: int           # Total mutations generated
  skipped: int         # Number of mutations skipped
  duplicated: int      # Number of duplicate mutations
  error: string        # Error type: "" | "timeout" | "parse_error" | "no_mutations" | "execution_error"
}
```

## Error Handling

### Timeout Errors

When a package exceeds the timeout:

```
[internal/complex_package] ⏱️  TIMEOUT after 28800s (8.0h)
```

Result: `error: "timeout"`, score: 0.0

### Execution Errors

When go-mutesting fails:

```
[internal/broken_package] ❌ Execution failed
```

Result: `error: "execution_error"`, score: 0.0

### No Mutations

When no mutations are generated:

```
[internal/empty_package] ⚠️  No mutations generated
```

Result: `error: "no_mutations"`, score: 0.0

### Parse Errors

When output format is unexpected, the parser uses safe defaults (0.0, 0) with try/catch blocks.

## Performance Characteristics

### Sequential vs Parallel

- **Sequential**: ~35 packages × 5 min/package = ~175 minutes
- **Parallel (4 workers)**: ~35 packages ÷ 4 × 5 min = ~44 minutes
- **Parallel (8 workers)**: ~35 packages ÷ 8 × 5 min = ~22 minutes

### Memory Usage

- Streaming pipeline architecture (low memory)
- No collection until final summary
- Each worker operates independently

### CPU Usage

- CPU-bound: mutation testing is computationally intensive
- Parallelism matches available cores
- Default (4 workers): conservative for 4-core systems
- Recommended: Set `--parallel` to number of CPU cores

## Integration with Taskfile

Add to `Taskfile.yml`:

```yaml
tasks:
  mutation:
    desc: Run mutation testing on all packages
    cmds:
      - nu scripts/mutation-test.nu

  mutation:verbose:
    desc: Run mutation testing with verbose output
    cmds:
      - nu scripts/mutation-test.nu --verbose

  mutation:package:
    desc: Run mutation testing on specific package
    cmds:
      - nu scripts/mutation-test.nu --package {{.PKG}}

  mutation:fast:
    desc: Run mutation testing with 8 parallel workers
    cmds:
      - nu scripts/mutation-test.nu --parallel 8
```

Usage:
```bash
task mutation
task mutation:verbose
task mutation:package PKG=internal/build
task mutation:fast
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Mutation Testing

on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly on Sunday at 2 AM
  workflow_dispatch:

jobs:
  mutation-test:
    runs-on: ubuntu-latest
    timeout-minutes: 480  # 8 hours

    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Setup Nushell
        run: |
          wget https://github.com/nushell/nushell/releases/download/0.99.0/nu-0.99.0-x86_64-linux-gnu.tar.gz
          tar xf nu-0.99.0-x86_64-linux-gnu.tar.gz
          sudo mv nu /usr/local/bin/

      - name: Install go-mutesting
        run: go install github.com/avito-tech/go-mutesting/cmd/go-mutesting@latest

      - name: Run Mutation Testing
        run: nu scripts/mutation-test.nu --parallel 8

      - name: Upload Results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: mutation-test-results
          path: .mutation-test-results/
```

## Troubleshooting

### Issue: "command not found: go-mutesting"

**Solution**: Install go-mutesting:
```bash
go install github.com/avito-tech/go-mutesting/cmd/go-mutesting@latest
```

### Issue: "command not found: timeout"

**Solution**: Install coreutils (usually pre-installed on Linux/macOS):
```bash
# Ubuntu/Debian
sudo apt-get install coreutils

# macOS
brew install coreutils
```

### Issue: All packages timeout

**Solution**: Increase timeout or reduce parallelism:
```bash
# Increase timeout to 16 hours
nu scripts/mutation-test.nu --timeout 57600

# Reduce parallelism to avoid resource contention
nu scripts/mutation-test.nu --parallel 2
```

### Issue: Low mutation scores across all packages

**Solution**: This indicates test suite quality issues. Focus on:
1. Adding edge case tests
2. Testing error handling paths
3. Validating boundary conditions
4. Property-based testing

## Best Practices

1. **Run Regularly**: Weekly mutation testing catches test quality regressions
2. **Start Small**: Test individual packages first before full suite
3. **Tune Parallelism**: Match `--parallel` to available CPU cores
4. **Monitor Timeouts**: Packages that timeout may need refactoring
5. **Track Trends**: Monitor mutation scores over time
6. **Prioritize Critical Paths**: Focus on high-risk packages first
7. **Integrate with CI**: Run on schedule, not on every commit

## References

- [go-mutesting](https://github.com/avito-tech/go-mutesting) - Mutation testing tool
- [Nushell Documentation](https://www.nushell.sh/book/) - Shell language reference
- [Mutation Testing Best Practices](https://mutation-testing.github.io/) - Industry standards
