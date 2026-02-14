# Architecture Decisions

## 1. Zero-Dependency Logging (Observer Pattern)

We chose not to import `log/slog` or any third-party logging library to keep `procio` lightweight and embeddable. Instead, we expose an `Observer` interface. This allows consumers to plug in their own telemetry stack without conflicting dependencies.

## 2. Windows Job Objects for Process Lifecycle

On Windows, we use **Job Objects** (`JOBOBJECT_EXTENDED_LIMIT_INFORMATION` with `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`) to ensure child processes are terminated when the parent process exits. this is the only robust way to match the behavior of Linux `Pdeathsig`.

- **Implementation**: We create a single global Job Object per `procio` instance (process) and assign all started child processes to it. The handle to this Job Object is held by the Go process itself; when the parent process exits or is killed, the Windows kernel automatically closes the handle, triggering the `KILL_ON_JOB_CLOSE` limit for all associated children. This is a "set and forget" strategy that requires no explicit cleanup from the user.

## 3. "Fake EOF" Detection on Windows

Windows Console Input (`CONIN$`) has a quirk where pressing `Ctrl+C` can sometimes be interpreted as an EOF (`0` bytes read) by `ReadFile` depending on the mode, instead of generating a signal immediately handled by the runtime.

- **Solution**: The `scan.Scanner` implements a heuristic. If it sees a 0-byte read (EOF), it checks if it should retry a few times (`threshold`) before accepting it as a true stream end. This filters out transient EOFs caused by signal handling interrupts.
