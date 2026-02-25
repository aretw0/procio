package procio_test

import (
	"context"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aretw0/procio"
	"github.com/aretw0/procio/proc"
	"github.com/aretw0/procio/scan"
)

// integrationObserver tracks all events across proc, termio, and scan
type integrationObserver struct {
	processStartedCount atomic.Int32
	processFailedCount  atomic.Int32
	ioErrorCount        atomic.Int32
	scanErrorCount      atomic.Int32
}

func (o *integrationObserver) OnProcessStarted(pid int) {
	o.processStartedCount.Add(1)
}

func (o *integrationObserver) OnProcessFailed(err error) {
	o.processFailedCount.Add(1)
}

func (o *integrationObserver) OnIOError(operation string, err error) {
	o.ioErrorCount.Add(1)
}

func (o *integrationObserver) OnScanError(err error) {
	o.scanErrorCount.Add(1)
}

func (o *integrationObserver) LogDebug(msg string, args ...any) {}
func (o *integrationObserver) LogInfo(msg string, args ...any)  {}
func (o *integrationObserver) LogWarn(msg string, args ...any)  {}
func (o *integrationObserver) LogError(msg string, args ...any) {}

// TestIntegration_ProcessAndScanner verifies that proc and scan can be composed
func TestIntegration_ProcessAndScanner(t *testing.T) {
	// Setup observer
	observer := &integrationObserver{}
	procio.SetObserver(observer)
	t.Cleanup(func() { procio.SetObserver(nil) })

	// Create a process that will run and exit cleanly
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use a simple command that exists on both platforms
	cmd := exec.CommandContext(ctx, "go", "version")

	// Start process
	if err := proc.Start(cmd); err != nil {
		t.Fatalf("proc.Start failed: %v", err)
	}

	// Wait for process in background
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Create a scanner from a string reader
	input := strings.NewReader("line1\nline2\nline3\n")
	lines := make([]string, 0, 3)

	scanner := scan.NewScanner(input,
		scan.WithLineHandler(func(line string) {
			lines = append(lines, line)
		}),
	)

	// Start scanner
	scanner.Start(ctx)

	// Verify scanner captured all lines
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}

	// Wait for process to complete
	select {
	case err := <-done:
		if err != nil && ctx.Err() == nil {
			t.Errorf("process failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("process did not complete in time")
	}

	// Verify observer received process started event
	if count := observer.processStartedCount.Load(); count != 1 {
		t.Errorf("expected 1 OnProcessStarted call, got %d", count)
	}
}

// TestIntegration_ContextCancellationPropagation verifies context cancellation
// is properly handled across all components
func TestIntegration_ContextCancellationPropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create a slow reader that will be interrupted
	slowReader := &slowReader{delay: 500 * time.Millisecond} // Longer delay
	lineReceived := make(chan string, 1)

	scanner := scan.NewScanner(slowReader,
		scan.WithLineHandler(func(line string) {
			lineReceived <- line
		}),
	)

	// Start scanner in background
	done := make(chan struct{})
	go func() {
		scanner.Start(ctx)
		close(done)
	}()

	// Cancel context immediately (before first read completes)
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Verify scanner stops
	select {
	case <-done:
		// Success - scanner stopped
	case <-time.After(2 * time.Second):
		t.Fatal("scanner did not stop after context cancellation")
	}

	// Note: Due to timing, a line might be received if read completed before cancel.
	// The important part is that scanner.Start() returned promptly after cancel.
}

// TestIntegration_ObserverReceivesAllEvents verifies the observer pattern
// works end-to-end across all primitives
func TestIntegration_ObserverReceivesAllEvents(t *testing.T) {
	observer := &integrationObserver{}
	procio.SetObserver(observer)
	t.Cleanup(func() { procio.SetObserver(nil) })

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// 1. Start a process (should trigger OnProcessStarted)
	cmd := exec.CommandContext(ctx, "go", "version")
	if err := proc.Start(cmd); err != nil {
		t.Fatalf("proc.Start failed: %v", err)
	}

	go func() { _ = cmd.Wait() }()

	// 2. Create scanner with error (should trigger OnScanError when no handler)
	errReader := &errorReader{err: io.ErrUnexpectedEOF}
	scanner := scan.NewScanner(errReader)
	scanner.Start(ctx)

	// Wait a bit for events to propagate
	time.Sleep(100 * time.Millisecond)

	// Verify observer received events
	if count := observer.processStartedCount.Load(); count < 1 {
		t.Errorf("expected at least 1 OnProcessStarted call, got %d", count)
	}

	if count := observer.scanErrorCount.Load(); count < 1 {
		t.Errorf("expected at least 1 OnScanError call, got %d", count)
	}
}

// slowReader simulates a slow I/O source
type slowReader struct {
	delay time.Duration
	data  []byte
	pos   int
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	if r.data == nil {
		r.data = []byte("slow data\n")
	}

	time.Sleep(r.delay)

	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// errorReader always returns an error
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
