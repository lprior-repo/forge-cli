# internal/agent

**Go Functional Writer Agent - AI-assisted code generation following functional programming best practices**

## Overview

The `agent` package provides an intelligent code generation and validation system that enforces functional programming principles, data-oriented design, and assertive programming patterns. It implements the complete "Coding Best Practices as Code v2.3" specification.

## Philosophy

This agent embodies three core paradigms:

### 1. Functional Thinking (Data, Calculations, Actions)

Every piece of code is categorized:
- **Data** - Inert, immutable structs with no methods
- **Calculations** - Pure functions (same input → same output)
- **Actions** - Impure I/O at the system edges

### 2. Data-Oriented Design

- Separate code from data
- Use simple, concrete structures
- Prefer functions over methods
- Design as data transformation pipelines

### 3. Assertive Programming

- **Validation** at boundaries (returns errors)
- **Assertions** in core logic (panics on bugs)
- Fail fast with guard clauses

## Architecture

```
┌─────────────────────────────────────────────┐
│           Agent (Orchestrator)              │
│  - Coordinates generation and validation    │
└─────────────────────────────────────────────┘
                    ↓
    ┌───────────────┴───────────────┐
    ↓                               ↓
┌────────────────┐          ┌──────────────┐
│   Generator    │          │  Validator   │
│  (Pure Core)   │          │ (Pure Core)  │
└────────────────┘          └──────────────┘
    ↓                               ↓
┌────────────────┐          ┌──────────────┐
│  CodeWriter    │          │  Analyzer    │
│ (Action Shell) │          │(Action Shell)│
└────────────────┘          └──────────────┘
```

## Core Types

### Data Structures (Immutable)

```go
// CodeSpec defines what code to generate (PURE DATA)
type CodeSpec struct {
    Type        CodeType              // Function, Struct, Package
    Name        string                // Symbol name
    Package     string                // Package name
    Inputs      []Parameter           // Function parameters
    Outputs     []Parameter           // Return values
    Paradigm    Paradigm              // Data, Calculation, or Action
    Validation  ValidationRules       // Guard clauses
    Assertions  []Assertion           // Internal invariants
    Dependencies []Dependency         // Injected dependencies
    Doc         string                // Documentation
}

// CodeType represents the kind of code
type CodeType string
const (
    TypeFunction   CodeType = "function"
    TypeStruct     CodeType = "struct"
    TypeInterface  CodeType = "interface"
    TypePackage    CodeType = "package"
)

// Paradigm categorizes code as Data, Calculation, or Action
type Paradigm string
const (
    ParadigmData        Paradigm = "data"        // Inert data
    ParadigmCalculation Paradigm = "calculation" // Pure function
    ParadigmAction      Paradigm = "action"      // I/O function
)

// Parameter represents function parameter or struct field
type Parameter struct {
    Name    string
    Type    string
    IsError bool // Special handling for error returns
}

// ValidationRules define boundary checks
type ValidationRules struct {
    RequireNonNil   []string // Parameter names
    RequireNonEmpty []string // String/slice parameter names
    RequirePositive []string // Numeric parameter names
    Custom          []GuardClause
}

// GuardClause is a custom validation
type GuardClause struct {
    Condition string // Go expression
    ErrorMsg  string
}

// Assertion defines internal invariant check
type Assertion struct {
    Condition string // Go expression that must be true
    Message   string // Panic message if false
}
```

### Generators (Pure Functions)

```go
// GenerateFunction creates Go function code from spec (PURE)
func GenerateFunction(spec CodeSpec) (string, error) {
    // 1. Validate spec
    if err := ValidateSpec(spec); err != nil {
        return "", err
    }

    // 2. Generate signature
    sig := generateSignature(spec)

    // 3. Generate body based on paradigm
    var body string
    switch spec.Paradigm {
    case ParadigmData:
        return "", errors.New("data paradigm should use GenerateStruct")
    case ParadigmCalculation:
        body = generateCalculationBody(spec)
    case ParadigmAction:
        body = generateActionBody(spec)
    }

    // 4. Compose
    return fmt.Sprintf("%s\n%s\n%s\n%s",
        generateDocComment(spec.Doc),
        sig,
        body,
        "}"), nil
}

// generateCalculationBody creates pure function body (PURE)
func generateCalculationBody(spec CodeSpec) string {
    var parts []string

    // Add validation guard clauses
    for _, guard := range spec.Validation.Custom {
        parts = append(parts, fmt.Sprintf(
            "\tif !(%s) {\n\t\treturn %s, fmt.Errorf(\"%s\")\n\t}",
            guard.Condition, zeroValues(spec.Outputs), guard.ErrorMsg,
        ))
    }

    // Add assertions
    for _, assertion := range spec.Assertions {
        parts = append(parts, fmt.Sprintf(
            "\tif !(%s) {\n\t\tpanic(\"%s\")\n\t}",
            assertion.Condition, assertion.Message,
        ))
    }

    // Placeholder for actual logic
    parts = append(parts, "\t// TODO: Implement pure calculation logic here")

    return strings.Join(parts, "\n")
}

// generateActionBody creates I/O function body (PURE)
func generateActionBody(spec CodeSpec) string {
    var parts []string

    // Actions should take context as first parameter
    if !hasContextParam(spec.Inputs) {
        parts = append(parts, "\t// WARNING: Actions should accept context.Context")
    }

    // Add validation
    for _, guard := range spec.Validation.Custom {
        parts = append(parts, fmt.Sprintf(
            "\tif !(%s) {\n\t\treturn %s, fmt.Errorf(\"%s\")\n\t}",
            guard.Condition, zeroValues(spec.Outputs), guard.ErrorMsg,
        ))
    }

    // Placeholder
    parts = append(parts, "\t// TODO: Implement I/O action here")

    return strings.Join(parts, "\n")
}
```

