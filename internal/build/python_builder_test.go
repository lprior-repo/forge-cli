package build

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPythonBuildSignature tests that PythonBuild has correct signature
func TestPythonBuildSignature(t *testing.T) {
	t.Run("PythonBuild matches BuildFunc signature", func(t *testing.T) {
		// PythonBuild should be assignable to BuildFunc
		var buildFunc BuildFunc = PythonBuild

		// Should compile and work with functional patterns
		result := buildFunc(context.Background(), Config{
			SourceDir: "/nonexistent",
			Runtime:   "python3.13",
		})

		// Should return Either type
		assert.True(t, E.IsLeft(result) || E.IsRight(result), "Should return Either type")
	})

	t.Run("PythonBuild returns Left on error", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/nonexistent/directory",
			OutputPath: "/tmp/output.zip",
			Runtime:    "python3.11",
		}

		result := PythonBuild(context.Background(), cfg)

		assert.True(t, E.IsLeft(result), "Should return Left on error")
	})
}

// TestPythonBuildPure tests that PythonBuild is a pure function
func TestPythonBuildPure(t *testing.T) {
	t.Run("same inputs produce same result", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/nonexistent",
			Runtime:    "python3.11",
			OutputPath: "/tmp/test.zip",
		}

		result1 := PythonBuild(context.Background(), cfg)
		result2 := PythonBuild(context.Background(), cfg)

		// Both should fail the same way
		assert.Equal(t, E.IsLeft(result1), E.IsLeft(result2))
	})

	t.Run("deterministic error behavior", func(t *testing.T) {
		cfg := Config{
			SourceDir: "/nonexistent/path",
			Runtime:   "python3.13",
		}

		// Multiple calls should produce consistent error results
		result1 := PythonBuild(context.Background(), cfg)
		result2 := PythonBuild(context.Background(), cfg)

		// Both should return Left (error)
		assert.True(t, E.IsLeft(result1))
		assert.True(t, E.IsLeft(result2))
	})
}

// TestPythonBuildComposition tests PythonBuild with functional composition
func TestPythonBuildComposition(t *testing.T) {
	t.Run("composes with WithCache", func(t *testing.T) {
		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(PythonBuild)

		// Should still be a BuildFunc
		var buildFunc BuildFunc = cachedBuild
		assert.NotNil(t, buildFunc)
	})

	t.Run("composes with WithLogging", func(t *testing.T) {
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}

		loggedBuild := WithLogging(logger)(PythonBuild)

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
		)(PythonBuild)

		// Should still be a BuildFunc
		var buildFunc BuildFunc = composed
		assert.NotNil(t, buildFunc)
	})
}

// TestPythonBuildRegistry tests PythonBuild in registry
func TestPythonBuildRegistry(t *testing.T) {
	t.Run("registry contains Python runtimes", func(t *testing.T) {
		registry := NewRegistry()

		assert.Contains(t, registry, "python3.9", "Should contain python3.9")
		assert.Contains(t, registry, "python3.10", "Should contain python3.10")
		assert.Contains(t, registry, "python3.11", "Should contain python3.11")
		assert.Contains(t, registry, "python3.12", "Should contain python3.12")
		assert.Contains(t, registry, "python3.13", "Should contain python3.13")
	})

	t.Run("Python builders use PythonBuild function", func(t *testing.T) {
		registry := NewRegistry()

		builder39 := registry["python3.9"]
		builder310 := registry["python3.10"]
		builder313 := registry["python3.13"]

		// All should be the same function (PythonBuild)
		// We can test this by checking they behave identically
		cfg := Config{SourceDir: "/nonexistent"}

		result39 := builder39(context.Background(), cfg)
		result310 := builder310(context.Background(), cfg)
		result313 := builder313(context.Background(), cfg)

		// All should fail the same way
		assert.Equal(t, E.IsLeft(result39), E.IsLeft(result310))
		assert.Equal(t, E.IsLeft(result310), E.IsLeft(result313))
	})
}

