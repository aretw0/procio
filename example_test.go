package procio_test

import (
	"fmt"

	"github.com/aretw0/procio"
)

// MyObserver implements procio.Observer
type MyObserver struct{}

func (o *MyObserver) OnProcessStarted(pid int)         { fmt.Printf("Started PID: %d\n", pid) }
func (o *MyObserver) OnProcessFailed(err error)        { fmt.Printf("Failed: %v\n", err) }
func (o *MyObserver) OnIOError(op string, err error)   { fmt.Printf("IO Error (%s): %v\n", op, err) }
func (o *MyObserver) OnScanError(err error)            { fmt.Printf("Scan Error: %v\n", err) }
func (o *MyObserver) LogDebug(msg string, args ...any) {}
func (o *MyObserver) LogWarn(msg string, args ...any)  { fmt.Printf("%s\n", msg) }
func (o *MyObserver) LogError(msg string, args ...any) {}

func ExampleSetObserver() {
	// Set a custom observer to receive lifecycle and IO hooks
	procio.SetObserver(&MyObserver{})

	// Emitting a log to demonstrate (internally used by procio packages)
	obs := procio.GetObserver()
	obs.LogWarn("Observer configured successfully")

	// Reset to default (noop) when done
	procio.SetObserver(nil)

	// Output:
	// Observer configured successfully
}
