// Package procio provides robust process and terminal primitives for Go applications.
//
// It serves as the foundation for the `lifecycle` library but stands alone as a
// "Hidden Gem" for solving two universal problems in Go:
//
//  1. Zombie Processes (via pkg/proc): Ensures child processes are terminated when
//     the parent dies, using platform-specific mechanisms like PDeathSig (Linux)
//     and Job Objects (Windows).
//
//  2. Windows I/O Resilience (via pkg/termio): Provides interruptible readers and
//     CONIN$ handling to prevent indefinite blocking on Stdin, a common issue
//     when developing cross-platform CLIs.
package procio