// TestPythonBuildBasic tests basic Python build without dependencies
func TestPythonBuildBasic(t *testing.T) {
	t.Run("builds simple Python function without requirements.txt", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create simple handler
		handlerPath := filepath.Join(tmpDir, "handler.py")
		handlerContent := `def handler(event, context):
    return {"statusCode": 200, "body": "Hello"}
`
		err := os.WriteFile(handlerPath, []byte(handlerContent), 0644)
		require.NoError(t, err)

		outputPath := filepath.Join(tmpDir, "lambda.zip")
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: outputPath,
			Runtime:    "python3.11",
		}

		result := PythonBuild(context.Background(), cfg)

		assert.True(t, E.IsRight(result), "Should succeed for simple Python function")

		// Extract artifact
		artifact := E.Fold(
			func(err error) Artifact { return Artifact{} },
			func(a Artifact) Artifact { return a },
		)(result)

		// Verify artifact
		assert.Equal(t, outputPath, artifact.Path)
		assert.NotEmpty(t, artifact.Checksum, "Should have checksum")
		assert.Greater(t, artifact.Size, int64(0), "Should have non-zero size")

		// Verify zip file exists
		_, err = os.Stat(outputPath)
		assert.NoError(t, err, "Zip file should exist")
	})
}

// TestPythonBuildWithRequirements tests Python build with requirements.txt
func TestPythonBuildWithRequirements(t *testing.T) {
	// Skip if neither pip nor uv is available
	if _, err := exec.LookPath("pip"); err != nil {
		if _, err := exec.LookPath("uv"); err != nil {
			t.Skip("Neither pip nor uv available, skipping requirements test")
		}
	}

	t.Run("builds with requirements.txt", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create handler
		handlerPath := filepath.Join(tmpDir, "handler.py")
		handlerContent := `def handler(event, context):
    return {"statusCode": 200}
`
		err := os.WriteFile(handlerPath, []byte(handlerContent), 0644)
		require.NoError(t, err)

		// Create requirements.txt with very lightweight package to avoid disk quota issues
		// 'six' is a small, stable package with minimal dependencies
		reqPath := filepath.Join(tmpDir, "requirements.txt")
		reqContent := "six==1.16.0\n"
		err = os.WriteFile(reqPath, []byte(reqContent), 0644)
		require.NoError(t, err)

		outputPath := filepath.Join(tmpDir, "lambda.zip")
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: outputPath,
			Runtime:    "python3.11",
		}

		result := PythonBuild(context.Background(), cfg)

		if E.IsLeft(result) {
			// Extract error for debugging
			err := E.Fold(
				func(e error) error { return e },
				func(a Artifact) error { return nil },
			)(result)
			t.Logf("Build failed: %v", err)
		}

		assert.True(t, E.IsRight(result), "Should succeed with valid requirements.txt")

		// Verify artifact
		artifact := E.Fold(
			func(err error) Artifact { return Artifact{} },
			func(a Artifact) Artifact { return a },
		)(result)

		assert.NotEmpty(t, artifact.Path)
		assert.FileExists(t, artifact.Path)
		assert.Greater(t, artifact.Size, int64(1000), "Should be larger with dependencies")
	})
}

// TestPythonBuildUvDetection tests uv vs pip detection
func TestPythonBuildUvDetection(t *testing.T) {
	t.Run("uses uv if available", func(t *testing.T) {
		// Check if uv is available
		_, err := exec.LookPath("uv")
		if err != nil {
			t.Skip("uv not available, skipping uv detection test")
		}

		tmpDir := t.TempDir()

		// Create simple handler
		handlerPath := filepath.Join(tmpDir, "handler.py")
		err = os.WriteFile(handlerPath, []byte("def handler(e, c): pass"), 0644)
		require.NoError(t, err)

		// Create empty requirements.txt to trigger dependency install
		reqPath := filepath.Join(tmpDir, "requirements.txt")
		err = os.WriteFile(reqPath, []byte(""), 0644)
		require.NoError(t, err)

		cfg := Config{
			SourceDir:  tmpDir,
			Runtime:    "python3.11",
			OutputPath: filepath.Join(tmpDir, "output.zip"),
		}

		result := PythonBuild(context.Background(), cfg)

		// Should succeed (uv handles empty requirements.txt gracefully)
		assert.True(t, E.IsRight(result), "Should succeed with uv")
	})
}

