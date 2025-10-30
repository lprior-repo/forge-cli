package build

import (
	"context"
	"errors"
	"fmt"
	"testing"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildFuncSignature tests that BuildFunc has correct type.
func TestBuildFuncSignature(t *testing.T) {
	t.Run("BuildFunc accepts context and config", func(t *testing.T) {
		buildFunc := func(_ context.Context, _ Config) E.Either[error, Artifact] {
			return E.Right[error](Artifact{Path: "/tmp/test"})
		}

		result := buildFunc(t.Context(), Config{SourceDir: "/tmp"})
		assert.True(t, E.IsRight(result), "Should return Right")
	})

	t.Run("BuildFunc can return error via Left", func(t *testing.T) {
		buildFunc := func(_ context.Context, _ Config) E.Either[error, Artifact] {
			return E.Left[Artifact](errors.New("build failed"))
		}

		result := buildFunc(t.Context(), Config{})
		assert.True(t, E.IsLeft(result), "Should return Left on error")
	})
}

// TestRegistry tests the build function registry.
func TestRegistry(t *testing.T) {
	t.Run("NewRegistry creates registry with default builders", func(t *testing.T) {
		registry := NewRegistry()

		assert.NotNil(t, registry)
		assert.Contains(t, registry, "go1.x")
		assert.Contains(t, registry, "python3.11")
		assert.Contains(t, registry, "nodejs20.x")
	})

	t.Run("Get returns Some for existing runtime", func(t *testing.T) {
		registry := NewRegistry()

		builderOpt := GetBuilder(registry, "go1.x")
		assert.True(t, O.IsSome(builderOpt), "Should return Some for go1.x")
	})

	t.Run("Get returns None for non-existent runtime", func(t *testing.T) {
		registry := NewRegistry()

		builderOpt := GetBuilder(registry, "rust")
		assert.True(t, O.IsNone(builderOpt), "Should return None for unsupported runtime")
	})

	t.Run("can add custom builder to registry", func(t *testing.T) {
		registry := NewRegistry()

		customBuilder := func(_ context.Context, _ Config) E.Either[error, Artifact] {
			return E.Right[error](Artifact{Path: "/custom"})
		}

		registry["custom"] = customBuilder

		builderOpt := GetBuilder(registry, "custom")
		assert.True(t, O.IsSome(builderOpt), "Should find custom builder")
	})
}

// TestWithCache tests the caching higher-order function.
func TestWithCache(t *testing.T) {
	t.Run("caches successful builds", func(t *testing.T) {
		callCount := 0
		mockBuild := func(_ context.Context, _ Config) E.Either[error, Artifact] {
			callCount++
			return E.Right[error](Artifact{Path: fmt.Sprintf("/build-%d", callCount)})
		}

		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(mockBuild)

		cfg := Config{SourceDir: "/test"}

		// First call - should execute
		result1 := cachedBuild(t.Context(), cfg)
		assert.True(t, E.IsRight(result1))
		assert.Equal(t, 1, callCount, "Should call build once")

		// Second call - should use cache
		result2 := cachedBuild(t.Context(), cfg)
		assert.True(t, E.IsRight(result2))
		assert.Equal(t, 1, callCount, "Should not call build again (cached)")
	})

	t.Run("does not cache failures", func(t *testing.T) {
		callCount := 0
		mockBuild := func(_ context.Context, _ Config) E.Either[error, Artifact] {
			callCount++
			return E.Left[Artifact](errors.New("build failed"))
		}

		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(mockBuild)

		cfg := Config{SourceDir: "/test"}

		// First call - should fail
		result1 := cachedBuild(t.Context(), cfg)
		assert.True(t, E.IsLeft(result1))
		assert.Equal(t, 1, callCount)

		// Second call - should try again (not cached)
		result2 := cachedBuild(t.Context(), cfg)
		assert.True(t, E.IsLeft(result2))
		assert.Equal(t, 2, callCount, "Should retry on failure")
	})

	t.Run("different configs have separate cache entries", func(t *testing.T) {
		callCount := 0
		mockBuild := func(_ context.Context, cfg Config) E.Either[error, Artifact] {
			callCount++
			return E.Right[error](Artifact{Path: cfg.SourceDir})
		}

		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(mockBuild)

		cfg1 := Config{SourceDir: "/test1"}
		cfg2 := Config{SourceDir: "/test2"}

		// Build both configs
		cachedBuild(t.Context(), cfg1)
		cachedBuild(t.Context(), cfg2)

		assert.Equal(t, 2, callCount, "Should build both configs")

		// Rebuild first config - should use cache
		cachedBuild(t.Context(), cfg1)
		assert.Equal(t, 2, callCount, "Should use cache for cfg1")
	})
}

// TestWithLogging tests the logging higher-order function.
func TestWithLogging(t *testing.T) {
	t.Run("logs successful builds", func(t *testing.T) {
		var infoLogs, errorLogs []string

		logger := &mockLogger{
			infoFn: func(msg string, _ ...interface{}) {
				infoLogs = append(infoLogs, msg)
			},
			errorFn: func(msg string, _ ...interface{}) {
				errorLogs = append(errorLogs, msg)
			},
		}

		mockBuild := func(_ context.Context, _ Config) E.Either[error, Artifact] {
			return E.Right[error](Artifact{Path: "/test"})
		}

		loggedBuild := WithLogging(logger)(mockBuild)
		loggedBuild(t.Context(), Config{Runtime: "go1.x"})

		assert.Contains(t, infoLogs, "Building", "Should log build start")
		assert.Contains(t, infoLogs, "Build succeeded", "Should log success")
		assert.Empty(t, errorLogs, "Should not log errors on success")
	})

	t.Run("logs build failures", func(t *testing.T) {
		var errorLogs []string

		logger := &mockLogger{
			infoFn: func(_ string, _ ...interface{}) {},
			errorFn: func(msg string, _ ...interface{}) {
				errorLogs = append(errorLogs, msg)
			},
		}

		mockBuild := func(_ context.Context, _ Config) E.Either[error, Artifact] {
			return E.Left[Artifact](errors.New("build failed"))
		}

		loggedBuild := WithLogging(logger)(mockBuild)
		loggedBuild(t.Context(), Config{})

		assert.Contains(t, errorLogs, "Build failed", "Should log failure")
	})
}

// TestCompose tests composing multiple decorators.
func TestCompose(t *testing.T) {
	t.Run("composes cache and logging", func(t *testing.T) {
		callCount := 0
		var logs []string

		mockBuild := func(_ context.Context, _ Config) E.Either[error, Artifact] {
			callCount++
			return E.Right[error](Artifact{Path: "/test"})
		}

		cache := NewMemoryCache()
		logger := &mockLogger{
			infoFn: func(msg string, _ ...interface{}) {
				logs = append(logs, msg)
			},
			errorFn: func(_ string, _ ...interface{}) {},
		}

		// Compose both decorators
		decorated := Compose(
			WithCache(cache),
			WithLogging(logger),
		)(mockBuild)

		cfg := Config{SourceDir: "/test"}

		// First call
		decorated(t.Context(), cfg)
		assert.Equal(t, 1, callCount)
		assert.Contains(t, logs, "Building")

		// Second call - should use cache
		logs = nil // Clear logs
		decorated(t.Context(), cfg)
		assert.Equal(t, 1, callCount, "Should use cache")
		// Logging happens before cache check, so we still see logs
	})

	t.Run("decorator order matters", func(t *testing.T) {
		// When cache wraps logging: Cache → Logging → Build
		// When logging wraps cache: Logging → Cache → Build

		callCount := 0
		mockBuild := func(_ context.Context, _ Config) E.Either[error, Artifact] {
			callCount++
			return E.Right[error](Artifact{Path: "/test"})
		}

		cache := NewMemoryCache()
		logger := &mockLogger{
			infoFn:  func(_ string, _ ...interface{}) {},
			errorFn: func(_ string, _ ...interface{}) {},
		}

		// Order 1: Cache first, then logging
		build1 := WithLogging(logger)(WithCache(cache)(mockBuild))
		cfg := Config{SourceDir: "/test1"}
		build1(t.Context(), cfg)
		build1(t.Context(), cfg)
		count1 := callCount

		callCount = 0

		// Order 2: Logging first, then cache
		build2 := WithCache(cache)(WithLogging(logger)(mockBuild))
		cfg2 := Config{SourceDir: "/test2"}
		build2(t.Context(), cfg2)
		build2(t.Context(), cfg2)
		count2 := callCount

		// Both should cache effectively
		assert.Equal(t, count1, count2, "Both orders should cache")
	})
}

// TestBuildAll tests building multiple configs in parallel.
func TestBuildAll(t *testing.T) {
	t.Run("builds all configs successfully", func(t *testing.T) {
		registry := Registry{
			"go1.x": func(_ context.Context, cfg Config) E.Either[error, Artifact] {
				return E.Right[error](Artifact{Path: cfg.SourceDir + "/bootstrap"})
			},
			"python3.11": func(_ context.Context, cfg Config) E.Either[error, Artifact] {
				return E.Right[error](Artifact{Path: cfg.SourceDir + "/lambda.zip"})
			},
		}

		configs := []Config{
			{SourceDir: "/api", Runtime: "go1.x"},
			{SourceDir: "/worker", Runtime: "python3.11"},
		}

		result := BuildAll(t.Context(), configs, registry)

		require.True(t, E.IsRight(result), "Should succeed")

		// Extract artifacts from Either using pattern matching
		var artifacts []Artifact
		if E.IsRight(result) {
			opt := E.ToOption(result)
			artifacts = O.GetOrElse(func() []Artifact { return nil })(opt)
		}
		assert.Len(t, artifacts, 2, "Should build 2 artifacts")
	})

	t.Run("fails if any build fails", func(t *testing.T) {
		registry := Registry{
			"go1.x": func(_ context.Context, _ Config) E.Either[error, Artifact] {
				return E.Left[Artifact](errors.New("go build failed"))
			},
		}

		configs := []Config{
			{SourceDir: "/api", Runtime: "go1.x"},
		}

		result := BuildAll(t.Context(), configs, registry)

		assert.True(t, E.IsLeft(result), "Should fail if any build fails")
	})

	t.Run("fails if runtime not found", func(t *testing.T) {
		registry := NewRegistry()

		configs := []Config{
			{SourceDir: "/api", Runtime: "rust"}, // Unsupported
		}

		result := BuildAll(t.Context(), configs, registry)

		assert.True(t, E.IsLeft(result), "Should fail for unsupported runtime")
	})
}

// Mock implementations for testing.

// mockLogger is a mock logger implementation for testing.
type mockLogger struct {
	infoFn  func(_ string, _ ...interface{})
	errorFn func(_ string, _ ...interface{})
}

// MemoryCache is a simple in-memory cache for testing.
type MemoryCache struct {
	cache map[string]Artifact
}

func (m *mockLogger) Info(msg string, _args ...interface{}) {
	if m.infoFn != nil {
		m.infoFn(msg, _args...)
	}
}

func (m *mockLogger) Error(msg string, _args ...interface{}) {
	if m.errorFn != nil {
		m.errorFn(msg, _args...)
	}
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		cache: make(map[string]Artifact),
	}
}

