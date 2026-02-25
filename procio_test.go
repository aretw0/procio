package procio

import (
	"errors"
	"testing"
)

type spyObserver struct {
	processStarted []int
	processFailed  []error
	ioErrors       []string
	scanErrors     []error
	debugMsgs      []string
}

func (o *spyObserver) OnProcessStarted(pid int)         { o.processStarted = append(o.processStarted, pid) }
func (o *spyObserver) OnProcessFailed(err error)        { o.processFailed = append(o.processFailed, err) }
func (o *spyObserver) OnIOError(op string, err error)   { o.ioErrors = append(o.ioErrors, op) }
func (o *spyObserver) OnScanError(err error)            { o.scanErrors = append(o.scanErrors, err) }
func (o *spyObserver) LogDebug(msg string, args ...any) { o.debugMsgs = append(o.debugMsgs, msg) }
func (o *spyObserver) LogInfo(msg string, args ...any)  {}
func (o *spyObserver) LogWarn(msg string, args ...any)  {}
func (o *spyObserver) LogError(msg string, args ...any) {}

func TestSetObserver_CustomObserver(t *testing.T) {
	spy := &spyObserver{}
	SetObserver(spy)
	defer SetObserver(nil)

	obs := GetObserver()
	obs.OnProcessStarted(42)
	obs.OnProcessFailed(errors.New("boom"))
	obs.OnIOError("test.op", errors.New("io fail"))
	obs.OnScanError(errors.New("scan fail"))
	obs.LogDebug("debug msg")

	if len(spy.processStarted) != 1 || spy.processStarted[0] != 42 {
		t.Errorf("OnProcessStarted: expected [42], got %v", spy.processStarted)
	}
	if len(spy.processFailed) != 1 {
		t.Errorf("OnProcessFailed: expected 1 call, got %d", len(spy.processFailed))
	}
	if len(spy.ioErrors) != 1 || spy.ioErrors[0] != "test.op" {
		t.Errorf("OnIOError: expected [test.op], got %v", spy.ioErrors)
	}
	if len(spy.scanErrors) != 1 {
		t.Errorf("OnScanError: expected 1 call, got %d", len(spy.scanErrors))
	}
	if len(spy.debugMsgs) != 1 || spy.debugMsgs[0] != "debug msg" {
		t.Errorf("LogDebug: expected ['debug msg'], got %v", spy.debugMsgs)
	}
}

func TestSetObserver_NilResetsToNoop(t *testing.T) {
	spy := &spyObserver{}
	SetObserver(spy)
	SetObserver(nil) // should reset to noop

	obs := GetObserver()

	// These should not panic (noop handles them)
	obs.OnProcessStarted(1)
	obs.OnProcessFailed(errors.New("x"))
	obs.OnIOError("op", errors.New("x"))
	obs.OnScanError(errors.New("x"))
	obs.LogDebug("x")
	obs.LogWarn("x")
	obs.LogError("x")

	// spy should have received nothing after reset
	if len(spy.processStarted) != 0 {
		t.Error("Expected noop observer, but spy was called")
	}
}

func TestNoopObserver_AllMethods(t *testing.T) {
	noop := noopObserver{}

	// Ensure no panics
	noop.OnProcessStarted(0)
	noop.OnProcessFailed(nil)
	noop.OnIOError("", nil)
	noop.OnScanError(nil)
	noop.LogDebug("")
	noop.LogWarn("")
	noop.LogError("")
}
