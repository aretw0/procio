package termio_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/aretw0/procio"
	"github.com/aretw0/procio/termio"
)

func TestUpgrade_NonFile(t *testing.T) {
	buf := bytes.NewBufferString("hello")
	upgraded, err := termio.Upgrade(buf)
	if err != nil {
		t.Fatalf("Upgrade failed: %v", err)
	}

	out := make([]byte, 5)
	n, err := upgraded.Read(out)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != 5 || string(out) != "hello" {
		t.Errorf("Unexpected read content: %s", out)
	}
}

func TestUpgrade_FileNonTerminal(t *testing.T) {
	f, err := os.CreateTemp("", "procio_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	upgraded, err := termio.Upgrade(f)
	if err != nil {
		t.Fatalf("Upgrade failed: %v", err)
	}

	if upgraded != f {
		t.Errorf("Expected original file object for non-terminal file")
	}
}

// upgradeObserver captures Observer calls to verify correct hook routing.
type upgradeObserver struct {
	ioOps         []string
	processFailed []error
}

func (o *upgradeObserver) OnProcessStarted(int)             {}
func (o *upgradeObserver) OnProcessFailed(err error)        { o.processFailed = append(o.processFailed, err) }
func (o *upgradeObserver) OnIOError(op string, err error)   { o.ioOps = append(o.ioOps, op) }
func (o *upgradeObserver) OnScanError(error)                {}
func (o *upgradeObserver) LogDebug(msg string, args ...any) {}
func (o *upgradeObserver) LogInfo(msg string, args ...any)  {}
func (o *upgradeObserver) LogWarn(msg string, args ...any)  {}
func (o *upgradeObserver) LogError(msg string, args ...any) {}

func TestUpgrade_DoesNotCallOnProcessFailed(t *testing.T) {
	obs := &upgradeObserver{}
	procio.SetObserver(obs)
	defer procio.SetObserver(nil)

	buf := bytes.NewBufferString("test")
	_, err := termio.Upgrade(buf)
	if err != nil {
		t.Fatalf("Upgrade failed: %v", err)
	}

	if len(obs.processFailed) > 0 {
		t.Errorf("OnProcessFailed should not be called by Upgrade, got %d calls", len(obs.processFailed))
	}
	if len(obs.ioOps) > 0 {
		t.Errorf("Expected no OnIOError for non-file reader, got %v", obs.ioOps)
	}
}
