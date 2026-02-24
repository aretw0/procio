package pty_test

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/aretw0/procio/pty"
)

func TestPTY_StartAndRead(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", "echo hello pty")
	} else {
		cmd = exec.CommandContext(ctx, "echo", "hello pty")
	}

	p, err := pty.StartPTY(cmd)
	if err != nil {
		t.Fatalf("StartPTY failed: %v", err)
	}
	defer p.Close()

	// Read from the controller in a background goroutine to avoid blocking forever
	// and to continuously consume until we find our expected output. Windows ConPTY
	// often emits escape sequences before the actual process output.
	outCh := make(chan string, 1)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 128)
		for {
			n, err := p.Controller.Read(buf)
			if n > 0 {
				sb.Write(buf[:n])
				if strings.Contains(sb.String(), "hello pty") {
					outCh <- sb.String()
					return
				}
			}
			if err != nil {
				outCh <- sb.String()
				return
			}
		}
	}()

	var out string
	select {
	case out = <-outCh:
	case <-ctx.Done():
		t.Fatal("Timeout waiting for expected output")
	}

	// Output may contain CR/LF depending on platform.
	if !strings.Contains(out, "hello pty") {
		t.Errorf("Expected output to contain 'hello pty', got %q", out)
	}

	// Wait for the command to finish.
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Logf("cmd.Wait returned error (expected with raw shell kills): %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Timeout waiting for command wait")
	}
}
