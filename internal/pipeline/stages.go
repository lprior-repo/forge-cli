package pipeline

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/stack"
	"github.com/samber/lo"
)

// DetectStacks finds all stacks in the project
// Pure functional approach - no OOP
func DetectStacks(ctx context.Context, s State) E.Either[error, State] {
	stacks, err := stack.FindStacks(s.ProjectDir)
	if err != nil {
		return E.Left[State](fmt.Errorf("failed to detect stacks: %w", err))
	}

	s.Stacks = stacks
	return E.Right[error](s)
}

// ValidateStacks validates all stacks using lo.Filter
// Pure functional approach - no methods
func ValidateStacks(ctx context.Context, s State) E.Either[error, State] {
	// Use lo.Filter to find invalid stacks
	invalid := lo.Filter(s.Stacks, func(st *stack.Stack, _ int) bool {
		return stack.ValidateStack(st) != nil
	})

	if len(invalid) > 0 {
		// Map to error messages
		errors := lo.Map(invalid, func(st *stack.Stack, _ int) string {
			return fmt.Sprintf("%s: %v", st.Name, stack.ValidateStack(st))
		})
		return E.Left[State](fmt.Errorf("validation failed: %v", errors))
	}

	return E.Right[error](s)
}

// FilterStacksByRuntime creates a stage that filters stacks by runtime
func FilterStacksByRuntime(runtime string) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		s.Stacks = lo.Filter(s.Stacks, func(st *stack.Stack, _ int) bool {
			return st.Runtime == runtime
		})
		return E.Right[error](s)
	}
}

// SortStacksByDependencies is a no-op - Terraform handles dependency ordering
// This function exists for backward compatibility
func SortStacksByDependencies(ctx context.Context, s State) E.Either[error, State] {
	// Terraform automatically handles dependency ordering via resource dependencies
	// No need for manual topological sorting
	return E.Right[error](s)
}

// GroupStacksByDepth is a no-op - Terraform handles parallel execution
// This function exists for backward compatibility
func GroupStacksByDepth(ctx context.Context, s State) E.Either[error, State] {
	// Terraform automatically handles parallel execution of independent resources
	// No need for manual grouping
	return E.Right[error](s)
}

// MapStacks applies a transformation to all stacks
func MapStacks(transform func(*stack.Stack) *stack.Stack) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		s.Stacks = lo.Map(s.Stacks, func(st *stack.Stack, _ int) *stack.Stack {
			return transform(st)
		})
		return E.Right[error](s)
	}
}

// FilterStacks applies a predicate to filter stacks
func FilterStacks(predicate func(*stack.Stack) bool) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		s.Stacks = lo.Filter(s.Stacks, func(st *stack.Stack, _ int) bool {
			return predicate(st)
		})
		return E.Right[error](s)
	}
}

// ReduceStacks reduces stacks to a single value
func ReduceStacks[T any](initial T, reducer func(T, *stack.Stack, int) T) func(context.Context, State) (T, error) {
	return func(ctx context.Context, s State) (T, error) {
		result := lo.Reduce(s.Stacks, func(acc T, st *stack.Stack, idx int) T {
			return reducer(acc, st, idx)
		}, initial)
		return result, nil
	}
}
