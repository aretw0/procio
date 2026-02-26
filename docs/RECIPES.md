# Recipes

Common patterns for using `procio`.

## 1. Running a Process with Cleanup Guarantee

```go
import (
    "context"
    "github.com/aretw0/procio/proc"
)

func main() {
    ctx := context.Background()
    // proc.NewCmd = exec.CommandContext + platform hygiene (Job Objects / Pdeathsig)
    cmd := proc.NewCmd(ctx, "sleep", "100")
    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }
    cmd.Wait() // wait or let context kill it
}
```

## 2. Reading Standard Input Safely

Reading from `stdin` in a way that respects `context.Context` cancellation (unblocking the read).

```go
import (
    "context"
    "os"
    "github.com/aretw0/procio/scan"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    scanner := scan.NewScanner(os.Stdin, scan.WithLineHandler(func(line string) {
        println("User wrote:", line)
    }))

    // Blocks until EOF or Context Cancelled
    scanner.Start(ctx) 
}
```

## 3. Configuring Logging

Connecting `procio` to `log/slog`.

```go
import (
    "log/slog"
    "github.com/aretw0/procio"
)

type SlogAdapter struct{}

func (s SlogAdapter) OnProcessStarted(pid int) { slog.Info("process started", "pid", pid) }
func (s SlogAdapter) OnProcessFailed(err error) { slog.Error("process failed", "err", err) }
func (s SlogAdapter) OnIOError(op string, err error) { slog.Error("io error", "op", op, "err", err) }
func (s SlogAdapter) OnScanError(err error) { slog.Error("scan error", "err", err) }
func (s SlogAdapter) LogDebug(msg string, args ...any) { slog.Debug(msg, args...) }
func (s SlogAdapter) LogWarn(msg string, args ...any) { slog.Warn(msg, args...) }
func (s SlogAdapter) LogError(msg string, args ...any) { slog.Error(msg, args...) }

func init() {
    procio.SetObserver(SlogAdapter{})
}
```

## 4. Subprocess with Deadline and Parent Cancellation (Chained Cancels)

When running subprocesses inside a larger application, derive the subprocess context from the application context. Cancelling the parent automatically cascades.

```go
import (
    "context"
    "log"
    "time"
    "github.com/aretw0/procio/proc"
)

func runWorker(appCtx context.Context) error {
    // Constrain the worker to at most 30s, but respect app shutdown.
    ctx, cancel := context.WithTimeout(appCtx, 30*time.Second)
    defer cancel()

    cmd := proc.NewCmd(ctx, "worker", "--mode", "batch")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        return fmt.Errorf("start worker: %w", err)
    }
    return cmd.Wait()
}
```

## 5. Process + Scanner Pipeline

Running a background process while reading its output through a `scan.Scanner`.

```go
import (
    "context"
    "fmt"
    "os/signal"
    "syscall"
    "github.com/aretw0/procio/proc"
    "github.com/aretw0/procio/scan"
)

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    cmd := proc.NewCmd(ctx, "tail", "-f", "/var/log/app.log")

    // Pipe stdout to a scanner for line-by-line processing.
    stdout, _ := cmd.StdoutPipe()
    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }

    scanner := scan.NewScanner(stdout, 
        scan.WithProcess(cmd), // v0.5.0+: deterministic EOF detection
        scan.WithLineHandler(func(line string) {
            fmt.Println("[log]", line)
        }),
    )
    scanner.Start(ctx) // blocks until process exits, ctx is done, or pipe closes

    cmd.Wait()
}
```

## 6. Running an Interactive Application in a PTY

When you need to wrap interactive tools like `vim`, `htop`, or `ssh` that require a real terminal.

```go
import (
    "context"
    "io"
    "os"
    "github.com/aretw0/procio/proc"
    "github.com/aretw0/procio/pty"
    "github.com/aretw0/procio/termio"
)

func main() {
    ctx := context.Background()
    cmd := exec.CommandContext(ctx, "vim")
    
    // Allocate PTY and attach it to the command
    p, err := pty.StartPTY(cmd)
    if err != nil {
        panic(err)
    }
    defer p.Controller.Close()

    // Put host terminal into raw mode to pass keys (like Ctrl+C) through
    console, _ := termio.NewConsole(os.Stdin)
    console.EnableRawMode()
    defer console.Restore()

    // Forward IO
    go io.Copy(p.Controller, os.Stdin)
    go io.Copy(os.Stdout, p.Controller)

    cmd.Wait()
}
```

## 7. Streaming Process Telemetry

Monitor basic CPU and memory usage of a background process without external dependencies.

```go
import (
    "context"
    "fmt"
    "time"
    "github.com/aretw0/procio/proc"
)

func main() {
    ctx := context.Background()
    cmd := proc.NewCmd(ctx, "long-running-task")
    cmd.Start()

    // Monitor emits metrics every second
    ch, _ := proc.Monitor(ctx, cmd, time.Second)

    for m := range ch {
        fmt.Printf("CPU: %.1f%% | Mem: %d KB\n", m.CPUPercent, m.MemRSS/1024)
    }
}
```

## 8. Integration with `lifecycle`

**Problem**: You want to run OS processes managed by `lifecycle.ProcessWorker` and connect `procio`'s telemetry to your existing `lifecycle.Observer`.

**Solution (v0.4.0+)**: Since `lifecycle.Observer` is a superset of `procio.Observer` (adds `OnGoroutinePanicked`), any `lifecycle.Observer` implementation satisfies `procio.Observer` directly — no wrapper needed.

```go
import (
    "context"
    "log/slog"
    "github.com/aretw0/procio"
    "github.com/aretw0/procio/proc"
)

// MyObserver satisfies both procio.Observer and lifecycle.Observer.
// The only extra method lifecycle requires is OnGoroutinePanicked.
type MyObserver struct{}

func (o *MyObserver) OnProcessStarted(pid int)           { slog.Info("started", "pid", pid) }
func (o *MyObserver) OnProcessFailed(err error)          { slog.Error("failed", "err", err) }
func (o *MyObserver) OnIOError(op string, err error)     { slog.Error("io", "op", op, "err", err) }
func (o *MyObserver) OnScanError(err error)              { slog.Error("scan", "err", err) }
func (o *MyObserver) LogDebug(msg string, args ...any)   { slog.Debug(msg, args...) }
func (o *MyObserver) LogInfo(msg string, args ...any)    { slog.Info(msg, args...) }
func (o *MyObserver) LogWarn(msg string, args ...any)    { slog.Warn(msg, args...) }
func (o *MyObserver) LogError(msg string, args ...any)   { slog.Error(msg, args...) }
func (o *MyObserver) OnGoroutinePanicked(v any, stack []byte) { /* lifecycle-specific */ }

// Compile-time check — this line fails to build if procio.Observer changes.
var _ procio.Observer = (*MyObserver)(nil)

func main() {
    procio.SetObserver(&MyObserver{})

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Use proc.NewCmd (not exec.Command) to bind context cancellation.
    cmd := proc.NewCmd(ctx, "long-running-service", "--flag")
    if err := cmd.Start(); err != nil {
        slog.Error("could not start", "err", err)
        return
    }
    cmd.Wait()
}
```

> **Pattern**: `lifecycle.ProcessWorker` is the reference implementation of the Worker Contract.
> Use it directly or follow its pattern for custom workers.
> See [Integration Points](TECHNICAL.md#integration-points-v040) for a deeper explanation.