### Validators (Pure Functions)

```go
// ValidationResult represents validation outcome (PURE DATA)
type ValidationResult struct {
    Valid      bool
    Violations []Violation
}

// Violation describes a rule violation (PURE DATA)
type Violation struct {
    Rule     string
    Severity Severity
    Message  string
    Location CodeLocation
}

// Severity levels
type Severity string
const (
    SeverityError   Severity = "error"
    SeverityWarning Severity = "warning"
    SeverityInfo    Severity = "info"
)

// ValidateCode checks code against best practices (PURE)
func ValidateCode(code string, rules Rules) ValidationResult {
    violations := []Violation{}

    // Parse code into AST
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, "", code, parser.ParseComments)
    if err != nil {
        return ValidationResult{
            Valid: false,
            Violations: []Violation{{
                Rule:     "syntax",
                Severity: SeverityError,
                Message:  err.Error(),
            }},
        }
    }

    // Check each rule
    violations = append(violations, checkPurityRules(node, rules)...)
    violations = append(violations, checkNamingRules(node, rules)...)
    violations = append(violations, checkComplexityRules(node, rules)...)
    violations = append(violations, checkErrorHandlingRules(node, rules)...)

    return ValidationResult{
        Valid:      len(filterBySeverity(violations, SeverityError)) == 0,
        Violations: violations,
    }
}

// checkPurityRules validates functional purity (PURE)
func checkPurityRules(node *ast.File, rules Rules) []Violation {
    violations := []Violation{}

    ast.Inspect(node, func(n ast.Node) bool {
        if fn, ok := n.(*ast.FuncDecl); ok {
            // Check for hidden I/O in calculations
            if isMarkedAsCalculation(fn) {
                if hasIOCalls(fn.Body) {
                    violations = append(violations, Violation{
                        Rule:     "predictable",
                        Severity: SeverityError,
                        Message:  "Calculation contains I/O operations",
                        Location: getLocation(fn),
                    })
                }
            }

            // Check for time.Now() in core logic
            if hasTimeNow(fn.Body) {
                violations = append(violations, Violation{
                    Rule:     "predictable",
                    Severity: SeverityWarning,
                    Message:  "Function uses time.Now() - inject clock dependency",
                    Location: getLocation(fn),
                })
            }
        }
        return true
    })

    return violations
}
```

## Usage

### Generate Pure Function

```go
import "github.com/lewis/forge/internal/agent"

// Specify what to generate
spec := agent.CodeSpec{
    Type:    agent.TypeFunction,
    Name:    "CalculateDiscount",
    Package: "pricing",
    Paradigm: agent.ParadigmCalculation,
    Inputs: []agent.Parameter{
        {Name: "price", Type: "decimal.Decimal"},
        {Name: "customerTier", Type: "string"},
    },
    Outputs: []agent.Parameter{
        {Name: "discountedPrice", Type: "decimal.Decimal"},
        {Name: "err", Type: "error", IsError: true},
    },
    Validation: agent.ValidationRules{
        Custom: []agent.GuardClause{
            {
                Condition: "price.GreaterThan(decimal.Zero)",
                ErrorMsg:  "price must be positive",
            },
            {
                Condition: `customerTier == "gold" || customerTier == "silver" || customerTier == "bronze"`,
                ErrorMsg:  "invalid customer tier",
            },
        },
    },
    Assertions: []agent.Assertion{
        {
            Condition: "discountedPrice.LessThanOrEqual(price)",
            Message:   "BUG: discounted price cannot exceed original price",
        },
    },
    Doc: "CalculateDiscount applies tier-based discounts to a price",
}

// Generate code
code, err := agent.GenerateFunction(spec)
if err != nil {
    log.Fatal(err)
}

fmt.Println(code)
```

**Generated output:**
```go
// CalculateDiscount applies tier-based discounts to a price
func CalculateDiscount(price decimal.Decimal, customerTier string) (decimal.Decimal, error) {
    // Validation (guard clauses)
    if !(price.GreaterThan(decimal.Zero)) {
        return decimal.Zero, fmt.Errorf("price must be positive")
    }
    if !(customerTier == "gold" || customerTier == "silver" || customerTier == "bronze") {
        return decimal.Zero, fmt.Errorf("invalid customer tier")
    }

    // TODO: Implement pure calculation logic here
    var discountedPrice decimal.Decimal

    // Assertion (internal invariant)
    if !(discountedPrice.LessThanOrEqual(price)) {
        panic("BUG: discounted price cannot exceed original price")
    }

    return discountedPrice, nil
}
```

