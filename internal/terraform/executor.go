package terraform

import (
	"context"
)

// Function types for terraform operations
type InitFunc func(ctx context.Context, dir string, opts ...InitOption) error
type PlanFunc func(ctx context.Context, dir string, opts ...PlanOption) (bool, error)
type ApplyFunc func(ctx context.Context, dir string, opts ...ApplyOption) error
type DestroyFunc func(ctx context.Context, dir string, opts ...DestroyOption) error
type OutputFunc func(ctx context.Context, dir string) (map[string]interface{}, error)
type ValidateFunc func(ctx context.Context, dir string) error

// Executor is a collection of terraform operation functions
type Executor struct {
	Init     InitFunc
	Plan     PlanFunc
	Apply    ApplyFunc
	Destroy  DestroyFunc
	Output   OutputFunc
	Validate ValidateFunc
}

// NewExecutor creates a real terraform executor using terraform-exec
func NewExecutor(tfPath string) Executor {
	return Executor{
		Init:     makeInitFunc(tfPath),
		Plan:     makePlanFunc(tfPath),
		Apply:    makeApplyFunc(tfPath),
		Destroy:  makeDestroyFunc(tfPath),
		Output:   makeOutputFunc(tfPath),
		Validate: makeValidateFunc(tfPath),
	}
}

// NewMockExecutor creates a mock executor for testing
func NewMockExecutor() Executor {
	return Executor{
		Init: func(ctx context.Context, dir string, opts ...InitOption) error {
			return nil
		},
		Plan: func(ctx context.Context, dir string, opts ...PlanOption) (bool, error) {
			return true, nil // hasChanges = true
		},
		Apply: func(ctx context.Context, dir string, opts ...ApplyOption) error {
			return nil
		},
		Destroy: func(ctx context.Context, dir string, opts ...DestroyOption) error {
			return nil
		},
		Output: func(ctx context.Context, dir string) (map[string]interface{}, error) {
			return make(map[string]interface{}), nil
		},
		Validate: func(ctx context.Context, dir string) error {
			return nil
		},
	}
}

