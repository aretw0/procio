// Package proc provides platform-agnostic process management with safety guarantees.
//
// It abstracts away OS-specific details for reliable process termination:
//   - Windows: Uses Job Objects to ensure child processes are killed when the parent exits.
//   - Linux: Uses Pdeathsig (SIGKILL) for the same guarantee.
//
// Usage:
//
//	ctx := context.Background()
//	cmd := proc.Start(ctx, "ping", "google.com")
//	cmd.Wait()
package proc
