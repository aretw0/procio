//go:build linux

package proc

import (
	"os/exec"
	"syscall"

	"github.com/aretw0/procio"
)

func start(cmd *exec.Cmd) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// Pdeathsig ensures that if the parent dies, the child receives this signal.
	cmd.SysProcAttr.Pdeathsig = syscall.SIGKILL

	procio.GetObserver().LogDebug("starting process with Pdeathsig", "command", cmd.Path)
	if err := cmd.Start(); err != nil {
		procio.GetObserver().OnProcessFailed(err)
		return err
	}
	procio.GetObserver().OnProcessStarted(cmd.Process.Pid)
	return nil
}
