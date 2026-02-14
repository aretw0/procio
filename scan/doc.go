// Package scan provides a robust, context-aware command and line scanner.
//
// It is designed to replace bufio.Scanner for interactive CLI applications,
// offering features like:
//   - Context cancellation support (cooperates with interruptible readers).
//   - "Fake EOF" detection for Windows consoles (filtering transient interrupts).
//   - Configurable buffering and line handling callbacks.
package scan
