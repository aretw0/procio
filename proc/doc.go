// Package proc provides platform-agnostic process management with safety guarantees.
//
// It abstracts away OS-specific details for reliable process termination:
//   - Windows: Uses Job Objects to ensure child processes are killed when the parent exits.
//   - Linux: Uses Pdeathsig (SIGKILL) for the same guarantee.
//
// Usage:
//
//	cmd := exec.Command("ping", "google.com")
//	_ = proc.Start(cmd)
//	cmd.Wait()
package proc
