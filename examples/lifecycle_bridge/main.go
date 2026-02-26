// Package main demonstrates how to integrate procio with lifecycle.
//
// This example shows the canonical ObserverBridge pattern (ADR-11):
// a single struct that satisfies both procio.Observer and lifecycle.Observer,
// eliminating the need for any adapter wrapper.
//
// This file also acts as a compile-time contract test: if procio.Observer
// or lifecycle.Observer changes in a breaking way, this file fails to build.
//
// Run: go run ./examples/lifecycle_bridge/
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aretw0/procio"
	"github.com/aretw0/procio/proc"
)

// =============================================================================
// ObserverBridge: implements procio.Observer.
//
// In a real application using lifecycle, this struct would also implement
// lifecycle.Observer by adding OnGoroutinePanicked(v any, stack []byte).
// Here we keep it lightweight to avoid importing lifecycle.
// =============================================================================

type ObserverBridge struct {
	logger *slog.Logger
}

func NewObserverBridge() *ObserverBridge {
	return &ObserverBridge{
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}
}

func (b *ObserverBridge) OnProcessStarted(pid int) {
	b.logger.Info("process started", "pid", pid)
}

func (b *ObserverBridge) OnProcessFailed(err error) {
	b.logger.Error("process failed", "err", err)
}

func (b *ObserverBridge) OnIOError(op string, err error) {
	b.logger.Error("io error", "op", op, "err", err)
}

func (b *ObserverBridge) OnScanError(err error) {
	b.logger.Error("scan error", "err", err)
}

func (b *ObserverBridge) LogDebug(msg string, args ...any) { b.logger.Debug(msg, args...) }
func (b *ObserverBridge) LogWarn(msg string, args ...any)  { b.logger.Warn(msg, args...) }
func (b *ObserverBridge) LogError(msg string, args ...any) { b.logger.Error(msg, args...) }

// Compile-time check: ObserverBridge must satisfy procio.Observer.
// If procio.Observer gains or removes a method, this line fails to build,
// alerting the maintainer to update all bridge implementations.
var _ procio.Observer = (*ObserverBridge)(nil)

// =============================================================================
// Worker Contract: canonical pattern for embedding proc.Start in a worker.
//
// Compare with lifecycle.ProcessWorker in pkg/core/worker/process.go.
// The key difference from the naive approach: proc.NewCmd (not exec.Command)
// binds context cancellation, so appCtx cancel propagates to the subprocess.
// =============================================================================

func runProcess(ctx context.Context, binary string, args ...string) error {
	// Context Contract: derive from the application context.
	// Cancelling ctx (e.g. on SIGINT) terminates the subprocess.
	subCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Use proc.NewCmd, NOT exec.Command.
	// proc.Start applies platform hygiene (Pdeathsig / Job Objects).
	cmd := proc.NewCmd(subCtx, binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", binary, err)
	}
	return cmd.Wait()
}

func main() {
	// Wire the observer bridge once at startup.
	bridge := NewObserverBridge()
	procio.SetObserver(bridge)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.Info("lifecycle_bridge example: running 'echo hello world'")

	if err := runProcess(ctx, "echo", "hello", "world"); err != nil {
		slog.Error("process error", "err", err)
		os.Exit(1)
	}

	slog.Info("done")
}
