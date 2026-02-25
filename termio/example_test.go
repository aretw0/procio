package termio_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aretw0/procio/termio"
)

func ExampleNewInterruptibleReader() {
	ctx, cancel := context.WithCancel(context.Background())

	// A reader that simulates continuous input
	reader := strings.NewReader("hello\nworld\n")

	// Wrap it with InterruptibleReader
	ir := termio.NewInterruptibleReader(reader, ctx.Done())

	// Read in a loop
	go func() {
		buf := make([]byte, 10)
		for {
			n, err := ir.Read(buf)
			if err != nil {
				if err == io.EOF {
					fmt.Println("EOF reached")
				} else {
					// Likely context cancellation propagated
					fmt.Printf("Read error: %v\n", err)
				}
				return
			}
			fmt.Printf("Read %d bytes\n", n)
		}
	}()

	// Process for a bit, then cancel the context
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for goroutine to yield
	time.Sleep(50 * time.Millisecond)

	// Since NewReader is fast, it might just hit EOF before the cancel.
	// This example demonstrates structure rather than precise timing.
}
