package pipeline

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"

	"github.com/lewis/forge/internal/build"
	"github.com/lewis/forge/internal/discovery"
)

// PURE: Returns events as data instead of printing to console.
func ConventionScanV2() EventStage {
	return func(ctx context.Context, s State) E.Either[error, StageResult] {
		// Pure functional call - no OOP
		functions, err := discovery.ScanFunctions(s.ProjectDir)
		if err != nil {
			return E.Left[StageResult](fmt.Errorf("failed to scan functions: %w", err))
		}

		if len(functions) == 0 {
			return E.Left[StageResult](errors.New("no functions found in src/functions/"))
		}

		// Build events (pure data)
		events := []StageEvent{
			NewEvent(EventLevelInfo, "==> Scanning for Lambda functions..."),
			NewEvent(EventLevelInfo, fmt.Sprintf("Found %d function(s):", len(functions))),
		}

		for _, fn := range functions {
			events = append(events, NewEvent(EventLevelInfo, fmt.Sprintf("  - %s (%s)", fn.Name, fn.Runtime)))
		}
		events = append(events, NewEvent(EventLevelInfo, ""))

		// Create new state (immutable)
		newState := State{
			ProjectDir: s.ProjectDir,
			Artifacts:  s.Artifacts,
			Outputs:    s.Outputs,
			Config:     functions,
		}

		return E.Right[error](StageResult{
			State:  newState,
			Events: events,
		})
	}
}

// ConventionStubsV2 creates an event-based stage that generates stub zip files.
func ConventionStubsV2() EventStage {
	return func(ctx context.Context, s State) E.Either[error, StageResult] {
		// Extract functions from state
		functions, ok := s.Config.([]discovery.Function)
		if !ok {
			return E.Left[StageResult](errors.New("invalid state: functions not found"))
		}

		buildDir := filepath.Join(s.ProjectDir, ".forge", "build")

		count, err := discovery.CreateStubZips(functions, buildDir)
		if err != nil {
			return E.Left[StageResult](fmt.Errorf("failed to create stub zips: %w", err))
		}

		// Build events
		events := []StageEvent{}
		if count > 0 {
			events = append(events, NewEvent(EventLevelInfo, fmt.Sprintf("Created %d stub zip(s)\n", count)))
		}

		return E.Right[error](StageResult{
			State:  s,
			Events: events,
		})
	}
}

// ConventionBuildV2 creates an event-based stage that builds all discovered functions.
func ConventionBuildV2() EventStage {
	return func(ctx context.Context, s State) E.Either[error, StageResult] {
		// Extract functions from state
		functions, ok := s.Config.([]discovery.Function)
		if !ok {
			return E.Left[StageResult](errors.New("invalid state: functions not found"))
		}

		registry := build.NewRegistry()
		buildDir := filepath.Join(s.ProjectDir, ".forge", "build")

		// Build events
		events := []StageEvent{
			NewEvent(EventLevelInfo, "==> Building Lambda functions..."),
		}

		// Build new artifacts map (immutable)
		newArtifacts := make(map[string]Artifact)
		// Copy existing artifacts
		for k, v := range s.Artifacts {
			newArtifacts[k] = v
		}

		// Build each function and add to new artifacts map
		for _, fn := range functions {
			events = append(events, NewEvent(EventLevelInfo, fmt.Sprintf("[%s] Building...", fn.Name)))

			// Get builder from registry (returns Option)
			builderOpt := build.GetBuilder(registry, fn.Runtime)
			if O.IsNone(builderOpt) {
				return E.Left[StageResult](fmt.Errorf("unsupported runtime: %s", fn.Runtime))
			}

			// Extract builder using Fold
			builder := O.Fold(
				func() build.BuildFunc { return nil },
				func(b build.BuildFunc) build.BuildFunc { return b },
			)(builderOpt)

			// Convert to build config with validation (returns Either)
			cfgResult := discovery.ToBuildConfig(fn, buildDir)

			// Handle config validation error
			if E.IsLeft(cfgResult) {
				err := E.Fold(
					func(e error) error { return e },
					func(c build.Config) error { return nil },
				)(cfgResult)
				return E.Left[StageResult](fmt.Errorf("invalid build config for %s: %w", fn.Name, err))
			}

			// Extract config
			cfg := E.Fold(
				func(error) build.Config { return build.Config{} },
				func(c build.Config) build.Config { return c },
			)(cfgResult)

			// Execute build (returns Either)
			result := builder(ctx, cfg)

			// Handle result using functional error handling
			if E.IsLeft(result) {
				err := E.Fold(
					func(e error) error { return e },
					func(a build.Artifact) error { return nil },
				)(result)
				return E.Left[StageResult](fmt.Errorf("failed to build %s: %w", fn.Name, err))
			}

			// Extract artifact
			artifact := E.Fold(
				func(e error) build.Artifact { return build.Artifact{} },
				func(a build.Artifact) build.Artifact { return a },
			)(result)

			// Add to new artifacts map (immutable)
			newArtifacts[fn.Name] = Artifact{
				Path:     artifact.Path,
				Checksum: artifact.Checksum,
				Size:     artifact.Size,
			}

			sizeMB := float64(artifact.Size) / 1024 / 1024
			events = append(events, NewEvent(EventLevelSuccess, fmt.Sprintf("[%s] Built: %s (%.2f MB)", fn.Name, filepath.Base(artifact.Path), sizeMB)))
		}

		events = append(events, NewEvent(EventLevelInfo, ""))

		// Return new State (immutable)
		return E.Right[error](StageResult{
			State: State{
				ProjectDir: s.ProjectDir,
				Artifacts:  newArtifacts,
				Outputs:    s.Outputs,
				Config:     s.Config,
			},
			Events: events,
		})
	}
}
