package spout

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SigMon registers the signal handler and run the handler in a standalone goroutine.
func (s *Spout) SigMon() {
	c := make(chan os.Signal, 10)
	signal.Notify(c,
		os.Interrupt,    // Ctrl-C
		syscall.SIGTERM, // kill
	)

	go s.sigHandler(c)
}

type UnsupportedSignalError string

func (e UnsupportedSignalError) Error() string {
	return fmt.Sprintf("Unsupported signal: %s", e)
}

func (s *Spout) sigHandler(c chan os.Signal) {
	for sig := range c {
		switch sig {
		case os.Interrupt:
			fallthrough
		case syscall.SIGTERM:
			signal.Stop(c)
			// Don't call os.Exit(), wait until all workers are closed.
			s.Stop()
		default:
			err := UnsupportedSignalError(sig.String())
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
