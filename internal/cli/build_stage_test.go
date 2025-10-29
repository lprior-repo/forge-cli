package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/pipeline"
	"github.com/lewis/forge/internal/stack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateBuildStageSuccessPaths tests successful build execution
func TestCreateBuildStageSuccessPaths(t *testing.T) {
	t.Run("successfully builds Go stack with bootstrap", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "api")
		err := os.MkdirAll(stackDir, 0755)
		require.NoError(t, err)

		// Create a minimal Go project
		mainGo := `package main

func main() {}
`
		err = os.WriteFile(filepath.Join(stackDir, "main.go"), []byte(mainGo), 0644)
		require.NoError(t, err)

		goMod := `module example.com/lambda

go 1.21
`
		err = os.WriteFile(filepath.Join(stackDir, "go.mod"), []byte(goMod), 0644)
		require.NoError(t, err)

		// Create build stage
		buildStage := createBuildStage()

		initialState := pipeline.State{
			ProjectDir: tmpDir,
			Stacks: []*stack.Stack{
				{
					Name:    "api",
					Path:    "api",
					AbsPath: stackDir,
					Runtime: "go1.x",
					Handler: ".",
				},
			},
		}

		result := buildStage(context.Background(), initialState)

		// Should succeed
		assert.True(t, E.IsRight(result), "Build should succeed")

		// Extract final state
		finalState := E.Fold(
			func(e error) pipeline.State { return pipeline.State{} },
			func(s pipeline.State) pipeline.State { return s },
		)(result)

		// Verify artifact was created
		assert.NotNil(t, finalState.Artifacts)
		assert.Contains(t, finalState.Artifacts, "api")
		artifact := finalState.Artifacts["api"]
		assert.NotEmpty(t, artifact.Path)
		assert.NotEmpty(t, artifact.Checksum)
		assert.Greater(t, artifact.Size, int64(0))
	})

	t.Run("successfully builds Python stack with lambda.zip", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "worker")
		err := os.MkdirAll(stackDir, 0755)
		require.NoError(t, err)

		// Create a minimal Python project
		handlerPy := `def handler(event, context):
    return {"statusCode": 200}
`
		err = os.WriteFile(filepath.Join(stackDir, "handler.py"), []byte(handlerPy), 0644)
		require.NoError(t, err)

		// Create build stage
		buildStage := createBuildStage()

		initialState := pipeline.State{
			ProjectDir: tmpDir,
			Stacks: []*stack.Stack{
				{
					Name:    "worker",
					Path:    "worker",
					AbsPath: stackDir,
					Runtime: "python3.12",
					Handler: "handler.handler",
				},
			},
		}

		result := buildStage(context.Background(), initialState)

		// Should succeed
		assert.True(t, E.IsRight(result), "Build should succeed")

		// Extract final state
		finalState := E.Fold(
			func(e error) pipeline.State { return pipeline.State{} },
			func(s pipeline.State) pipeline.State { return s },
		)(result)

		// Verify artifact was created
		assert.NotNil(t, finalState.Artifacts)
		assert.Contains(t, finalState.Artifacts, "worker")
		artifact := finalState.Artifacts["worker"]
		assert.NotEmpty(t, artifact.Path)
		assert.Contains(t, artifact.Path, "lambda.zip")
	})

	t.Run("successfully builds Node.js stack", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "frontend")
		err := os.MkdirAll(stackDir, 0755)
		require.NoError(t, err)

		// Create a minimal Node.js project
		indexJs := `exports.handler = async (event) => {
    return { statusCode: 200 };
};
`
		err = os.WriteFile(filepath.Join(stackDir, "index.js"), []byte(indexJs), 0644)
		require.NoError(t, err)

		packageJSON := `{
  "name": "lambda",
  "version": "1.0.0",
  "main": "index.js"
}
`
		err = os.WriteFile(filepath.Join(stackDir, "package.json"), []byte(packageJSON), 0644)
		require.NoError(t, err)

		// Create build stage
		buildStage := createBuildStage()

		initialState := pipeline.State{
			ProjectDir: tmpDir,
			Stacks: []*stack.Stack{
				{
					Name:    "frontend",
					Path:    "frontend",
					AbsPath: stackDir,
					Runtime: "nodejs20.x",
					Handler: "index.handler",
				},
			},
		}

		result := buildStage(context.Background(), initialState)

		// Should succeed
		assert.True(t, E.IsRight(result), "Build should succeed")

		// Extract final state
		finalState := E.Fold(
			func(e error) pipeline.State { return pipeline.State{} },
			func(s pipeline.State) pipeline.State { return s },
		)(result)

		// Verify artifact was created
		assert.NotNil(t, finalState.Artifacts)
		assert.Contains(t, finalState.Artifacts, "frontend")
	})

	t.Run("builds multiple stacks in sequence", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create two stacks
		apiDir := filepath.Join(tmpDir, "api")
		workerDir := filepath.Join(tmpDir, "worker")
		err := os.MkdirAll(apiDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(workerDir, 0755)
		require.NoError(t, err)

		// API stack (Go)
		err = os.WriteFile(filepath.Join(apiDir, "main.go"), []byte("package main\nfunc main() {}"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(apiDir, "go.mod"), []byte("module example.com/api\ngo 1.21"), 0644)
		require.NoError(t, err)

		// Worker stack (Python)
		err = os.WriteFile(filepath.Join(workerDir, "handler.py"), []byte("def handler(event, context): pass"), 0644)
		require.NoError(t, err)

		buildStage := createBuildStage()

		initialState := pipeline.State{
			ProjectDir: tmpDir,
			Stacks: []*stack.Stack{
				{Name: "api", Path: "api", AbsPath: apiDir, Runtime: "go1.x", Handler: "."},
				{Name: "worker", Path: "worker", AbsPath: workerDir, Runtime: "python3.12", Handler: "handler.handler"},
			},
		}

		result := buildStage(context.Background(), initialState)

		assert.True(t, E.IsRight(result), "Should build all stacks successfully")

		finalState := E.Fold(
			func(e error) pipeline.State { return pipeline.State{} },
			func(s pipeline.State) pipeline.State { return s },
		)(result)

		// Both artifacts should be present
		assert.Len(t, finalState.Artifacts, 2)
		assert.Contains(t, finalState.Artifacts, "api")
		assert.Contains(t, finalState.Artifacts, "worker")
	})
}
