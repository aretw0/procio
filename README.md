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

### starting a process safely

```go
import "github.com/aretw0/procio/proc"

cmd := exec.Command("long-running-worker")
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

## Observability

`procio` is opinionated about specific mechanisms but unopinionated about logging/metrics.
You can inject your own observer:

```go
import "github.com/aretw0/procio"

procio.SetObserver(myObserver)
```
