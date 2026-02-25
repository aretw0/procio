package pty_test

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/aretw0/procio/pty"
)

func ExampleStartPTY() {
	// A basic demonstration of creating an interactive PTY session.
	// Note: PTY behavior varies wildly between Windows/Linux.

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "echo", "interactive")

	// StartPTY attaches the command's std streams to a new terminal
	pt, err := pty.StartPTY(cmd)
	if err != nil {
		fmt.Printf("Failed to start PTY: %v\n", err)
		return
	}
	defer pt.Controller.Close()

	// You can now Read/Write from pt.Controller to interact
	// with the child process as if it were a terminal.

	_ = cmd.Wait()
	fmt.Println("PTY session handled")
}
