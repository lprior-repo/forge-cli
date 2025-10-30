package pipeline

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
)

// State carries data through the pipeline
type State struct {
	ProjectDir string
	Artifacts  map[string]Artifact
	Outputs    map[string]interface{}
	Config     interface{}
}

// Artifact represents a built artifact
type Artifact struct {
	Path     string
	Checksum string
	Size     int64
}

// Stage is a function that transforms state
// Uses Either monad for error handling
type Stage func(context.Context, State) E.Either[error, State]

// Pipeline composes stages functionally
type Pipeline struct {
	stages []Stage
}

// New creates a new pipeline from stages
func New(stages ...Stage) Pipeline {
	return Pipeline{stages: stages}
}

// Run executes all stages in order using functional composition
// Pure function approach - pipeline is immutable data
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
			return E.Left[State](fmt.Errorf("unexpected None in pipeline"))
		}

		state := O.GetOrElse(func() State { return State{} })(opt)
		result = stage(ctx, state)
	}

	return result
}

// Chain composes multiple pipelines into one
func Chain(pipelines ...Pipeline) Pipeline {
	var stages []Stage
	for _, p := range pipelines {
		stages = append(stages, p.stages...)
	}
	return Pipeline{stages: stages}
}

// Parallel runs stages in parallel and combines results
// Uses fp-go's parallel execution patterns
func Parallel(stages ...Stage) Stage {
	return func(ctx context.Context, s State) E.Either[error, State] {
		// TODO: Implement true parallel execution using goroutines and channels
		// For now, run sequentially
		result := E.Right[error](s)
		for _, stage := range stages {
			if E.IsLeft(result) {
				return result
			}

			opt := E.ToOption(result)
			state := O.GetOrElse(func() State { return State{} })(opt)
			result = stage(ctx, state)
		}
		return result
	}
}