func (c *MemoryCache) Get(cfg Config) (Artifact, bool) {
	key := fmt.Sprintf("%s-%s", cfg.SourceDir, cfg.Runtime)
	artifact, ok := c.cache[key]
	return artifact, ok
}

func (c *MemoryCache) Set(cfg Config, artifact Artifact) {
	key := fmt.Sprintf("%s-%s", cfg.SourceDir, cfg.Runtime)
	c.cache[key] = artifact
}

// BenchmarkBuildFunctions benchmarks different build patterns.
func BenchmarkBuildFunctions(b *testing.B) {
	mockBuild := func(_ context.Context, _ Config) E.Either[error, Artifact] {
		return E.Right[error](Artifact{Path: "/test"})
	}

	b.Run("Plain", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mockBuild(b.Context(), Config{})
		}
	})

	b.Run("WithCache", func(b *testing.B) {
		cache := NewMemoryCache()
		cached := WithCache(cache)(mockBuild)
		cfg := Config{SourceDir: "/test"}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cached(b.Context(), cfg)
		}
	})

	b.Run("WithLogging", func(b *testing.B) {
		logger := &mockLogger{
			infoFn:  func(_ string, _ ...interface{}) {},
			errorFn: func(_ string, _ ...interface{}) {},
		}
		logged := WithLogging(logger)(mockBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logged(b.Context(), Config{})
		}
	})

	b.Run("Composed", func(b *testing.B) {
		cache := NewMemoryCache()
		logger := &mockLogger{
			infoFn:  func(_ string, _ ...interface{}) {},
			errorFn: func(_ string, _ ...interface{}) {},
		}

		composed := Compose(
			WithCache(cache),
			WithLogging(logger),
		)(mockBuild)

		cfg := Config{SourceDir: "/test"}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			composed(b.Context(), cfg)
		}
	})
}
