//go:build integration
// +build integration

package build

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationGoBuild tests building a real Go Lambda function.
func TestIntegrationGoBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// Write minimal Go Lambda function
	goCode := `package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
	Name string ` + "`json:\"name\"`" + `
}

type Response struct {
	Message string ` + "`json:\"message\"`" + `
}

func handler(ctx context.Context, event Event) (Response, error) {
	return Response{Message: "Hello " + event.Name}, nil
}

func main() {
	lambda.Start(handler)
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(goCode), 0o644)
	require.NoError(t, err)

	// Create go.mod
	goMod := `module test-lambda

go 1.22

require github.com/aws/aws-lambda-go v1.47.0
`
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644)
	require.NoError(t, err)

	// Run go mod download to populate go.sum
	// This is normally done by the developer before building
	// We skip this in the test since it requires network access
	// Instead, we'll test with a simple function that doesn't need external deps

	t.Run("builds Go Lambda function", func(t *testing.T) {
		// Rewrite to not use external dependencies
		simpleGoCode := `package main

func main() {
	println("Hello World")
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(simpleGoCode), 0o644)
		require.NoError(t, err)

		simpleGoMod := `module test-lambda
go 1.22
`
		err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(simpleGoMod), 0o644)
		require.NoError(t, err)
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: filepath.Join(tmpDir, "bootstrap"),
			Runtime:    "go1.x",
		}

		result := GoBuild(t.Context(), cfg)

		if E.IsLeft(result) {
			err := E.Fold(
				func(e error) error { return e },
				func(_ Artifact) error { return nil },
			)(result)
			t.Logf("Build failed with error: %v", err)
		}

		require.True(t, E.IsRight(result), "Go build should succeed")

		// Extract artifact
		artifact := E.Fold(
			func(e error) Artifact { return Artifact{} },
			func(_ Artifact) Artifact { return a },
		)(result)

		assert.FileExists(t, artifact.Path)
		assert.Greater(t, artifact.Size, int64(0))
		assert.NotEmpty(t, artifact.Checksum)

		// Verify binary is executable
		info, err := os.Stat(artifact.Path)
		require.NoError(t, err)
		assert.True(t, info.Mode()&0111 != 0, "Binary should be executable")
	})

	t.Run("builds with custom output path", func(t *testing.T) {
		customPath := filepath.Join(tmpDir, "custom-output")
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: customPath,
			Runtime:    "go1.x",
		}

		result := GoBuild(t.Context(), cfg)

		require.True(t, E.IsRight(result), "Build with custom path should succeed")

		artifact := E.Fold(
			func(e error) Artifact { return Artifact{} },
			func(_ Artifact) Artifact { return a },
		)(result)

		assert.Equal(t, customPath, artifact.Path)
		assert.FileExists(t, artifact.Path)
	})
}

