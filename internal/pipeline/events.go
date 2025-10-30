package pipeline

import "fmt"

// EventLevel represents the severity/type of an event
type EventLevel string

const (
	// EventLevelInfo represents informational events
	EventLevelInfo EventLevel = "info"
	// EventLevelSuccess represents successful operation events
	EventLevelSuccess EventLevel = "success"
	// EventLevelWarning represents warning events
	EventLevelWarning EventLevel = "warning"
	// EventLevelError represents error events
	EventLevelError EventLevel = "error"
)

// StageEvent represents an event that occurred during pipeline execution
// PURE: Immutable data structure
type StageEvent struct {
	Level   EventLevel             // Event severity level
	Message string                 // Human-readable message
	Data    map[string]interface{} // Optional structured data
}

// StageResult combines state with events emitted during stage execution
// PURE: Immutable data structure
type StageResult struct {
	State  State        // The transformed state
	Events []StageEvent // Events emitted during execution
}

// NewEvent creates a new stage event
// PURE: Constructor function
func NewEvent(level EventLevel, message string) StageEvent {
	return StageEvent{
		Level:   level,
		Message: message,
		Data:    nil,
	}
}

// NewEventWithData creates a new stage event with structured data
// PURE: Constructor function
func NewEventWithData(level EventLevel, message string, data map[string]interface{}) StageEvent {
	return StageEvent{
		Level:   level,
		Message: message,
		Data:    data,
	}
}

// PrintEvents renders events to stdout
// ACTION: Performs I/O (console output)
func PrintEvents(events []StageEvent) {
	for _, event := range events {
		switch event.Level {
		case EventLevelInfo:
			fmt.Println(event.Message)
		case EventLevelSuccess:
			fmt.Printf("✓ %s\n", event.Message)
		case EventLevelWarning:
			fmt.Printf("⚠ %s\n", event.Message)
		case EventLevelError:
			fmt.Printf("✗ %s\n", event.Message)
		default:
			fmt.Println(event.Message)
		}
	}
}

// CollectEvents extracts all events from a stage result
// PURE: Data extraction
func CollectEvents(result StageResult) []StageEvent {
	return result.Events
}
