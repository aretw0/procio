//go:build ignore

// telemetry demonstrates real-time CPU and memory monitoring of a child process
// using proc.Monitor.
//
// Usage:
//
//	go run examples/telemetry/main.go [command [args...]]
//
// If no command is specified, it defaults to a platform-appropriate busy loop.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aretw0/procio/proc"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = defaultWorkload()
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cmd := proc.NewCmd(ctx, args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := proc.Start(cmd); err != nil {
		log.Fatalf("start: %v", err)
	}

	ch, err := proc.Monitor(ctx, cmd, time.Second)
	if err != nil {
		log.Fatalf("monitor: %v", err)
	}

	fmt.Printf("%-8s  %-10s  %-12s\n", "PID", "CPU%", "MEM (KB)")
	fmt.Println("----------------------------------------")
	for m := range ch {
		fmt.Printf("%-8d  %-10.2f  %-12d\n", m.PID, m.CPUPercent, m.MemRSS/1024)
	}

	_ = cmd.Wait()
}

func defaultWorkload() []string {
	// A simple, portable busy-loop command.
	if _, err := os.Stat("/bin/sh"); err == nil {
		return []string{"sh", "-c", "while true; do :; done"}
	}
	// Fallback for Windows: busy ping loop.
	return []string{"cmd", "/c", "for /l %i in () do echo ."}
}
