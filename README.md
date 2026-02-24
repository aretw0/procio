# procio

[![Go Report Card](https://goreportcard.com/badge/github.com/aretw0/procio)](https://goreportcard.com/report/github.com/aretw0/procio)
[![Go Reference](https://pkg.go.dev/badge/github.com/aretw0/procio.svg)](https://pkg.go.dev/github.com/aretw0/procio)
[![License](https://img.shields.io/github/license/aretw0/procio.svg?color=red)](./LICENSE)
[![Release](https://img.shields.io/github/release/aretw0/procio.svg?branch=main)](https://github.com/aretw0/procio/releases)

`procio` is a lightweight, standalone set of composable primitives for safe process lifecycle and interactive I/O in Go.

It provides three core primitives:

- **proc**: Leak-free process management (ensures child processes die when parent dies).
- **termio**: Interruptible terminal I/O (handling interrupts and safe terminal handles).
- **scan**: Robust input scanning with protection against "Fake EOF" signals on Windows.

## Installation

```bash
go get github.com/aretw0/procio
```

## Usage

### Starting a Process Safely

```go
import "github.com/aretw0/procio/proc"

cmd := proc.NewCmd(ctx, "long-running-worker")
// Uses Pdeathsig (Linux) or Job Objects (Windows) to enforce cleanup
err := proc.Start(cmd)
```

### Reading Input Robustly

```go
import "github.com/aretw0/procio/scan"

scanner := scan.NewScanner(os.Stdin)
scanner.Start(ctx) // Handles transient interrupts
```

### Enabling TTY Cancellation

For interactive CLIs that need Ctrl+C cancellation support:

```go
import (
    "context"
    "github.com/aretw0/procio/scan"
)

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

scanner := scan.NewScanner(os.Stdin,
    scan.WithInterruptible(), // Enables context cancellation via Ctrl+C
    scan.WithLineHandler(func(line string) {
        fmt.Println("Got:", line)
    }),
)

scanner.Start(ctx) // Returns when context is cancelled or EOF
```

### Chained Cancels

`proc.NewCmd` integrates naturally with derived contexts, so cancellation hierarchies work as expected:

```go
// appCtx controls the whole application lifetime.
appCtx, appCancel := context.WithCancel(context.Background())
defer appCancel()

// subCtx adds a deadline for a specific subprocess.
subCtx, subCancel := context.WithTimeout(appCtx, 10*time.Second)
defer subCancel()

cmd := proc.NewCmd(subCtx, "worker")
if err := proc.Start(cmd); err != nil {
    log.Fatal(err)
}
cmd.Wait()
// worker is terminated when subCtx expires OR when appCtx is cancelled —
// whichever comes first. Platform hygiene (Job Objects / Pdeathsig) is
// still applied regardless of which signal arrives first.
```

### Advanced Features

`procio` provides primitives for advanced process control:

#### Pseudo-Terminals (PTY)

Wrap interactive applications:

```go
import "github.com/aretw0/procio/pty"

cmd := proc.NewCmd(ctx, "vim")
p, err := pty.StartPTY(cmd)
// Forward p.Controller to/from host Stdin/Stdout
```

#### Streaming Telemetry

Monitor processes in real-time (Linux & Windows):

```go
ch, err := proc.Monitor(ctx, cmd, time.Second)
for m := range ch {
    fmt.Printf("CPU: %.1f%% Mem: %d KB\n", m.CPUPercent, m.MemRSS/1024)
}
```

## Observability

`procio` is opinionated about specific mechanisms but unopinionated about logging/metrics.
You can inject your own observer:

```go
import "github.com/aretw0/procio"

procio.SetObserver(myObserver)
```

See [docs/RECIPES.md](./docs/RECIPES.md) for a complete `log/slog` adapter example.

## License

This project is licensed under the terms of the [AGPL-3.0](./LICENSE).
