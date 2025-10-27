package stack

import (
	"fmt"
	"strings"
)

// Graph manages stack dependencies and execution order
type Graph struct {
	stacks []*Stack
	adj    map[string][]string // adjacency list: stack path -> dependent paths
}

// NewGraph creates a new dependency graph from stacks
func NewGraph(stacks []*Stack) (*Graph, error) {
	g := &Graph{
		stacks: stacks,
		adj:    make(map[string][]string),
	}

	// Build adjacency list
	for _, stack := range stacks {
		g.adj[stack.Path] = stack.Dependencies
	}

	// Validate: check for cycles and missing dependencies
	if err := g.validate(); err != nil {
		return nil, err
	}

	return g, nil
}

// TopologicalSort returns stacks in deployment order
// Stacks with no dependencies come first, respecting the dependency chain
func (g *Graph) TopologicalSort() ([]*Stack, error) {
	// Kahn's algorithm for topological sorting
	inDegree := make(map[string]int)
	for _, stack := range g.stacks {
		inDegree[stack.Path] = 0
	}

	// Calculate in-degrees
	for _, deps := range g.adj {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	// Queue for stacks with no dependencies
	var queue []string
	for _, stack := range g.stacks {
		if inDegree[stack.Path] == 0 {
			queue = append(queue, stack.Path)
		}
	}

	var sorted []string
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		// For each dependent of current
		for _, stack := range g.stacks {
			if stack.Path == current {
				for _, dep := range stack.Dependencies {
					inDegree[dep]--
					if inDegree[dep] == 0 {
						queue = append(queue, dep)
					}
				}
			}
		}
	}

	// If sorted doesn't contain all stacks, there's a cycle
	if len(sorted) != len(g.stacks) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	// Reverse to get deployment order (dependencies first)
	for i := 0; i < len(sorted)/2; i++ {
		j := len(sorted) - 1 - i
		sorted[i], sorted[j] = sorted[j], sorted[i]
	}

	// Convert paths back to Stack objects
	result := make([]*Stack, 0, len(sorted))
	stackMap := make(map[string]*Stack)
	for _, stack := range g.stacks {
		stackMap[stack.Path] = stack
	}
	for _, path := range sorted {
		result = append(result, stackMap[path])
	}

	return result, nil
}

// GetParallel returns groups of stacks that can be deployed in parallel
// Each group contains stacks with no dependencies on each other
func (g *Graph) GetParallel() ([][]*Stack, error) {
	sorted, err := g.TopologicalSort()
	if err != nil {
		return nil, err
	}

	// Group by depth level
	levels := make(map[int][]*Stack)
	depth := make(map[string]int)

	// Calculate depth for each stack
	for _, stack := range sorted {
		maxDepth := 0
		for _, dep := range stack.Dependencies {
			if d, ok := depth[dep]; ok {
				if d+1 > maxDepth {
					maxDepth = d + 1
				}
			}
		}
		depth[stack.Path] = maxDepth
		levels[maxDepth] = append(levels[maxDepth], stack)
	}

	// Convert to ordered slice of groups
	maxLevel := 0
	for level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	result := make([][]*Stack, maxLevel+1)
	for level := 0; level <= maxLevel; level++ {
		result[level] = levels[level]
	}

	return result, nil
}

// validate checks for missing dependencies and cycles
func (g *Graph) validate() error {
	stackPaths := make(map[string]bool)
	for _, stack := range g.stacks {
		stackPaths[stack.Path] = true
	}

	// Check for missing dependencies
	for _, stack := range g.stacks {
		for _, dep := range stack.Dependencies {
			if !stackPaths[dep] {
				return fmt.Errorf("stack %s depends on non-existent stack %s", stack.Name, dep)
			}
		}
	}

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(string) bool
	hasCycle = func(path string) bool {
		visited[path] = true
		recStack[path] = true

		for _, dep := range g.adj[path] {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}

		recStack[path] = false
		return false
	}

	for _, stack := range g.stacks {
		if !visited[stack.Path] {
			if hasCycle(stack.Path) {
				cycle := g.findCycle()
				return fmt.Errorf("circular dependency detected: %s", cycle)
			}
		}
	}

	return nil
}

// findCycle attempts to find and format a cycle for error messages
func (g *Graph) findCycle() string {
	// Simple DFS to find any cycle
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycle []string

	var dfs func(string) bool
	dfs = func(path string) bool {
		visited[path] = true
		recStack[path] = true
		cycle = append(cycle, path)

		for _, dep := range g.adj[path] {
			if !visited[dep] {
				if dfs(dep) {
					return true
				}
			} else if recStack[dep] {
				// Found cycle
				idx := 0
				for i, p := range cycle {
					if p == dep {
						idx = i
						break
					}
				}
				cycle = cycle[idx:]
				cycle = append(cycle, dep)
				return true
			}
		}

		cycle = cycle[:len(cycle)-1]
		recStack[path] = false
		return false
	}

	for _, stack := range g.stacks {
		if !visited[stack.Path] {
			if dfs(stack.Path) {
				return strings.Join(cycle, " -> ")
			}
		}
	}

	return "unknown cycle"
}
