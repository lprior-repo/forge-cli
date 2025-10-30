// Package agent provides AI-assisted Go code generation following functional
// programming best practices. It implements Data/Calculations/Actions separation,
// data-oriented design, and assertive programming patterns.
package agent

import "go/token"

// CodeType represents the kind of code to generate
type CodeType string

const (
	TypeFunction  CodeType = "function"
	TypeStruct    CodeType = "struct"
	TypeInterface CodeType = "interface"
	TypePackage   CodeType = "package"
)

// Paradigm categorizes code according to functional thinking
type Paradigm string

const (
	ParadigmData        Paradigm = "data"        // Inert, immutable data
	ParadigmCalculation Paradigm = "calculation" // Pure function (no side effects)
	ParadigmAction      Paradigm = "action"      // I/O function (side effects)
)

// CodeSpec defines what code to generate (PURE DATA - immutable)
type CodeSpec struct {
	Type         CodeType          // What kind of code
	Name         string            // Symbol name
	Package      string            // Package name
	Inputs       []Parameter       // Function parameters or struct fields
	Outputs      []Parameter       // Return values (functions only)
	Paradigm     Paradigm          // Data, Calculation, or Action
	Validation   ValidationRules   // Guard clauses at boundaries
	Assertions   []Assertion       // Internal invariants
	Dependencies []Dependency      // Injected dependencies
	Doc          string            // Documentation comment
	Visibility   Visibility        // Exported or unexported
}

// Parameter represents a function parameter or struct field
type Parameter struct {
	Name    string
	Type    string
	IsError bool   // Special handling for error returns
	Doc     string // Field/parameter documentation
}

// Visibility determines if a symbol is exported
type Visibility string

const (
	VisibilityExported   Visibility = "exported"
	VisibilityUnexported Visibility = "unexported"
)

// ValidationRules define boundary checks (guard clauses)
type ValidationRules struct {
	RequireNonNil   []string      // Parameter names that must not be nil
	RequireNonEmpty []string      // String/slice parameters that must not be empty
	RequirePositive []string      // Numeric parameters that must be > 0
	Custom          []GuardClause // Custom validation logic
}

// GuardClause is a custom validation check
type GuardClause struct {
	Condition string // Go boolean expression
	ErrorMsg  string // Error message if condition is false
}

// Assertion defines an internal invariant check
type Assertion struct {
	Condition string // Go boolean expression that must be true
	Message   string // Panic message if assertion fails
}

// Dependency represents an injected dependency (for Actions)
type Dependency struct {
	Name string // Parameter name
	Type string // Interface type
	Doc  string // Why this dependency exists
}

// ValidationResult represents code validation outcome (PURE DATA)
type ValidationResult struct {
	Valid      bool
	Violations []Violation
}

// Violation describes a rule violation (PURE DATA)
type Violation struct {
	Rule     string       // Rule name from specification
	Severity Severity     // Error, warning, or info
	Message  string       // Human-readable description
	Location CodeLocation // Where the violation occurs
	Fix      string       // Optional: suggested fix
}

// Severity levels for violations
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// CodeLocation pinpoints a violation in source code
type CodeLocation struct {
	File   string
	Line   int
	Column int
}

// Rules contains all validation rules (PURE DATA)
type Rules struct {
	Philosophy RuleSet
	Structure  RuleSet
	Safety     RuleSet
	Idioms     RuleSet
	Security   RuleSet
}

// RuleSet is a collection of related rules
type RuleSet struct {
	Name        string
	Description string
	Rules       []Rule
}

// Rule defines a single validation rule
type Rule struct {
	Name        string
	Description string
	Severity    Severity
	Checker     RuleChecker // Function that checks the rule
}

// RuleChecker is a function that validates a rule (PURE)
type RuleChecker func(code string) []Violation

// GeneratedCode represents the output of code generation
type GeneratedCode struct {
	Package string            // Package declaration
	Imports []string          // Import statements
	Code    string            // Generated code
	Tests   string            // Generated tests (optional)
	Metrics CodeMetrics       // Complexity metrics
}

// CodeMetrics tracks code quality metrics
type CodeMetrics struct {
	Lines              int
	CyclomaticComplexity int
	CognitiveComplexity  int
	ParameterCount     int
	NestingDepth       int
}

// RefactorSuggestion represents an improvement suggestion
type RefactorSuggestion struct {
	Rule        string // Which rule triggered this
	Description string // What to improve
	Before      string // Current code
	After       string // Suggested code
	Location    CodeLocation
}

// AgentMode determines the agent's behavior
type AgentMode string

const (
	ModeGenerate AgentMode = "generate" // Generate new code
	ModeValidate AgentMode = "validate" // Validate existing code
	ModeRefactor AgentMode = "refactor" // Suggest improvements
)

// Agent orchestrates code generation and validation (ACTION layer)
type Agent struct {
	Mode  AgentMode
	Rules Rules
}

// Position wraps go/token.Position for consistency
type Position struct {
	Filename string
	Offset   int
	Line     int
	Column   int
}

// FromTokenPosition converts go/token.Position to our Position
func FromTokenPosition(pos token.Position) Position {
	return Position{
		Filename: pos.Filename,
		Offset:   pos.Offset,
		Line:     pos.Line,
		Column:   pos.Column,
	}
}