### Generate Data Structure

```go
spec := agent.CodeSpec{
    Type:    agent.TypeStruct,
    Name:    "Order",
    Package: "domain",
    Paradigm: agent.ParadigmData,
    Fields: []agent.Parameter{
        {Name: "ID", Type: "string"},
        {Name: "CustomerID", Type: "string"},
        {Name: "Items", Type: "[]OrderItem"},
        {Name: "Total", Type: "decimal.Decimal"},
        {Name: "Status", Type: "OrderStatus"},
        {Name: "CreatedAt", Type: "time.Time"},
    },
    Doc: "Order represents a customer order (immutable data)",
}

code, _ := agent.GenerateStruct(spec)
```

**Generated output:**
```go
// Order represents a customer order (immutable data)
type Order struct {
    ID         string
    CustomerID string
    Items      []OrderItem
    Total      decimal.Decimal
    Status     OrderStatus
    CreatedAt  time.Time
}
```

### Validate Existing Code

```go
code := `
func ProcessPayment(orderID string) error {
    // BAD: Uses time.Now() directly
    timestamp := time.Now()

    // BAD: No validation
    result := chargeCard(orderID, timestamp)

    // BAD: Ignores error
    _ = result

    return nil
}
`

rules := agent.DefaultRules()
result := agent.ValidateCode(code, rules)

if !result.Valid {
    for _, v := range result.Violations {
        fmt.Printf("[%s] %s: %s\n", v.Severity, v.Rule, v.Message)
    }
}
```

**Output:**
```
[error] predictable: Function uses time.Now() - inject clock dependency
[error] errors_explicit: Error return value is ignored
[error] errors_fail_fast: Missing input validation
```

## Agent Modes

### 1. Generator Mode

Generate new code from specifications:
```go
agent := agent.New(agent.ModeGenerate)
code, err := agent.Generate(spec)
```

### 2. Validator Mode

Validate existing code:
```go
agent := agent.New(agent.ModeValidate)
result := agent.Validate(existingCode)
```

### 3. Refactor Mode

Suggest improvements to existing code:
```go
agent := agent.New(agent.ModeRefactor)
suggestions := agent.Analyze(existingCode)
for _, s := range suggestions {
    fmt.Printf("- %s\n", s.Description)
    fmt.Printf("  Before: %s\n", s.Before)
    fmt.Printf("  After:  %s\n", s.After)
}
```

## Rules Engine

The agent enforces 30+ rules from the specification:

### Philosophy Rules
- Composable
- Predictable (Purity)
- Idiomatic
- Domain-Centric
- Simplicity (KISS, DRY, YAGNI)

### Structure Rules
- Clear naming conventions
- Package cohesion
- Minimal API surface
- Standard project layout

### Safety Rules
- Explicit error handling
- Fail-fast validation
- Resource cleanup
- HTTP client management

### Security Rules
- Input validation
- Secrets management

## Files

- **`types.go`** - Core data structures (CodeSpec, ValidationResult, etc.)
- **`generator.go`** - Pure code generation functions
- **`validator.go`** - Pure validation functions
- **`analyzer.go`** - AST analysis utilities
- **`rules.go`** - Rule definitions and default rulesets
- **`agent.go`** - Main orchestrator (Action layer)
- **`writer.go`** - File I/O actions

## Testing

```go
func TestGeneratePureFunction(t *testing.T) {
    spec := agent.CodeSpec{
        Type:     agent.TypeFunction,
        Name:     "Add",
        Paradigm: agent.ParadigmCalculation,
        Inputs: []agent.Parameter{
            {Name: "a", Type: "int"},
            {Name: "b", Type: "int"},
        },
        Outputs: []agent.Parameter{
            {Name: "sum", Type: "int"},
        },
    }

    code, err := agent.GenerateFunction(spec)

    assert.NoError(t, err)
    assert.Contains(t, code, "func Add(a int, b int) int")
    assert.NotContains(t, code, "time.Now()") // No hidden I/O
}
```

## Design Principles

1. **Self-documenting** - Generated code explains its paradigm and rules
2. **Fail-safe** - Invalid specs are rejected at generation time
3. **Composable** - Combine generators for complex code
4. **Pure core** - All generation logic is deterministic
5. **Extensible** - Add custom rules and generators

## Future Enhancements

- [ ] AI-powered code completion
- [ ] Automatic refactoring suggestions
- [ ] Integration with golangci-lint
- [ ] IDE plugin (VS Code, GoLand)
- [ ] GitHub Actions integration
- [ ] Mutation testing for generated code
- [ ] Benchmark generation
- [ ] Property-based test generation
