package proc_test

import (
	"context"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/aretw0/procio/proc"
)

func TestMonitor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := getSleepCmd(ctx)
	if err := proc.Start(cmd); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	ch, err := proc.Monitor(ctx, cmd, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Monitor failed: %v", err)
	}

	var metricsReceived int
	for m := range ch {
		metricsReceived++
		if m.PID != cmd.Process.Pid {
			t.Errorf("Expected PID %d, got %d", cmd.Process.Pid, m.PID)
		}
		// CPU and RSS might be 0 for a sleeping short-lived process, but they shouldn't be negative.
		if m.CPUPercent < 0 {
			t.Errorf("Expected CPUPercent >= 0, got %f", m.CPUPercent)
		}
		if m.MemRSS < 0 {
			t.Errorf("Expected MemRSS >= 0, got %d", m.MemRSS)
		}
		if metricsReceived >= 3 {
			// We got enough samples, let's stop.
			cancel()
		}
	}

	if metricsReceived == 0 && runtime.GOOS != "darwin" && runtime.GOOS != "windows" { // Fallbacks return immediately
		t.Errorf("No metrics received on supported platform")
	}

	_ = cmd.Wait()
}

func TestMonitor_NotStarted(t *testing.T) {
	cmd := proc.NewCmd(context.Background(), "echo")
	_, err := proc.Monitor(context.Background(), cmd, time.Second)
	if err == nil {
		t.Error("Expected error calling Monitor on unstarted process, got nil")
	}
}

func getSleepCmd(ctx context.Context) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return proc.NewCmd(ctx, "powershell", "-Command", "Start-Sleep -Seconds 3")
	}
	return proc.NewCmd(ctx, "sleep", "3")
}
