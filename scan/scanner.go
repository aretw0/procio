package scan

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/aretw0/procio"
	"github.com/aretw0/procio/termio"
)

// Scanner reads lines from an io.Reader (like Stdin) with robust error handling
// and "Fake EOF" protection for Windows environments.
type Scanner struct {
	r          io.Reader
	bufSize    int
	backoff    time.Duration
	eofCount   int
	threshold  int
	unsafe     bool
	onLine     func(line string)
	onClear    func()
	onError    func(err error)
}

// Option configures the Scanner.
type Option func(*Scanner)

// WithBufferSize sets the size of the internal read buffer.
func WithBufferSize(size int) Option {
	return func(s *Scanner) {
		if size <= 0 {
			size = 1024
		}
		s.bufSize = size
	}
}

// WithBackoff configures the duration to wait before retrying interruptions or errors.
func WithBackoff(d time.Duration) Option {
	return func(s *Scanner) {
		s.backoff = d
	}
}

// WithForceExitThreshold sets the number of consecutive EOFs required to stop scanning.
func WithThreshold(n int) Option {
	return func(s *Scanner) {
		s.threshold = n
	}
}

// WithUnsafeMode disables the EOF threshold check.
func WithUnsafeMode(unsafe bool) Option {
	return func(s *Scanner) {
		s.unsafe = unsafe
	}
}

// WithLineHandler sets the callback for complete lines.
func WithLineHandler(fn func(line string)) Option {
	return func(s *Scanner) {
		s.onLine = fn
	}
}

// WithClearHandler sets the callback for when the line buffer is cleared (e.g. on interrupt).
func WithClearHandler(fn func()) Option {
	return func(s *Scanner) {
		s.onClear = fn
	}
}

// WithErrorHandler sets the callback for non-fatal errors.
func WithErrorHandler(fn func(err error)) Option {
	return func(s *Scanner) {
		s.onError = fn
	}
}

// NewScanner creates a new robust scanner.
func NewScanner(r io.Reader, opts ...Option) *Scanner {
	s := &Scanner{
		r:         r,
		bufSize:   1024,
		backoff:   100 * time.Millisecond,
		threshold: 3,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Start runs the read loop until context cancellation or persistent failure.
func (s *Scanner) Start(ctx context.Context) {
	buffer := make([]byte, s.bufSize)
	var lineBuilder strings.Builder

	for {
		if ctx.Err() != nil {
			return
		}

		n, err := s.r.Read(buffer)

		// Always process read bytes first, even if error occurred
		if n > 0 {
			s.eofCount = 0 // Reset EOF count if we got data
			s.processChunk(buffer[:n], &lineBuilder)
		}

		// Handle Context Cancellation (Priority)
		if ctx.Err() != nil {
			return
		}

		if err != nil {
			if shouldStop := s.handleReadError(err, &lineBuilder); shouldStop {
				// Flush remaining buffer if any
				if lineBuilder.Len() > 0 {
					line := strings.TrimSpace(lineBuilder.String())
					if line != "" && s.onLine != nil {
						s.onLine(line)
					}
					lineBuilder.Reset()
				}
				return
			}
			continue
		}
	}
}

func (s *Scanner) handleReadError(err error, lineBuilder *strings.Builder) bool {
	// Special handling for EOF: Do NOT reset the buffer, as we might want to flush it
	// if we decide to stop.
	if errors.Is(err, io.EOF) {
		return s.handleEOF()
	}

	if termio.IsInterrupted(err) {
		// On Windows, Ctrl+C can cause a transient EOF or another interrupted error.
		// We treat these as Interruptions.
		lineBuilder.Reset()
		if s.onClear != nil {
			s.onClear()
		}
		return s.handleEOF()
	}

	// Other errors: Log and retry with backoff
	if s.onError != nil {
		s.onError(err)
	} else {
		procio.GetObserver().OnScanError(err)
		procio.GetObserver().LogWarn("scan read error", "error", err)
	}
	time.Sleep(s.backoff)
	return false
}

func (s *Scanner) handleEOF() bool {
	s.eofCount++

	// threshold exceeded: stop the source
	if s.eofCount > s.threshold && !s.unsafe {
		return true
	}

	time.Sleep(s.backoff)
	return false
}

func (s *Scanner) processChunk(chunk []byte, lineBuilder *strings.Builder) {
	for _, b := range chunk {
		if b == '\r' {
			continue // Ignore Carriage Return (CRLF handling)
		}
		if b == '\n' {
			// Line complete
			cmd := strings.TrimSpace(lineBuilder.String())
			lineBuilder.Reset()

			if s.onLine != nil {
				s.onLine(cmd)
			}
		} else {
			lineBuilder.WriteByte(b)
		}
	}
}
