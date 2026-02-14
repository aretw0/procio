# procio

`procio` is a lightweight, standalone Go module for robust process lifecycle management and interactive I/O.

It provides three core primitives:

- **proc**: Leak-free process management (ensures child processes die when parent dies).
- **termio**: Interactive terminal I/O (handling interrupts, raw mode, and virtual terminals).
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
scanner.Start(ctx) // Respects context cancellation and handles transient interrupts
```

## Observability

`procio` is opinionated about specific mechanisms but unopinionated about logging/metrics.
You can inject your own observer:

```go
import "github.com/aretw0/procio"

procio.SetObserver(myObserver)
```
