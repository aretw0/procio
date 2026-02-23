# Planning: procio

## Roadmap

### v0.1.0 (Initial Extracted Release)

- [x] Extract `proc` package behavior.
- [x] Extract `termio` package behavior.
- [x] Extract `scan` logic.
- [x] Decouple dependencies (remove `pkg/core/log`).
- [x] Stabilize API for initial release.

### v0.1.1 (Composable Primitives & Consolidation)

- [x] **Observer Refactoring:** Generalized interface for standalone I/O events (`OnIOError`, `OnScanError`).
- [x] **Composability:** Added `scan.WithInterruptible()` for automatic TTY/Context integration.
- [x] **Messaging Alignment:** Updated docs to prioritize "composable primitives" positioning.
- [x] **Practical Examples:** Added `composition`, `interruptible`, and `observer` examples.
- [x] **Test Coverage:** Added end-to-end integration tests.

### v0.1.2 (Ergonomic Context API)

**Focus**: Close the ergonomic gap identified during the `lifecycle v1.7.1` ecosystem audit: callers must currently combine `exec.CommandContext(ctx, ...)` with `proc.Start(cmd)` in two separate steps. This is an easily-forgettable pattern that can lead to missing process hygiene attributes.

- [ ] **`proc.NewCmd(ctx, name, args...)`**: Convenience constructor that returns `*exec.Cmd` pre-configured with:
  - Context-linked cancellation (`exec.CommandContext` semantics).
  - Platform hygiene attributes (`PDeathSig`/`Job Objects`) applied automatically.
- [ ] **Examples**: Update all `examples/` to prefer `proc.NewCmd` over the two-step pattern.
- [ ] **Documentation**: Add a "Chained Cancels" section to `README.md` showing how `proc.NewCmd(ctx, ...)` integrates cleanly with derived contexts.

### v0.2.0 (Advanced Features)

- [ ] **PTY Support:** Pseudo-terminal primitives for running interactive applications (vim, htop) within processes.
- [ ] **Enhanced Windows Console API:** Improved console mode handling and event processing.
- [ ] **Streaming Telemetry:** Real-time process metrics (CPU, memory) via callbacks or channels, without external dependencies.

### v0.3.0 (Integration & Field Validation)

- [ ] **Integration with `lifecycle` v1.5+:** Provide an optional adapter so `lifecycle` can compose `procio` primitives.
- [ ] **Production Feedback Loop:** Gather real-world usage data and address edge cases.
- [ ] **API Stabilization:** Final breaking changes before v1.0 freeze.

### v1.0.0 (Stable Release)

- [ ] **API Freeze:** Semantic versioning guarantees (no breaking changes in 1.x).
- [ ] **Complete Documentation:** All use cases documented, including platform-specific edge cases.
- [ ] **Test Coverage:** >85% coverage across all critical packages.
- [ ] **Production-Ready:** Proven stability in at least one external production environment.
