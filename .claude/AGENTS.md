# Claude Code Agents - Usage Guide

This directory contains specialized Claude Code agents for the Forge project. These agents are designed to assist with specific development tasks while maintaining code quality and following functional programming principles.

## Available Agents

### 1. code-reviewer
**Purpose**: Comprehensive code reviews focusing on correctness, security, performance, and testing

**When to use**:
- After writing significant code changes
- Before creating pull requests
- When debugging quality issues
- During code refactoring

**Invocation**:
```
Use the code-reviewer agent to review my recent changes
```

**What it does**:
1. **Automated Verification**:
   - Runs `go test ./...` to verify all tests pass
   - Checks coverage with `task coverage:check` (90% minimum)
   - Runs `golangci-lint` for code quality
   - Executes `go test -race` to detect race conditions

2. **Manual Review**:
   - Correctness & reliability (logic errors, edge cases)
   - Security (input validation, path traversal, secrets)
   - Performance (algorithmic complexity, memory usage)
   - Testing (coverage gaps, edge case tests)
   - Code quality (function complexity, naming, duplication)
   - Go-specific issues (scanner.Err(), defer, context cancellation)

3. **Structured Output**:
   - Executive summary with verdict
   - Critical issues (must fix before merge)
   - Code quality issues
   - Performance concerns
   - Testing gaps
   - Positive highlights
   - Prioritized recommendations with effort estimates

**Example Output Structure**:
```xml
<code_review>
  <automated_verification>
    <tests status="PASS">All 156 tests passing</tests>
    <coverage status="PASS">92.3%</coverage>
    <linter status="PASS">Zero issues</linter>
    <race_detection status="FAIL">4 races found</race_detection>
  </automated_verification>

  <executive_summary>
    <verdict>Request Changes</verdict>
    <critical_issues_count>1</critical_issues_count>
    <estimated_effort>2-3 hours</estimated_effort>
  </executive_summary>

  <critical_issues>
    <!-- Detailed issues with code examples and fixes -->
  </critical_issues>

  <recommendations>
    1. Fix race condition in Spinner (Critical, 1 hour)
    2. Add input validation (High, 30 min)
    <!-- ... -->
  </recommendations>
</code_review>
```

### 2. go-expert
**Purpose**: Expert Go development with functional programming, TDD, and production-ready code

**When to use**:
- Writing new Go code
- Fixing bugs in Go code
- Refactoring Go code to functional style
- Adding comprehensive tests

**Invocation**:
```
Use the go-expert agent to implement [feature]
```

**What it focuses on**:
- Functional programming patterns (pure functions, immutability)
- Monadic error handling (Either/Option from fp-go)
- Table-driven tests with 90%+ coverage
- Concurrency safety (sync primitives, race-free code)
- Idiomatic Go (proper error handling, defer, context)

**Example patterns it teaches**:
```go
// Monadic error handling
func ProcessData(input string) E.Either[error, Result] {
    if input == "" {
        return E.Left[Result](errors.New("empty input"))
    }
    return E.Right[error](Result{Data: input})
}

// Table-driven tests
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Result
        wantErr bool
    }{
        // ... test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test logic
        })
    }
}
```

### 3. nushell-writer
**Purpose**: Production-ready Nushell scripts with functional programming and type safety

**When to use**:
- Writing build scripts
- Creating automation tools
- Developing deployment scripts
- Building CLI utilities in Nushell

**Invocation**:
```
Use the nushell-writer agent to create a script that [task]
```

**What it focuses on**:
- Explicit type signatures for all functions
- Pure functional patterns (no mutation)
- Streaming pipelines (lazy evaluation)
- Comprehensive testing and error handling
- Production quality (logging, monitoring, graceful failures)

## Agent Architecture

### File Structure
```
.claude/
└── agents/
    ├── code-reviewer.md    # Code review specialist
    ├── go-expert.md        # Go development expert
    └── nushell-writer.md   # Nushell scripting expert
```

### Agent Format

Each agent is a Markdown file with YAML frontmatter:

```markdown
---
name: agent-name
description: Clear description of when and how to use this agent
tools: Read, Write, Edit, Bash, Grep, Glob
model: sonnet
---

# Agent Name

Agent instructions and expertise...
```

**Required fields**:
- `name`: Lowercase with hyphens (e.g., `code-reviewer`)
- `description`: When to use, what it does (include "Use PROACTIVELY" for automatic delegation)
- `tools`: Comma-separated list of allowed tools
- `model`: `sonnet`, `opus`, or `haiku` (optional, defaults to sonnet)

### Invocation Methods

#### 1. Explicit Natural Language
Directly request the agent by name:
```
Use the code-reviewer agent to review my changes
Have the go-expert agent implement this feature
Ask the nushell-writer agent to create a deployment script
```

#### 2. Automatic/Proactive
Claude Code automatically selects agents based on:
- Task description matching agent's description
- "Use PROACTIVELY" keyword in description
- Current context and required tools

For example, the code-reviewer agent has:
> description: "Use PROACTIVELY after writing significant code..."

This means Claude will automatically suggest using it after you write code.

## Best Practices

### For Agent Users

1. **Be Specific**: Clearly describe what you want the agent to do
   - ❌ "Review my code"
   - ✅ "Use the code-reviewer agent to review the UI package for race conditions and test coverage"