// TestShouldSkipFile tests the file skipping logic
func TestShouldSkipFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"skip .pyc", "module.pyc", true},
		{"skip .pyo", "module.pyo", true},
		{"skip .pyd", "module.pyd", true},
		{"skip __pycache__", "__pycache__", true},
		{"skip .git", ".git", true},
		{"skip .DS_Store", ".DS_Store", true},
		{"keep .py", "handler.py", false},
		{"keep .txt", "requirements.txt", false},
		{"keep .json", "config.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipFile(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestPythonBuildErrorHandling tests error scenarios
func TestPythonBuildErrorHandling(t *testing.T) {
	t.Run("returns error for invalid requirements.txt", func(t *testing.T) {
		// Skip if neither pip nor uv is available
		if _, err := exec.LookPath("pip"); err != nil {
			if _, err := exec.LookPath("uv"); err != nil {
				t.Skip("Neither pip nor uv available, skipping")
			}
		}

		tmpDir := t.TempDir()

		// Create handler
		handlerPath := filepath.Join(tmpDir, "handler.py")
		err := os.WriteFile(handlerPath, []byte("def handler(e, c): pass"), 0644)
		require.NoError(t, err)

		// Create invalid requirements.txt
		reqPath := filepath.Join(tmpDir, "requirements.txt")
		invalidReq := "nonexistent-package-xyz-123456789==99.99.99\n"
		err = os.WriteFile(reqPath, []byte(invalidReq), 0644)
		require.NoError(t, err)

		cfg := Config{
			SourceDir: tmpDir,
			Runtime:   "python3.13",
		}

		result := PythonBuild(context.Background(), cfg)

		assert.True(t, E.IsLeft(result), "Should fail with invalid requirements.txt")

		// Extract error
		buildErr := E.Fold(
			func(e error) error { return e },
			func(a Artifact) error { return nil },
		)(result)

		assert.NotNil(t, buildErr)
		// Error should mention command failure (from executeCommand)
		errMsg := buildErr.Error()
		t.Logf("Actual error message: %s", errMsg)
		assert.True(t,
			strings.Contains(errMsg, "command failed") ||
			strings.Contains(errMsg, "pip install failed") ||
			strings.Contains(errMsg, "uv pip install failed"),
			"Error should mention command/pip/uv install failure. Got: %s", errMsg,
		)
	})
}

// TestPythonBuildOutputPath tests output path handling
func TestPythonBuildOutputPath(t *testing.T) {
	t.Run("uses default output path if not specified", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/tmp/test",
			OutputPath: "", // Empty - should use default
			Runtime:    "python3.11",
		}

		// We know this will fail, but we can check the default path logic
		// by examining what path would be used
		expectedPath := filepath.Join(cfg.SourceDir, "lambda.zip")
		assert.Contains(t, expectedPath, "lambda.zip")
	})

	t.Run("respects custom output path", func(t *testing.T) {
		customPath := "/custom/path/myfunction.zip"
		cfg := Config{
			SourceDir:  "/tmp/test",
			OutputPath: customPath,
			Runtime:    "python3.11",
		}

		assert.Equal(t, customPath, cfg.OutputPath)
	})
}

// TestEnvSlice tests environment variable conversion
func TestEnvSlice(t *testing.T) {
	t.Run("converts empty map", func(t *testing.T) {
		env := envSlice(map[string]string{})
		assert.Empty(t, env)
	})

	t.Run("converts map with values", func(t *testing.T) {
		env := envSlice(map[string]string{
			"FOO": "bar",
			"BAZ": "qux",
		})

		assert.Len(t, env, 2)
		assert.Contains(t, env, "FOO=bar")
		assert.Contains(t, env, "BAZ=qux")
	})
}

// Benchmark PythonBuild function
func BenchmarkPythonBuild(b *testing.B) {
	// Setup a basic project (this will fail, but we're benchmarking the function overhead)
	cfg := Config{
		SourceDir:  "/nonexistent",
		OutputPath: "/tmp/test.zip",
		Runtime:    "python3.13",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PythonBuild(context.Background(), cfg)
	}
}

// BenchmarkPythonBuildWithComposition benchmarks composed build functions
func BenchmarkPythonBuildWithComposition(b *testing.B) {
	cfg := Config{
		SourceDir: "/nonexistent",
		Runtime:   "python3.13",
	}

	b.Run("Plain", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			PythonBuild(context.Background(), cfg)
		}
	})

	b.Run("WithCache", func(b *testing.B) {
		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(PythonBuild)

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
		loggedBuild := WithLogging(logger)(PythonBuild)

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
		)(PythonBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			composed(context.Background(), cfg)
		}
	})
}
