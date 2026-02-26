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

### v0.2.0 (Ergonomic Context API)

**Focus**: Close the ergonomic gap identified during the `lifecycle v1.7.1` ecosystem audit: callers must currently combine `exec.CommandContext(ctx, ...)` with `proc.Start(cmd)` in two separate steps. This is an easily-forgettable pattern that can lead to missing process hygiene attributes.

- [x] **`proc.NewCmd(ctx, name, args...)`**: Convenience constructor that returns `*exec.Cmd` pre-configured with:
  - Context-linked cancellation (`exec.CommandContext` semantics).
  - Platform hygiene attributes (`PDeathSig`/`Job Objects`) applied automatically.
- [x] **Examples**: Update all `examples/` to prefer `proc.NewCmd` over the two-step pattern.
- [x] **Documentation**: Add a "Chained Cancels" section to `README.md` showing how `proc.NewCmd(ctx, ...)` integrates cleanly with derived contexts.

### v0.3.0 (Advanced Features)

- [x] **PTY Support:** Pseudo-terminal primitives for running interactive applications (vim, htop) within processes.
- [x] **Enhanced Windows Console API:** Improved console mode handling and event processing.
- [x] **Streaming Telemetry:** Real-time process metrics (CPU, memory) via callbacks or channels, without external dependencies.

### v0.4.0 (Integration)

**Focus**: Make `procio` ready for idiomatic consumption by the ecosystem. This release closes the DX gap between `procio` and its primary consumer (`lifecycle`), establishing clear integration contracts so downstream projects (`loam`, `trellis`) know exactly how to compose these primitives.

- [x] **`Observer.LogInfo`**: Add `LogInfo(msg string, args ...any)` to the `Observer` interface. Aligns `procio.Observer` with `lifecycle.Observer` (which already exposes `LogInfo`). A `lifecycle.Observer` implementation now satisfies `procio.Observer` directly without requiring an adapter wrapper. **Breaking change from v0.1.x; `noopObserver` updated accordingly.**
- [x] **Integration Recipe**: Add `Recipe 8: Integration with lifecycle` to `RECIPES.md`, demonstrating `proc.NewCmd` inside a `lifecycle.Worker`, `ObserverBridge` wiring, and context chaining as cleanup guarantee.
- [x] **Ecosystem ADR**: Add ADR-11 to `DECISIONS.md` documenting the integration hierarchy: `lifecycle` is the sole intended direct consumer of `procio` in the ecosystem; `loam` and `trellis` should compose `procio` primitives via `lifecycle.ProcessWorker`.
- [x] **Integration Points** section in `TECHNICAL.md`: Formally describe the three integration boundaries (Observer Bridge, Context Contract, Worker Contract).
- [x] **`examples/lifecycle_bridge/`**: Compilable example acting as a compile-time contract test for the `ObserverBridge` pattern.

### v0.4.1 (API Safety & Decoupling)

- [x] **API Safety:** Enveloped `*exec.Cmd` returned by `proc.NewCmd` into a proprietary `*proc.Cmd`, overriding `Start`, `Run`, `Output` and `CombinedOutput` to implicitly apply platform hygiene (preventing accidental raw calls).
- [x] **Interface Decoupling:** Removed `LogInfo` from the `Observer` interface to keep the telemetry contract strictly bound to process/io events and decoupled from opinionated logging levels.

### v0.4.2 (Patch): IOObserver Composition

**Focus**: Refactor `Observer` to support granular composition. This allows consumers like `lifecycle` to implement clean "Feature Discovery" for I/O hooks without forcing a fat interface implementation.

- [x] **`IOObserver` Interface**: Extract `OnIOError` and `OnScanError` into a dedicated `IOObserver` interface.
- [x] **Composition**: Update `Observer` to embed `IOObserver`. This keeps `procio.SetObserver(Observer)` backward compatible while allowing `lifecycle` to type-assert against `IOObserver` specifically.

### v0.5.0 (Production Feedback Loop)

- [ ] **Production Feedback Loop:** Gather real-world usage data and address edge cases.

### v0.6.0 (API Stabilization)

- [ ] **API Stabilization:** Final breaking changes before v1.0 freeze.

### v1.0.0 (Stable Release)

- [ ] **API Freeze:** Semantic versioning guarantees (no breaking changes in 1.x).
- [ ] **Complete Documentation:** All use cases documented, including platform-specific edge cases.
- [ ] **Test Coverage:** >85% coverage across all critical packages.
- [ ] **Production-Ready:** Proven stability in at least one external production environment.
