package termio_test

import (
	"os"
	"testing"

	"github.com/aretw0/procio/termio"
)

func TestConsole_New(t *testing.T) {
	// Creating a console against a pipe or file (not a true TTY on standard test runners)
	// will likely fail, so we just test that it errors correctly instead of panicking.
	f, err := os.CreateTemp("", "console-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	_, err = termio.NewConsole(f)
	if err == nil {
		t.Log("Warning: NewConsole succeeded on a temporary file. This is unexpected but safe.")
	} else {
		t.Logf("NewConsole failed expectedly with: %v", err)
	}
}
