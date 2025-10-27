package stack

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectorFindStacks(t *testing.T) {
	// Use testdata from project root
	testdataDir, err := filepath.Abs("../../testdata")
	require.NoError(t, err)

	tests := []struct {
		name      string
		projectDir string
		wantCount int
		wantNames []string
	}{
		{
			name:       "basic project",
			projectDir: filepath.Join(testdataDir, "basic"),
			wantCount:  1,
			wantNames:  []string{"api"},
		},
		{
			name:       "multi-function project",
			projectDir: filepath.Join(testdataDir, "multi-function"),
			wantCount:  3,
			wantNames:  []string{"shared", "api", "worker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewDetector(tt.projectDir)
			stacks, err := detector.FindStacks()

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(stacks))

			// Verify stack names
			names := make([]string, len(stacks))
			for i, s := range stacks {
				names[i] = s.Name
			}

			for _, wantName := range tt.wantNames {
				assert.Contains(t, names, wantName)
			}
		})
	}
}

func TestStackValidate(t *testing.T) {
	tests := []struct {
		name    string
		stack   *Stack
		wantErr bool
	}{
		{
			name: "valid Go stack",
			stack: &Stack{
				Name:    "test",
				Runtime: "go1.x",
			},
			wantErr: false,
		},
		{
			name: "valid Python stack",
			stack: &Stack{
				Name:    "test",
				Runtime: "python3.11",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			stack: &Stack{
				Runtime: "go1.x",
			},
			wantErr: true,
		},
		{
			name: "missing runtime",
			stack: &Stack{
				Name: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid runtime",
			stack: &Stack{
				Name:    "test",
				Runtime: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stack.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetBuildTarget(t *testing.T) {
	tests := []struct {
		name    string
		runtime string
		want    string
	}{
		{
			name:    "Go runtime",
			runtime: "go1.x",
			want:    "bootstrap",
		},
		{
			name:    "Python runtime",
			runtime: "python3.11",
			want:    "lambda.zip",
		},
		{
			name:    "Node runtime",
			runtime: "nodejs20.x",
			want:    "lambda.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stack{Runtime: tt.runtime}
			got := s.GetBuildTarget()
			assert.Equal(t, tt.want, got)
		})
	}
}
