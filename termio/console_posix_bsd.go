//go:build darwin || freebsd || openbsd || netbsd || dragonfly

package termio

import "golang.org/x/sys/unix"

const (
	ioctlGetTermios = unix.TIOCGETA
	ioctlSetTermios = unix.TIOCSETA
)
