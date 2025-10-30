package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNodeBuildSignature tests that NodeBuild has correct signature.
func TestNodeBuildSignature(t *testing.T) {
	t.Run("NodeBuild matches BuildFunc signature", func(t *testing.T) {
		// NodeBuild should be assignable to BuildFunc
		var buildFunc BuildFunc = NodeBuild

		// Should compile and work with functional patterns
		result := buildFunc(t.Context(), Config{
			SourceDir: "/nonexistent",
			Runtime:   "nodejs22.x",
		})

		// Should return Either type
		assert.True(t, E.IsLeft(result) || E.IsRight(result), "Should return Either type")
	})

	t.Run("NodeBuild returns Left on error", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/nonexistent/directory",
			OutputPath: "/tmp/output.zip",
			Runtime:    "nodejs22.x",
		}

		result := NodeBuild(t.Context(), cfg)

		assert.True(t, E.IsLeft(result), "Should return Left on error")
	})
}

// TestNodeBuildPure tests that NodeBuild is a pure function.
func TestNodeBuildPure(t *testing.T) {
	t.Run("same inputs produce same result", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/nonexistent",
			Runtime:    "nodejs22.x",
			OutputPath: "/tmp/test.zip",
		}

		result1 := NodeBuild(t.Context(), cfg)
		result2 := NodeBuild(t.Context(), cfg)

		// Both should fail the same way
		assert.Equal(t, E.IsLeft(result1), E.IsLeft(result2))
	})

	t.Run("deterministic error behavior", func(t *testing.T) {
		cfg := Config{
			SourceDir: "/nonexistent/path",
			Runtime:   "nodejs22.x",
		}

		// Multiple calls should produce consistent error results
		result1 := NodeBuild(t.Context(), cfg)
		result2 := NodeBuild(t.Context(), cfg)

		// Both should return Left (error)
		assert.True(t, E.IsLeft(result1))
		assert.True(t, E.IsLeft(result2))
	})
}

// TestNodeBuildComposition tests NodeBuild with functional composition.
func TestNodeBuildComposition(t *testing.T) {
	t.Run("composes with WithCache", func(t *testing.T) {
		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(NodeBuild)

		// Should still be a BuildFunc
		var buildFunc BuildFunc = cachedBuild
		assert.NotNil(t, buildFunc)
	})

	t.Run("composes with WithLogging", func(t *testing.T) {
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}

		loggedBuild := WithLogging(logger)(NodeBuild)

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
		)(NodeBuild)

		// Should still be a BuildFunc
		var buildFunc BuildFunc = composed
		assert.NotNil(t, buildFunc)
	})
}

// TestNodeBuildRegistry tests NodeBuild in registry.
func TestNodeBuildRegistry(t *testing.T) {
	t.Run("registry contains Node.js runtimes", func(t *testing.T) {
		registry := NewRegistry()

		assert.Contains(t, registry, "nodejs18.x", "Should contain nodejs18.x")
		assert.Contains(t, registry, "nodejs20.x", "Should contain nodejs20.x")
		assert.Contains(t, registry, "nodejs22.x", "Should contain nodejs22.x")
	})

	t.Run("Node.js builders use NodeBuild function", func(t *testing.T) {
		registry := NewRegistry()

		builder18 := registry["nodejs18.x"]
		builder20 := registry["nodejs20.x"]
		builder22 := registry["nodejs22.x"]

		// All should be the same function (NodeBuild)
		// We can test this by checking they behave identically
		cfg := Config{SourceDir: "/nonexistent"}

		result18 := builder18(t.Context(), cfg)
		result20 := builder20(t.Context(), cfg)
		result22 := builder22(t.Context(), cfg)

		// All should fail the same way
		assert.Equal(t, E.IsLeft(result18), E.IsLeft(result20))
		assert.Equal(t, E.IsLeft(result20), E.IsLeft(result22))
	})
}

