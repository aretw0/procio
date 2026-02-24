//go:build !windows

package termio

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

type posixConsole struct {
	fd       int
	origTerm unix.Termios
}

func newConsoleImpl(f *os.File) (consoleImpl, error) {
	fd := int(f.Fd())
	term, err := unix.IoctlGetTermios(fd, ioctlGetTermios)
	if err != nil {
		return nil, fmt.Errorf("console: tcgetattr: %w", err)
	}
	return &posixConsole{fd: fd, origTerm: *term}, nil
}

func (c *posixConsole) enableRawMode() error {
	raw := c.origTerm
	// Apply cfmakeraw semantics:
	//   - IGNBRK,BRKINT,PARMRK,ISTRIP,INLCR,IGNCR,ICRNL,IXON → off
	//   - OPOST → off
	//   - CS8 → on; ECHO,ECHONL,ICANON,ISIG,IEXTEN → off
	raw.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP |
		unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8
	raw.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(c.fd, ioctlSetTermios, &raw); err != nil {
		return fmt.Errorf("console: tcsetattr (raw): %w", err)
	}
	return nil
}

func (c *posixConsole) restore() error {
	if err := unix.IoctlSetTermios(c.fd, ioctlSetTermios, &c.origTerm); err != nil {
		return fmt.Errorf("console: tcsetattr (restore): %w", err)
	}
	return nil
}

func (c *posixConsole) size() (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(c.fd, unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, fmt.Errorf("console: TIOCGWINSZ: %w", err)
	}
	return int(ws.Col), int(ws.Row), nil
}
