package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopologicalSort(t *testing.T) {
	tests := []struct {
		name    string
		stacks  []*Stack
		want    []string
		wantErr bool
	}{
		{
			name: "no dependencies",
			stacks: []*Stack{
				{Name: "a", Path: "a"},
				{Name: "b", Path: "b"},
			},
			want:    []string{"a", "b"},
			wantErr: false,
		},
		{
			name: "linear dependency",
			stacks: []*Stack{
				{Name: "b", Path: "b", Dependencies: []string{"a"}},
				{Name: "a", Path: "a"},
			},
			want:    []string{"a", "b"},
			wantErr: false,
		},
		{
			name: "tree dependency",
			stacks: []*Stack{
				{Name: "shared", Path: "shared"},
				{Name: "api", Path: "api", Dependencies: []string{"shared"}},
				{Name: "worker", Path: "worker", Dependencies: []string{"shared"}},
			},
			want:    []string{"shared", "api", "worker"},
			wantErr: false,
		},
		{
			name: "circular dependency",
			stacks: []*Stack{
				{Name: "a", Path: "a", Dependencies: []string{"b"}},
				{Name: "b", Path: "b", Dependencies: []string{"a"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := NewGraph(tt.stacks)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			sorted, err := graph.TopologicalSort()
			require.NoError(t, err)

			// Extract names
			names := make([]string, len(sorted))
			for i, s := range sorted {
				names[i] = s.Name
			}

			// Check that dependencies come before dependents
			if len(tt.want) > 0 {
				for _, stack := range sorted {
					for _, dep := range stack.Dependencies {
						depIdx := -1
						stackIdx := -1
						for i, s := range sorted {
							if s.Path == dep {
								depIdx = i
							}
							if s.Path == stack.Path {
								stackIdx = i
							}
						}
						if depIdx != -1 && stackIdx != -1 {
							assert.Less(t, depIdx, stackIdx,
								"dependency %s should come before %s", dep, stack.Name)
						}
					}
				}
			}
		})
	}
}

func TestGetParallel(t *testing.T) {
	stacks := []*Stack{
		{Name: "shared", Path: "shared"},
		{Name: "api", Path: "api", Dependencies: []string{"shared"}},
		{Name: "worker", Path: "worker", Dependencies: []string{"shared"}},
	}

	graph, err := NewGraph(stacks)
	require.NoError(t, err)

	groups, err := graph.GetParallel()
	require.NoError(t, err)

	// Should have 2 levels:
	// Level 0: shared
	// Level 1: api, worker (can run in parallel)
	assert.Equal(t, 2, len(groups))
	assert.Equal(t, 1, len(groups[0])) // shared
	assert.Equal(t, 2, len(groups[1])) // api and worker
}
