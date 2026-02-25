# Technical Architecture: procio

## Module Structure

```text
procio/
├── procio.go       # Facade, Configuration, Observer Interface
├── proc/           # Process lifecycle (Pdeathsig, Job Objects, NewCmd)
├── termio/         # Terminal I/O (Raw mode, Interruptible Readers)
└── scan/           # High-level Scanner (Lines, Commands, Fake EOF)
```

## Primitives

### `proc`

| Function | Purpose |
|---|---|
| `NewCmd(ctx, name, args...)` | Recommended constructor. Wraps `exec.CommandContext`; returns `*exec.Cmd` pre-linked to `ctx`. |
| `Start(cmd)` | Applies platform hygiene and launches. Must be called after `NewCmd`. |
| `Monitor(ctx, cmd, interval)`| Streams real-time `Metrics` (CPU%, MemRSS) from the running process to a channel. |

### `pty`

| Struct / Function | Purpose |
|---|---|
| `PTY` | Holds the `Controller` and `Worker` file descriptors for a pseudo-terminal pair. |
| `StartPTY(cmd)` | Starts a command attached to a new PTY, returning the `PTY` instance for I/O interaction. |

**Platform implementations:**

- **Linux**: Sets `syscall.SysProcAttr{Pdeathsig: SIGKILL}` before calling `cmd.Start()`.
- **Windows**: Global Job Object initialized once per process via `sync.Once`. Each started process is assigned to the Job via `AssignProcessToJobObject`. Fail-closed: if assignment fails, the child is killed immediately.
- **Fallback**: Wraps `cmd.Start()`. Returns error if `StrictMode = true`.

### `termio`

- **Console**: Abstract wrapper over OS terminal modes. Allows calling `EnableRawMode()`, `Restore()`, and `Size()`.
- **IsInterrupted**: Centralized logic to detect cancellation across platforms (Ctrl+C, Context Done, EOF).
- **InterruptibleReader**: Wraps `io.Reader` to select on `ctx.Done()` before and after blocking calls.

### `scan`

- **Scanner**: A context-aware replacement for `bufio.Scanner`.
- **Fake EOF**: On Windows, Ctrl+C can trigger a transient EOF. The scanner implements a counter-based heuristic (`threshold`) to differentiate transient interrupts from true stream end.

## Observability

`procio` uses a "push" model via the `Observer` interface defined in the root package. It does not import any logging libraries. See [DECISIONS.md](./DECISIONS.md#1-zero-dependency-logging-observer-pattern).

## Context Flow

```text
application context (appCtx)
    └── subprocess context (subCtx = WithTimeout(appCtx, ...))
            └── proc.NewCmd(subCtx, "worker")
                    └── proc.Start(cmd)  ← applies OS hygiene
```

Cancelling `appCtx` propagates to `subCtx`, which terminates the process via `exec.CommandContext` internals. Platform hygiene (Job Objects / Pdeathsig) acts as a safety net for abnormal parent exits.

## PTY Pipeline Flow

When utilizing `pty.StartPTY`, the data flows through the pseudo-terminal layer before hitting the child process, forcing it to act interactively.

```mermaid
graph LR
    A[Host App] <-->|Read / Write| B(PTY Controller)
    B <-->|Kernel Forwarding| C(PTY Worker)
    C <-->|stdin/stdout/stderr| D[Child Process]
    
    style B fill:#e1f5fe,stroke:#01579b
    style C fill:#e1f5fe,stroke:#01579b
```

## Integration Points (v0.4.0)

`procio` exposes three stable boundaries for integration with higher-level frameworks (see also [ADR-11](./DECISIONS.md#11-integration-hierarchy-in-the-ecosystem-v040)).

### 1. Observer Bridge

The `Observer` interface (root package) is the **telemetry contract**. Consumers implement it in their own package; `procio` never imports any logger. Since v0.4.0, `lifecycle.Observer` is a drop-in implementation — it satisfies `procio.Observer` directly with no glue code:

```go
var _ procio.Observer = (*MyObserver)(nil) // compile-time check
procio.SetObserver(myObserver)
```

Full `ObserverBridge` example: see `examples/lifecycle_bridge/` and `lifecycle/docs/RECIPES.md`.

### 2. Context Contract

`proc.NewCmd(ctx, name, args...)` is the **process creation contract**. Context-linked cancellation and platform hygiene (Job Objects / Pdeathsig) are guaranteed when the caller derives the subprocess context from the application context:

```text
appCtx (lifecycle.SignalContext or signal.Context)
    └── subCtx = context.WithTimeout(appCtx, 30s)
            └── proc.NewCmd(subCtx, "worker")
                    └── proc.Start(cmd)  ← applies OS hygiene
```

### 3. Worker Contract

The canonical pattern for embedding `proc.Start` inside a `lifecycle.Worker`:

```go
func (w *MyWorker) Start(ctx context.Context) error {
    // Use proc.NewCmd, NOT exec.Command — binds context cancellation.
    w.cmd = proc.NewCmd(ctx, w.binary, w.args...)
    w.cmd.Stdout = w.stdout
    return proc.Start(w.cmd)
}
```

Reference implementation: `lifecycle.ProcessWorker` in `pkg/core/worker/process.go`.
