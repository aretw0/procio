package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aretw0/procio"
	"github.com/aretw0/procio/proc"
	"github.com/aretw0/procio/scan"
)

// customObserver implements procio.Observer with custom logging
type customObserver struct {
	logger *log.Logger
}

func (o *customObserver) OnProcessStarted(pid int) {
	o.logger.Printf("[PROCESS] Started: PID %d", pid)
}

func (o *customObserver) OnProcessFailed(err error) {
	o.logger.Printf("[PROCESS] Failed: %v", err)
}

func (o *customObserver) OnIOError(operation string, err error) {
	o.logger.Printf("[IO] Error in %s: %v", operation, err)
}

func (o *customObserver) OnScanError(err error) {
	o.logger.Printf("[SCAN] Error: %v", err)
}

func (o *customObserver) LogDebug(msg string, args ...any) {
	o.logger.Printf("[DEBUG] "+msg, args...)
}

func (o *customObserver) LogInfo(msg string, args ...any) {
	o.logger.Printf("[INFO] "+msg, args...)
}

func (o *customObserver) LogWarn(msg string, args ...any) {
	o.logger.Printf("[WARN] "+msg, args...)
}

func (o *customObserver) LogError(msg string, args ...any) {
	o.logger.Printf("[ERROR] "+msg, args...)
}

func main() {
	fmt.Println("=== Custom Observer Example ===")
	fmt.Println("This example shows how to implement and inject a custom observer.")

	// Create and inject custom observer
	observer := &customObserver{
		logger: log.New(os.Stdout, "[procio] ", log.Ltime),
	}
	procio.SetObserver(observer)

	// 1. Process lifecycle event
	fmt.Println("1. Starting a process (will generate OnProcessStarted event)...")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := proc.NewCmd(ctx, "ping", "127.0.0.1")
	if err := cmd.Start(); err != nil {
		fmt.Printf("Process start failed (expected on some systems): %v\n", err)
	} else {
		go func() {
			_ = cmd.Wait()
		}()
	}

	time.Sleep(500 * time.Millisecond)

	// 2. Scanner with observer hooks
	fmt.Println("\n2. Starting scanner (type a line or Ctrl+C)...")
	fmt.Println("WithInterruptible() will upgrade the terminal (OnIOError if it fails).")

	scanCtx, scanCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer scanCancel()

	scanner := scan.NewScanner(os.Stdin,
		scan.WithInterruptible(), // May trigger OnIOError during terminal upgrade
		scan.WithLineHandler(func(line string) {
			fmt.Printf("Got line: %s\n", line)
		}),
	)

	scanner.Start(scanCtx)

	fmt.Println("\n✓ Observer received telemetry from process and I/O operations.")
	fmt.Println("Check the [procio] prefixed logs above to see observer calls.")
}
