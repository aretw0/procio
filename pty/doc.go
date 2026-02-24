// Package pty provides pseudo-terminal (PTY) primitives for running interactive
// applications within child processes.
//
// A pseudo-terminal is a pair of virtual file descriptors that behave like a
// real terminal. It allows programs that expect to be connected to a terminal
// (vim, htop, bash, etc.) to run as subprocesses of a Go application.
//
// # Terminology
//
// The PTY pair uses the following naming convention:
//   - Controller: the application side. Read process output from it; write
//     input to it.
//   - Worker: the process side. Attached to the child process as its stdin,
//     stdout, and stderr.
//
// # Platform Support
//
//   - Linux, macOS, BSDs: uses openpty(3) via [golang.org/x/sys/unix].
//   - Windows 10+ (build 17763+): uses the ConPTY API (CreatePseudoConsole).
//   - Other platforms: [StartPTY] returns [ErrUnsupported].
//
// # Usage
//
//	cmd := proc.NewCmd(ctx, "vim", "file.txt")
//	pty, err := pty.StartPTY(cmd)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer pty.Controller.Close()
//
//	// pty.Controller is an *os.File; use io.Copy for streaming I/O.
//	go io.Copy(os.Stdout, pty.Controller)
//	io.Copy(pty.Controller, os.Stdin)
//	cmd.Wait()
package pty
