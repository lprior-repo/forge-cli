package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSpinner(t *testing.T) {
	t.Run("creates spinner with message", func(t *testing.T) {
		buf := &bytes.Buffer{}
		spinner := NewSpinner(buf, "Loading...")

		assert.NotNil(t, spinner)
		assert.Equal(t, "Loading...", spinner.message)
		assert.Equal(t, buf, spinner.writer)
		assert.False(t, spinner.active)
	})
}

func TestSpinnerStartStop(t *testing.T) {
	t.Run("starts and stops spinner", func(t *testing.T) {
		buf := &bytes.Buffer{}
		spinner := NewSpinner(buf, "Processing")

		spinner.Start()
		assert.True(t, spinner.active)

		// Let it spin for a bit
		time.Sleep(150 * time.Millisecond)

		spinner.Stop()
		assert.False(t, spinner.active)

		// Should have written some frames
		output := buf.String()
		assert.NotEmpty(t, output)
	})

	t.Run("stop clears the spinner line", func(t *testing.T) {
		buf := &bytes.Buffer{}
		spinner := NewSpinner(buf, "Working")

		spinner.Start()
		time.Sleep(150 * time.Millisecond)
		spinner.Stop()

		output := buf.String()
		// Should contain carriage returns for clearing
		assert.Contains(t, output, "\r")
	})

	t.Run("handles stop without start", func(t *testing.T) {
		buf := &bytes.Buffer{}
		spinner := NewSpinner(buf, "Test")

		// Should not panic
		assert.NotPanics(t, func() {
			spinner.Stop()
		})
	})
}

func TestSpinnerFrames(t *testing.T) {
	t.Run("uses spinner frames", func(t *testing.T) {
		buf := &bytes.Buffer{}
		spinner := NewSpinner(buf, "Loading")

		assert.Len(t, spinner.frames, 10)
		assert.Equal(t, "⠋", spinner.frames[0])
		assert.Equal(t, "⠙", spinner.frames[1])
	})
}

func TestNewProgressBar(t *testing.T) {
	t.Run("creates progress bar with total and prefix", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bar := NewProgressBar(buf, 100, "Building")

		assert.NotNil(t, bar)
		assert.Equal(t, 100, bar.total)
		assert.Equal(t, "Building", bar.prefix)
		assert.Equal(t, buf, bar.writer)
		assert.Equal(t, 0, bar.current)
	})
}

func TestProgressBarUpdate(t *testing.T) {
	t.Run("updates progress to specific value", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bar := NewProgressBar(buf, 100, "Progress")

		bar.Update(50)

		assert.Equal(t, 50, bar.current)
		output := buf.String()
		assert.Contains(t, output, "Progress")
		assert.Contains(t, output, "50/100")
		assert.Contains(t, output, "50%")
	})

	t.Run("updates to 0%", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bar := NewProgressBar(buf, 10, "Test")

		bar.Update(0)

		output := buf.String()
		assert.Contains(t, output, "0/10")
		assert.Contains(t, output, "0%")
	})

	t.Run("updates to 100%", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bar := NewProgressBar(buf, 10, "Test")

		bar.Update(10)

		output := buf.String()
		assert.Contains(t, output, "10/10")
		assert.Contains(t, output, "100%")
	})
}

func TestProgressBarIncrement(t *testing.T) {
	t.Run("increments progress by 1", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bar := NewProgressBar(buf, 5, "Building")

		bar.Increment()
		assert.Equal(t, 1, bar.current)

		bar.Increment()
		assert.Equal(t, 2, bar.current)

		bar.Increment()
		assert.Equal(t, 3, bar.current)
	})
}

func TestProgressBarComplete(t *testing.T) {
	t.Run("marks progress as complete", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bar := NewProgressBar(buf, 10, "Deploy")

		bar.Update(5)
		bar.Complete()

		assert.Equal(t, 10, bar.current)
		output := buf.String()
		assert.Contains(t, output, "10/10")
		assert.Contains(t, output, "100%")
		// Should have a newline at the end
		assert.True(t, strings.HasSuffix(output, "\n"))
	})
}

func TestProgressBarRender(t *testing.T) {
	t.Run("renders progress bar with filled and empty sections", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bar := NewProgressBar(buf, 100, "Test")

		bar.Update(25)

		output := buf.String()
		// Should contain filled blocks (█) and empty blocks (░)
		assert.Contains(t, output, "█")
		assert.Contains(t, output, "░")
		assert.Contains(t, output, "[")
		assert.Contains(t, output, "]")
	})

	t.Run("renders with custom width", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bar := NewProgressBar(buf, 100, "Custom")
		bar.width = 20

		bar.Update(50)

		output := buf.String()
		// Bar should be visible
		assert.Contains(t, output, "█")
		assert.Contains(t, output, "░")
	})
}

func TestProgressBarSequence(t *testing.T) {
	t.Run("progresses through multiple updates", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bar := NewProgressBar(buf, 4, "Steps")

		// Simulate 4 steps
		for i := 1; i <= 4; i++ {
			buf.Reset()
			bar.Update(i)
			output := buf.String()
			assert.Contains(t, output, "Steps")
		}

		assert.Equal(t, 4, bar.current)
	})
}

// BenchmarkSpinner tests spinner performance
func BenchmarkSpinner(b *testing.B) {
	buf := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spinner := NewSpinner(buf, "Testing")
		spinner.Start()
		time.Sleep(10 * time.Millisecond)
		spinner.Stop()
		buf.Reset()
	}
}

// BenchmarkProgressBar tests progress bar performance
func BenchmarkProgressBar(b *testing.B) {
	buf := &bytes.Buffer{}
	bar := NewProgressBar(buf, 100, "Benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		bar.Update(i % 100)
	}
}
