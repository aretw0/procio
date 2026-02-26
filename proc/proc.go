package proc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
)

// StrictMode if true, will cause Start to return an error on unsupported platforms
// instead of just logging a warning. Default is false.
var StrictMode bool

// Cmd is a type that wraps exec.Cmd to provide safe execution defaults.
// It ensures that platform hygiene attributes (Pdeathsig on Linux, Job Objects
// on Windows) are applied seamlessly when starting the process.
type Cmd struct {
	*exec.Cmd
}

// NewCmd creates a new Cmd pre-configured with context-linked cancellation
// (exec.CommandContext semantics). It is the recommended entry point when the
// caller holds a context.
//
// Unlike the standard library exec.CommandContext, methods like Start() and Run()
// on this returned Cmd will automatically apply platform safety mechanisms to
// prevent child process leaks.
func NewCmd(ctx context.Context, name string, args ...string) *Cmd {
	return &Cmd{
		Cmd: exec.CommandContext(ctx, name, args...),
	}
}

// Start starts the specified command but ensures that the child process
// is killed if the parent process (this process) dies.
//
// On Linux, it uses SysProcAttr.Pdeathsig (SIGKILL).
// On Windows, it uses Job Objects (JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE).
//
// On other platforms, it falls back to cmd.Start() and logs a warning,
// unless StrictMode is set to true, in which case it returns an error.
func (c *Cmd) Start() error {
	return start(c.Cmd)
}

// Run starts the specified command and waits for it to complete.
// It applies the same platform hygiene attributes as Start.
func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

// Output runs the command and returns its standard output.
// It applies the same platform hygiene attributes as Start.
func (c *Cmd) Output() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = io.Discard
	err := c.Run()
	return b.Bytes(), err
}

// CombinedOutput runs the command and returns its combined standard output and standard error.
// It applies the same platform hygiene attributes as Start.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	err := c.Run()
	return b.Bytes(), err
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
