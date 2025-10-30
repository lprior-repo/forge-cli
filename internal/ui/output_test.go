package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOutput(t *testing.T) {
	t.Run("creates output with writer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		assert.NotNil(t, out)
		assert.Equal(t, buf, out.writer)
	})
}

func TestDefaultOutput(t *testing.T) {
	t.Run("creates output with stdout", func(t *testing.T) {
		out := DefaultOutput()

		assert.NotNil(t, out)
	})
}

func TestOutputSuccess(t *testing.T) {
	t.Run("writes success message with checkmark", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Success("deployment completed")

		output := buf.String()
		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "deployment completed")
	})

	t.Run("formats success message with arguments", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Success("built %d functions", 5)

		output := buf.String()
		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "built 5 functions")
	})
}

func TestOutputError(t *testing.T) {
	t.Run("writes error message with X mark", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Error("build failed")

		output := buf.String()
		assert.Contains(t, output, "✗")
		assert.Contains(t, output, "build failed")
	})

	t.Run("formats error message with arguments", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Error("failed to build %s", "my-function")

		output := buf.String()
		assert.Contains(t, output, "✗")
		assert.Contains(t, output, "failed to build my-function")
	})
}

func TestOutputWarning(t *testing.T) {
	t.Run("writes warning message with warning icon", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Warning("no functions found")

		output := buf.String()
		assert.Contains(t, output, "⚠")
		assert.Contains(t, output, "no functions found")
	})

	t.Run("formats warning message with arguments", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Warning("found %d warnings", 3)

		output := buf.String()
		assert.Contains(t, output, "⚠")
		assert.Contains(t, output, "found 3 warnings")
	})
}

func TestOutputInfo(t *testing.T) {
	t.Run("writes info message with info icon", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Info("scanning functions")

		output := buf.String()
		assert.Contains(t, output, "ℹ")
		assert.Contains(t, output, "scanning functions")
	})

	t.Run("formats info message with arguments", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Info("found %d functions", 2)

		output := buf.String()
		assert.Contains(t, output, "ℹ")
		assert.Contains(t, output, "found 2 functions")
	})
}

func TestOutputPrint(t *testing.T) {
	t.Run("writes plain message", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Print("plain text")

		output := buf.String()
		assert.Contains(t, output, "plain text")
		assert.Contains(t, output, "\n")
	})

	t.Run("formats message with arguments", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Print("value: %d", 42)

		output := buf.String()
		assert.Contains(t, output, "value: 42")
	})
}

func TestOutputDim(t *testing.T) {
	t.Run("writes dimmed message", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Dim("additional info")

		output := buf.String()
		assert.Contains(t, output, "additional info")
	})

	t.Run("formats dimmed message with arguments", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Dim("output: %s", "/tmp/build")

		output := buf.String()
		assert.Contains(t, output, "output: /tmp/build")
	})
}

func TestOutputHeader(t *testing.T) {
	t.Run("writes header with formatting", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Header("Building Functions")

		output := buf.String()
		assert.Contains(t, output, "=== Building Functions ===")
		// Should have blank lines before and after
		lines := strings.Split(output, "\n")
		assert.GreaterOrEqual(t, len(lines), 3)
	})
}

func TestOutputStep(t *testing.T) {
	t.Run("writes step indicator with numbers", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Step(2, 5, "Building my-function")

		output := buf.String()
		assert.Contains(t, output, "[2/5]")
		assert.Contains(t, output, "Building my-function")
	})

	t.Run("handles first step", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Step(1, 10, "Starting build")

		output := buf.String()
		assert.Contains(t, output, "[1/10]")
		assert.Contains(t, output, "Starting build")
	})

	t.Run("handles last step", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Step(3, 3, "Final step")

		output := buf.String()
		assert.Contains(t, output, "[3/3]")
		assert.Contains(t, output, "Final step")
	})
}

func TestOutputMultipleMessages(t *testing.T) {
	t.Run("writes multiple messages in sequence", func(t *testing.T) {
		buf := &bytes.Buffer{}
		out := NewOutput(buf)

		out.Header("Deployment")
		out.Info("Starting deployment")
		out.Success("Deployed function 1")
		out.Success("Deployed function 2")
		out.Warning("Function 3 has warnings")
		out.Error("Function 4 failed")

		output := buf.String()
		assert.Contains(t, output, "=== Deployment ===")
		assert.Contains(t, output, "Starting deployment")
		assert.Contains(t, output, "Deployed function 1")
		assert.Contains(t, output, "Deployed function 2")
		assert.Contains(t, output, "Function 3 has warnings")
		assert.Contains(t, output, "Function 4 failed")
	})
}

// BenchmarkOutput tests performance of output operations.
func BenchmarkOutputSuccess(b *testing.B) {
	buf := &bytes.Buffer{}
	out := NewOutput(buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		out.Success("operation completed")
	}
}

func BenchmarkOutputFormatted(b *testing.B) {
	buf := &bytes.Buffer{}
	out := NewOutput(buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		out.Success("built %d functions in %s", 10, "/tmp/build")
	}
}
