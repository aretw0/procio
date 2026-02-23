package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aretw0/procio/proc"
	"github.com/aretw0/procio/scan"
)

func main() {
	// 1. Process Management Guarantee
	// This command will be killed even if this main process panics or is killed.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fmt.Println("Starting a long-running process (2s timeout)...")
	// Using a command that works on both Windows and Linux to show execution.
	// On Windows, 'ping -n 1 127.0.0.1' takes a second.
	// On Linux, 'ping -c 1 127.0.0.1' takes a second.
	cmd := proc.NewCmd(ctx, "ping", "127.0.0.1")
	if err := proc.Start(cmd); err != nil {
		fmt.Printf("Note: Could not start 'ping' (expected on some restricted environments): %v\n", err)
	} else {
		fmt.Println("Process started. It will be killed by the OS if this app crashes.")
		// We wait for it in the background to show it's tracked.
		go func() {
			_ = cmd.Wait()
			fmt.Println("\nBackground process finished.")
		}()
	}

	fmt.Println("\n--- Interactive Input Demo ---")
	fmt.Println("Type something (or Ctrl+C to quit):")
	scanner := scan.NewScanner(os.Stdin,
		scan.WithLineHandler(func(line string) {
			fmt.Printf("Echo: %s\n", line)
		}),
	)

	// scanner.Start blocks until EOF or Context Cancelled (e.g. via Signal)
	scanner.Start(context.Background())

	fmt.Println("\nExiting. Bye!")
}
