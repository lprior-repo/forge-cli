package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/stack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectStacks tests stack detection stage
func TestDetectStacks(t *testing.T) {
	t.Run("detects stacks in project", func(t *testing.T) {
		// Create test project
		tmpDir := t.TempDir()

		// Create stack directories
		stack1 := filepath.Join(tmpDir, "api")
		stack2 := filepath.Join(tmpDir, "worker")
		require.NoError(t, os.MkdirAll(stack1, 0755))
		require.NoError(t, os.MkdirAll(stack2, 0755))

		// Create stack.forge.hcl files
		stackConfig := `stack {
  name    = "test"
  runtime = "go1.x"
}`
		require.NoError(t, os.WriteFile(filepath.Join(stack1, "stack.forge.hcl"), []byte(stackConfig), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(stack2, "stack.forge.hcl"), []byte(stackConfig), 0644))

		state := State{ProjectDir: tmpDir}
		result := DetectStacks(context.Background(), state)

		assert.True(t, E.IsRight(result), "Should succeed")
		newState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Len(t, newState.Stacks, 2, "Should find 2 stacks")
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		state := State{ProjectDir: "/non/existent/path"}
		result := DetectStacks(context.Background(), state)

		assert.True(t, E.IsLeft(result), "Should fail")
	})
}

// TestValidateStacks tests stack validation stage
func TestValidateStacks(t *testing.T) {
	t.Run("validates all stacks successfully", func(t *testing.T) {
		stacks := []*stack.Stack{
			{Name: "api", Path: "api", AbsPath: "/tmp/api", Runtime: "go1.x"},
			{Name: "worker", Path: "worker", AbsPath: "/tmp/worker", Runtime: "python3.11"},
		}

		state := State{Stacks: stacks}
		result := ValidateStacks(context.Background(), state)

		assert.True(t, E.IsRight(result), "Should succeed with valid stacks")
	})

	t.Run("returns error for invalid stacks", func(t *testing.T) {
		stacks := []*stack.Stack{
			{Name: "", Path: "", Runtime: "go1.x"}, // Invalid: no name or path
			{Name: "worker", Path: "worker", AbsPath: "/tmp/worker", Runtime: "python3.11"},
		}

		state := State{Stacks: stacks}
		result := ValidateStacks(context.Background(), state)

		assert.True(t, E.IsLeft(result), "Should fail with invalid stacks")
	})
}

// TestFilterStacksByRuntime tests runtime filtering
func TestFilterStacksByRuntime(t *testing.T) {
	t.Run("filters stacks by runtime", func(t *testing.T) {
		stacks := []*stack.Stack{
			{Name: "api", Path: "/tmp/api", Runtime: "go1.x"},
			{Name: "worker", Path: "/tmp/worker", Runtime: "python3.11"},
			{Name: "lambda", Path: "/tmp/lambda", Runtime: "go1.x"},
		}

		state := State{Stacks: stacks}
		stage := FilterStacksByRuntime("go1.x")
		result := stage(context.Background(), state)

		assert.True(t, E.IsRight(result), "Should succeed")
		newState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Len(t, newState.Stacks, 2, "Should have 2 Go stacks")
		for _, st := range newState.Stacks {
			assert.Equal(t, "go1.x", st.Runtime)
		}
	})

	t.Run("returns empty when no stacks match", func(t *testing.T) {
		stacks := []*stack.Stack{
			{Name: "api", Path: "/tmp/api", Runtime: "go1.x"},
		}

		state := State{Stacks: stacks}
		stage := FilterStacksByRuntime("python3.11")
		result := stage(context.Background(), state)

		assert.True(t, E.IsRight(result), "Should succeed")
		newState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Empty(t, newState.Stacks, "Should have no stacks")
	})
}

// TestSortStacksByDependencies tests no-op pass-through (Terraform handles sorting)
func TestSortStacksByDependencies(t *testing.T) {
	t.Run("passes through stacks unchanged", func(t *testing.T) {
		stacks := []*stack.Stack{
			{Name: "frontend", Path: "frontend", AbsPath: "/tmp/frontend", Runtime: "nodejs20.x", Dependencies: []string{"api"}},
			{Name: "api", Path: "api", AbsPath: "/tmp/api", Runtime: "go1.x", Dependencies: []string{"database"}},
			{Name: "database", Path: "database", AbsPath: "/tmp/database", Runtime: "go1.x"},
		}

		state := State{Stacks: stacks}
		result := SortStacksByDependencies(context.Background(), state)

		assert.True(t, E.IsRight(result), "Should succeed")
		newState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		// Terraform handles dependency ordering, so we just pass through unchanged
		require.Len(t, newState.Stacks, 3)
		assert.Equal(t, "frontend", newState.Stacks[0].Name)
		assert.Equal(t, "api", newState.Stacks[1].Name)
		assert.Equal(t, "database", newState.Stacks[2].Name)
	})

	t.Run("handles empty stack list", func(t *testing.T) {
		state := State{Stacks: []*stack.Stack{}}
		result := SortStacksByDependencies(context.Background(), state)

		assert.True(t, E.IsRight(result), "Should succeed with empty list")
	})
}

