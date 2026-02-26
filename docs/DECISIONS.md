# Architecture Decisions

## 1. Zero-Dependency Logging (Observer Pattern)

We chose not to import `log/slog` or any third-party logging library to keep `procio` lightweight and embeddable. Instead, we expose an `Observer` interface. This allows consumers to plug in their own telemetry stack without conflicting dependencies.

## 2. Windows Job Objects for Process Lifecycle

On Windows, we use **Job Objects** (`JOBOBJECT_EXTENDED_LIMIT_INFORMATION` with `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`) to ensure child processes are terminated when the parent process exits. this is the only robust way to match the behavior of Linux `Pdeathsig`.

- **Implementation**: We create a single global Job Object per `procio` instance (process) and assign all started child processes to it. The handle to this Job Object is held by the Go process itself; when the parent process exits or is killed, the Windows kernel automatically closes the handle, triggering the `KILL_ON_JOB_CLOSE` limit for all associated children. This is a "set and forget" strategy that requires no explicit cleanup from the user.

## 3. "Fake EOF" Detection on Windows

Windows Console Input (`CONIN$`) has a quirk where pressing `Ctrl+C` can sometimes be interpreted as an EOF (`0` bytes read) by `ReadFile` depending on the mode, instead of generating a signal immediately handled by the runtime.

- **Solution**: The `scan.Scanner` implements a heuristic. If it sees a 0-byte read (EOF), it checks if it should retry a few times (`threshold`) before accepting it as a true stream end. This filters out transient EOFs caused by signal handling interrupts.

## 4. `proc.NewCmd` as Ergonomic Entry Point (v0.2.0)

**Context:** The v0.1.x API required two separate steps to start a process with both context-linked cancellation and platform hygiene:

```go
// error-prone two-step pattern
cmd := exec.CommandContext(ctx, name, args...)
err := proc.Start(cmd)
```

The risk: if a caller mistakenly uses `exec.Command` (no context) instead of `exec.CommandContext`, the platform hygiene is applied but cancellation propagation is silently lost.

**Decision:** Introduce `proc.NewCmd(ctx, name, args...)` as the recommended constructor. It is a thin wrapper over `exec.CommandContext` that signals intent ("I have a context, use it") without hiding the two-phase model (construct then start).

**Rejected alternatives:**

| Alternative | Why rejected |
|---|---|
| `proc.Start(ctx, name, args...)` — combine start + launch | Prevents configuring `Stdout`, `Env`, `Dir` before launching. Breaks the builder pattern. |
| `proc.StartCmd(cmd) error` accepting context-less cmd | Doesn't solve the ergonomic problem; caller still needs to remember `CommandContext`. |
| Return a custom struct wrapping `*exec.Cmd` | Adds indirection with no benefit; `*exec.Cmd` is already the standard Go type. |

**Constraint preserved:** `proc.Start(cmd)` remains unchanged. `NewCmd` is purely additive.

## 5. PTY Terminology (`controller`/`worker`)

We chose the terms `controller` and `worker` instead of the traditional `master`/`slave` or `pty`/`tty` for pseudo-terminals. This aligns with modern, inclusive language guidelines while accurately describing the role of each end: the `controller` reads/writes the I/O stream, and the `worker` acts as the terminal device for the child process.

## 6. PTY Implementation (Windows ConPTY vs POSIX `openpty`)

On Windows, we use the ConPTY API (`CreatePseudoConsole`), which is the modern standard (Windows 10+) for pseudo-terminals. This replaces brittle legacy hacks using named pipes and hidden console windows. On POSIX, rather than linking `cgo` for libc's `openpty`, we directly open `/dev/ptmx` and use ioctls (`TIOCGPTN`/`TIOCPTYGNAME` for name discovery, `TIOCSPTLCK`/`TIOCPTYUNLK` for unlocking, and `TIOCPTYGRANT` on Darwin) to allocate the worker. This keeps the library `CGO_ENABLED=0` capable and pure Go. Because the ioctl surface differs between platforms, the implementation is split across `pty_linux.go`, `pty_darwin.go`, and `pty_bsd.go`.

## 7. Zero-Dependency Telemetry

For process metrics (CPU/Memory), we implemented direct OS scraping instead of leveraging comprehensive but heavy third-party libraries (like `shirou/gopsutil`).

- **Linux**: Direct parsing of `/proc/[pid]/stat` and `/proc/[pid]/statm`.
- **Windows**: Syscalls to `GetProcessTimes` and `GetProcessMemoryInfo` via `psapi.dll`.

