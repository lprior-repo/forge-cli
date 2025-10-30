package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExitError tests ExitError error formatting.
func TestExitError(t *testing.T) {
	t.Run("Error method formats correctly", func(t *testing.T) {
		err := &ExitError{
			Command:  "apply",
			ExitCode: 1,
			Stdout:   "some output",
			Stderr:   "error message",
		}

		expected := "terraform apply failed with exit code 1\nstderr: error message"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Unwrap returns exec.ExitError", func(t *testing.T) {
		err := &ExitError{
			Command:  "plan",
			ExitCode: 2,
		}

		unwrapped := err.Unwrap()
		assert.Error(t, unwrapped)
	})
}

// TestStateLockError tests StateLockError.
func TestStateLockError(t *testing.T) {
	t.Run("Error method includes lock ID", func(t *testing.T) {
		err := &StateLockError{
			Message: "state is locked",
			LockID:  "abc123",
		}

		errMsg := err.Error()
		assert.Contains(t, errMsg, "terraform state locked")
		assert.Contains(t, errMsg, "abc123")
	})
}

// TestNoChangesError tests NoChangesError.
func TestNoChangesError(t *testing.T) {
	t.Run("Error method returns correct message", func(t *testing.T) {
		err := &NoChangesError{}
		assert.Equal(t, "no changes detected", err.Error())
	})
}

// TestValidationError tests ValidationError.
func TestValidationError(t *testing.T) {
	t.Run("Error with file and line", func(t *testing.T) {
		err := &ValidationError{
			Message: "invalid syntax",
			File:    "main.tf",
			Line:    42,
		}

		errMsg := err.Error()
		assert.Contains(t, errMsg, "validation error at main.tf:42")
		assert.Contains(t, errMsg, "invalid syntax")
	})

	t.Run("Error without file location", func(t *testing.T) {
		err := &ValidationError{
			Message: "invalid configuration",
		}

		errMsg := err.Error()
		assert.Equal(t, "validation error: invalid configuration", errMsg)
	})

	t.Run("Error with file but no line", func(t *testing.T) {
		err := &ValidationError{
			Message: "invalid configuration",
			File:    "main.tf",
			Line:    0,
		}

		errMsg := err.Error()
		assert.Equal(t, "validation error: invalid configuration", errMsg)
	})

	t.Run("Error with file and line 1", func(t *testing.T) {
		err := &ValidationError{
			Message: "syntax error",
			File:    "main.tf",
			Line:    1,
		}

		errMsg := err.Error()
		assert.Contains(t, errMsg, "validation error at main.tf:1")
		assert.Contains(t, errMsg, "syntax error")
	})
}

// TestParseTerraformError tests error parsing logic.
func TestParseTerraformError(t *testing.T) {
	t.Run("Detects state lock error with Error locking state", func(t *testing.T) {
		stderr := `Error locking state: some message
ID: 12345-abcde-67890
more details`

		err := ParseTerraformError(stderr, 1, "apply")

		lockErr, ok := err.(*StateLockError)
		assert.True(t, ok, "Should return StateLockError")
		assert.Equal(t, "12345-abcde-67890", lockErr.LockID)
	})

	t.Run("Detects state lock error with Error acquiring", func(t *testing.T) {
		stderr := `Error acquiring the state lock
ID: xyz-123
details`

		err := ParseTerraformError(stderr, 1, "plan")

		lockErr, ok := err.(*StateLockError)
		assert.True(t, ok, "Should return StateLockError")
		assert.Equal(t, "xyz-123", lockErr.LockID)
	})

	t.Run("Detects no changes error", func(t *testing.T) {
		stderr := "No changes. Your infrastructure matches the configuration."

		err := ParseTerraformError(stderr, 0, "plan")

		_, ok := err.(*NoChangesError)
		assert.True(t, ok, "Should return NoChangesError")
	})

	t.Run("Detects no changes with lowercase", func(t *testing.T) {
		stderr := "terraform detected no changes to your infrastructure"

		err := ParseTerraformError(stderr, 0, "plan")

		_, ok := err.(*NoChangesError)
		assert.True(t, ok, "Should return NoChangesError")
	})

	t.Run("Detects validation error with Invalid", func(t *testing.T) {
		stderr := "Error: Invalid configuration syntax"

		err := ParseTerraformError(stderr, 1, "validate")

		valErr, ok := err.(*ValidationError)
		assert.True(t, ok, "Should return ValidationError")
		assert.Contains(t, valErr.Message, "Invalid")
	})

	t.Run("Detects validation error with Unsupported", func(t *testing.T) {
		stderr := "Error: Unsupported attribute"

		err := ParseTerraformError(stderr, 1, "validate")

		valErr, ok := err.(*ValidationError)
		assert.True(t, ok, "Should return ValidationError")
		assert.Contains(t, valErr.Message, "Unsupported")
	})

	t.Run("Returns generic ExitError for unknown errors", func(t *testing.T) {
		stderr := "Some generic terraform error"

		err := ParseTerraformError(stderr, 1, "apply")

		exitErr, ok := err.(*ExitError)
		assert.True(t, ok, "Should return ExitError")
		assert.Equal(t, 1, exitErr.ExitCode)
		assert.Equal(t, "apply", exitErr.Command)
		assert.Equal(t, stderr, exitErr.Stderr)
	})
}

// TestExtractLockID tests lock ID extraction.
func TestExtractLockID(t *testing.T) {
	t.Run("Extracts lock ID correctly", func(t *testing.T) {
		stderr := `Error locking state
ID: 12345-abc-67890
More text`

		lockID := extractLockID(stderr)
		assert.Equal(t, "12345-abc-67890", lockID)
	})

	t.Run("Returns empty string when ID not found", func(t *testing.T) {
		stderr := "Error locking state\nNo ID present"

		lockID := extractLockID(stderr)
		assert.Equal(t, "", lockID)
	})

	t.Run("Handles ID at end of string without newline", func(t *testing.T) {
		stderr := "ID: test-lock-id"

		lockID := extractLockID(stderr)
		// Function returns empty if no newline found
		assert.Equal(t, "", lockID)
	})

	t.Run("Handles ID at end with newline", func(t *testing.T) {
		stderr := "ID: test-lock-id\n"

		lockID := extractLockID(stderr)
		assert.Equal(t, "test-lock-id", lockID)
	})

	t.Run("Trims whitespace from lock ID", func(t *testing.T) {
		stderr := "ID:   lock-with-spaces  \n"

		lockID := extractLockID(stderr)
		assert.Equal(t, "lock-with-spaces", lockID)
	})
}
