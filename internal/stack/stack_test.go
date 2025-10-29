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
		name       string
		projectDir string
		wantCount  int
		wantNames  []string
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
			stacks, err := FindStacks(tt.projectDir)

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

func TestDetectorDependencyResolution(t *testing.T) {
	testdataDir, err := filepath.Abs("../../testdata")
	require.NoError(t, err)

	t.Run("resolves relative dependencies", func(t *testing.T) {
		stacks, err := FindStacks(filepath.Join(testdataDir, "multi-function"))
		require.NoError(t, err)

		// Find the api stack which has a dependency on shared
		var apiStack *Stack
		for _, s := range stacks {
			if s.Name == "api" {
				apiStack = s
				break
			}
		}
		require.NotNil(t, apiStack, "api stack should exist")

		// Verify dependencies are parsed
		require.Len(t, apiStack.Dependencies, 1, "api should have 1 dependency")
		assert.Equal(t, "shared", apiStack.Dependencies[0], "api should depend on shared")
	})

	t.Run("worker stack has correct dependency", func(t *testing.T) {
		stacks, err := FindStacks(filepath.Join(testdataDir, "multi-function"))
		require.NoError(t, err)

		// Find the worker stack which has a dependency on shared
		var workerStack *Stack
		for _, s := range stacks {
			if s.Name == "worker" {
				workerStack = s
				break
			}
		}
		require.NotNil(t, workerStack, "worker stack should exist")

		// Verify dependencies are parsed - worker depends on shared
		require.Len(t, workerStack.Dependencies, 1, "worker should have 1 dependency")
		assert.Equal(t, "shared", workerStack.Dependencies[0], "worker should depend on shared")
	})

	t.Run("stack with no dependencies", func(t *testing.T) {
		stacks, err := FindStacks(filepath.Join(testdataDir, "multi-function"))
		require.NoError(t, err)

		// Find the shared stack which has no dependencies
		var sharedStack *Stack
		for _, s := range stacks {
			if s.Name == "shared" {
				sharedStack = s
				break
			}
		}
		require.NotNil(t, sharedStack, "shared stack should exist")

		// Verify no dependencies
		assert.Empty(t, sharedStack.Dependencies, "shared should have no dependencies")
	})
}

func TestStackValidate(t *testing.T) {
	tests := []struct {
		name    string
		stack   *Stack
		wantErr bool
		errMsg  string
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
			errMsg:  "stack name is required",
		},
		{
			name: "empty name",
			stack: &Stack{
				Name:    "",
				Runtime: "go1.x",
			},
			wantErr: true,
			errMsg:  "stack name is required",
		},
		{
			name: "missing runtime",
			stack: &Stack{
				Name: "test",
			},
			wantErr: true,
			errMsg:  "stack runtime is required",
		},
		{
			name: "empty runtime",
			stack: &Stack{
				Name:    "test",
				Runtime: "",
			},
			wantErr: true,
			errMsg:  "stack runtime is required",
		},
		{
			name: "invalid runtime",
			stack: &Stack{
				Name:    "test",
				Runtime: "invalid",
			},
			wantErr: true,
			errMsg:  "unsupported runtime: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStack(tt.stack)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
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
			got := GetBuildTarget(s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNeedsBuild(t *testing.T) {
	tests := []struct {
		name    string
		runtime string
		want    bool
	}{
		{
			name:    "Go runtime needs build",
			runtime: "go1.x",
			want:    true,
		},
		{
			name:    "Python 3.11 needs build",
			runtime: "python3.11",
			want:    true,
		},
		{
			name:    "Python 3.12 needs build",
			runtime: "python3.12",
			want:    true,
		},
		{
			name:    "Python 3.13 needs build",
			runtime: "python3.13",
			want:    true,
		},
		{
			name:    "Node.js 20.x needs build",
			runtime: "nodejs20.x",
			want:    true,
		},
		{
			name:    "Node.js 18.x needs build",
			runtime: "nodejs18.x",
			want:    true,
		},
		{
			name:    "provided.al2 does not need build",
			runtime: "provided.al2",
			want:    false,
		},
		{
			name:    "provided.al2023 does not need build",
			runtime: "provided.al2023",
			want:    false,
		},
		{
			name:    "unknown runtime does not need build",
			runtime: "unknown",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stack{Runtime: tt.runtime}
			got := NeedsBuild(s)
			assert.Equal(t, tt.want, got, "Runtime %s should NeedsBuild=%v", tt.runtime, tt.want)
		})
	}
}

func TestIsValidRuntime(t *testing.T) {
	t.Run("valid runtimes", func(t *testing.T) {
		validRuntimes := []string{
			"go1.x",
			"python3.11",
			"python3.12",
			"python3.13",
			"nodejs20.x",
			"nodejs18.x",
			"provided.al2",
			"provided.al2023",
		}

		for _, runtime := range validRuntimes {
			t.Run(runtime, func(t *testing.T) {
				assert.True(t, isValidRuntime(runtime), "Runtime %s should be valid", runtime)
			})
		}
	})

	t.Run("invalid runtimes", func(t *testing.T) {
		invalidRuntimes := []string{
			"",
			"go2.x",
			"python3.10",
			"nodejs16.x",
			"java11",
			"ruby3.2",
			"invalid",
			"PYTHON3.11", // case sensitive
		}

		for _, runtime := range invalidRuntimes {
			t.Run(runtime, func(t *testing.T) {
				assert.False(t, isValidRuntime(runtime), "Runtime %s should be invalid", runtime)
			})
		}
	})
}

func TestStackStructure(t *testing.T) {
	t.Run("stack with dependencies", func(t *testing.T) {
		stack := &Stack{
			Name:         "api",
			Path:         "stacks/api",
			AbsPath:      "/project/stacks/api",
			Dependencies: []string{"database", "shared"},
			Runtime:      "go1.x",
			Handler:      "stacks/api",
			Description:  "API Gateway handler",
		}

		assert.Equal(t, "api", stack.Name)
		assert.Len(t, stack.Dependencies, 2)
		assert.Contains(t, stack.Dependencies, "database")
		assert.Contains(t, stack.Dependencies, "shared")
	})

	t.Run("stack with no dependencies", func(t *testing.T) {
		stack := &Stack{
			Name:         "simple",
			Runtime:      "python3.11",
			Dependencies: []string{},
		}

		assert.Equal(t, "simple", stack.Name)
		assert.Empty(t, stack.Dependencies)
		assert.NoError(t, ValidateStack(stack))
	})

	t.Run("stack with nil dependencies", func(t *testing.T) {
		stack := &Stack{
			Name:         "simple",
			Runtime:      "python3.11",
			Dependencies: nil,
		}

		assert.Equal(t, "simple", stack.Name)
		assert.Nil(t, stack.Dependencies)
		assert.NoError(t, ValidateStack(stack))
	})
}
