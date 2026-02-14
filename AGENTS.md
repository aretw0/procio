# AGENTS.md

## Project Metadata

- **Name**: procio
- **Purpose**: Robust Process IO and SignalingPrimitives for Go.
- **Philosophy**: Platform-agnostic safety. Leak-free process execution.

## Key Primitives

- **proc**: Process management (`Start`). Guarantees child death on parent exit.
- **termio**: Terminal I/O. `InterruptibleReader` for non-blocking reads on Windows/Linux.
- **scan**: Context-aware scanner. Replaces `bufio.Scanner` for interactive CLIs.

## Development Rules

- **Testing**: Must pass on Windows and Linux.
- **Design**: functional options pattern for configuration.
