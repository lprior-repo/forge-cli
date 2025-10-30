package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// Output provides colored terminal output.
type Output struct {
	writer io.Writer

	// Color functions
	success *color.Color
	error   *color.Color
	warning *color.Color
	info    *color.Color
	dim     *color.Color
}

// NewOutput creates a new output instance.
func NewOutput(w io.Writer) *Output {
	return &Output{
		writer:  w,
		success: color.New(color.FgGreen),
		error:   color.New(color.FgRed),
		warning: color.New(color.FgYellow),
		info:    color.New(color.FgCyan),
		dim:     color.New(color.Faint),
	}
}

// DefaultOutput creates output writing to stdout.
func DefaultOutput() *Output {
	return NewOutput(os.Stdout)
}

// Success prints a success message in green.
func (o *Output) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = o.success.Fprintf(o.writer, "✓ %s\n", msg)
}

// Error prints an error message in red.
func (o *Output) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = o.error.Fprintf(o.writer, "✗ %s\n", msg)
}

// Warning prints a warning message in yellow.
func (o *Output) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = o.warning.Fprintf(o.writer, "⚠ %s\n", msg)
}

// Info prints an info message in cyan.
func (o *Output) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = o.info.Fprintf(o.writer, "ℹ %s\n", msg)
}

// Print prints a regular message.
func (o *Output) Print(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.writer, format+"\n", args...)
}

// Dim prints a dimmed message.
func (o *Output) Dim(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = o.dim.Fprintln(o.writer, msg)
}

// Header prints a section header.
func (o *Output) Header(text string) {
	_, _ = fmt.Fprintln(o.writer)
	_, _ = fmt.Fprintf(o.writer, "=== %s ===\n", text)
	_, _ = fmt.Fprintln(o.writer)
}

// Step prints a step in a process.
func (o *Output) Step(step, total int, message string) {
	_, _ = o.dim.Fprintf(o.writer, "[%d/%d] ", step, total)
	_, _ = fmt.Fprintln(o.writer, message)
}
