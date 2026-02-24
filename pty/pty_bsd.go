//go:build freebsd || openbsd || netbsd || dragonfly

package pty

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

func ioctlGrantpt(fd int) error {
	// On BSDs, permissions are managed automatically when opening /dev/ptmx.
	return nil
}

func ioctlUnlockpt(fd int) error {
	// BSDs do not require a specific unlock ioctl (unlike Linux TIOCSPTLCK).
	return nil
}

func ioctlPtsname(fd int) (string, error) {
	// On BSDs, we use TIOCPTYGNAME to get the worker device path.
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
