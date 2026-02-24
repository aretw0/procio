//go:build windows

package proc

import (
	"context"
	"errors"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var errProcessNotStarted = errors.New("proc: Monitor called before process was started (cmd.Process is nil)")

// processMemoryCounters is PROCESS_MEMORY_COUNTERS from psapi.h.
// golang.org/x/sys/windows does not expose this struct in v0.40.
type processMemoryCounters struct {
	Cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

var (
	modPsapi                 = windows.NewLazySystemDLL("psapi.dll")
	procGetProcessMemoryInfo = modPsapi.NewProc("GetProcessMemoryInfo")
)

func getProcessMemoryInfo(h windows.Handle) (processMemoryCounters, error) {
	var mem processMemoryCounters
	mem.Cb = uint32(unsafe.Sizeof(mem))
	r, _, err := procGetProcessMemoryInfo.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(&mem)),
		uintptr(mem.Cb),
	)
	if r == 0 {
		return mem, err
	}
	return mem, nil
}

// monitorLoop is the Windows telemetry sampler.
// Uses GetProcessTimes for CPU and psapi.GetProcessMemoryInfo for RSS.
func monitorLoop(ctx context.Context, pid int, interval time.Duration, ch chan<- Metrics) {
	defer close(ch)

	handle, err := windows.OpenProcess(
		windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		uint32(pid),
	)
	if err != nil {
		return
	}
	defer windows.CloseHandle(handle)

	var prevKernel, prevUser windows.Filetime
	var prevTime time.Time

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			var creation, exit, kernel, user windows.Filetime
			if err := windows.GetProcessTimes(handle, &creation, &exit, &kernel, &user); err != nil {
				return
			}

			mem, err := getProcessMemoryInfo(handle)
			if err != nil {
				return
			}

			m := Metrics{
				PID:    pid,
				MemRSS: int64(mem.WorkingSetSize),
			}

			if !prevTime.IsZero() {
				elapsed := t.Sub(prevTime).Seconds()
				// FILETIME ticks are 100 nanoseconds each.
				kernelDelta := filetimeDiff(kernel, prevKernel)
				userDelta := filetimeDiff(user, prevUser)
				cpuSeconds := float64(kernelDelta+userDelta) * 1e-7
				m.CPUPercent = (cpuSeconds / elapsed) * 100.0
			}

			prevKernel = kernel
			prevUser = user
			prevTime = t

			select {
			case ch <- m:
			case <-ctx.Done():
				return
			}
		}
	}
}

// filetimeDiff returns the difference of two FILETIME values in 100-ns ticks.
func filetimeDiff(a, b windows.Filetime) uint64 {
	av := uint64(a.HighDateTime)<<32 | uint64(a.LowDateTime)
	bv := uint64(b.HighDateTime)<<32 | uint64(b.LowDateTime)
	if av < bv {
		return 0
	}
	return av - bv
}
