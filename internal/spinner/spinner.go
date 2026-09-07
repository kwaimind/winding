// Package spinner shows a terminal spinner while a long-running step works.
package spinner

import (
	"fmt"
	"os"
	"time"
)

var frames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner animates a message on stdout until Stop is called. It is a no-op
// when stdout isn't a terminal, so piped/CI output stays clean.
type Spinner struct {
	stop chan struct{}
	done chan struct{}
}

// Start begins animating message and returns the Spinner. Callers must call
// Stop once the work it describes has finished.
func Start(message string) *Spinner {
	s := &Spinner{stop: make(chan struct{}), done: make(chan struct{})}
	if !isTerminal() {
		close(s.done)
		return s
	}
	go s.run(message)
	return s
}

func (s *Spinner) run(message string) {
	defer close(s.done)
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	for i := 0; ; i++ {
		select {
		case <-s.stop:
			fmt.Print("\r\033[K")
			return
		case <-ticker.C:
			fmt.Printf("\r%s %s", frames[i%len(frames)], message)
		}
	}
}

// Stop halts the animation and clears its line.
func (s *Spinner) Stop() {
	select {
	case <-s.done:
		return
	default:
		close(s.stop)
		<-s.done
	}
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
