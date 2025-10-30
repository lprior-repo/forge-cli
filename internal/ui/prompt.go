package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Prompter provides interactive prompts.
type Prompter struct {
	reader  io.Reader
	writer  io.Writer
	output  *Output
	scanner *bufio.Scanner
}

// NewPrompter creates a new prompter.
func NewPrompter(r io.Reader, w io.Writer) *Prompter {
	return &Prompter{
		reader:  r,
		writer:  w,
		output:  NewOutput(w),
		scanner: bufio.NewScanner(r),
	}
}

// Returns true if user confirms, false otherwise.
func (p *Prompter) Confirm(message string) bool {
	p.output.Warning("%s (y/N): ", message)

	if !p.scanner.Scan() {
		return false
	}

	response := strings.ToLower(strings.TrimSpace(p.scanner.Text()))
	return response == "y" || response == "yes"
}

// Requires typing "yes" explicitly for safety.
func (p *Prompter) ConfirmDestruction(message, resource string) bool {
	p.output.Error("DESTRUCTIVE ACTION")
	p.output.Warning("%s", message)
	p.output.Print("")
	p.output.Print("Resource: %s", resource)
	p.output.Print("Type 'yes' to confirm: ")

	if !p.scanner.Scan() {
		return false
	}

	response := strings.TrimSpace(p.scanner.Text())
	return response == "yes"
}

// Input prompts for text input.
func (p *Prompter) Input(message string) string {
	_, _ = fmt.Fprintf(p.writer, "%s: ", message)

	if !p.scanner.Scan() {
		return ""
	}

	return strings.TrimSpace(p.scanner.Text())
}

// Select prompts the user to select from a list of options
// Returns (index, value, error). Returns error if max attempts exceeded or no input received.
func (p *Prompter) Select(message string, options []string) (int, string, error) {
	const maxAttempts = 3

	p.output.Print(message)
	for i, opt := range options {
		p.output.Print("  %d) %s", i+1, opt)
	}
	p.output.Print("")

	for attempt := 0; attempt < maxAttempts; attempt++ {
		input := p.Input(fmt.Sprintf("Select option (1-%d)", len(options)))

		if input == "" {
			return 0, "", errors.New("no input received")
		}

		var selection int
		_, err := fmt.Sscanf(input, "%d", &selection)
		if err == nil && selection >= 1 && selection <= len(options) {
			return selection - 1, options[selection-1], nil
		}

		if attempt < maxAttempts-1 {
			p.output.Error("Invalid selection. Please choose a number between 1 and %d", len(options))
		}
	}

	return 0, "", fmt.Errorf("maximum selection attempts (%d) exceeded", maxAttempts)
}
