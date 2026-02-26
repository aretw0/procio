package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/aretw0/procio/proc"
	"github.com/aretw0/procio/scan"
)

func main() {
	fmt.Println("=== Composition Example: Process + Scanner ===")
	fmt.Println("This example shows how to compose proc and scan primitives.")
	fmt.Println("A background process runs while you interact via scanner.")

	// Setup signal-based cancellation
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// 1. Start a background process
	fmt.Println("Starting background process...")
	cmd := startBackgroundProcess(ctx)
	if cmd != nil {
		defer func() {
			if cmd.Process != nil {
				fmt.Println("\nTerminating background process...")
			}
		}()
	}

	// 2. Interactive scanner in foreground
	fmt.Println("\nInteractive mode (type 'status' to check process, Ctrl+C to quit):")
	scanner := scan.NewScanner(os.Stdin,
		scan.WithInterruptible(),
		scan.WithLineHandler(func(line string) {
			handleCommand(line, cmd)
		}),
	)

	scanner.Start(ctx)

	fmt.Println("\n✓ Application stopped gracefully.")
}

func startBackgroundProcess(ctx context.Context) *proc.Cmd {
	// On Windows: timeout 30
	// On Linux: sleep 30
	var cmd *proc.Cmd
	if _, err := exec.LookPath("timeout"); err == nil {
		cmd = proc.NewCmd(ctx, "timeout", "30")
	} else {
		cmd = proc.NewCmd(ctx, "sleep", "30")
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Falha ao iniciar comando: %v\n", err)
		return nil
	}

	fmt.Println("✓ Background process started (will run for ~30s or until cancelled)")

	// Monitor in background
	go func() {
		err := cmd.Wait()
		if err != nil && ctx.Err() == nil {
			fmt.Printf("\n⚠️  Background process exited with error: %v\n", err)
		} else if ctx.Err() == nil {
			fmt.Println("\n✓ Background process completed successfully")
		}
	}()

	return cmd
}

func handleCommand(line string, cmd *proc.Cmd) {
	switch line {
	case "status":
		if cmd == nil {
			fmt.Println("→ No background process running")
		} else if cmd.Process == nil {
			fmt.Println("→ Background process not started")
		} else {
			fmt.Printf("→ Background process running (PID: %d)\n", cmd.Process.Pid)
		}
	case "help":
		fmt.Println("→ Commands: status, help")
	case "":
		// Ignore empty lines
	default:
		fmt.Printf("→ Unknown command: '%s' (try 'help')\n", line)
	}
}
