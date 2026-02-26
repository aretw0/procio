package procio

// IOObserver defines hooks for low-level process I/O events.
type IOObserver interface {
	OnIOError(op string, err error)
	OnScanError(err error)
}

// Observer allows external packages to plug in observability (logs, metrics)
// without coupling this module to specific implementations.
type Observer interface {
	IOObserver
	OnProcessStarted(pid int)
	OnProcessFailed(err error)
	LogDebug(msg string, args ...any)
	LogWarn(msg string, args ...any)
	LogError(msg string, args ...any)
}

var globalObserver Observer = noopObserver{}

// SetObserver configures the global observer for process events.
func SetObserver(o Observer) {
	if o != nil {
		globalObserver = o
	} else {
		globalObserver = noopObserver{}
	}
}

// GetObserver returns the current global observer.
// Useful for sub-packages to access the shared observer.
func GetObserver() Observer {
	return globalObserver
}

type noopObserver struct{}

func (noopObserver) OnProcessStarted(int)    {}
func (noopObserver) OnProcessFailed(error)   {}
func (noopObserver) OnIOError(string, error) {}
func (noopObserver) OnScanError(error)       {}
func (noopObserver) LogDebug(string, ...any) {}
func (noopObserver) LogWarn(string, ...any)  {}
func (noopObserver) LogError(string, ...any) {}
