package proc

import (
	"context"
	"os/exec"
)

// StrictMode if true, will cause Start to return an error on unsupported platforms
// instead of just logging a warning. Default is false.
var StrictMode bool

// NewCmd creates an *exec.Cmd pre-configured with context-linked cancellation
// (exec.CommandContext semantics). It is the recommended entry point when the
// caller holds a context, replacing the error-prone two-step pattern:
//
//	// before
//	cmd := exec.CommandContext(ctx, name, args...)
//	err := proc.Start(cmd)
//
//	// after
//	cmd := proc.NewCmd(ctx, name, args...)
//	err := proc.Start(cmd)
//
// NewCmd does NOT call Start. The caller must still call proc.Start(cmd) to
// apply platform hygiene attributes (Pdeathsig on Linux, Job Objects on
// Windows) and launch the process. This separation allows configuring Stdout,
// Stdin, Env, Dir, etc. before starting.
func NewCmd(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// Start starts the specified command but ensures that the child process
// is killed if the parent process (this process) dies.
//
// On Linux, it uses SysProcAttr.Pdeathsig (SIGKILL).
// On Windows, it uses Job Objects (JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE).
//
// On other platforms, it falls back to cmd.Start() and logs a warning,
// unless StrictMode is set to true, in which case it returns an error.
//
// This is a safer alternative to cmd.Start() for long-running child processes.
func Start(cmd *exec.Cmd) error {
	return start(cmd)
}
