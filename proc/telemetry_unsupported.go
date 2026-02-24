//go:build !linux && !windows

package proc

import (
	"context"
	"errors"
	"time"
)

var errProcessNotStarted = errors.New("proc: Monitor called before process was started (cmd.Process is nil)")

func monitorLoop(_ context.Context, _ int, _ time.Duration, ch chan<- Metrics) {
	close(ch)
}
