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

**Platform implementations:**

- **Linux**: Sets `syscall.SysProcAttr{Pdeathsig: SIGKILL}` before calling `cmd.Start()`.
- **Windows**: Global Job Object initialized once per process via `sync.Once`. Each started process is assigned to the Job via `AssignProcessToJobObject`. Fail-closed: if assignment fails, the child is killed immediately.
- **Fallback**: Wraps `cmd.Start()`. Returns error if `StrictMode = true`.

### `termio`

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
