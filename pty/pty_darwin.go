//go:build darwin || freebsd || openbsd || netbsd || dragonfly

package pty

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

func ioctlUnlockpt(fd int) error {
	// On many BSDs/Darwin, /dev/ptmx doesn't strictly need a TIOCSPTLCK unlock.
	return nil
}

func ioctlPtsname(fd int) (string, error) {
	// On Darwin/BSD, we use TIOCPTYGNAME to get the worker device path.
	const TIOCPTYGNAME = 0x40807453
	var buf [128]byte
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), TIOCPTYGNAME, uintptr(unsafe.Pointer(&buf[0]))); errno != 0 {
		return "", errno
	}

	// The buffer contains a null-terminated string.
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i]), nil
		}
	}
	return string(buf[:]), nil
}
