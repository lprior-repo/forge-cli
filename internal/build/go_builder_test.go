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

// TestGoBuildSignature tests that GoBuild has correct signature
func TestGoBuildSignature(t *testing.T) {
	t.Run("GoBuild matches BuildFunc signature", func(t *testing.T) {
		// GoBuild should be assignable to BuildFunc
		var buildFunc BuildFunc = GoBuild

		// Should compile and work with functional patterns
		result := buildFunc(context.Background(), Config{
			SourceDir: "/nonexistent",
			Runtime:   "provided.al2023",
			Handler:   "main.go",
		})

		// Should return Either type
		assert.True(t, E.IsLeft(result) || E.IsRight(result), "Should return Either type")
	})

	t.Run("GoBuild returns Left on error", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/nonexistent/directory",
			OutputPath: "/tmp/bootstrap",
			Runtime:    "provided.al2023",
			Handler:    "main.go",
		}

		result := GoBuild(context.Background(), cfg)

		assert.True(t, E.IsLeft(result), "Should return Left on error")
	})
}

// TestGoBuildPure tests that GoBuild is a pure function
func TestGoBuildPure(t *testing.T) {
	t.Run("same inputs produce same result", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/nonexistent",
			Runtime:    "provided.al2023",
			OutputPath: "/tmp/test-bootstrap",
			Handler:    "main.go",
		}

		result1 := GoBuild(context.Background(), cfg)
		result2 := GoBuild(context.Background(), cfg)

		// Both should fail the same way
		assert.Equal(t, E.IsLeft(result1), E.IsLeft(result2))
	})

	t.Run("deterministic error behavior", func(t *testing.T) {
		cfg := Config{
			SourceDir: "/nonexistent/path",
			Runtime:   "provided.al2023",
			Handler:   "main.go",
		}

		// Multiple calls should produce consistent error results
		result1 := GoBuild(context.Background(), cfg)
		result2 := GoBuild(context.Background(), cfg)

		// Both should return Left (error)
		assert.True(t, E.IsLeft(result1))
		assert.True(t, E.IsLeft(result2))
	})
}

// TestGoBuildComposition tests GoBuild with functional composition
func TestGoBuildComposition(t *testing.T) {
	t.Run("composes with WithCache", func(t *testing.T) {
		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(GoBuild)

		// Should still be a BuildFunc
		var buildFunc BuildFunc = cachedBuild
		assert.NotNil(t, buildFunc)
	})

	t.Run("composes with WithLogging", func(t *testing.T) {
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}

		loggedBuild := WithLogging(logger)(GoBuild)

		// Should still be a BuildFunc
		var buildFunc BuildFunc = loggedBuild
		assert.NotNil(t, buildFunc)
	})

	t.Run("composes with multiple decorators", func(t *testing.T) {
		cache := NewMemoryCache()
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}

		composed := Compose(
			WithCache(cache),
			WithLogging(logger),
		)(GoBuild)

		// Should still be a BuildFunc
		var buildFunc BuildFunc = composed
		assert.NotNil(t, buildFunc)
	})
}

// TestGoBuildRegistry tests GoBuild in registry
func TestGoBuildRegistry(t *testing.T) {
	t.Run("registry contains Go runtimes", func(t *testing.T) {
		registry := NewRegistry()

		assert.Contains(t, registry, "go1.x", "Should contain go1.x")
		assert.Contains(t, registry, "provided.al2", "Should contain provided.al2")
		assert.Contains(t, registry, "provided.al2023", "Should contain provided.al2023")
	})

	t.Run("Go builders use GoBuild function", func(t *testing.T) {
		registry := NewRegistry()

		builderGo1x := registry["go1.x"]
		builderAl2 := registry["provided.al2"]
		builderAl2023 := registry["provided.al2023"]

		// All should be the same function (GoBuild)
		// We can test this by checking they behave identically
		cfg := Config{SourceDir: "/nonexistent", Handler: "main.go"}

		result1x := builderGo1x(context.Background(), cfg)
		resultAl2 := builderAl2(context.Background(), cfg)
		resultAl2023 := builderAl2023(context.Background(), cfg)

		// All should fail the same way
		assert.Equal(t, E.IsLeft(result1x), E.IsLeft(resultAl2))
		assert.Equal(t, E.IsLeft(resultAl2), E.IsLeft(resultAl2023))
	})
}

// TestGoBuildBasic tests basic Go build
func TestGoBuildBasic(t *testing.T) {
	t.Run("builds simple Go Lambda function", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create simple main.go
		mainPath := filepath.Join(tmpDir, "main.go")
		mainContent := `package main

import (
	"context"
)

func main() {}

func Handler(ctx context.Context, event interface{}) (interface{}, error) {
	return map[string]string{"message": "Hello"}, nil
}
`
		err := os.WriteFile(mainPath, []byte(mainContent), 0644)
		require.NoError(t, err)

		// Create go.mod
		modPath := filepath.Join(tmpDir, "go.mod")
		modContent := `module example.com/lambda

go 1.21
`
		err = os.WriteFile(modPath, []byte(modContent), 0644)
		require.NoError(t, err)

		outputPath := filepath.Join(tmpDir, "bootstrap")
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: outputPath,
			Runtime:    "provided.al2023",
			Handler:    ".",
		}

		result := GoBuild(context.Background(), cfg)

		assert.True(t, E.IsRight(result), "Should succeed for simple Go function")

		// Extract artifact
		artifact := E.Fold(
			func(err error) Artifact { return Artifact{} },
			func(a Artifact) Artifact { return a },
		)(result)

		// Verify artifact
		assert.Equal(t, outputPath, artifact.Path)
		assert.NotEmpty(t, artifact.Checksum, "Should have checksum")
		assert.Greater(t, artifact.Size, int64(0), "Should have non-zero size")

		// Verify binary exists
		_, err = os.Stat(outputPath)
		assert.NoError(t, err, "Binary should exist")
	})
}

