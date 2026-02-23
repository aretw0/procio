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
    if err := proc.Start(cmd); err != nil {
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

    if err := proc.Start(cmd); err != nil {
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
    if err := proc.Start(cmd); err != nil {
        log.Fatal(err)
    }

    scanner := scan.NewScanner(stdout, scan.WithLineHandler(func(line string) {
        fmt.Println("[log]", line)
    }))
    scanner.Start(ctx) // blocks until ctx is done or pipe closes

    cmd.Wait()
}
```
