package scan

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

type mockProcess struct {
	alive atomic.Bool
}

func (p *mockProcess) IsAlive() bool {
	return p.alive.Load()
}

type fakeEOFReader struct {
	data  []byte
	pos   int
	count int
}

func (r *fakeEOFReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	// Simulate "Fake EOF" every 3 reads
	r.count++
	if r.count%3 == 0 {
		return 0, io.EOF
	}

	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func TestScanner_DeterministicEOF(t *testing.T) {
	data := "line1\nline2\nline3\n"
	reader := &fakeEOFReader{data: []byte(data)}
	proc := &mockProcess{}
	proc.alive.Store(true)

	lines := make([]string, 0)
	s := NewScanner(reader,
		WithLineHandler(func(line string) {
			lines = append(lines, line)
		}),
		WithThreshold(1), // Low threshold to test resilience
		WithBackoff(1*time.Millisecond),
		WithProcess(proc),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 1. First run: Process is alive, should handle fake EOFs and keep going
	go func() {
		time.Sleep(50 * time.Millisecond)
		proc.alive.Store(false) // Finally kill it
	}()

	s.Start(ctx)

	if len(lines) < 3 {
		t.Errorf("expected 3 lines, got %d. lines: %v", len(lines), lines)
	}
	if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestScanner_ImmediateEOFOnProcessDeath(t *testing.T) {
	data := "line1\n"
	reader := &fakeEOFReader{data: []byte(data)}
	proc := &mockProcess{}
	proc.alive.Store(false) // Process already dead

	lines := make([]string, 0)
	s := NewScanner(reader,
		WithLineHandler(func(line string) {
			lines = append(lines, line)
		}),
		WithThreshold(10), // High threshold shouldn't matter if process is dead
		WithProcess(proc),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	s.Start(ctx)
	elapsed := time.Since(start)

	// Since it's dead, it should exit immediately on the first EOF (which our reader gives on 3rd read or end)
	// Actually our reader gives EOF on 3rd read.
	if elapsed > 30*time.Millisecond {
		t.Errorf("expected quick exit, took %v", elapsed)
	}
}
