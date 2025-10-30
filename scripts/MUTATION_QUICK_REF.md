# Mutation Testing Quick Reference

## Essential Commands

```bash
# Default (4 workers, 8h timeout)
nu scripts/mutation-test.nu

# Single package
nu scripts/mutation-test.nu -p internal/build

# Fast (8 workers)
nu scripts/mutation-test.nu -j 8

# Short timeout (1h)
nu scripts/mutation-test.nu -t 3600

# Verbose errors
nu scripts/mutation-test.nu -v

# Combination
nu scripts/mutation-test.nu -p internal/cli -v -t 1800
```

## Exit Codes

| Score | Exit | Status |
|-------|------|--------|
| ≥85% | 0 | EXCELLENT |
| 75-85% | 0 | GOOD (aim for 85%) |
| 65-75% | 1 | FAIR (needs work) |
| <65% | 1 | NEEDS IMPROVEMENT |

## Performance Guide

| Workers | Packages | Time |
|---------|----------|------|
| 1 | 35 | ~175 min |
| 4 | 35 | ~44 min |
| 8 | 35 | ~22 min |

## Error Types

| Error | Meaning |
|-------|---------|
| TIMEOUT | Package exceeded time limit |
| ERROR | go-mutesting failed to run |
| NO MUTATIONS | Package has no mutable code |
| PARSE ERROR | Output format unexpected |

## Package List (35 total)

**Core** (8): build, cli, config, discovery, scaffold, pipeline, terraform, state

**Generators** (6): generators, generators/dynamodb, generators/python, generators/s3, generators/sns, generators/sqs

**Lingon** (1): lingon

**TF Modules** (15): tfmodules, tfmodules/apigatewayv2, tfmodules/appconfig, tfmodules/appsync, tfmodules/cloudfront, tfmodules/dynamodb, tfmodules/eventbridge, tfmodules/lambda, tfmodules/s3, tfmodules/secretsmanager, tfmodules/sns, tfmodules/sqs, tfmodules/ssm, tfmodules/stepfunctions

**UI** (1): ui

## Taskfile Integration

```bash
task mutation                # Default run
task mutation:verbose        # With verbose output
task mutation:package PKG=internal/build  # Single package
task mutation:fast          # 8 workers
```

## Recommended Workflow

1. Test single package first: `nu scripts/mutation-test.nu -p internal/build -v`
2. Verify results make sense
3. Run full suite: `nu scripts/mutation-test.nu -j 8`
4. Monitor scores weekly
5. Focus on packages <75%
