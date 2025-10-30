package agent

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// GenerateFunction creates a Go function from a CodeSpec (PURE)
func GenerateFunction(spec CodeSpec) (string, error) {
	// Validate spec first
	if err := ValidateSpec(spec); err != nil {
		return "", fmt.Errorf("invalid spec: %w", err)
	}

	if spec.Type != TypeFunction {
		return "", fmt.Errorf("spec type must be 'function', got '%s'", spec.Type)
	}

	var parts []string

	// 1. Documentation comment
	if spec.Doc != "" {
		parts = append(parts, generateDocComment(spec.Doc, spec.Name))
	}

	// 2. Function signature
	sig := generateSignature(spec)
	parts = append(parts, sig+" {")

	// 3. Function body based on paradigm
	body := generateBody(spec)
	parts = append(parts, body)

	// 4. Closing brace
	parts = append(parts, "}")

	return strings.Join(parts, "\n"), nil
}

// GenerateStruct creates a Go struct from a CodeSpec (PURE)
func GenerateStruct(spec CodeSpec) (string, error) {
	if err := ValidateSpec(spec); err != nil {
		return "", fmt.Errorf("invalid spec: %w", err)
	}

	if spec.Type != TypeStruct {
		return "", fmt.Errorf("spec type must be 'struct', got '%s'", spec.Type)
	}

	var parts []string

	// Documentation
	if spec.Doc != "" {
		parts = append(parts, generateDocComment(spec.Doc, spec.Name))
	}

	// Struct declaration
	parts = append(parts, fmt.Sprintf("type %s struct {", spec.Name))

	// Fields
	for _, field := range spec.Inputs {
		fieldLine := fmt.Sprintf("\t%s %s", field.Name, field.Type)
		if field.Doc != "" {
			fieldLine += " // " + field.Doc
		}
		parts = append(parts, fieldLine)
	}

	parts = append(parts, "}")

	return strings.Join(parts, "\n"), nil
}

// GenerateInterface creates a Go interface from a CodeSpec (PURE)
func GenerateInterface(spec CodeSpec) (string, error) {
	if err := ValidateSpec(spec); err != nil {
		return "", fmt.Errorf("invalid spec: %w", err)
	}

	if spec.Type != TypeInterface {
		return "", fmt.Errorf("spec type must be 'interface', got '%s'", spec.Type)
	}

	var parts []string

	// Documentation
	if spec.Doc != "" {
		parts = append(parts, generateDocComment(spec.Doc, spec.Name))
	}

	// Interface declaration
	parts = append(parts, fmt.Sprintf("type %s interface {", spec.Name))

	// Methods (stored in Inputs for interfaces)
	for _, method := range spec.Inputs {
		parts = append(parts, fmt.Sprintf("\t%s", method.Type)) // Type contains full method signature
	}

	parts = append(parts, "}")

	return strings.Join(parts, "\n"), nil
}

// ValidateSpec checks if a CodeSpec is valid (PURE)
func ValidateSpec(spec CodeSpec) error {
	if spec.Name == "" {
		return fmt.Errorf("name is required")
	}

	if spec.Type == "" {
		return fmt.Errorf("type is required")
	}

	// Check visibility matches name convention
	if spec.Visibility == VisibilityExported && !isExported(spec.Name) {
		return fmt.Errorf("exported symbol '%s' must start with uppercase", spec.Name)
	}

	if spec.Visibility == VisibilityUnexported && isExported(spec.Name) {
		return fmt.Errorf("unexported symbol '%s' must start with lowercase", spec.Name)
	}

	// Function-specific validation
	if spec.Type == TypeFunction {
		if spec.Paradigm == "" {
			return fmt.Errorf("paradigm is required for functions")
		}

		// Calculations must not have dependencies (they're pure)
		if spec.Paradigm == ParadigmCalculation && len(spec.Dependencies) > 0 {
			return fmt.Errorf("calculations cannot have dependencies (must be pure)")
		}

		// Actions should have context.Context as first parameter
		if spec.Paradigm == ParadigmAction && !hasContextParam(spec.Inputs) {
			return fmt.Errorf("actions should accept context.Context as first parameter")
		}
	}

	return nil
}

// generateDocComment creates a documentation comment (PURE)
func generateDocComment(doc, name string) string {
	// Go convention: doc comment starts with the symbol name
	if !strings.HasPrefix(doc, name) {
		doc = name + " " + strings.ToLower(doc[:1]) + doc[1:]
	}

	lines := strings.Split(doc, "\n")
	commented := lo.Map(lines, func(line string, _ int) string {
		return "// " + line
	})

	return strings.Join(commented, "\n")
}

// generateSignature creates function signature (PURE)
func generateSignature(spec CodeSpec) string {
	// Parameters
	params := lo.Map(spec.Inputs, func(p Parameter, _ int) string {
		return fmt.Sprintf("%s %s", p.Name, p.Type)
	})

	paramsStr := strings.Join(params, ", ")

	// Return values
	var returnsStr string
	if len(spec.Outputs) == 0 {
		returnsStr = ""
	} else if len(spec.Outputs) == 1 {
		returnsStr = " " + spec.Outputs[0].Type
	} else {
		returns := lo.Map(spec.Outputs, func(p Parameter, _ int) string {
			if p.Name != "" {
				return fmt.Sprintf("%s %s", p.Name, p.Type)
			}
			return p.Type
		})
		returnsStr = " (" + strings.Join(returns, ", ") + ")"
	}

	return fmt.Sprintf("func %s(%s)%s", spec.Name, paramsStr, returnsStr)
}

