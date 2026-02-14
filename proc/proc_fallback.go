//go:build !windows && !linux

package proc

import (
	"errors"
	"os/exec"
	"runtime"

	"github.com/aretw0/procio"
)

func start(cmd *exec.Cmd) error {
	// Fallback for macOS, BSD, etc. where Pdeathsig/JobObjects aren't available
	// or implemented yet.
	if StrictMode {
		return errors.New("process hygiene not supported on " + runtime.GOOS)
	}

	procio.GetObserver().LogWarn("process hygiene is not supported on this platform, falling back to standard cmd.Start()",
		"os", runtime.GOOS,
		"arch", runtime.GOARCH)

	if err := cmd.Start(); err != nil {
		procio.GetObserver().OnProcessFailed(err)
		return err
	}
	procio.GetObserver().OnProcessStarted(cmd.Process.Pid)
	return nil
}
