# Planning: procio

## Roadmap

### v0.1.0 (Initial Extracted Release)

- [x] Extract `proc` package behavior.
- [x] Extract `termio` package behavior.
- [x] Extract `scan` logic.
- [x] Decouple dependencies (remove `pkg/core/log`).
- [x] Stabilize API for initial release.

### v0.2.0 (Advanced Features)

- [ ] **PTY Support:** Pseudo-terminal primitives for running interactive applications (vim, htop) within processes.
- [ ] **Enhanced Windows Console API:** Improved console mode handling and event processing.
- [ ] **Streaming Telemetry:** Real-time process metrics (CPU, memory) via callbacks or channels, without external dependencies.

### v0.3.0 (Integration & Field Validation)

- [ ] **Integration with `lifecycle` v2:** Establish `procio` as the primary process and I/O engine.
- [ ] **Production Feedback Loop:** Gather real-world usage data and address edge cases.
- [ ] **API Stabilization:** Final breaking changes before v1.0 freeze.

### v1.0.0 (Stable Release)

- [ ] **API Freeze:** Semantic versioning guarantees (no breaking changes in 1.x).
- [ ] **Complete Documentation:** All use cases documented, including platform-specific edge cases.
- [ ] **Test Coverage:** >85% coverage across all critical packages.
- [ ] **Production-Ready:** Proven stability in at least one external production environment.
