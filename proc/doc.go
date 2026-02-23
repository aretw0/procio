// Package proc provides platform-agnostic process management with safety guarantees.
//
// It abstracts away OS-specific details for reliable process termination:
//   - Windows: Uses Job Objects to ensure child processes are killed when the parent exits.
//   - Linux: Uses Pdeathsig (SIGKILL) for the same guarantee.
//
// # Recommended Usage
//
// Use [NewCmd] to create a command with context-linked cancellation and then
// call [Start] to apply platform hygiene and launch the process:
//
//	cmd := proc.NewCmd(ctx, "worker", "--config", "prod.yaml")
//	cmd.Stdout = os.Stdout // configure before starting
//	if err := proc.Start(cmd); err != nil {
//	    log.Fatal(err)
//	}
//	cmd.Wait()
//
// # Legacy Usage
//
// If you must construct the command yourself (e.g. to set SysProcAttr),
// use exec.CommandContext directly and pass to Start:
//
//	cmd := exec.CommandContext(ctx, "ping", "google.com")
//	_ = proc.Start(cmd)
//	cmd.Wait()
package proc