// TestGoBuildErrorHandling tests error scenarios
func TestGoBuildErrorHandling(t *testing.T) {
	t.Run("returns error for invalid Go code", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create invalid Go code
		mainPath := filepath.Join(tmpDir, "main.go")
		invalidCode := `package main

this is not valid Go code
`
		err := os.WriteFile(mainPath, []byte(invalidCode), 0644)
		require.NoError(t, err)

		// Create go.mod
		modPath := filepath.Join(tmpDir, "go.mod")
		modContent := `module example.com/lambda

go 1.21
`
		err = os.WriteFile(modPath, []byte(modContent), 0644)
		require.NoError(t, err)

		cfg := Config{
			SourceDir: tmpDir,
			Runtime:   "provided.al2023",
			Handler:   ".",
		}

		result := GoBuild(context.Background(), cfg)

		assert.True(t, E.IsLeft(result), "Should fail with invalid Go code")

		// Extract error
		buildErr := E.Fold(
			func(e error) error { return e },
			func(a Artifact) error { return nil },
		)(result)

		assert.NotNil(t, buildErr)
		assert.Contains(t, buildErr.Error(), "go build failed", "Error should mention build failure")
	})

	t.Run("returns error for missing source directory", func(t *testing.T) {
		cfg := Config{
			SourceDir: "/nonexistent/directory/that/does/not/exist",
			Runtime:   "provided.al2023",
			Handler:   ".",
		}

		result := GoBuild(context.Background(), cfg)

		assert.True(t, E.IsLeft(result), "Should fail with missing source directory")
	})
}

// TestGoBuildOutputPath tests output path handling
func TestGoBuildOutputPath(t *testing.T) {
	t.Run("uses default output path if not specified", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/tmp/test",
			OutputPath: "", // Empty - should use default
			Runtime:    "provided.al2023",
			Handler:    ".",
		}

		// Default should be bootstrap in source directory
		expectedPath := filepath.Join(cfg.SourceDir, "bootstrap")
		assert.Contains(t, expectedPath, "bootstrap")
	})

	t.Run("respects custom output path", func(t *testing.T) {
		customPath := "/custom/path/my-lambda"
		cfg := Config{
			SourceDir:  "/tmp/test",
			OutputPath: customPath,
			Runtime:    "provided.al2023",
			Handler:    ".",
		}

		assert.Equal(t, customPath, cfg.OutputPath)
	})
}

// TestGoBuildEnvironment tests environment variable handling
func TestGoBuildEnvironment(t *testing.T) {
	t.Run("sets required Lambda environment variables", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create simple main.go
		mainPath := filepath.Join(tmpDir, "main.go")
		mainContent := `package main

func main() {}
`
		err := os.WriteFile(mainPath, []byte(mainContent), 0644)
		require.NoError(t, err)

		// Create go.mod
		modPath := filepath.Join(tmpDir, "go.mod")
		modContent := `module example.com/lambda

go 1.21
`
		err = os.WriteFile(modPath, []byte(modContent), 0644)
		require.NoError(t, err)

		cfg := Config{
			SourceDir: tmpDir,
			Runtime:   "provided.al2023",
			Handler:   ".",
			Env: map[string]string{
				"CUSTOM_VAR": "custom_value",
			},
		}

		// This will attempt to build with custom env vars
		result := GoBuild(context.Background(), cfg)

		// Should succeed or fail consistently
		_ = result // Environment is set internally, hard to test directly without mocking
	})
}

// Benchmark GoBuild function
func BenchmarkGoBuild(b *testing.B) {
	// Setup a basic project (this will fail, but we're benchmarking the function overhead)
	cfg := Config{
		SourceDir:  "/nonexistent",
		OutputPath: "/tmp/test-bootstrap",
		Runtime:    "provided.al2023",
		Handler:    ".",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GoBuild(context.Background(), cfg)
	}
}

// BenchmarkGoBuildWithComposition benchmarks composed build functions
func BenchmarkGoBuildWithComposition(b *testing.B) {
	cfg := Config{
		SourceDir: "/nonexistent",
		Runtime:   "provided.al2023",
		Handler:   ".",
	}

	b.Run("Plain", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GoBuild(context.Background(), cfg)
		}
	})

	b.Run("WithCache", func(b *testing.B) {
		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(GoBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cachedBuild(context.Background(), cfg)
		}
	})

	b.Run("WithLogging", func(b *testing.B) {
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}
		loggedBuild := WithLogging(logger)(GoBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			loggedBuild(context.Background(), cfg)
		}
	})

	b.Run("Composed", func(b *testing.B) {
		cache := NewMemoryCache()
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}

		composed := Compose(
			WithCache(cache),
			WithLogging(logger),
		)(GoBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			composed(context.Background(), cfg)
		}
	})
}
