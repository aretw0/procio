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
    // If main() panics or exits, 'sleep 100' will be killed by the OS (Job Object/Pdeathsig)
    cmd := proc.Start(ctx, "sleep", "100")
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
func (s SlogAdapter) LogDebug(msg string, args ...any) { slog.Debug(msg, args...) }
func (s SlogAdapter) LogWarn(msg string, args ...any) { slog.Warn(msg, args...) }
func (s SlogAdapter) LogError(msg string, args ...any) { slog.Error(msg, args...) }

func init() {
    procio.SetObserver(SlogAdapter{})
}
```