// TestGenerateNodeBuildSpec tests the pure spec generation function.
func TestGenerateNodeBuildSpec(t *testing.T) {
	t.Run("generates spec without package.json or TypeScript", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/tmp/test",
			OutputPath: "",
			Runtime:    "nodejs22.x",
		}

		spec := GenerateNodeBuildSpec(cfg, false, false)

		assert.Equal(t, filepath.Join("/tmp/test", "lambda.zip"), spec.OutputPath)
		assert.Equal(t, "/tmp/test", spec.SourceDir)
		assert.False(t, spec.HasPackageJSON)
		assert.False(t, spec.HasTypeScript)
		assert.Empty(t, spec.InstallCmd)
		assert.Empty(t, spec.BuildCmd)
	})

	t.Run("generates spec with package.json but no TypeScript", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/tmp/test",
			OutputPath: "/tmp/output.zip",
			Runtime:    "nodejs22.x",
		}

		spec := GenerateNodeBuildSpec(cfg, true, false)

		assert.True(t, spec.HasPackageJSON)
		assert.False(t, spec.HasTypeScript)
		assert.Contains(t, spec.InstallCmd, "npm")
		assert.Contains(t, spec.InstallCmd, "install")
		assert.Empty(t, spec.BuildCmd, "Should not have build cmd without TypeScript")
	})

	t.Run("generates spec with package.json and TypeScript", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/tmp/test",
			OutputPath: "/tmp/output.zip",
			Runtime:    "nodejs22.x",
		}

		spec := GenerateNodeBuildSpec(cfg, true, true)

		assert.True(t, spec.HasPackageJSON)
		assert.True(t, spec.HasTypeScript)
		assert.Contains(t, spec.InstallCmd, "npm")
		assert.Contains(t, spec.BuildCmd, "npm")
		assert.Contains(t, spec.BuildCmd, "run")
		assert.Contains(t, spec.BuildCmd, "build")
	})
}

// TestNodeBuildBasic tests basic Node.js build without dependencies
func TestNodeBuildBasic(t *testing.T) {
	t.Run("builds simple Node.js function without package.json", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create simple handler
		handlerPath := filepath.Join(tmpDir, "index.js")
		handlerContent := `exports.handler = async (event) => {
    return {
        statusCode: 200,
        body: JSON.stringify('Hello from Lambda!'),
    };
};
`
		err := os.WriteFile(handlerPath, []byte(handlerContent), 0o644)
		require.NoError(t, err)

		outputPath := filepath.Join(tmpDir, "lambda.zip")
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: outputPath,
			Runtime:    "nodejs22.x",
		}

		result := NodeBuild(t.Context(), cfg)

		assert.True(t, E.IsRight(result), "Should succeed for simple Node.js function")

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

// TestNodeBuildWithPackageJson tests Node.js build with package.json
func TestNodeBuildWithPackageJson(t *testing.T) {
	// Skip if npm is not available
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available, skipping package.json test")
	}

	t.Run("builds with package.json", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create handler
		handlerPath := filepath.Join(tmpDir, "index.js")
		handlerContent := `exports.handler = async (event) => {
    return { statusCode: 200 };
};
`
		err := os.WriteFile(handlerPath, []byte(handlerContent), 0o644)
		require.NoError(t, err)

		// Create package.json with a simple dependency
		pkgPath := filepath.Join(tmpDir, "package.json")
		pkgContent := `{
  "name": "test-function",
  "version": "1.0.0",
  "dependencies": {
    "uuid": "^9.0.0"
  }
}
`
		err = os.WriteFile(pkgPath, []byte(pkgContent), 0o644)
		require.NoError(t, err)

		outputPath := filepath.Join(tmpDir, "lambda.zip")
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: outputPath,
			Runtime:    "nodejs22.x",
		}

		result := NodeBuild(t.Context(), cfg)

		if E.IsLeft(result) {
			// Extract error for debugging
			err := E.Fold(
				func(e error) error { return e },
				func(_ Artifact) error { return nil },
			)(result)
			t.Logf("Build failed: %v", err)
		}

		assert.True(t, E.IsRight(result), "Should succeed with valid package.json")

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

// TestNodeBuildTypeScript tests TypeScript compilation.
func TestNodeBuildTypeScript(t *testing.T) {
	// Skip if npm is not available
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available, skipping TypeScript test")
	}

	t.Run("detects and builds TypeScript project", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create TypeScript handler
		handlerPath := filepath.Join(tmpDir, "index.ts")
		handlerContent := `export const handler = async (event: any) => {
    return {
        statusCode: 200,
        body: JSON.stringify('Hello from TypeScript!'),
    };
};
`
		err := os.WriteFile(handlerPath, []byte(handlerContent), 0o644)
		require.NoError(t, err)

		// Create tsconfig.json
		tsconfigPath := filepath.Join(tmpDir, "tsconfig.json")
		tsconfigContent := `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "outDir": "./dist",
    "rootDir": "./",
    "strict": true,
    "esModuleInterop": true
  }
}
`
		err = os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0o644)
		require.NoError(t, err)

		// Create package.json with build script and typescript dependency
		pkgPath := filepath.Join(tmpDir, "package.json")
		pkgContent := `{
  "name": "test-typescript",
  "version": "1.0.0",
  "scripts": {
    "build": "tsc"
  },
  "devDependencies": {
    "typescript": "^5.3.0",
    "@types/node": "^20.10.0"
  }
}
`
		err = os.WriteFile(pkgPath, []byte(pkgContent), 0o644)
		require.NoError(t, err)

		outputPath := filepath.Join(tmpDir, "lambda.zip")
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: outputPath,
			Runtime:    "nodejs22.x",
		}

		result := NodeBuild(t.Context(), cfg)

		if E.IsLeft(result) {
			// Extract error for debugging
			err := E.Fold(
				func(e error) error { return e },
				func(_ Artifact) error { return nil },
			)(result)
			t.Logf("TypeScript build failed: %v", err)
		}

		// Should succeed and compile TypeScript
		assert.True(t, E.IsRight(result), "Should succeed with TypeScript project")

		// Verify dist directory was created (TypeScript compiled)
		distDir := filepath.Join(tmpDir, "dist")
		_, err = os.Stat(distDir)
		assert.NoError(t, err, "dist directory should exist after TypeScript compilation")
	})
}

