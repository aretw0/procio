package scan

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
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
	// Simulate "Fake EOF" (transient error treated as EOF by some readers, but we handle it via Mock)
	// Actually, Fake EOF logic in scanner relies on count > threshold.
	// Since standard io.Reader returns io.EOF, we can simulate persistent vs transient EOF?
	// Scanner treats io.EOF as error? No, Read returns n=0, err=EOF.
	// wait, Scanner.Start:
	// n, err := s.r.Read(buffer)
	// if err != nil { if s.handleReadError(...) return }
	// handleReadError -> sleeps if not interrupted?
	//
	// If err is io.EOF, it is treated as "Error".
	// Wait, is io.EOF treated as error in Go? Yes.
	// So handleReadError gets io.EOF.
	// handleReadError calls onError(err) and sleeps.
	//
	// Wait, standard Scanner stops on EOF.
	// OUR robust scanner treats EOF as an error and retries unless threshold exceeded.
	// This is specifically for "Fake EOF" on Windows.

	// So if we send 4 EOFs, it should stop (default threshold 3).

	r := &MockReader{
		chunks: [][]byte{},
		err:    io.EOF,
	}

	scanner := NewScanner(r,
		WithBackoff(1*time.Millisecond),
		WithThreshold(3),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()
	scanner.Start(ctx)
	duration := time.Since(start)

	// Should have slept 3-4 times * 1ms.
	// If it blocked forever, ctx timeout would trigger (1s).
	// If it returned immediately, something is wrong.

	if duration > 500*time.Millisecond {
		t.Error("Scanner took too long, probably didn't hit threshold")
	}
}
