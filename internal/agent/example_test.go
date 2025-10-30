package agent_test

import (
	"fmt"
	"testing"

	"github.com/lewis/forge/internal/agent"
	"github.com/stretchr/testify/assert"
)

// ExampleGenerateFunction_pureCalculation demonstrates generating a pure calculation
func ExampleGenerateFunction_pureCalculation() {
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
				Message:   "discounted price cannot exceed original price",
			},
		},
		Doc: "CalculateDiscount applies tier-based discounts to a price",
		Visibility: agent.VisibilityExported,
	}

	code, _ := agent.GenerateFunction(spec)
	fmt.Println(code)

	// Output:
	// // CalculateDiscount applies tier-based discounts to a price
	// func CalculateDiscount(price decimal.Decimal, customerTier string) (discountedPrice decimal.Decimal, err error) {
	// 	// CALCULATION (Pure Function - no side effects)
	//
	// 	// Validation (guard clauses)
	// 	if !(price.GreaterThan(decimal.Zero)) {
	// 		return decimal.Decimal{}, fmt.Errorf("price must be positive")
	// 	}
	// 	if !(customerTier == "gold" || customerTier == "silver" || customerTier == "bronze") {
	// 		return decimal.Decimal{}, fmt.Errorf("invalid customer tier")
	// 	}
	//
	// 	// TODO: Implement pure calculation logic
	// 	// Same inputs must always produce same outputs
	// 	// No I/O, no time.Now(), no randomness, no global state
	//
	// 	return decimal.Decimal{}, nil
	//
	// 	// Assertions (internal invariants)
	// 	if !(discountedPrice.LessThanOrEqual(price)) {
	// 		panic("BUG: discounted price cannot exceed original price")
	// 	}
	// }
}

// ExampleGenerateFunction_action demonstrates generating an I/O action
func ExampleGenerateFunction_action() {
	spec := agent.CodeSpec{
		Type:    agent.TypeFunction,
		Name:    "SaveOrder",
		Package: "repository",
		Paradigm: agent.ParadigmAction,
		Inputs: []agent.Parameter{
			{Name: "ctx", Type: "context.Context"},
			{Name: "order", Type: "*Order"},
		},
		Outputs: []agent.Parameter{
			{Name: "", Type: "error", IsError: true},
		},
		Validation: agent.ValidationRules{
			RequireNonNil: []string{"order"},
		},
		Doc: "SaveOrder persists an order to the database",
		Visibility: agent.VisibilityExported,
	}

	code, _ := agent.GenerateFunction(spec)
	fmt.Println(code)

	// Output:
	// // SaveOrder persists an order to the database
	// func SaveOrder(ctx context.Context, order *Order) error {
	// 	// ACTION (I/O Function - has side effects)
	//
	// 	// Validation (guard clauses)
	// 	if order == nil {
	// 		return fmt.Errorf("order cannot be nil")
	// 	}
	//
	// 	// TODO: Implement I/O action
	// 	// Call pure calculations for business logic
	// 	// Perform I/O at the edges
	//
	// 	return nil
	// }
}

// ExampleGenerateStruct demonstrates generating a data structure
func ExampleGenerateStruct() {
	spec := agent.CodeSpec{
		Type:    agent.TypeStruct,
		Name:    "Order",
		Package: "domain",
		Paradigm: agent.ParadigmData,
		Inputs: []agent.Parameter{
			{Name: "ID", Type: "string", Doc: "Unique identifier"},
			{Name: "CustomerID", Type: "string"},
			{Name: "Items", Type: "[]OrderItem"},
			{Name: "Total", Type: "decimal.Decimal"},
			{Name: "Status", Type: "OrderStatus"},
			{Name: "CreatedAt", Type: "time.Time"},
		},
		Doc: "Order represents a customer order (immutable data)",
		Visibility: agent.VisibilityExported,
	}

	code, _ := agent.GenerateStruct(spec)
	fmt.Println(code)

	// Output:
	// // Order represents a customer order (immutable data)
	// type Order struct {
	// 	ID         string // Unique identifier
	// 	CustomerID string
	// 	Items      []OrderItem
	// 	Total      decimal.Decimal
	// 	Status     OrderStatus
	// 	CreatedAt  time.Time
	// }
}

