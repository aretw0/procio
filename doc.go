// Package procio is the root package for robust process I/O and signaling primitives.
//
// It provides a platform-agnostic way to handle process execution and terminal input
// with safety guarantees (like leak-free process termination using Job Objects on Windows
// and Pdeathsig on Linux).
//
// # Subpackages
//
//   - proc: Process management and lifecycle guarantees.
//   - scan: Context-aware parsing of input streams (Scanner).
//   - termio: Terminal I/O utilities and interruptible readers.
//
// # Observability
//
// procio does not depend on any logging library. Instead, it exposes an Observer interface
// that you can implement to bridge logs and metrics to your preferred system.
package procio
