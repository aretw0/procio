//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package pty

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

func startPTY(cmd *exec.Cmd) (*PTY, error) {
	controller, worker, err := openpty()
	if err != nil {
		return nil, err
	}

	// Attach the worker to the child's stdio so the process gets a real tty.
	cmd.Stdin = worker
	cmd.Stdout = worker
	cmd.Stderr = worker

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &unix.SysProcAttr{}
	}
	// New session + controlling terminal: the child becomes the session leader
	// and the worker fd is set as its controlling terminal via TIOCSCTTY.
	cmd.SysProcAttr.Setsid = true
	cmd.SysProcAttr.Setctty = true
	cmd.SysProcAttr.Ctty = 0 // fd 0 (stdin) inside the child

	if err := cmd.Start(); err != nil {
		_ = controller.Close()
		_ = worker.Close()
		return nil, fmt.Errorf("pty: start process: %w", err)
	}

	// Close the worker end in the parent. The child's inherited descriptors
	// keep it alive. Closing here prevents stale references that would delay
	// EOF detection after the child exits.
	_ = worker.Close()

	return &PTY{
		Controller: controller,
		Worker:     worker,
	}, nil
}

// openpty allocates a new PTY pair using /dev/ptmx (POSIX).
// Returns (controller, worker, error).
func openpty() (*os.File, *os.File, error) {
	// Open the controller side (/dev/ptmx).
	controllerFD, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("pty: open /dev/ptmx: %w", err)
	}

	// ioctlGrantpt: on Linux this is a no-op; on Darwin it adjusts permissions.
	if err := ioctlGrantpt(controllerFD); err != nil {
		_ = unix.Close(controllerFD)
		return nil, nil, fmt.Errorf("pty: grantpt: %w", err)
	}

	// unlockpt: grants the worker side exclusive access.
	if err := ioctlUnlockpt(controllerFD); err != nil {
		_ = unix.Close(controllerFD)
		return nil, nil, fmt.Errorf("pty: unlockpt: %w", err)
	}

	// ptsname: get the path to the worker side (e.g. /dev/pts/3).
	workerPath, err := ioctlPtsname(controllerFD)
	if err != nil {
		_ = unix.Close(controllerFD)
		return nil, nil, fmt.Errorf("pty: ptsname: %w", err)
	}

	workerFD, err := unix.Open(workerPath, unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		_ = unix.Close(controllerFD)
		return nil, nil, fmt.Errorf("pty: open worker %s: %w", workerPath, err)
	}

	return os.NewFile(uintptr(controllerFD), "pty-controller"),
		os.NewFile(uintptr(workerFD), workerPath),
		nil
}
