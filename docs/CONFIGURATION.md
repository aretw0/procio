# Configuration

Usage of `procio` is generally configuration-free for simple cases. However, global behaviors can be tuned via the `procio` facade or package-level variables.

## Global Observability

`procio` does not use any logging library by default. It uses an `Observer` interface to allow you to plug in your own logging/metrics solution.

```go
type Observer interface {
   OnProcessStarted(pid int)
   OnProcessFailed(err error)
   OnIOError(op string, err error)
   OnScanError(err error)
   LogDebug(msg string, args ...any)
   LogWarn(msg string, args ...any)
   LogError(msg string, args ...any)
}
```

To configure:

```go
import "github.com/aretw0/procio"

procio.SetObserver(myLogger)
```

## Process strictness

### `proc.StrictMode`

By default (`false`), `proc.Start` will attempt to use OS-specific features (like Job Objects on Windows or Pdeathsig on Linux) but will degrade gracefully if they fail or are unavailable, logging a warning to the Observer.

If set to `true`, `proc.Start` will return an error if it cannot guarantee that the child process will be killed when the parent dies.

```go
import "github.com/aretw0/procio/proc"

proc.StrictMode = true
```

## Scanner Tuning

The `scan.Scanner` has internal defaults that can be adjusted via options:

- `WithBufferSize(size int)`: Sets the read buffer size.
- `WithThreshold(count int)`: Sets the consecutive EOF threshold for Windows "fake EOF" detection.
- `WithProcess(cmd *proc.Cmd)`: Binds the scanner to a process for deterministic EOF detection (v0.5.0+).
- `WithProcessLiveness(fn func() bool)`: Binds any liveness check for deterministic EOF (v0.5.0+).