2. **Provide Context**: Include relevant file paths and scope
   - ❌ "Fix this"
   - ✅ "Use the go-expert agent to fix the race condition in internal/ui/progress.go:39"

3. **Review Agent Output**: Agents provide recommendations, not automatic fixes
   - Read the structured output carefully
   - Prioritize critical issues first
   - Verify automated checks pass

4. **Iterative Workflow**:
   ```
   1. Write code
   2. Request code-reviewer agent
   3. Fix critical issues
   4. Re-run automated verification
   5. Address remaining issues
   6. Final review
   ```

### For Agent Developers

1. **Clear Descriptions**: Make it obvious when to use the agent
2. **Specific Tools**: Only request tools the agent needs
3. **Examples**: Include code examples for common patterns
4. **Structured Output**: Use consistent XML/markdown structure
5. **Actionable**: Every issue should have a specific fix
6. **Prioritized**: Order recommendations by severity and effort

## Integration with Development Workflow

### Pre-Commit
```bash
# Before committing
1. Write code
2. "Use code-reviewer agent to check my changes"
3. Fix any critical issues
4. Run: task test
5. Run: task coverage:check
6. Run: task lint
7. Commit
```

### Pull Request
```bash
# Before creating PR
1. "Use code-reviewer agent for comprehensive review"
2. Address all critical and high-priority issues
3. Verify automated checks pass:
   - Tests: 100% pass rate
   - Coverage: ≥90%
   - Linter: Zero issues
   - Race detection: Clean
4. Create PR with review summary
```

### Bug Fixing
```bash
# When fixing bugs
1. "Use go-expert agent to fix [bug description]"
2. Agent provides TDD approach:
   - Write failing test first
   - Implement fix
   - Verify test passes
3. "Use code-reviewer agent to verify fix"
4. Check no regressions introduced
```

## Troubleshooting

### Agent Not Responding
**Issue**: Agent doesn't seem to activate

**Solutions**:
1. Use explicit invocation: "Use the [agent-name] agent to..."
2. Check agent frontmatter is valid YAML
3. Verify agent file is in `.claude/agents/`
4. Ensure description field is clear and specific

### Agent Missing Tools
**Issue**: Agent can't perform required actions

**Solution**: Update agent's `tools` field:
```yaml
tools: Read, Write, Edit, Bash, Grep, Glob
```

### Agent Output Too Generic
**Issue**: Agent provides vague recommendations

**Solution**: Be more specific in your request:
- Include file paths
- Specify what to check (security, performance, etc.)
- Provide context about the change

## Examples

### Example 1: Code Review After UI Changes
```
User: I just added a new UI package with colored output and progress bars.
      Use the code-reviewer agent to review it comprehensively.

Agent: [Runs automated verification]
       [Provides structured review]

       Found:
       - ✅ Tests: 27 passing (100%)
       - ✅ Coverage: 97.5%
       - ❌ Race conditions: 4 found

       Critical Issue:
       Race in Spinner tests - buffer accessed while goroutine writes

       Fix: Add sync.WaitGroup to wait for goroutine completion
       [Shows exact code fix]

       Estimated effort: 30 minutes
```

### Example 2: Implementing New Feature
```
User: I need to implement namespace support for ephemeral environments.
      Use the go-expert agent to help.

Agent: [Provides TDD approach]

       1. Write tests first:
       func TestNamespacePrefix(t *testing.T) {
           tests := []struct {
               namespace string
               resource  string
               want      string
           }{
               {"pr-123", "api", "pr-123-api"},
               {"", "api", "api"},
           }
           // ... test implementation
       }

       2. Implement function using pure functional approach
       3. Verify with Either monad for error handling
       4. Check coverage ≥90%
```

### Example 3: Creating Build Script
```
User: Create a Nushell script to automate the build and deploy process.
      Use the nushell-writer agent.

Agent: [Provides production-ready script]

       #!/usr/bin/env nu

       # build-deploy.nu - Automated build and deployment

       def main [
           environment: string,
           --dry-run (-n),
           --verbose (-v)
       ]: nothing -> record {
           # Type-safe, functional implementation
           # Comprehensive error handling
           # Streaming for efficiency
           # Tests included
       }
```

## Monitoring Agent Effectiveness

Track these metrics to ensure agents are helpful:

1. **Time Saved**: How much faster is development with agents?
2. **Issues Caught**: How many bugs found before production?
3. **Code Quality**: Coverage, lint scores, test pass rates
4. **Learning**: Are developers learning better patterns?

## Future Enhancements

Planned agent improvements:

- [ ] **test-writer**: Specialized in generating comprehensive tests
- [ ] **refactor-expert**: Focuses on code refactoring to functional patterns
- [ ] **security-auditor**: Deep security analysis
- [ ] **performance-optimizer**: Algorithmic and performance improvements
- [ ] **documentation-writer**: Generates godoc, READMEs, API docs

## Contributing

To add a new agent:

1. Create `.claude/agents/your-agent.md`
2. Add proper YAML frontmatter
3. Include clear usage examples
4. Document when to use it
5. Test with real scenarios
6. Update this document

## Support

For issues or questions:
- Check agent descriptions and examples
- Review this usage guide
- Test with simple scenarios first
- Provide specific context in requests
