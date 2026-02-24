//go:build linux

package pty

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

func ioctlUnlockpt(fd int) error {
	// TIOCSPTLCK expects a pointer to a C int (int32), not a Go int (int64 on amd64).
	// unix.IoctlSetInt passes unsafe.Pointer(&goInt) which is 8 bytes — the kernel
	// reads a 4-byte int from that address and sees the high bits of the Go int,
	// causing EFAULT or a wrong value. We must pass a *int32 explicitly.
	var zero int32
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.TIOCSPTLCK, uintptr(unsafe.Pointer(&zero))); errno != 0 {
		return errno
	}
	return nil
}

func ioctlGrantpt(_ int) error {
	// grantpt is a no-op on Linux since the kernel devpts automatically handles permissions.
	return nil
}

func ioctlPtsname(fd int) (string, error) {
	var n uint32
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.TIOCGPTN, uintptr(unsafe.Pointer(&n))); errno != 0 {
		// Fallback: try to derive from fd number via procfs on Linux.
		data, err := os.ReadFile("/proc/self/fd/" + strconv.Itoa(fd))
		if err != nil {
			return "", fmt.Errorf("TIOCGPTN: %w", errno)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return fmt.Sprintf("/dev/pts/%d", n), nil
}
