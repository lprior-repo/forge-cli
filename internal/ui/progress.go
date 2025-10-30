package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Spinner provides a simple progress spinner
type Spinner struct {
	writer  io.Writer
	message string
	frames  []string
	delay   time.Duration
	active  bool
	done    chan bool
	once    sync.Once
}

// NewSpinner creates a new spinner with a message
func NewSpinner(w io.Writer, message string) *Spinner {
	return &Spinner{
		writer:  w,
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		delay:   100 * time.Millisecond,
		done:    make(chan bool),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.active = true
	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				// Clear the line
				fmt.Fprintf(s.writer, "\r%s\r", strings.Repeat(" ", len(s.message)+3))
				return
			default:
				frame := s.frames[i%len(s.frames)]
				fmt.Fprintf(s.writer, "\r%s %s", frame, s.message)
				i++
				time.Sleep(s.delay)
			}
		}
	}()
}

// Stop stops the spinner (safe to call multiple times)
func (s *Spinner) Stop() {
	s.once.Do(func() {
		if s.active {
			s.active = false
			close(s.done)
		}
	})
}

// ProgressBar provides a simple progress bar
type ProgressBar struct {
	writer  io.Writer
	total   int
	current int
	width   int
	prefix  string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(w io.Writer, total int, prefix string) *ProgressBar {
	return &ProgressBar{
		writer: w,
		total:  total,
		width:  40,
		prefix: prefix,
	}
}

// Update updates the progress bar
func (p *ProgressBar) Update(current int) {
	p.current = current
	p.render()
}

// Increment increments the progress by 1
func (p *ProgressBar) Increment() {
	p.current++
	p.render()
}

// Complete marks the progress as complete
func (p *ProgressBar) Complete() {
	p.current = p.total
	p.render()
	fmt.Fprintln(p.writer)
}

func (p *ProgressBar) render() {
	percent := float64(p.current) / float64(p.total)
	filled := int(percent * float64(p.width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	fmt.Fprintf(p.writer, "\r%s [%s] %d/%d (%.0f%%)",
		p.prefix, bar, p.current, p.total, percent*100)
}