// TestIntegrationGoBuildFailure tests Go build failures.
func TestIntegrationGoBuildFailure(t *testing.T) {
	tmpDir := t.TempDir()

	// Write invalid Go code
	invalidGoCode := `package main

func main() {
	// Missing import, invalid syntax
	lambda.Start(handler)
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(invalidGoCode), 0o644)
	require.NoError(t, err)

	// Create go.mod
	goMod := `module test-lambda
go 1.22
`
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644)
	require.NoError(t, err)

	t.Run("fails with invalid Go code", func(t *testing.T) {
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: filepath.Join(tmpDir, "bootstrap"),
			Runtime:    "go1.x",
		}

		result := GoBuild(t.Context(), cfg)

		assert.True(t, E.IsLeft(result), "Build should fail with invalid code")
	})
}

// TestIntegrationPythonBuild tests building Python Lambda function.
func TestIntegrationPythonBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// Write minimal Python Lambda function
	pythonCode := `import json

def handler(event, context):
    name = event.get('name', 'World')
    return {
        'statusCode': 200,
        'body': json.dumps({'message': f'Hello {name}'})
    }
`
	err := os.WriteFile(filepath.Join(tmpDir, "lambda_function.py"), []byte(pythonCode), 0o644)
	require.NoError(t, err)

	t.Run("builds Python Lambda function", func(t *testing.T) {
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: filepath.Join(tmpDir, "lambda.zip"),
			Runtime:    "python3.11",
			Handler:    "lambda_function.handler",
		}

		result := PythonBuild(t.Context(), cfg)

		require.True(t, E.IsRight(result), "Python build should succeed")

		artifact := E.Fold(
			func(e error) Artifact { return Artifact{} },
			func(_ Artifact) Artifact { return a },
		)(result)

		assert.FileExists(t, artifact.Path)
		assert.Greater(t, artifact.Size, int64(0))
		assert.NotEmpty(t, artifact.Checksum)
		assert.Equal(t, ".zip", filepath.Ext(artifact.Path))
	})

	// Skip dependency test - requires network access to pip install
	t.Run("includes requirements.txt dependencies", func(t *testing.T) {
		t.Skip("Skipping Python dependency test - requires network access for pip install")
	})
}

// TestIntegrationNodeBuild tests building Node.js Lambda function
func TestIntegrationNodeBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// Write minimal Node.js Lambda function
	nodeCode := `exports.handler = async (event) => {
    const name = event.name || 'World';
    return {
        statusCode: 200,
        body: JSON.stringify({
            message: ` + "`Hello ${name}`" + `
        })
    };
};
`
	err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(nodeCode), 0o644)
	require.NoError(t, err)

	t.Run("builds Node.js Lambda function", func(t *testing.T) {
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: filepath.Join(tmpDir, "lambda.zip"),
			Runtime:    "nodejs20.x",
			Handler:    "index.handler",
		}

		result := NodeBuild(t.Context(), cfg)

		require.True(t, E.IsRight(result), "Node build should succeed")

		artifact := E.Fold(
			func(e error) Artifact { return Artifact{} },
			func(_ Artifact) Artifact { return a },
		)(result)

		assert.FileExists(t, artifact.Path)
		assert.Greater(t, artifact.Size, int64(0))
		assert.NotEmpty(t, artifact.Checksum)
	})

	// Skip dependency test - requires network access to npm install
	t.Run("includes package.json dependencies", func(t *testing.T) {
		t.Skip("Skipping Node.js dependency test - requires network access for npm install")
	})
}

// TestIntegrationBuildAll tests building multiple configs.
func TestIntegrationBuildAll(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go function (simplified, no external deps)
	goDir := filepath.Join(tmpDir, "go-function")
	err := os.MkdirAll(goDir, 0o755)
	require.NoError(t, err)

	goCode := `package main
func main() {
	println("Hello from Go")
}
`
	err = os.WriteFile(filepath.Join(goDir, "main.go"), []byte(goCode), 0o644)
	require.NoError(t, err)

	goMod := `module test
go 1.22
`
	err = os.WriteFile(filepath.Join(goDir, "go.mod"), []byte(goMod), 0o644)
	require.NoError(t, err)

	// Create Python function
	pyDir := filepath.Join(tmpDir, "py-function")
	err = os.MkdirAll(pyDir, 0o755)
	require.NoError(t, err)

	pyCode := `def handler(event, context):
    return {'message': 'Hello from Python'}
`
	err = os.WriteFile(filepath.Join(pyDir, "lambda_function.py"), []byte(pyCode), 0o644)
	require.NoError(t, err)

	t.Run("builds multiple functions in parallel", func(t *testing.T) {
		configs := []Config{
			{
				SourceDir:  goDir,
				OutputPath: filepath.Join(goDir, "bootstrap"),
				Runtime:    "go1.x",
			},
			{
				SourceDir:  pyDir,
				OutputPath: filepath.Join(pyDir, "lambda.zip"),
				Runtime:    "python3.11",
				Handler:    "lambda_function.handler",
			},
		}

		registry := NewRegistry()
		result := BuildAll(t.Context(), configs, registry)

		require.True(t, E.IsRight(result), "BuildAll should succeed")

		artifacts := E.Fold(
			func(e error) []Artifact { return nil },
			func(a []Artifact) []Artifact { return a },
		)(result)

		assert.Len(t, artifacts, 2, "Should build both functions")
		assert.FileExists(t, artifacts[0].Path)
		assert.FileExists(t, artifacts[1].Path)
	})
}

// TestIntegrationWithCache tests caching functionality.
func TestIntegrationWithCache(t *testing.T) {
	tmpDir := t.TempDir()

	goCode := `package main
func main() {
	println("Hello")
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(goCode), 0o644)
	require.NoError(t, err)

	goMod := `module test
go 1.22
`
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644)
	require.NoError(t, err)

	t.Run("cache speeds up second build", func(t *testing.T) {
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: filepath.Join(tmpDir, "bootstrap"),
			Runtime:    "go1.x",
		}

		cache := &memoryCache{cache: make(map[string]Artifact)}
		cachedBuild := WithCache(cache)(GoBuild)

		// First build - should execute
		result1 := cachedBuild(t.Context(), cfg)
		require.True(t, E.IsRight(result1))

		// Second build - should use cache (much faster)
		result2 := cachedBuild(t.Context(), cfg)
		require.True(t, E.IsRight(result2))

		// Both should return same artifact
		artifact1 := E.Fold(
			func(e error) Artifact { return Artifact{} },
			func(_ Artifact) Artifact { return a },
		)(result1)

		artifact2 := E.Fold(
			func(e error) Artifact { return Artifact{} },
			func(_ Artifact) Artifact { return a },
		)(result2)

		assert.Equal(t, artifact1.Checksum, artifact2.Checksum)
	})
}

// Simple cache implementation for testing.
type memoryCache struct {
	cache map[string]Artifact
}

func (c *memoryCache) Get(cfg Config) (Artifact, bool) {
	key := cfg.SourceDir + "-" + cfg.Runtime
	artifact, ok := c.cache[key]
	return artifact, ok
}

func (c *memoryCache) Set(cfg Config, artifact Artifact) {
	key := cfg.SourceDir + "-" + cfg.Runtime
	c.cache[key] = artifact
}
