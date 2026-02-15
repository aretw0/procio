package scan

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aretw0/procio"
	"github.com/aretw0/procio/termio"
)

func TestScanner_Basic(t *testing.T) {
	input := "line1\nline2\nline3"
	reader := strings.NewReader(input)

	var received []string

	scanner := NewScanner(reader,
		WithLineHandler(func(line string) {
			received = append(received, line)
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	scanner.Start(ctx)

	if len(received) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(received))
	}
	if received[0] != "line1" || received[1] != "line2" || received[2] != "line3" {
		t.Errorf("Unexpected content: %v", received)
	}
}

// MockReader allows distinct chunks and simulated errors
type MockReader struct {
	chunks [][]byte
	err    error
	index  int
}

func (m *MockReader) Read(p []byte) (n int, err error) {
	if m.index >= len(m.chunks) {
		if m.err != nil {
			return 0, m.err
		}
		return 0, io.EOF
	}
	chunk := m.chunks[m.index]
	m.index++
	copy(p, chunk)
	return len(chunk), nil
}

type countingEOFReader struct {
	count *int
}

func (r *countingEOFReader) Read(_ []byte) (int, error) {
	*r.count = *r.count + 1
	return 0, io.EOF
}

func TestScanner_PartialReads(t *testing.T) {
	r := &MockReader{
		chunks: [][]byte{
			[]byte("hel"),
			[]byte("lo\n"),
			[]byte("wor"),
			[]byte("ld\n"),
		},
	}

	var received []string
	scanner := NewScanner(r,
		WithLineHandler(func(line string) {
			received = append(received, line)
		}),
	)

	scanner.Start(context.Background())

	if len(received) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(received))
	}
	if received[0] != "hello" {
		t.Errorf("Expected 'hello', got '%s'", received[0])
	}
	if received[1] != "world" {
		t.Errorf("Expected 'world', got '%s'", received[1])
	}
}

func TestScanner_ContextCancel(t *testing.T) {
	// Make it block forever
	blockingReader, writer := io.Pipe()
	defer writer.Close()

	scanner := NewScanner(blockingReader)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		scanner.Start(ctx)
		close(done)
	}()

	// Cancel immediately
	cancel()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Scanner did not exit on context cancellation")
	}
}

func TestScanner_FakeEOF(t *testing.T) {
	// Retry EOF to handle transient "fake EOF" on Windows; stop after threshold + 1.
	readCount := 0
	r := &countingEOFReader{count: &readCount}

	scanner := NewScanner(r,
		WithBackoff(0),
		WithThreshold(3),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	scanner.Start(ctx)

	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("Scanner did not stop after EOF threshold")
	}
	if readCount != 4 {
		t.Errorf("Expected 4 EOF reads, got %d", readCount)
	}
}

// --- WithInterruptible tests ---

