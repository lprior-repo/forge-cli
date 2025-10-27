package terraform

import (
	"fmt"
	"os/exec"
	"strings"
)

// ExitError wraps exec.ExitError with Terraform context
type ExitError struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("terraform %s failed with exit code %d\nstderr: %s",
		e.Command, e.ExitCode, e.Stderr)
}

// Unwrap for errors.Is and errors.As
func (e *ExitError) Unwrap() error {
	return &exec.ExitError{}
}

// StateLockError represents a Terraform state lock error
type StateLockError struct {
	Message string
	LockID  string
}

func (e *StateLockError) Error() string {
	return fmt.Sprintf("terraform state locked: %s (lock ID: %s)", e.Message, e.LockID)
}

// NoChangesError indicates terraform detected no changes
type NoChangesError struct{}

func (e *NoChangesError) Error() string {
	return "no changes detected"
}

// ValidationError represents a terraform validation error
type ValidationError struct {
	Message string
	File    string
	Line    int
}

func (e *ValidationError) Error() string {
	if e.File != "" && e.Line > 0 {
		return fmt.Sprintf("validation error at %s:%d: %s", e.File, e.Line, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ParseTerraformError attempts to parse terraform-specific errors from stderr
func ParseTerraformError(stderr string, exitCode int, command string) error {
	// Check for state lock errors
	if strings.Contains(stderr, "Error locking state") || strings.Contains(stderr, "Error acquiring the state lock") {
		lockID := extractLockID(stderr)
		return &StateLockError{
			Message: stderr,
			LockID:  lockID,
		}
	}

	// Check for no changes
	if strings.Contains(stderr, "No changes") || strings.Contains(stderr, "no changes") {
		return &NoChangesError{}
	}

	// Check for validation errors
	if strings.Contains(stderr, "Error: Invalid") || strings.Contains(stderr, "Error: Unsupported") {
		return &ValidationError{
			Message: stderr,
		}
	}

	// Generic exit error
	return &ExitError{
		Command:  command,
		ExitCode: exitCode,
		Stderr:   stderr,
	}
}

// extractLockID attempts to extract the lock ID from error message
func extractLockID(stderr string) string {
	// Simple extraction - can be enhanced with regex
	if idx := strings.Index(stderr, "ID: "); idx != -1 {
		start := idx + 4
		if end := strings.IndexAny(stderr[start:], "\n\r"); end != -1 {
			return strings.TrimSpace(stderr[start : start+end])
		}
	}
	return ""
}