// TestGroupStacksByDepth tests no-op pass-through (Terraform handles grouping)
func TestGroupStacksByDepth(t *testing.T) {
	t.Run("passes through stacks unchanged", func(t *testing.T) {
		stacks := []*stack.Stack{
			{Name: "database", Path: "database", AbsPath: "/tmp/database", Runtime: "go1.x"},
			{Name: "api", Path: "api", AbsPath: "/tmp/api", Runtime: "go1.x", Dependencies: []string{"database"}},
			{Name: "frontend", Path: "frontend", AbsPath: "/tmp/frontend", Runtime: "nodejs20.x", Dependencies: []string{"api"}},
		}

		state := State{Stacks: stacks}
		result := GroupStacksByDepth(context.Background(), state)

		assert.True(t, E.IsRight(result), "Should succeed")
		newState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		// Terraform handles parallel execution, so we just pass through unchanged
		assert.Len(t, newState.Stacks, 3)
	})

	t.Run("handles empty stack list", func(t *testing.T) {
		state := State{Stacks: []*stack.Stack{}}
		result := GroupStacksByDepth(context.Background(), state)

		assert.True(t, E.IsRight(result), "Should succeed with empty list")
	})
}

// TestMapStacks tests stack transformation
func TestMapStacks(t *testing.T) {
	t.Run("applies transformation to all stacks", func(t *testing.T) {
		stacks := []*stack.Stack{
			{Name: "api", Path: "/tmp/api", Runtime: "go1.x"},
			{Name: "worker", Path: "/tmp/worker", Runtime: "python3.11"},
		}

		state := State{Stacks: stacks}
		// Add a prefix to all stack names
		stage := MapStacks(func(st *stack.Stack) *stack.Stack {
			st.Name = "prefix-" + st.Name
			return st
		})
		result := stage(context.Background(), state)

		assert.True(t, E.IsRight(result), "Should succeed")
		newState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Len(t, newState.Stacks, 2)
		assert.Equal(t, "prefix-api", newState.Stacks[0].Name)
		assert.Equal(t, "prefix-worker", newState.Stacks[1].Name)
	})
}

// TestFilterStacks tests stack filtering with custom predicate
func TestFilterStacks(t *testing.T) {
	t.Run("filters stacks by custom predicate", func(t *testing.T) {
		stacks := []*stack.Stack{
			{Name: "api", Path: "/tmp/api", Runtime: "go1.x"},
			{Name: "worker", Path: "/tmp/worker", Runtime: "python3.11"},
			{Name: "lambda", Path: "/tmp/lambda", Runtime: "go1.x"},
		}

		state := State{Stacks: stacks}
		// Filter stacks that contain "api" in name
		stage := FilterStacks(func(st *stack.Stack) bool {
			return st.Name == "api" || st.Name == "worker"
		})
		result := stage(context.Background(), state)

		assert.True(t, E.IsRight(result), "Should succeed")
		newState := E.Fold(
			func(e error) State { return State{} },
			func(s State) State { return s },
		)(result)

		assert.Len(t, newState.Stacks, 2)
	})
}

// TestReduceStacks tests stack reduction
func TestReduceStacks(t *testing.T) {
	t.Run("reduces stacks to single value", func(t *testing.T) {
		stacks := []*stack.Stack{
			{Name: "api", Path: "/tmp/api", Runtime: "go1.x"},
			{Name: "worker", Path: "/tmp/worker", Runtime: "python3.11"},
			{Name: "lambda", Path: "/tmp/lambda", Runtime: "go1.x"},
		}

		state := State{Stacks: stacks}
		// Count total stacks
		reducer := ReduceStacks(0, func(acc int, st *stack.Stack, idx int) int {
			return acc + 1
		})
		count, err := reducer(context.Background(), state)

		assert.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("concatenates stack names", func(t *testing.T) {
		stacks := []*stack.Stack{
			{Name: "api", Path: "/tmp/api", Runtime: "go1.x"},
			{Name: "worker", Path: "/tmp/worker", Runtime: "python3.11"},
		}

		state := State{Stacks: stacks}
		reducer := ReduceStacks("", func(acc string, st *stack.Stack, idx int) string {
			if acc == "" {
				return st.Name
			}
			return acc + "," + st.Name
		})
		names, err := reducer(context.Background(), state)

		assert.NoError(t, err)
		assert.Equal(t, "api,worker", names)
	})
}