func TestScanner_WithInterruptible_Basic(t *testing.T) {
	input := "alpha\nbeta\n"
	reader := strings.NewReader(input)

	var received []string
	scanner := NewScanner(reader,
		WithInterruptible(),
		WithLineHandler(func(line string) {
			received = append(received, line)
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	scanner.Start(ctx)

	if len(received) != 2 {
		t.Errorf("Expected 2 lines, got %d: %v", len(received), received)
	}
	if len(received) == 2 && (received[0] != "alpha" || received[1] != "beta") {
		t.Errorf("Unexpected content: %v", received)
	}
}

func TestScanner_WithInterruptible_ContextCancel(t *testing.T) {
	// InterruptibleReader checks ctx before/after each Read call.
	// Use a slow reader so the read returns and the post-check (or next pre-check) fires.
	sr := &slowChunkReader{
		chunks: []slowChunk{
			{data: "first\n", delay: 10 * time.Millisecond},
			{data: "", delay: 200 * time.Millisecond}, // will be interrupted
		},
	}

	var received []string
	scanner := NewScanner(sr,
		WithInterruptible(),
		WithLineHandler(func(line string) {
			received = append(received, line)
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		scanner.Start(ctx)
		close(done)
	}()

	// Let the first chunk arrive, then cancel during or before the second read
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success — interruptible reader respected cancellation
		if len(received) != 1 || received[0] != "first" {
			t.Errorf("Expected [first], got %v", received)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Scanner with WithInterruptible did not exit on context cancellation")
	}
}

type slowChunk struct {
	data  string
	delay time.Duration
}

type slowChunkReader struct {
	chunks []slowChunk
	index  int
}

func (r *slowChunkReader) Read(p []byte) (int, error) {
	if r.index >= len(r.chunks) {
		return 0, io.EOF
	}
	c := r.chunks[r.index]
	r.index++
	time.Sleep(c.delay)
	if c.data == "" {
		return 0, io.EOF
	}
	return copy(p, c.data), nil
}

func TestScanner_WithInterruptible_Interrupted(t *testing.T) {
	// Simulate an ErrInterrupted from InterruptibleReader
	cancel := make(chan struct{})
	close(cancel) // pre-cancelled

	base := strings.NewReader("data\n")
	ir := termio.NewInterruptibleReader(base, cancel)

	var received []string
	scanner := NewScanner(ir,
		WithLineHandler(func(line string) {
			received = append(received, line)
		}),
		WithBackoff(1*time.Millisecond),
		WithThreshold(1),
	)

	ctx, ctxCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer ctxCancel()
	scanner.Start(ctx)

	// Should have stopped quickly due to ErrInterrupted being treated as interruption
	if len(received) != 0 {
		t.Errorf("Expected no lines (pre-cancelled reader), got %v", received)
	}
}

// --- Observer integration tests ---

type baseObserver struct{}

func (baseObserver) OnProcessStarted(int)    {}
func (baseObserver) OnProcessFailed(error)   {}
func (baseObserver) OnIOError(string, error) {}
func (baseObserver) OnScanError(error)       {}
func (baseObserver) LogDebug(string, ...any) {}
func (baseObserver) LogWarn(string, ...any)  {}
func (baseObserver) LogError(string, ...any) {}

type testObserver struct {
	baseObserver
	scanErrors []error
	ioErrors   []string
}

func (o *testObserver) OnScanError(err error)            { o.scanErrors = append(o.scanErrors, err) }
func (o *testObserver) OnIOError(op string, err error)   { o.ioErrors = append(o.ioErrors, op) }
func (o *testObserver) LogWarn(msg string, args ...any)  {}
func (o *testObserver) LogError(msg string, args ...any) {}

func TestScanner_ObserverOnScanError_NoHandler(t *testing.T) {
	// When no WithErrorHandler is set, scan errors should go to the Observer.
	obs := &testObserver{}
	procio.SetObserver(obs)
	defer procio.SetObserver(nil)

	customErr := errors.New("custom disk error")
	callCount := 0
	r := &errorAfterReader{
		data:    "line\n",
		err:     customErr,
		maxErrs: 2,
		count:   &callCount,
	}

	scanner := NewScanner(r,
		WithBackoff(1*time.Millisecond),
		WithThreshold(1),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	scanner.Start(ctx)

	if len(obs.scanErrors) == 0 {
		t.Error("Expected OnScanError to be called, got 0 calls")
	}
	for _, err := range obs.scanErrors {
		if !errors.Is(err, customErr) {
			t.Errorf("Expected custom error, got %v", err)
		}
	}
}

func TestScanner_ObserverNotCalled_WithHandler(t *testing.T) {
	// When WithErrorHandler IS set, Observer.OnScanError should NOT be called.
	obs := &testObserver{}
	procio.SetObserver(obs)
	defer procio.SetObserver(nil)

	customErr := errors.New("handled error")
	callCount := 0
	r := &errorAfterReader{
		data:    "line\n",
		err:     customErr,
		maxErrs: 2,
		count:   &callCount,
	}

	var handlerCalled bool
	scanner := NewScanner(r,
		WithBackoff(1*time.Millisecond),
		WithThreshold(1),
		WithErrorHandler(func(err error) {
			handlerCalled = true
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	scanner.Start(ctx)

	if !handlerCalled {
		t.Error("Expected WithErrorHandler to be called")
	}
	if len(obs.scanErrors) > 0 {
		t.Errorf("Observer.OnScanError should not be called when handler is set, got %d calls", len(obs.scanErrors))
	}
}

// errorAfterReader returns data once, then returns err up to maxErrs times, then EOF.
type errorAfterReader struct {
	data    string
	err     error
	maxErrs int
	count   *int
	sent    bool
}

func (r *errorAfterReader) Read(p []byte) (int, error) {
	if !r.sent {
		r.sent = true
		n := copy(p, r.data)
		return n, nil
	}
	*r.count++
	if *r.count <= r.maxErrs {
		return 0, r.err
	}
	return 0, io.EOF
}
