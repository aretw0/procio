package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aretw0/procio/scan"
)

func main() {
	fmt.Println("=== Interruptible Scanner Example ===")
	fmt.Println("This example shows how to use WithInterruptible() for Ctrl+C support.")
	fmt.Println("Type lines and press Enter. Press Ctrl+C to exit gracefully.")

	// Create a context that cancels on OS signals (Ctrl+C)
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Create scanner with interruptible mode enabled
	scanner := scan.NewScanner(os.Stdin,
		scan.WithInterruptible(), // Enables context cancellation
		scan.WithLineHandler(func(line string) {
			fmt.Printf("→ You typed: %s\n", line)
		}),
	)

	// Start will:
	// 1. Upgrade stdin to a safe terminal handle (CONIN$ on Windows)
	// 2. Create an interruptible reader that checks context before/after reads
	// 3. Return when context is cancelled or EOF is reached
	scanner.Start(ctx)

	fmt.Println("\n✓ Scanner stopped gracefully. Goodbye!")
}
