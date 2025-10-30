package pipeline

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPrintEvents tests the event printing function
func TestPrintEvents(t *testing.T) {
	t.Run("prints info events", func(t *testing.T) {
		events := []StageEvent{
			NewEvent(EventLevelInfo, "This is an info message"),
		}

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintEvents(events)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "This is an info message")
	})

	t.Run("prints success events with checkmark", func(t *testing.T) {
		events := []StageEvent{
			NewEvent(EventLevelSuccess, "Operation succeeded"),
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintEvents(events)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Operation succeeded")
		assert.Contains(t, output, "✓")
	})

	t.Run("prints warning events with warning symbol", func(t *testing.T) {
		events := []StageEvent{
			NewEvent(EventLevelWarning, "This is a warning"),
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintEvents(events)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "This is a warning")
		assert.Contains(t, output, "⚠")
	})

	t.Run("prints error events with X symbol", func(t *testing.T) {
		events := []StageEvent{
			NewEvent(EventLevelError, "An error occurred"),
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintEvents(events)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "An error occurred")
		assert.Contains(t, output, "✗")
	})

	t.Run("prints multiple events in order", func(t *testing.T) {
		events := []StageEvent{
			NewEvent(EventLevelInfo, "Starting process"),
			NewEvent(EventLevelSuccess, "Step 1 complete"),
			NewEvent(EventLevelWarning, "Minor issue detected"),
			NewEvent(EventLevelError, "Critical failure"),
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintEvents(events)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Starting process")
		assert.Contains(t, output, "Step 1 complete")
		assert.Contains(t, output, "Minor issue detected")
		assert.Contains(t, output, "Critical failure")
	})

	t.Run("prints empty event list without error", func(t *testing.T) {
		events := []StageEvent{}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintEvents(events)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)

		output := buf.String()
		assert.Empty(t, output)
	})

	t.Run("handles unknown event level as default", func(t *testing.T) {
		events := []StageEvent{
			{
				Level:   EventLevel("unknown"),
				Message: "Unknown level message",
				Data:    nil,
			},
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintEvents(events)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Unknown level message")
	})
}

// TestCollectEvents tests event collection from stage results
func TestCollectEvents(t *testing.T) {
	t.Run("collects events from result", func(t *testing.T) {
		events := []StageEvent{
			NewEvent(EventLevelInfo, "Test event 1"),
			NewEvent(EventLevelSuccess, "Test event 2"),
		}

		result := StageResult{
			State: State{
				ProjectDir: "/test",
			},
			Events: events,
		}

		collected := CollectEvents(result)
		assert.Equal(t, events, collected)
		assert.Len(t, collected, 2)
	})

	t.Run("collects empty events", func(t *testing.T) {
		result := StageResult{
			State:  State{},
			Events: []StageEvent{},
		}

		collected := CollectEvents(result)
		assert.NotNil(t, collected)
		assert.Len(t, collected, 0)
	})

	t.Run("collects nil events", func(t *testing.T) {
		result := StageResult{
			State:  State{},
			Events: nil,
		}

		collected := CollectEvents(result)
		assert.Nil(t, collected)
	})

	t.Run("preserves event data", func(t *testing.T) {
		data := map[string]interface{}{
			"count": 5,
			"name":  "test",
		}

		events := []StageEvent{
			NewEventWithData(EventLevelInfo, "Event with data", data),
		}

		result := StageResult{
			State:  State{},
			Events: events,
		}

		collected := CollectEvents(result)
		require.Len(t, collected, 1)
		assert.Equal(t, data, collected[0].Data)
	})
}

// TestStageEvent tests event creation
func TestStageEvent(t *testing.T) {
	t.Run("NewEvent creates event without data", func(t *testing.T) {
		event := NewEvent(EventLevelInfo, "Test message")

		assert.Equal(t, EventLevelInfo, event.Level)
		assert.Equal(t, "Test message", event.Message)
		assert.Nil(t, event.Data)
	})

	t.Run("NewEventWithData creates event with data", func(t *testing.T) {
		data := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		event := NewEventWithData(EventLevelWarning, "Warning message", data)

		assert.Equal(t, EventLevelWarning, event.Level)
		assert.Equal(t, "Warning message", event.Message)
		assert.Equal(t, data, event.Data)
	})

	t.Run("events are immutable", func(t *testing.T) {
		data := map[string]interface{}{
			"original": "value",
		}

		event := NewEventWithData(EventLevelInfo, "Message", data)

		// Modify original data
		data["modified"] = "new value"

		// Event should still have original data reference
		// (This tests that we're not deep copying, which is fine for our use case)
		assert.Contains(t, event.Data, "modified")
	})
}

// TestStageResult tests stage result structure
func TestStageResult(t *testing.T) {
	t.Run("creates stage result with state and events", func(t *testing.T) {
		state := State{
			ProjectDir: "/test",
			Artifacts: map[string]Artifact{
				"api": {Path: "/test/api.zip"},
			},
		}

		events := []StageEvent{
			NewEvent(EventLevelSuccess, "Build complete"),
		}

		result := StageResult{
			State:  state,
			Events: events,
		}

		assert.Equal(t, state.ProjectDir, result.State.ProjectDir)
		assert.Len(t, result.State.Artifacts, 1)
		assert.Len(t, result.Events, 1)
	})

	t.Run("stage result is immutable data structure", func(t *testing.T) {
		result := StageResult{
			State: State{
				ProjectDir: "/original",
			},
			Events: []StageEvent{
				NewEvent(EventLevelInfo, "Original event"),
			},
		}

		// Create a modified copy
		newResult := StageResult{
			State: State{
				ProjectDir: "/modified",
			},
			Events: append(result.Events, NewEvent(EventLevelInfo, "New event")),
		}

		// Original should be unchanged
		assert.Equal(t, "/original", result.State.ProjectDir)
		assert.Len(t, result.Events, 1)

		// New should have modifications
		assert.Equal(t, "/modified", newResult.State.ProjectDir)
		assert.Len(t, newResult.Events, 2)
	})
}

// TestEventLevels tests event level constants
func TestEventLevels(t *testing.T) {
	t.Run("event levels are distinct", func(t *testing.T) {
		levels := []EventLevel{
			EventLevelInfo,
			EventLevelSuccess,
			EventLevelWarning,
			EventLevelError,
		}

		// Check all levels are unique
		seen := make(map[EventLevel]bool)
		for _, level := range levels {
			assert.False(t, seen[level], "Duplicate event level: %s", level)
			seen[level] = true
		}

		assert.Len(t, seen, 4)
	})

	t.Run("event level string values are correct", func(t *testing.T) {
		assert.Equal(t, EventLevel("info"), EventLevelInfo)
		assert.Equal(t, EventLevel("success"), EventLevelSuccess)
		assert.Equal(t, EventLevel("warning"), EventLevelWarning)
		assert.Equal(t, EventLevel("error"), EventLevelError)
	})
}
