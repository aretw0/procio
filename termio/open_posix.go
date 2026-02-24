//go:build !windows

package termio

import (
	"io"
	"os"

	"golang.org/x/term"
)

// Open returns the standard input reader.
func Open() (io.ReadCloser, error) {
	return os.Stdin, nil
}

// IsTerminal checks if the given file is a terminal.
// This is used to decide whether to apply terminal enhancements like raw mode.
func IsTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}
