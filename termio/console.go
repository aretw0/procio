package termio

import "os"

// Console provides platform-specific terminal mode management.
//
// It abstracts raw mode toggling and terminal size queries across
// Windows (Console API) and POSIX (termios) platforms.
//
// Acquire a Console with [NewConsole]. Always defer [Console.Restore]
// to return the terminal to its original state.
type Console struct {
	f    *os.File
	impl consoleImpl
}

// consoleImpl is the platform-specific engine behind Console.
type consoleImpl interface {
	enableRawMode() error
	restore() error
	size() (width, height int, err error)
}

// NewConsole wraps f as a Console. f must be a file descriptor that is
// connected to a terminal (e.g. os.Stdin, os.Stdout, or a PTY controller).
//
// Returns an error if f is not a recognized terminal file descriptor or if
// the underlying terminal state cannot be read.
func NewConsole(f *os.File) (*Console, error) {
	impl, err := newConsoleImpl(f)
	if err != nil {
		return nil, err
	}
	return &Console{f: f, impl: impl}, nil
}

// EnableRawMode puts the terminal into raw mode: input is passed through
// character by character, echoing is disabled, and signal generation from
// special keys (Ctrl+C, Ctrl+Z, etc.) is suppressed.
//
// Call [Console.Restore] (typically via defer) to return to the previous mode.
func (c *Console) EnableRawMode() error {
	return c.impl.enableRawMode()
}

// Restore returns the terminal to the state it had when [NewConsole] was called.
func (c *Console) Restore() error {
	return c.impl.restore()
}

// Size returns the current width and height of the terminal in columns and rows.
func (c *Console) Size() (width, height int, err error) {
	return c.impl.size()
}

// Fd returns the underlying file descriptor.
func (c *Console) Fd() uintptr {
	return c.f.Fd()
}