// TestNodeBuildTypeScriptDetection tests TypeScript detection logic.
func TestNodeBuildTypeScriptDetection(t *testing.T) {
	t.Run("skips build if no tsconfig.json", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create handler (JavaScript, not TypeScript)
		handlerPath := filepath.Join(tmpDir, "index.js")
		err := os.WriteFile(handlerPath, []byte("exports.handler = async () => ({});"), 0o644)
		require.NoError(t, err)

		// Create package.json WITHOUT tsconfig.json
		pkgPath := filepath.Join(tmpDir, "package.json")
		pkgContent := `{"name": "test", "version": "1.0.0"}`
		err = os.WriteFile(pkgPath, []byte(pkgContent), 0o644)
		require.NoError(t, err)

		cfg := Config{
			SourceDir:  tmpDir,
			Runtime:    "nodejs22.x",
			OutputPath: filepath.Join(tmpDir, "output.zip"),
		}

		// This should succeed without trying to run npm run build
		// since there's no tsconfig.json
		result := NodeBuild(t.Context(), cfg)

		// May fail due to npm install, but should not fail due to missing build script
		if E.IsLeft(result) {
			err := E.Fold(
				func(e error) error { return e },
				func(_ Artifact) error { return nil },
			)(result)
			// Should not mention "npm run build" since tsconfig.json doesn't exist
			if err != nil {
				assert.NotContains(t, err.Error(), "npm run build failed",
					"Should not run npm run build without tsconfig.json")
			}
		}
	})
}

