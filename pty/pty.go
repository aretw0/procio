package pty

import (
	"errors"
	"os"
	"os/exec"
)

// ErrUnsupported is returned by [StartPTY] on platforms where
// pseudo-terminal allocation is not implemented.
var ErrUnsupported = errors.New("pty: PTY is not supported on this platform")

// PTY holds the two ends of a pseudo-terminal pair.
type PTY struct {
	// Controller is the application-side end of the PTY.
	// Read from it to receive process output; write to it to send input.
	// Call Close() on the PTY instead of closing this file directly.
	Controller *os.File

	// Worker is the process-side end of the PTY.
	// It is passed to the child process as its stdio; it is closed in the
	// parent by [StartPTY] after the process starts.
	Worker *os.File

	// Platform-specific teardown.
	close func() error
}

// Close closes the PTY controller and tears down associated kernel objects
// (like Windows PseudoConsoles). It returns the first error encountered, if any.
func (p *PTY) Close() error {
	if p.close != nil {
		return p.close()
	}
	if p.Controller != nil {
		return p.Controller.Close()
	}
	return nil
}

// StartPTY starts cmd attached to a newly allocated pseudo-terminal.
// It returns a [PTY] whose Controller the caller should use to read
// output from and write input to the process.
//
// StartPTY does NOT apply proc package hygiene attributes (Job Objects /
// Pdeathsig). Use [proc.NewCmd] and [proc.Start] for that; StartPTY
// simply replaces cmd.Start():
//
//	cmd := proc.NewCmd(ctx, "vim", "file.txt")
//	pty, err := pty.StartPTY(cmd)
//
// On platforms where PTY is not supported, StartPTY returns [ErrUnsupported].
func StartPTY(cmd *exec.Cmd) (*PTY, error) {
	return startPTY(cmd)
}
