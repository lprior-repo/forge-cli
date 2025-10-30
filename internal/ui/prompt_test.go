package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPrompter(t *testing.T) {
	t.Run("creates prompter with reader and writer", func(t *testing.T) {
		reader := strings.NewReader("")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		assert.NotNil(t, prompter)
		assert.Equal(t, reader, prompter.reader)
		assert.Equal(t, writer, prompter.writer)
		assert.NotNil(t, prompter.output)
	})
}

func TestPrompterConfirm(t *testing.T) {
	t.Run("returns true for 'y' response", func(t *testing.T) {
		reader := strings.NewReader("y\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Confirm("Continue?")

		assert.True(t, result)
		assert.Contains(t, writer.String(), "Continue?")
		assert.Contains(t, writer.String(), "(y/N)")
	})

	t.Run("returns true for 'yes' response", func(t *testing.T) {
		reader := strings.NewReader("yes\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Confirm("Proceed?")

		assert.True(t, result)
	})

	t.Run("returns true for 'Y' response (case insensitive)", func(t *testing.T) {
		reader := strings.NewReader("Y\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Confirm("Deploy?")

		assert.True(t, result)
	})

	t.Run("returns true for 'YES' response", func(t *testing.T) {
		reader := strings.NewReader("YES\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Confirm("Apply changes?")

		assert.True(t, result)
	})

	t.Run("returns false for 'n' response", func(t *testing.T) {
		reader := strings.NewReader("n\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Confirm("Continue?")

		assert.False(t, result)
	})

	t.Run("returns false for 'no' response", func(t *testing.T) {
		reader := strings.NewReader("no\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Confirm("Proceed?")

		assert.False(t, result)
	})

	t.Run("returns false for empty response", func(t *testing.T) {
		reader := strings.NewReader("\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Confirm("Continue?")

		assert.False(t, result)
	})

	t.Run("returns false for random text", func(t *testing.T) {
		reader := strings.NewReader("maybe\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Confirm("Deploy?")

		assert.False(t, result)
	})

	t.Run("trims whitespace from response", func(t *testing.T) {
		reader := strings.NewReader("  yes  \n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Confirm("Continue?")

		assert.True(t, result)
	})

	t.Run("handles EOF", func(t *testing.T) {
		reader := strings.NewReader("")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Confirm("Continue?")

		assert.False(t, result)
	})
}

func TestPrompterConfirmDestruction(t *testing.T) {
	t.Run("returns true for exact 'yes' response", func(t *testing.T) {
		reader := strings.NewReader("yes\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.ConfirmDestruction("Destroy all resources?", "/project")

		assert.True(t, result)
		output := writer.String()
		assert.Contains(t, output, "DESTRUCTIVE ACTION")
		assert.Contains(t, output, "Destroy all resources?")
		assert.Contains(t, output, "/project")
	})

	t.Run("returns false for 'y' (requires full 'yes')", func(t *testing.T) {
		reader := strings.NewReader("y\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.ConfirmDestruction("Delete everything?", "/tmp")

		assert.False(t, result)
	})

	t.Run("returns false for 'YES' (case sensitive)", func(t *testing.T) {
		reader := strings.NewReader("YES\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.ConfirmDestruction("Destroy?", "/data")

		assert.False(t, result)
	})

	t.Run("returns false for 'no'", func(t *testing.T) {
		reader := strings.NewReader("no\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.ConfirmDestruction("Delete?", "/home")

		assert.False(t, result)
	})

	t.Run("returns false for empty response", func(t *testing.T) {
		reader := strings.NewReader("\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.ConfirmDestruction("Destroy?", "/path")

		assert.False(t, result)
	})

	t.Run("trims whitespace but still requires exact 'yes'", func(t *testing.T) {
		reader := strings.NewReader("  yes  \n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.ConfirmDestruction("Delete?", "/test")

		assert.True(t, result)
	})

	t.Run("displays resource information", func(t *testing.T) {
		reader := strings.NewReader("yes\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		prompter.ConfirmDestruction("Remove infrastructure?", "/my-project")

		output := writer.String()
		assert.Contains(t, output, "Resource: /my-project")
	})

	t.Run("handles EOF", func(t *testing.T) {
		reader := strings.NewReader("")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.ConfirmDestruction("Destroy?", "/path")

		assert.False(t, result)
	})
}

func TestPrompterInput(t *testing.T) {
	t.Run("returns user input", func(t *testing.T) {
		reader := strings.NewReader("my-project\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Input("Project name")

		assert.Equal(t, "my-project", result)
		assert.Contains(t, writer.String(), "Project name:")
	})

	t.Run("trims whitespace", func(t *testing.T) {
		reader := strings.NewReader("  value  \n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Input("Enter value")

		assert.Equal(t, "value", result)
	})

	t.Run("returns empty string for empty input", func(t *testing.T) {
		reader := strings.NewReader("\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Input("Name")

		assert.Equal(t, "", result)
	})

	t.Run("handles EOF", func(t *testing.T) {
		reader := strings.NewReader("")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		result := prompter.Input("Name")

		assert.Equal(t, "", result)
	})
}

func TestPrompterSelect(t *testing.T) {
	t.Run("returns selected option", func(t *testing.T) {
		reader := strings.NewReader("2\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		options := []string{"Go", "Python", "Node.js"}
		index, value, err := prompter.Select("Choose runtime", options)

		assert.NoError(t, err)
		assert.Equal(t, 1, index)
		assert.Equal(t, "Python", value)
		output := writer.String()
		assert.Contains(t, output, "Choose runtime")
		assert.Contains(t, output, "1) Go")
		assert.Contains(t, output, "2) Python")
		assert.Contains(t, output, "3) Node.js")
	})

	t.Run("returns first option for selection 1", func(t *testing.T) {
		reader := strings.NewReader("1\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		options := []string{"Option A", "Option B"}
		index, value, err := prompter.Select("Select", options)

		assert.NoError(t, err)
		assert.Equal(t, 0, index)
		assert.Equal(t, "Option A", value)
	})

	t.Run("returns last option for valid selection", func(t *testing.T) {
		reader := strings.NewReader("3\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		options := []string{"First", "Second", "Third"}
		index, value, err := prompter.Select("Pick one", options)

		assert.NoError(t, err)
		assert.Equal(t, 2, index)
		assert.Equal(t, "Third", value)
	})

	t.Run("retries on invalid selection then succeeds", func(t *testing.T) {
		// 2 invalid attempts (invalid number, then out of range) then valid (2) - within 3 attempt limit
		reader := strings.NewReader("abc\n5\n2\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		options := []string{"A", "B", "C"}
		index, value, err := prompter.Select("Choose", options)

		assert.NoError(t, err)
		assert.Equal(t, 1, index)
		assert.Equal(t, "B", value)
		output := writer.String()
		// Should show error messages for invalid selections
		assert.Contains(t, output, "Invalid selection")
	})

	t.Run("shows error for out of range selection", func(t *testing.T) {
		reader := strings.NewReader("10\n1\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		options := []string{"One", "Two"}
		prompter.Select("Select", options)

		output := writer.String()
		assert.Contains(t, output, "Invalid selection")
		assert.Contains(t, output, "between 1 and 2")
	})

	t.Run("shows error for non-numeric input", func(t *testing.T) {
		reader := strings.NewReader("abc\n1\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)

		options := []string{"X", "Y"}
		prompter.Select("Pick", options)

		output := writer.String()
		assert.Contains(t, output, "Invalid selection")
	})
}

// BenchmarkConfirm tests confirm performance.
func BenchmarkConfirm(b *testing.B) {
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader("yes\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)
		prompter.Confirm("Continue?")
	}
}

// BenchmarkConfirmDestruction tests destructive confirm performance.
func BenchmarkConfirmDestruction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader("yes\n")
		writer := &bytes.Buffer{}
		prompter := NewPrompter(reader, writer)
		prompter.ConfirmDestruction("Destroy?", "/test")
	}
}
