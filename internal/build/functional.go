package build

import (
	"context"
	"fmt"

	A "github.com/IBM/fp-go/array"
	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
	"github.com/samber/lo"
)

// BuildFunc is the core abstraction - a pure function that builds an artifact
type BuildFunc func(context.Context, Config) E.Either[error, Artifact]

// Registry is a map of runtime to build function
type Registry map[string]BuildFunc

// NewRegistry creates a functional builder registry with latest Lambda runtimes
func NewRegistry() Registry {
	return Registry{
		// Go runtimes (use provided.al2023 for latest)
		"go1.x":           GoBuild,
		"provided.al2":    GoBuild,
		"provided.al2023": GoBuild,

		// Python runtimes (3.9 to 3.13)
		"python3.9":  PythonBuild,
		"python3.10": PythonBuild,
		"python3.11": PythonBuild,
		"python3.12": PythonBuild,
		"python3.13": PythonBuild,

		// Node.js runtimes (18.x, 20.x, 22.x)
		"nodejs18.x": NodeBuild,
		"nodejs20.x": NodeBuild,
		"nodejs22.x": NodeBuild,

		// Java runtimes (11, 17, 21)
		"java11": JavaBuild,
		"java17": JavaBuild,
		"java21": JavaBuild,
	}
}

// GetBuilder retrieves a builder for the given runtime (returns Option type)
// Pure function - no methods, takes Registry as parameter
func GetBuilder(r Registry, runtime string) O.Option[BuildFunc] {
	if builder, ok := r[runtime]; ok {
		return O.Some(builder)
	}
	return O.None[BuildFunc]()
}

// BuildAll builds multiple configs using functional sequence pattern
// Uses Array.Reduce to convert []Either[error, Artifact] to Either[error, []Artifact]
// Automatically short-circuits on first error (railway-oriented programming)
func BuildAll(ctx context.Context, configs []Config, registry Registry) E.Either[error, []Artifact] {
	// Map configs to build results ([]Either[error, Artifact])
	buildResults := A.Map(func(cfg Config) E.Either[error, Artifact] {
		builderOpt := GetBuilder(registry, cfg.Runtime)

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
	})(configs)

	// Use functional sequence to convert []Either[error, Artifact] → Either[error, []Artifact]
	// This automatically short-circuits on first error
	return sequenceEithers(buildResults)
}

// sequenceEithers converts []Either[E, A] to Either[E, []A]
// Short-circuits on first Left (error), otherwise collects all Right values
// PURE: Calculation - deterministic transformation
func sequenceEithers(eithers []E.Either[error, Artifact]) E.Either[error, []Artifact] {
	return A.Reduce(
		func(acc E.Either[error, []Artifact], item E.Either[error, Artifact]) E.Either[error, []Artifact] {
			// Use E.Chain for railway-oriented programming
			// If acc is Left (error), it stays Left
			// If acc is Right and item is Left, it becomes Left
			// If both are Right, append item to acc
			return E.Chain(func(artifacts []Artifact) E.Either[error, []Artifact] {
				return E.Map[error](func(artifact Artifact) []Artifact {
					return append(artifacts, artifact)
				})(item)
			})(acc)
		},
		E.Right[error]([]Artifact{}), // Start with empty Right
	)(eithers)
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

			// Store in cache if successful (use Fold to handle both cases)
			E.Fold(
				func(err error) error {
					// Left case (error) - do nothing
					return err
				},
				func(artifact Artifact) error {
					// Right case (success) - cache the artifact
					cache.Set(cfg, artifact)
					return nil
				},
			)(result)

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

			// Log based on result using Fold
			E.Fold(
				func(err error) error {
					log.Error("Build failed", "error", "build error")
					return err
				},
				func(artifact Artifact) error {
					log.Info("Build succeeded")
					return nil
				},
			)(result)

			return result
		}
	}
}

// Compose multiple decorators functionally
func Compose(decorators ...func(BuildFunc) BuildFunc) func(BuildFunc) BuildFunc {
	return func(build BuildFunc) BuildFunc {
		// Reverse decorators for right-to-left composition (like mathematical composition)
		reversed := make([]func(BuildFunc) BuildFunc, len(decorators))
		for i := range decorators {
			reversed[i] = decorators[len(decorators)-1-i]
		}

		// Apply decorators right-to-left
		return lo.Reduce(reversed, func(acc BuildFunc, decorator func(BuildFunc) BuildFunc, _ int) BuildFunc {
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
