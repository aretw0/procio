package proc_test

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/aretw0/procio/proc"
)

func ExampleStart() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start a process. proc.Start wraps the standard exec.Cmd to provide
	// platform-specific guarantees (Job Objects on Windows, Pdeathsig on Linux)
	// that ensure child processes are terminated when the parent process exits.
	cmd := exec.CommandContext(ctx, "sleep", "1")

	// Start the process with hygiene guarantees.
	if err := proc.Start(cmd); err != nil {
		fmt.Println("Error starting process:", err)
		return
	}

	// Wait for process to finish or context to be cancelled.
	_ = cmd.Wait()
}

func ExampleNewCmd() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// NewCmd combines exec.CommandContext with automatic platform hygiene.
	// It is the recommended entry point: no need to import os/exec directly.
	cmd := proc.NewCmd(ctx, "sleep", "1")
	if err := proc.Start(cmd); err != nil {
		fmt.Println("Error starting process:", err)
		return
	}

	// Wait for process to finish or context to be cancelled.
	_ = cmd.Wait()
}
