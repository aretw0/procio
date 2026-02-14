package scan_test

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aretw0/procio/scan"
)

func ExampleScanner_Start() {
	// Simulate input
	input := "hello\nworld"
	reader := strings.NewReader(input)

	// Create scanner
	scanner := scan.NewScanner(reader,
		scan.WithLineHandler(func(line string) {
			fmt.Printf("Received: %s\n", line)
		}),
	)

	// Block until EOF
	scanner.Start(context.Background())

	// Output:
	// Received: hello
	// Received: world
}

func ExampleNewScanner() {
	// Read from Stdin with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scanner := scan.NewScanner(os.Stdin,
		scan.WithLineHandler(func(line string) {
			fmt.Println("User typed:", line)
		}),
	)

	// Run in a goroutine if you need non-blocking start
	go scanner.Start(ctx)
}

func ExampleNewScanner_interruptible() {
	// WithInterruptible combines termio.Upgrade + InterruptibleReader
	// so context cancellation works even on blocking terminal reads.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scanner := scan.NewScanner(os.Stdin,
		scan.WithInterruptible(),
		scan.WithLineHandler(func(line string) {
			fmt.Println("User typed:", line)
		}),
	)

	go scanner.Start(ctx)
}