This maintains `procio`'s core philosophy of having zero external dependencies (aside from `golang.org/x/sys`), keeping it lightweight and suitable for seamless embedding into CLI tools.

## 8. ConPTY Requires `STARTF_USESTDHANDLES` to Prevent stdout Bypass

When creating a child process with `CreateProcess` and a `ProcThreadAttributeList` containing a `PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE`, the `StartupInfoEx.Flags` field **must** include `STARTF_USESTDHANDLES` with zeroed std handle fields. Without it, Windows falls back to inheriting the parent's stdio handles — including any anonymous pipes set up by `go test` to capture test output. The child writes directly to those pipes, bypassing the ConPTY entirely, so `p.Controller` never receives any bytes. Setting `STARTF_USESTDHANDLES` with zero handles forces the child to use only the attached PseudoConsole, reproducing the same behavior in `go test`, CI runners, and interactive terminals.

## 9. `TIOCSPTLCK` Requires a `*int32`, Not a Go `int`

`unix.IoctlSetInt` passes `unsafe.Pointer(&value)` where `value` is a Go `int` (8 bytes on amd64). `TIOCSPTLCK` is defined as `_IOW('T', 0x31, int)` where `int` is the C int — 4 bytes. The kernel validates the argument size via `_IOC_SIZE` and returns `EFAULT` on mismatch. The fix is to call `unix.Syscall` directly with a `*int32`:

```go
var zero int32
unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.TIOCSPTLCK, uintptr(unsafe.Pointer(&zero)))
```

`unix.IoctlSetInt` is incorrect for any ioctl that expects a C `int` pointer on 64-bit platforms. Use `Syscall` with the correctly-sized type.

## 11. Integration Hierarchy in the Ecosystem (v0.4.0)

**Context:** The `procio` library is used by multiple projects in the same ecosystem (`lifecycle`, `loam`, `trellis`). Without a documented integration contract, each project risks using `procio` in inconsistent ways (e.g., calling `proc.Start` directly instead of going through `lifecycle.ProcessWorker`).

**Decision:** Establish `lifecycle` as the **sole intended direct consumer** of `procio` in the ecosystem. Downstream projects (`loam`, `trellis`) must compose `procio` primitives **exclusively via `lifecycle.ProcessWorker`**, never by importing `procio` packages directly in application-level code.

The integration hierarchy is:

```text
procio          (OS Mechanics layer)
    ↓ consumed directly by
lifecycle       (Policy Engine layer)
    ↓ consumed directly by
loam / trellis  (Application layer — use lifecycle.ProcessWorker)
```

**Rationale:**

1. **Single responsibility**: `procio` handles "how to start a process safely"; `lifecycle` handles "when and why to start it, and what to do when it fails". Mixing these concerns in `loam`/`trellis` creates tight coupling.
2. **No adapter package needed**: The bridge between `procio` and `lifecycle` is achieved via the `Observer` interface (caller owns the adapter, `procio` does not). This keeps `procio` dependency-free and avoids circular imports.
3. **`Observer.LogInfo` alignment**: First introduced in v0.4.0, but later reversed in **v0.4.1**, `procio.Observer` is strictly decoupled from opinionated logging levels. `lifecycle.Observer` simply operates as a superset (having `LogInfo`, `OnGoroutinePanicked`, etc.) and thus implicitly satisfies `procio.Observer` natively with no extra wrapper.

**Consequences:**

- `loam` and `trellis` should have `procio` only as an **indirect** dependency (via `lifecycle`). Any direct `procio` import in these projects is a design smell.
- The `ObserverBridge` pattern (documented in `lifecycle/docs/RECIPES.md` and `procio/docs/RECIPES.md`) is the canonical way to unify telemetry from both layers.
- Future `procio` releases must remain backwards-compatible with `lifecycle`'s usage surface (`proc.Start`, `proc.NewCmd`, `termio.InterruptibleReader`, `scan.Scanner`).

On Linux, `grantpt()` is effectively a no-op because the kernel `devpts` filesystem automatically sets correct ownership and permissions on the worker device when `/dev/ptmx` is opened. On Darwin (macOS), this step is **mandatory**: the worker side (`/dev/ttysN`) starts with restrictive root-only permissions and requires the `TIOCPTYGRANT` ioctl on the controller fd before it can be opened by the calling process. Omitting it causes `open worker: permission denied`. A subsequent `TIOCPTYUNLK` ioctl is also needed to unlock the worker. This distinction drove the split of the POSIX PTY implementation into platform-specific files (`pty_linux.go`, `pty_darwin.go`, `pty_bsd.go`).