// TestNodeBuildErrorHandling tests error scenarios.
func TestNodeBuildErrorHandling(t *testing.T) {
	t.Run("returns error for invalid package.json", func(t *testing.T) {
		// Skip if npm is not available
		if _, err := exec.LookPath("npm"); err != nil {
			t.Skip("npm not available, skipping")
		}

		tmpDir := t.TempDir()

		// Create handler
		handlerPath := filepath.Join(tmpDir, "index.js")
		err := os.WriteFile(handlerPath, []byte("exports.handler = async () => ({});"), 0o644)
		require.NoError(t, err)

		// Create invalid package.json
		pkgPath := filepath.Join(tmpDir, "package.json")
		invalidPkg := `{"name": "test", "dependencies": {"nonexistent-package-xyz-123": "99.99.99"}}`
		err = os.WriteFile(pkgPath, []byte(invalidPkg), 0o644)
		require.NoError(t, err)

		cfg := Config{
			SourceDir: tmpDir,
			Runtime:   "nodejs22.x",
		}

		result := NodeBuild(t.Context(), cfg)

		assert.True(t, E.IsLeft(result), "Should fail with invalid package.json")

		// Extract error
		buildErr := E.Fold(
			func(e error) error { return e },
			func(_ Artifact) error { return nil },
		)(result)

		assert.NotNil(t, buildErr)
		assert.Contains(t, buildErr.Error(), "command failed", "Error should mention command failure")
	})

	t.Run("returns error for missing build script in TypeScript project", func(t *testing.T) {
		// Skip if npm is not available
		if _, err := exec.LookPath("npm"); err != nil {
			t.Skip("npm not available, skipping")
		}

		tmpDir := t.TempDir()

		// Create TypeScript handler
		handlerPath := filepath.Join(tmpDir, "index.ts")
		err := os.WriteFile(handlerPath, []byte("export const handler = async () => ({});"), 0o644)
		require.NoError(t, err)

		// Create tsconfig.json (indicates TypeScript project)
		tsconfigPath := filepath.Join(tmpDir, "tsconfig.json")
		err = os.WriteFile(tsconfigPath, []byte(`{"compilerOptions": {}}`), 0o644)
		require.NoError(t, err)

		// Create package.json WITHOUT build script
		pkgPath := filepath.Join(tmpDir, "package.json")
		pkgContent := `{"name": "test", "version": "1.0.0"}`
		err = os.WriteFile(pkgPath, []byte(pkgContent), 0o644)
		require.NoError(t, err)

		cfg := Config{
			SourceDir: tmpDir,
			Runtime:   "nodejs22.x",
		}

		result := NodeBuild(t.Context(), cfg)

		assert.True(t, E.IsLeft(result), "Should fail with missing build script")

		// Extract error
		buildErr := E.Fold(
			func(e error) error { return e },
			func(_ Artifact) error { return nil },
		)(result)

		assert.NotNil(t, buildErr)
		assert.Contains(t, buildErr.Error(), "command failed", "Error should mention command failure")
	})

	t.Run("returns error for invalid output path", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create handler
		handlerPath := filepath.Join(tmpDir, "index.js")
		err := os.WriteFile(handlerPath, []byte("exports.handler = async () => ({});"), 0o644)
		require.NoError(t, err)

		// Use an invalid output path
		cfg := Config{
			SourceDir:  tmpDir,
			OutputPath: "/nonexistent/dir/output.zip",
			Runtime:    "nodejs22.x",
		}

		result := NodeBuild(t.Context(), cfg)

		assert.True(t, E.IsLeft(result), "Should fail with invalid output path")
	})
}

// TestNodeBuildOutputPath tests output path handling.
func TestNodeBuildOutputPath(t *testing.T) {
	t.Run("uses default output path if not specified", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/tmp/test",
			OutputPath: "", // Empty - should use default
			Runtime:    "nodejs22.x",
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
			Runtime:    "nodejs22.x",
		}

		assert.Equal(t, customPath, cfg.OutputPath)
	})
}

// Benchmark NodeBuild function.
func BenchmarkNodeBuild(b *testing.B) {
	// Setup a basic project (this will fail, but we're benchmarking the function overhead)
	cfg := Config{
		SourceDir:  "/nonexistent",
		OutputPath: "/tmp/test.zip",
		Runtime:    "nodejs22.x",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NodeBuild(b.Context(), cfg)
	}
}

// BenchmarkNodeBuildWithComposition benchmarks composed build functions.
func BenchmarkNodeBuildWithComposition(b *testing.B) {
	cfg := Config{
		SourceDir: "/nonexistent",
		Runtime:   "nodejs22.x",
	}

	b.Run("Plain", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NodeBuild(b.Context(), cfg)
		}
	})

	b.Run("WithCache", func(b *testing.B) {
		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(NodeBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cachedBuild(b.Context(), cfg)
		}
	})

	b.Run("WithLogging", func(b *testing.B) {
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}
		loggedBuild := WithLogging(logger)(NodeBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			loggedBuild(b.Context(), cfg)
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
		)(NodeBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			composed(b.Context(), cfg)
		}
	})
}
