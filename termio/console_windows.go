//go:build windows

package termio

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

type windowsConsole struct {
	fd       windows.Handle
	origMode uint32
}

func newConsoleImpl(f *os.File) (consoleImpl, error) {
	fd := windows.Handle(f.Fd())
	var origMode uint32
	if err := windows.GetConsoleMode(fd, &origMode); err != nil {
		return nil, fmt.Errorf("console: GetConsoleMode: %w", err)
	}
	return &windowsConsole{fd: fd, origMode: origMode}, nil
}

func (c *windowsConsole) enableRawMode() error {
	// Raw mode on Windows:
	//   - Disable line input, echo, processed input (e.g. Ctrl+C handled in-app).
	//   - Enable virtual terminal input so ANSI escape sequences pass through.
	rawMode := c.origMode&^
		(windows.ENABLE_LINE_INPUT|
			windows.ENABLE_ECHO_INPUT|
			windows.ENABLE_PROCESSED_INPUT) |
		windows.ENABLE_VIRTUAL_TERMINAL_INPUT

	if err := windows.SetConsoleMode(c.fd, rawMode); err != nil {
		return fmt.Errorf("console: SetConsoleMode (raw): %w", err)
	}
	return nil
}

func (c *windowsConsole) restore() error {
	if err := windows.SetConsoleMode(c.fd, c.origMode); err != nil {
		return fmt.Errorf("console: SetConsoleMode (restore): %w", err)
	}
	return nil
}

func (c *windowsConsole) size() (width, height int, err error) {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(c.fd, &info); err != nil {
		return 0, 0, fmt.Errorf("console: GetConsoleScreenBufferInfo: %w", err)
	}
	w := int(info.Window.Right-info.Window.Left) + 1
	h := int(info.Window.Bottom-info.Window.Top) + 1
	return w, h, nil
}
