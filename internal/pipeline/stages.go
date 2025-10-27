package pipeline

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/stack"
	"github.com/samber/lo"
)

// DetectStacks finds all stacks in the project
func DetectStacks(ctx context.Context, s State) E.Either[error, State] {
	detector := stack.NewDetector(s.ProjectDir)
	stacks, err := detector.FindStacks()
	if err != nil {
		return E.Left[State](fmt.Errorf("failed to detect stacks: %w", err))
	}

	s.Stacks = stacks
	return E.Right[error](s)
}

// ValidateStacks validates all stacks using lo.Filter
func ValidateStacks(ctx context.Context, s State) E.Either[error, State] {
	// Use lo.Filter to find invalid stacks
	invalid := lo.Filter(s.Stacks, func(st *stack.Stack, _ int) bool {
		return st.Validate() != nil
	})

	if len(invalid) > 0 {
		// Map to error messages
		errors := lo.Map(invalid, func(st *stack.Stack, _ int) string {
			return fmt.Sprintf("%s: %v", st.Name, st.Validate())
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

// SortStacksByDependencies topologically sorts stacks
func SortStacksByDependencies(ctx context.Context, s State) E.Either[error, State] {
	if len(s.Stacks) == 0 {
		return E.Right[error](s)
	}

	graph, err := stack.NewGraph(s.Stacks)
	if err != nil {
		return E.Left[State](fmt.Errorf("failed to build dependency graph: %w", err))
	}

	sorted, err := graph.TopologicalSort()
	if err != nil {
		return E.Left[State](fmt.Errorf("failed to sort stacks: %w", err))
	}

	s.Stacks = sorted
	return E.Right[error](s)
}

// GroupStacksByDepth groups stacks by dependency depth for parallel execution
func GroupStacksByDepth(ctx context.Context, s State) E.Either[error, State] {
	if len(s.Stacks) == 0 {
		return E.Right[error](s)
	}

	graph, err := stack.NewGraph(s.Stacks)
	if err != nil {
		return E.Left[State](fmt.Errorf("failed to build dependency graph: %w", err))
	}

	groups, err := graph.GetParallel()
	if err != nil {
		return E.Left[State](fmt.Errorf("failed to group stacks: %w", err))
	}

	// Flatten groups back to single list for now
	// In real implementation, we'd process each group in parallel
	s.Stacks = lo.Flatten(groups)
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