// generateBody creates function body based on paradigm (PURE)
func generateBody(spec CodeSpec) string {
	var parts []string

	switch spec.Paradigm {
	case ParadigmData:
		return "\t// Data paradigm: use GenerateStruct instead"

	case ParadigmCalculation:
		parts = append(parts, "\t// CALCULATION (Pure Function - no side effects)")
		parts = append(parts, "")

		// Add validation guard clauses
		if len(spec.Validation.Custom) > 0 || len(spec.Validation.RequireNonNil) > 0 {
			parts = append(parts, "\t// Validation (guard clauses)")
			parts = append(parts, generateValidation(spec)...)
			parts = append(parts, "")
		}

		// Add TODO for implementation
		parts = append(parts, "\t// TODO: Implement pure calculation logic")
		parts = append(parts, "\t// Same inputs must always produce same outputs")
		parts = append(parts, "\t// No I/O, no time.Now(), no randomness, no global state")
		parts = append(parts, "")

		// Add placeholder return
		parts = append(parts, generatePlaceholderReturn(spec))

		// Add assertions before return
		if len(spec.Assertions) > 0 {
			parts = append(parts, "")
			parts = append(parts, "\t// Assertions (internal invariants)")
			parts = append(parts, generateAssertions(spec)...)
		}

	case ParadigmAction:
		parts = append(parts, "\t// ACTION (I/O Function - has side effects)")
		parts = append(parts, "")

		// Validation
		if len(spec.Validation.Custom) > 0 || len(spec.Validation.RequireNonNil) > 0 {
			parts = append(parts, "\t// Validation (guard clauses)")
			parts = append(parts, generateValidation(spec)...)
			parts = append(parts, "")
		}

		// TODO for implementation
		parts = append(parts, "\t// TODO: Implement I/O action")
		parts = append(parts, "\t// Call pure calculations for business logic")
		parts = append(parts, "\t// Perform I/O at the edges")
		parts = append(parts, "")

		// Placeholder return
		parts = append(parts, generatePlaceholderReturn(spec))
	}

	return strings.Join(parts, "\n")
}

// generateValidation creates guard clause code (PURE)
func generateValidation(spec CodeSpec) []string {
	var lines []string

	// NonNil checks
	for _, param := range spec.Validation.RequireNonNil {
		lines = append(lines, fmt.Sprintf(
			"\tif %s == nil {\n\t\treturn %s, fmt.Errorf(\"%s cannot be nil\")\n\t}",
			param, zeroValues(spec.Outputs), param,
		))
	}

	// NonEmpty checks
	for _, param := range spec.Validation.RequireNonEmpty {
		lines = append(lines, fmt.Sprintf(
			"\tif len(%s) == 0 {\n\t\treturn %s, fmt.Errorf(\"%s cannot be empty\")\n\t}",
			param, zeroValues(spec.Outputs), param,
		))
	}

	// Positive checks
	for _, param := range spec.Validation.RequirePositive {
		lines = append(lines, fmt.Sprintf(
			"\tif %s <= 0 {\n\t\treturn %s, fmt.Errorf(\"%s must be positive\")\n\t}",
			param, zeroValues(spec.Outputs), param,
		))
	}

	// Custom guards
	for _, guard := range spec.Validation.Custom {
		lines = append(lines, fmt.Sprintf(
			"\tif !(%s) {\n\t\treturn %s, fmt.Errorf(\"%s\")\n\t}",
			guard.Condition, zeroValues(spec.Outputs), guard.ErrorMsg,
		))
	}

	return lines
}

// generateAssertions creates assertion code (PURE)
func generateAssertions(spec CodeSpec) []string {
	return lo.Map(spec.Assertions, func(a Assertion, _ int) string {
		return fmt.Sprintf(
			"\tif !(%s) {\n\t\tpanic(\"BUG: %s\")\n\t}",
			a.Condition, a.Message,
		)
	})
}

// generatePlaceholderReturn creates a placeholder return statement (PURE)
func generatePlaceholderReturn(spec CodeSpec) string {
	if len(spec.Outputs) == 0 {
		return ""
	}

	values := zeroValues(spec.Outputs)
	return "\treturn " + values
}

// zeroValues generates zero values for return parameters (PURE)
func zeroValues(outputs []Parameter) string {
	if len(outputs) == 0 {
		return ""
	}

	zeros := lo.Map(outputs, func(p Parameter, _ int) string {
		return zeroValue(p.Type)
	})

	return strings.Join(zeros, ", ")
}

// zeroValue returns the zero value for a Go type (PURE)
func zeroValue(typ string) string {
	switch {
	case typ == "error":
		return "nil"
	case typ == "string":
		return `""`
	case strings.HasPrefix(typ, "*"):
		return "nil"
	case strings.HasPrefix(typ, "[]"):
		return "nil"
	case strings.HasPrefix(typ, "map["):
		return "nil"
	case typ == "bool":
		return "false"
	case strings.Contains(typ, "int") || strings.Contains(typ, "float"):
		return "0"
	default:
		// Assume struct, use zero value syntax
		return typ + "{}"
	}
}

// hasContextParam checks if inputs include context.Context (PURE)
func hasContextParam(inputs []Parameter) bool {
	if len(inputs) == 0 {
		return false
	}
	return inputs[0].Type == "context.Context"
}

// isExported checks if a name is exported (PURE)
func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Exported names start with uppercase letter
	r := rune(name[0])
	return r >= 'A' && r <= 'Z'
}
