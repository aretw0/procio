# Technical Architecture: procio

## Module Structure

```text
procio/
├── procio.go       # Facade, Configuration, Observer Interface
├── proc/           # Process attributes (Pdeathsig, Job Objects)
├── termio/         # Terminal I/O (Raw mode, Interruptible Readers)
└── scan/           # High-level Scanner (Lines, Commands, Fake EOF)
```

## Primitives

### `proc`

- **Linux**: Uses `syscall.SysProcAttr{Pdeathsig: SIGKILL}`.
- **Windows**: Uses Job Objects (`JOBOBJECT_EXTENDED_LIMIT_INFORMATION` with `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`). Global Job Object is initialized once per process.
- **Fallback**: Wraps `cmd.Start()`.

### `termio`

- **IsInterrupted**: Centralized logic to detect cancellation across platforms (Ctrl+C, Context Done, EOF).
- **InterruptibleReader**: Wraps `io.Reader` to select on `ctx.Done()` before and after blocking calls.

### `scan`

- **Scanner**: A context-aware replacement for `bufio.Scanner`.
- **Fake EOF**: On Windows, Ctrl+C can trigger a transient EOF. The scanner implements a counter-based heuristic (`threshold`) to differentiate transient interrupts from true stream end.

## Observability

`procio` uses a "push" model via the `Observer` interface defined in the root package. It does not import any logging libraries.
