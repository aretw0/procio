//go:build linux

package termio

import "golang.org/x/sys/unix"

const (
	ioctlGetTermios = unix.TCGETS
	ioctlSetTermios = unix.TCSETS
)
