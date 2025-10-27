package build

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
	"github.com/samber/lo"
)

// BuildFunc is the core abstraction - a pure function that builds an artifact
type BuildFunc func(context.Context, Config) E.Either[error, Artifact]

// Registry is a map of runtime to build function
type Registry map[string]BuildFunc

// NewRegistry creates a functional builder registry
func NewRegistry() Registry {
	return Registry{
		"go1.x":       GoBuild,
		"python3.11":  PythonBuild,
		"python3.12":  PythonBuild,
		"nodejs20.x":  NodeBuild,
		"nodejs18.x":  NodeBuild,
	}
}

// Get retrieves a builder for the given runtime (returns Option type)
func (r Registry) Get(runtime string) O.Option[BuildFunc] {
	if builder, ok := r[runtime]; ok {
		return O.Some(builder)
	}
	return O.None[BuildFunc]()
}

// BuildAll builds multiple configs in parallel using functional patterns
func BuildAll(ctx context.Context, configs []Config, registry Registry) E.Either[error, []Artifact] {
	// Map configs to build results
	results := lo.Map(configs, func(cfg Config, _ int) E.Either[error, Artifact] {
		builderOpt := registry.Get(cfg.Runtime)

		return O.Fold(
			// None case: runtime not found
			func() E.Either[error, Artifact] {
				return E.Left[Artifact](fmt.Errorf("unsupported runtime: %s", cfg.Runtime))
			},
			// Some case: execute builder
			func(builder BuildFunc) E.Either[error, Artifact] {
				return builder(ctx, cfg)
			},
		)(builderOpt)
	})

	// Check if any failed
	failures := lo.Filter(results, func(r E.Either[error, Artifact], _ int) bool {
		return E.IsLeft(r)
	})

	if len(failures) > 0 {
		// Extract first error
		firstError := E.Fold(
			func(err error) error { return err },
			func(a Artifact) error { return nil },
		)(failures[0])
		return E.Left[[]Artifact](firstError)
	}

	// Extract all artifacts
	artifacts := lo.Map(results, func(r E.Either[error, Artifact], _ int) Artifact {
		return E.Fold(
			func(err error) Artifact { return Artifact{} },
			func(a Artifact) Artifact { return a },
		)(r)
	})

	return E.Right[error](artifacts)
}

// WithCache is a higher-order function that adds caching to a builder
func WithCache(cache Cache) func(BuildFunc) BuildFunc {
	return func(build BuildFunc) BuildFunc {
		return func(ctx context.Context, cfg Config) E.Either[error, Artifact] {
			// Check cache
			if artifact, ok := cache.Get(cfg); ok {
				return E.Right[error](artifact)
			}

			// Build
			result := build(ctx, cfg)

			// Store in cache if successful
			if E.IsRight(result) {
				artifact := E.ToOption(result)
				if O.IsSome(artifact) {
					cache.Set(cfg, O.GetOrElse(func() Artifact { return Artifact{} })(artifact))
				}
			}

			return result
		}
	}
}

// WithLogging is a higher-order function that adds logging to a builder
func WithLogging(log Logger) func(BuildFunc) BuildFunc {
	return func(build BuildFunc) BuildFunc {
		return func(ctx context.Context, cfg Config) E.Either[error, Artifact] {
			log.Info("Building", "runtime", cfg.Runtime, "source", cfg.SourceDir)

			result := build(ctx, cfg)

			// Log based on result
			if E.IsLeft(result) {
				log.Error("Build failed", "error", "build error")
			} else {
				log.Info("Build succeeded")
			}

			return result
		}
	}
}

// Compose multiple decorators functionally
func Compose(decorators ...func(BuildFunc) BuildFunc) func(BuildFunc) BuildFunc {
	return func(build BuildFunc) BuildFunc {
		// Apply decorators right-to-left (like mathematical composition)
		return lo.Reduce(lo.Reverse(decorators), func(acc BuildFunc, decorator func(BuildFunc) BuildFunc, _ int) BuildFunc {
			return decorator(acc)
		}, build)
	}
}

// Cache interface for caching build artifacts
type Cache interface {
	Get(Config) (Artifact, bool)
	Set(Config, Artifact)
}

// Logger interface for logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}
