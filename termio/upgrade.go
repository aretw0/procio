package termio

import (
	"io"
	"os"

	"github.com/aretw0/procio"
	"golang.org/x/term"
)

// Upgrade checks if the provided reader is a file-based terminal.
// If it is, it upgrades it to a safe terminal reader (like CONIN$ on Windows) using Open().
// If it is not (e.g. pipe, file, buffer), it returns the original reader.
func Upgrade(r io.Reader) (io.Reader, error) {
	if f, ok := r.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		procio.GetObserver().LogDebug("upgrading terminal to raw/conin handle", "fd", f.Fd())
		newR, err := Open() // Open returns (io.ReadCloser, error)
		if err == nil {
			// metrics.GetProvider().IncTerminalUpgrade(true) - Metrics removed for now or should be on Observer?
			// The Observer interface doesn't have specific metric methods like IncTerminalUpgrade.
			// for now we just LogDebug or maybe add a generic metric method later?
			// The user requirement was "remove dependency".
			// I'll skip the metric call for now as the Observer interface is generic.
		} else {
			procio.GetObserver().OnProcessFailed(err) // Reusing OnProcessFailed or just log error?
			procio.GetObserver().LogError("failed to upgrade terminal", "error", err)
		}
		return newR, err
	}
	return r, nil
}
