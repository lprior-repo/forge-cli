package pipeline

import (
	"context"
	"errors"

	A "github.com/IBM/fp-go/array"
	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
)

// State carries data through the pipeline.
type State struct {
	ProjectDir string
	Artifacts  map[string]Artifact
	Outputs    map[string]interface{}
	Config     interface{}
}

// Artifact represents a built artifact.
type Artifact struct {
	Path     string
	Checksum string
	Size     int64
}

// Uses Either monad for error handling.
type Stage func(context.Context, State) E.Either[error, State]

// Uses Either monad for error handling and returns StageResult with events.
type EventStage func(context.Context, State) E.Either[error, StageResult]

// Pipeline composes stages functionally.
type Pipeline struct {
	stages []Stage
}

// EventPipeline composes event-based stages.
type EventPipeline struct {
	stages []EventStage
}

// New creates a new pipeline from stages.
func New(stages ...Stage) Pipeline {
	return Pipeline{stages: stages}
}

// NewEventPipeline creates a new event-based pipeline from stages.
func NewEventPipeline(stages ...EventStage) EventPipeline {
	return EventPipeline{stages: stages}
}

// Pure function approach - pipeline is immutable data.
func Run(p Pipeline, ctx context.Context, initial State) E.Either[error, State] {
	// Start with initial state wrapped in Right (success)
	result := E.Right[error](initial)

	// Chain all stages - manually check and proceed
	for _, stage := range p.stages {
		if E.IsLeft(result) {
			return result // Short-circuit on error
		}

		// Extract state and run next stage
		opt := E.ToOption(result)
		if O.IsNone(opt) {
			return E.Left[State](errors.New("unexpected None in pipeline"))
		}

		state := O.GetOrElse(func() State { return State{} })(opt)
		result = stage(ctx, state)
	}

	return result
}

// PURE: Functional composition of event stages with event collection.
func RunWithEvents(p EventPipeline, ctx context.Context, initial State) E.Either[error, StageResult] {
	// Start with empty events and initial state
	allEvents := []StageEvent{}
	currentState := initial

	for _, stage := range p.stages {
		// Run the stage
		result := stage(ctx, currentState)

		// Check for errors
		if E.IsLeft(result) {
			// Return error with events collected so far
			return E.MapLeft[StageResult](func(err error) error {
				// Print events before erroring
				PrintEvents(allEvents)
				return err
			})(result)
		}

		// Extract the stage result
		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Collect events
		allEvents = append(allEvents, stageResult.Events...)

		// Update current state
		currentState = stageResult.State
	}

	// Return final result with all collected events
	return E.Right[error](StageResult{
		State:  currentState,
		Events: allEvents,
	})
}

// Chain composes multiple pipelines into one.
func Chain(pipelines ...Pipeline) Pipeline {
	var stages []Stage
	for _, p := range pipelines {
		stages = append(stages, p.stages...)
	}
	return Pipeline{stages: stages}
}

// NOTE: Future enhancement - true parallel execution with goroutines (see Parallel).
func Sequential(stages ...Stage) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		// Use A.Reduce for functional sequential composition
		return A.Reduce(
			func(acc E.Either[error, State], stage Stage) E.Either[error, State] {
				// Use E.Chain for automatic error short-circuiting
				return E.Chain(func(state State) E.Either[error, State] {
					return stage(ctx, state)
				})(acc)
			},
			E.Right[error](s),
		)(stages)
	}
}

// Parallel is an alias for Sequential (true parallel execution not yet implemented)
// DEPRECATED: Use Sequential for clarity. True parallel execution is a future enhancement.
// See: https://github.com/lewis/forge/issues/XXX
func Parallel(stages ...Stage) Stage {
	return Sequential(stages...)
}
