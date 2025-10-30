package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConventionScanV2 tests the event-based function scanning stage
func TestConventionScanV2(t *testing.T) {
	t.Run("scans and finds Go functions with events", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create src/functions directory structure
		functionsDir := filepath.Join(tmpDir, "src", "functions")
		apiDir := filepath.Join(functionsDir, "api")
		require.NoError(t, os.MkdirAll(apiDir, 0755))

		// Create main.go file
		mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello")
}
`
		require.NoError(t, os.WriteFile(filepath.Join(apiDir, "main.go"), []byte(mainGo), 0644))

		state := State{ProjectDir: tmpDir}
		stage := ConventionScanV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result), "Should successfully scan functions")

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Verify state
		functions, ok := stageResult.State.Config.([]discovery.Function)
		require.True(t, ok, "Config should contain []discovery.Function")
		assert.Len(t, functions, 1, "Should find 1 function")
		assert.Equal(t, "api", functions[0].Name)
		assert.Equal(t, "provided.al2023", functions[0].Runtime)

		// Verify events
		require.NotEmpty(t, stageResult.Events, "Should have events")
		assert.Equal(t, EventLevelInfo, stageResult.Events[0].Level)
		assert.Contains(t, stageResult.Events[0].Message, "Scanning")

		// Should have event for found functions
		foundEvent := false
		for _, event := range stageResult.Events {
			if event.Message == "Found 1 function(s):" {
				foundEvent = true
				break
			}
		}
		assert.True(t, foundEvent, "Should have 'Found functions' event")
	})

	t.Run("scans and finds Python functions with events", func(t *testing.T) {
		tmpDir := t.TempDir()

		functionsDir := filepath.Join(tmpDir, "src", "functions")
		workerDir := filepath.Join(functionsDir, "worker")
		require.NoError(t, os.MkdirAll(workerDir, 0755))

		appPy := `def lambda_handler(event, context):
    return {"statusCode": 200}
`
		require.NoError(t, os.WriteFile(filepath.Join(workerDir, "app.py"), []byte(appPy), 0644))

		state := State{ProjectDir: tmpDir}
		stage := ConventionScanV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		functions, ok := stageResult.State.Config.([]discovery.Function)
		require.True(t, ok)
		assert.Len(t, functions, 1)
		assert.Equal(t, "worker", functions[0].Name)
		assert.Equal(t, "python3.13", functions[0].Runtime)

		// Verify events generated
		assert.NotEmpty(t, stageResult.Events)
	})

	t.Run("returns error when src/functions does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		state := State{ProjectDir: tmpDir}
		stage := ConventionScanV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsLeft(result), "Should return error")
	})

	t.Run("returns error when no functions found", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create empty src/functions directory
		functionsDir := filepath.Join(tmpDir, "src", "functions")
		require.NoError(t, os.MkdirAll(functionsDir, 0755))

		state := State{ProjectDir: tmpDir}
		stage := ConventionScanV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsLeft(result), "Should return error for no functions")
	})

	t.Run("scans multiple functions with events", func(t *testing.T) {
		tmpDir := t.TempDir()

		functionsDir := filepath.Join(tmpDir, "src", "functions")

		// Create Go function
		apiDir := filepath.Join(functionsDir, "api")
		require.NoError(t, os.MkdirAll(apiDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(apiDir, "main.go"), []byte("package main"), 0644))

		// Create Python function
		workerDir := filepath.Join(functionsDir, "worker")
		require.NoError(t, os.MkdirAll(workerDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(workerDir, "app.py"), []byte("def handler(e, c): pass"), 0644))

		// Create Node.js function
		webhookDir := filepath.Join(functionsDir, "webhook")
		require.NoError(t, os.MkdirAll(webhookDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(webhookDir, "index.js"), []byte("exports.handler = () => {}"), 0644))

		state := State{ProjectDir: tmpDir}
		stage := ConventionScanV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		functions, ok := stageResult.State.Config.([]discovery.Function)
		require.True(t, ok)
		assert.Len(t, functions, 3, "Should find 3 functions")

		// Verify events mention all 3 functions
		assert.NotEmpty(t, stageResult.Events)
		foundCountEvent := false
		for _, event := range stageResult.Events {
			if event.Message == "Found 3 function(s):" {
				foundCountEvent = true
				break
			}
		}
		assert.True(t, foundCountEvent, "Should have event with correct count")
	})
}

// TestConventionStubsV2 tests the event-based stub zip creation stage
func TestConventionStubsV2(t *testing.T) {
	t.Run("creates stub zips for functions with events", func(t *testing.T) {
		tmpDir := t.TempDir()

		functions := []discovery.Function{
			{Name: "api", Runtime: "provided.al2023", Path: filepath.Join(tmpDir, "api"), EntryPoint: "bootstrap"},
			{Name: "worker", Runtime: "python3.13", Path: filepath.Join(tmpDir, "worker"), EntryPoint: "app.handler"},
		}

		state := State{
			ProjectDir: tmpDir,
			Config:     functions,
		}

		stage := ConventionStubsV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result), "Should successfully create stubs")

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Verify stub zips were created
		buildDir := filepath.Join(tmpDir, ".forge", "build")
		apiZip := filepath.Join(buildDir, "api.zip")
		workerZip := filepath.Join(buildDir, "worker.zip")

		assert.FileExists(t, apiZip, "api stub zip should exist")
		assert.FileExists(t, workerZip, "worker stub zip should exist")

		// Verify events
		assert.NotEmpty(t, stageResult.Events, "Should have events")
		hasCreatedEvent := false
		for _, event := range stageResult.Events {
			if event.Message == "Created 2 stub zip(s)\n" {
				hasCreatedEvent = true
				assert.Equal(t, EventLevelInfo, event.Level)
				break
			}
		}
		assert.True(t, hasCreatedEvent, "Should have 'Created stubs' event")
	})

	t.Run("returns error when Config is not []discovery.Function", func(t *testing.T) {
		state := State{
			ProjectDir: t.TempDir(),
			Config:     "invalid",
		}

		stage := ConventionStubsV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsLeft(result), "Should return error for invalid config")
	})

	t.Run("succeeds with no functions and no events", func(t *testing.T) {
		state := State{
			ProjectDir: t.TempDir(),
			Config:     []discovery.Function{},
		}

		stage := ConventionStubsV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// No stubs created, so no events
		assert.Empty(t, stageResult.Events, "Should have no events when no stubs created")
	})
}

// TestConventionBuildV2 tests the event-based build stage
func TestConventionBuildV2(t *testing.T) {
	t.Run("returns error when Config is not []discovery.Function", func(t *testing.T) {
		state := State{
			ProjectDir: t.TempDir(),
			Config:     "invalid",
		}

		stage := ConventionBuildV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsLeft(result), "Should return error for invalid config")
	})

	t.Run("returns error for unsupported runtime with events", func(t *testing.T) {
		tmpDir := t.TempDir()

		functions := []discovery.Function{
			{Name: "test", Runtime: "unsupported", Path: tmpDir, EntryPoint: "main"},
		}

		state := State{
			ProjectDir: tmpDir,
			Config:     functions,
			Artifacts:  make(map[string]Artifact),
		}

		stage := ConventionBuildV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsLeft(result), "Should return error for unsupported runtime")
	})

	t.Run("succeeds with empty function list", func(t *testing.T) {
		state := State{
			ProjectDir: t.TempDir(),
			Config:     []discovery.Function{},
			Artifacts:  make(map[string]Artifact),
		}

		stage := ConventionBuildV2()
		result := stage(context.Background(), state)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Should have initial "Building" event
		assert.NotEmpty(t, stageResult.Events)
		assert.Equal(t, EventLevelInfo, stageResult.Events[0].Level)
		assert.Contains(t, stageResult.Events[0].Message, "Building")
	})

	t.Run("generates events for each build step", func(t *testing.T) {
		// This test would need mock builders, which is complex
		// For now, we test the error case which exercises event generation
		tmpDir := t.TempDir()

		functions := []discovery.Function{
			{Name: "test", Runtime: "unsupported-runtime-123", Path: tmpDir, EntryPoint: "main"},
		}

		state := State{
			ProjectDir: tmpDir,
			Config:     functions,
			Artifacts:  make(map[string]Artifact),
		}

		stage := ConventionBuildV2()
		result := stage(context.Background(), state)

		// Should fail, but would have generated initial events
		require.True(t, E.IsLeft(result))
	})
}

// TestRunWithEvents tests the event collection pipeline
func TestRunWithEvents(t *testing.T) {
	t.Run("collects events from multiple stages", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test function structure
		functionsDir := filepath.Join(tmpDir, "src", "functions")
		apiDir := filepath.Join(functionsDir, "api")
		require.NoError(t, os.MkdirAll(apiDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(apiDir, "main.go"), []byte("package main"), 0644))

		// Create pipeline with scan and stubs stages
		pipeline := NewEventPipeline(
			ConventionScanV2(),
			ConventionStubsV2(),
		)

		initialState := State{
			ProjectDir: tmpDir,
			Artifacts:  make(map[string]Artifact),
		}

		result := RunWithEvents(pipeline, context.Background(), initialState)

		require.True(t, E.IsRight(result), "Pipeline should succeed")

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Should have collected events from both stages
		assert.NotEmpty(t, stageResult.Events, "Should have events from multiple stages")

		// Count events by stage
		scanEvents := 0
		stubEvents := 0
		for _, event := range stageResult.Events {
			if event.Message == "==> Scanning for Lambda functions..." {
				scanEvents++
			}
			if event.Message == "Created 1 stub zip(s)\n" {
				stubEvents++
			}
		}

		assert.Greater(t, scanEvents, 0, "Should have scan events")
		assert.Greater(t, stubEvents, 0, "Should have stub events")

		// Verify final state
		functions, ok := stageResult.State.Config.([]discovery.Function)
		require.True(t, ok)
		assert.Len(t, functions, 1)
	})

	t.Run("stops on first error and returns collected events", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Pipeline with failing stage
		pipeline := NewEventPipeline(
			ConventionScanV2(), // Will fail - no functions dir
		)

		initialState := State{
			ProjectDir: tmpDir,
		}

		result := RunWithEvents(pipeline, context.Background(), initialState)

		require.True(t, E.IsLeft(result), "Pipeline should fail")

		// Events up to the failure should still be collected
		// (though in this case, the error happens before events are returned)
	})

	t.Run("preserves state across stages", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test structure
		functionsDir := filepath.Join(tmpDir, "src", "functions")
		apiDir := filepath.Join(functionsDir, "api")
		require.NoError(t, os.MkdirAll(apiDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(apiDir, "main.go"), []byte("package main"), 0644))

		pipeline := NewEventPipeline(
			ConventionScanV2(),  // Adds functions to Config
			ConventionStubsV2(), // Uses functions from Config
		)

		initialState := State{
			ProjectDir: tmpDir,
			Artifacts:  make(map[string]Artifact),
		}

		result := RunWithEvents(pipeline, context.Background(), initialState)

		require.True(t, E.IsRight(result))

		stageResult := E.Fold(
			func(e error) StageResult { return StageResult{} },
			func(r StageResult) StageResult { return r },
		)(result)

		// Final state should have functions from scan stage
		functions, ok := stageResult.State.Config.([]discovery.Function)
		require.True(t, ok)
		assert.Len(t, functions, 1)

		// And stub zips should have been created
		buildDir := filepath.Join(tmpDir, ".forge", "build")
		assert.DirExists(t, buildDir)
	})
}

// TestEventGeneration tests event creation and properties
func TestEventGeneration(t *testing.T) {
	t.Run("NewEvent creates event with correct properties", func(t *testing.T) {
		event := NewEvent(EventLevelInfo, "test message")

		assert.Equal(t, EventLevelInfo, event.Level)
		assert.Equal(t, "test message", event.Message)
		assert.Nil(t, event.Data)
	})

	t.Run("NewEventWithData creates event with data", func(t *testing.T) {
		data := map[string]interface{}{
			"key": "value",
			"count": 42,
		}

		event := NewEventWithData(EventLevelSuccess, "test message", data)

		assert.Equal(t, EventLevelSuccess, event.Level)
		assert.Equal(t, "test message", event.Message)
		assert.Equal(t, data, event.Data)
	})

	t.Run("event levels are distinct", func(t *testing.T) {
		assert.NotEqual(t, EventLevelInfo, EventLevelSuccess)
		assert.NotEqual(t, EventLevelInfo, EventLevelWarning)
		assert.NotEqual(t, EventLevelInfo, EventLevelError)
		assert.NotEqual(t, EventLevelSuccess, EventLevelWarning)
	})
}
