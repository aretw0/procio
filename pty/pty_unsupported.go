//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly && !windows

package pty

import "os/exec"

func startPTY(_ *exec.Cmd) (*PTY, error) {
	return nil, ErrUnsupported
}