// TestGenerateFunction_validation tests validation logic
func TestGenerateFunction_validation(t *testing.T) {
	t.Run("requires name", func(t *testing.T) {
		spec := agent.CodeSpec{
			Type:    agent.TypeFunction,
			Package: "test",
		}

		_, err := agent.GenerateFunction(spec)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("requires paradigm for functions", func(t *testing.T) {
		spec := agent.CodeSpec{
			Type:    agent.TypeFunction,
			Name:    "Test",
			Package: "test",
		}

		_, err := agent.GenerateFunction(spec)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "paradigm is required")
	})

	t.Run("calculations cannot have dependencies", func(t *testing.T) {
		spec := agent.CodeSpec{
			Type:     agent.TypeFunction,
			Name:     "Test",
			Package:  "test",
			Paradigm: agent.ParadigmCalculation,
			Dependencies: []agent.Dependency{
				{Name: "db", Type: "Database"},
			},
		}

		_, err := agent.GenerateFunction(spec)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "calculations cannot have dependencies")
	})

	t.Run("actions should have context parameter", func(t *testing.T) {
		spec := agent.CodeSpec{
			Type:     agent.TypeFunction,
			Name:     "Test",
			Package:  "test",
			Paradigm: agent.ParadigmAction,
			Inputs: []agent.Parameter{
				{Name: "data", Type: "string"},
			},
		}

		_, err := agent.GenerateFunction(spec)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context.Context")
	})
}

// TestGenerateFunction_guardClauses tests guard clause generation
func TestGenerateFunction_guardClauses(t *testing.T) {
	spec := agent.CodeSpec{
		Type:     agent.TypeFunction,
		Name:     "Divide",
		Package:  "math",
		Paradigm: agent.ParadigmCalculation,
		Inputs: []agent.Parameter{
			{Name: "a", Type: "int"},
			{Name: "b", Type: "int"},
		},
		Outputs: []agent.Parameter{
			{Name: "result", Type: "int"},
			{Name: "err", Type: "error"},
		},
		Validation: agent.ValidationRules{
			Custom: []agent.GuardClause{
				{
					Condition: "b != 0",
					ErrorMsg:  "division by zero",
				},
			},
		},
		Doc:        "Divide divides two integers",
		Visibility: agent.VisibilityExported,
	}

	code, err := agent.GenerateFunction(spec)

	assert.NoError(t, err)
	assert.Contains(t, code, "if !(b != 0)")
	assert.Contains(t, code, "division by zero")
}

// TestGenerateFunction_assertions tests assertion generation
func TestGenerateFunction_assertions(t *testing.T) {
	spec := agent.CodeSpec{
		Type:     agent.TypeFunction,
		Name:     "Square",
		Package:  "math",
		Paradigm: agent.ParadigmCalculation,
		Inputs: []agent.Parameter{
			{Name: "x", Type: "int"},
		},
		Outputs: []agent.Parameter{
			{Name: "result", Type: "int"},
		},
		Assertions: []agent.Assertion{
			{
				Condition: "result >= 0",
				Message:   "square cannot be negative",
			},
		},
		Doc:        "Square calculates x squared",
		Visibility: agent.VisibilityExported,
	}

	code, err := agent.GenerateFunction(spec)

	assert.NoError(t, err)
	assert.Contains(t, code, "if !(result >= 0)")
	assert.Contains(t, code, `panic("BUG: square cannot be negative")`)
}

// TestGenerateStruct tests struct generation
func TestGenerateStruct(t *testing.T) {
	spec := agent.CodeSpec{
		Type:    agent.TypeStruct,
		Name:    "User",
		Package: "domain",
		Paradigm: agent.ParadigmData,
		Inputs: []agent.Parameter{
			{Name: "ID", Type: "string"},
			{Name: "Email", Type: "string"},
			{Name: "CreatedAt", Type: "time.Time"},
		},
		Doc:        "User represents a system user",
		Visibility: agent.VisibilityExported,
	}

	code, err := agent.GenerateStruct(spec)

	assert.NoError(t, err)
	assert.Contains(t, code, "type User struct {")
	assert.Contains(t, code, "ID")
	assert.Contains(t, code, "string")
	assert.Contains(t, code, "Email")
	assert.Contains(t, code, "CreatedAt")
	assert.Contains(t, code, "time.Time")
}

// TestVisibility tests exported vs unexported naming
func TestVisibility(t *testing.T) {
	t.Run("exported function must start with uppercase", func(t *testing.T) {
		spec := agent.CodeSpec{
			Type:       agent.TypeFunction,
			Name:       "calculate", // lowercase
			Package:    "test",
			Paradigm:   agent.ParadigmCalculation,
			Visibility: agent.VisibilityExported, // mismatch!
		}

		_, err := agent.GenerateFunction(spec)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exported symbol")
	})

	t.Run("unexported function must start with lowercase", func(t *testing.T) {
		spec := agent.CodeSpec{
			Type:       agent.TypeFunction,
			Name:       "Calculate", // uppercase
			Package:    "test",
			Paradigm:   agent.ParadigmCalculation,
			Visibility: agent.VisibilityUnexported, // mismatch!
		}

		_, err := agent.GenerateFunction(spec)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexported symbol")
	})
}
