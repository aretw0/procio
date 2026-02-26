package proc

import (
	"context"
	"os/exec"
	"time"
)

// Metrics is a snapshot of resource usage for a running process.
type Metrics struct {
	// PID is the process ID this snapshot belongs to.
	PID int

	// CPUPercent is the estimated CPU usage percentage since the previous
	// snapshot (0–100 per logical core; may exceed 100 on multi-core systems).
	CPUPercent float64

	// MemRSS is the Resident Set Size in bytes: the portion of memory
	// currently held in RAM.
	MemRSS int64
}

// Monitor polls the resource usage of cmd.Process at the given interval and
// sends each [Metrics] snapshot to the returned channel.
//
// Monitor returns immediately. The sampling goroutine runs until:
//   - ctx is cancelled, or
//   - the channel's consumer stops reading (the channel is unbuffered).
//
// The returned channel is closed when monitoring stops. Monitor returns an
// error only if the process has not been started yet (cmd.Process == nil).
//
// Example:
//
//	cmd := proc.NewCmd(ctx, "worker", "--config", "prod.yaml")
//	if err := cmd.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	ch, err := proc.Monitor(ctx, cmd.Cmd, time.Second)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for m := range ch {
//	    fmt.Printf("pid=%d cpu=%.1f%% mem=%d KB\n", m.PID, m.CPUPercent, m.MemRSS/1024)
//	}
func Monitor(ctx context.Context, cmd *exec.Cmd, interval time.Duration) (<-chan Metrics, error) {
	if cmd.Process == nil {
		return nil, errProcessNotStarted
	}
	// Safety floor: prevent tight loops that could exhaust CPU.
	if interval < 10*time.Millisecond {
		interval = 10 * time.Millisecond
	}
	ch := make(chan Metrics)
	go monitorLoop(ctx, cmd.Process.Pid, interval, ch)
	return ch, nil
}
