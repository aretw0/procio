//go:build linux

package proc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var errProcessNotStarted = errors.New("proc: Monitor called before process was started (cmd.Process is nil)")

// monitorLoop is the Linux telemetry sampler.
// It reads /proc/[pid]/stat and /proc/[pid]/statm periodically.
func monitorLoop(ctx context.Context, pid int, interval time.Duration, ch chan<- Metrics) {
	defer close(ch)

	var prevCPU uint64
	var prevTime time.Time

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			m, cpu, err := readProcStat(pid)
			if err != nil {
				// Process likely exited.
				return
			}

			if !prevTime.IsZero() {
				elapsed := t.Sub(prevTime).Seconds()
				cpuDelta := float64(cpu - prevCPU)
				// Convert kernel jiffies (typically 100/s) to percent.
				clkTck := float64(100) // _SC_CLK_TCK; 100 is correct on all modern Linux
				m.CPUPercent = (cpuDelta / clkTck / elapsed) * 100.0
			}
			prevCPU = cpu
			prevTime = t

			select {
			case ch <- m:
			case <-ctx.Done():
				return
			}
		}
	}
}

// readProcStat parses /proc/[pid]/stat and /proc/[pid]/statm.
// Returns a partial Metrics (PID, MemRSS) and totalCPU ticks (utime+stime).
func readProcStat(pid int) (Metrics, uint64, error) {
	m := Metrics{PID: pid}

	// --- /proc/[pid]/stat ---
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return m, 0, err
	}
	// The comm field (field 2) is enclosed in parentheses and may contain spaces.
	// Find the last ')' to skip it safely.
	lastParen := strings.LastIndex(string(data), ")")
	if lastParen < 0 {
		return m, 0, fmt.Errorf("proc/stat: malformed: %s", statPath)
	}
	fields := strings.Fields(string(data)[lastParen+1:])
	// After the closing paren, fields are offset by 2 (fields 3..):
	// index 0 = field 3  (state)
	// index 11 = field 14 (utime)
	// index 12 = field 15 (stime)
	if len(fields) < 13 {
		return m, 0, fmt.Errorf("proc/stat: too few fields in %s", statPath)
	}
	utime, _ := strconv.ParseUint(fields[11], 10, 64)
	stime, _ := strconv.ParseUint(fields[12], 10, 64)
	totalCPU := utime + stime

	// --- /proc/[pid]/statm ---
	statmPath := fmt.Sprintf("/proc/%d/statm", pid)
	f, err := os.Open(statmPath)
	if err != nil {
		return m, totalCPU, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Scan()
	statmFields := strings.Fields(sc.Text())
	// Field 1 = resident pages.
	if len(statmFields) >= 2 {
		pages, _ := strconv.ParseInt(statmFields[1], 10, 64)
		const pageSize = 4096 // getconf PAGESIZE; 4096 is universal on modern Linux
		m.MemRSS = pages * pageSize
	}

	return m, totalCPU, nil
}
