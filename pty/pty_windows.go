//go:build windows

package pty

import (
	"fmt"
	"os"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

func startPTY(cmd *exec.Cmd) (*PTY, error) {
	// Create two anonymous pipes bridging the application and the ConPTY:
	//   inputRead / inputWrite  — application writes → ConPTY → process stdin
	//   outputRead / outputWrite — process stdout/stderr → ConPTY → application reads
	var (
		inputRead, inputWrite   windows.Handle
		outputRead, outputWrite windows.Handle
	)

	sa := &windows.SecurityAttributes{InheritHandle: 0}
	if err := windows.CreatePipe(&inputRead, &inputWrite, sa, 0); err != nil {
		return nil, fmt.Errorf("pty: create input pipe: %w", err)
	}
	if err := windows.CreatePipe(&outputRead, &outputWrite, sa, 0); err != nil {
		_ = windows.CloseHandle(inputRead)
		_ = windows.CloseHandle(inputWrite)
		return nil, fmt.Errorf("pty: create output pipe: %w", err)
	}

	// Default console size; can be changed later via ResizePseudoConsole.
	size := windows.Coord{X: 80, Y: 24}
	var hpc windows.Handle
	if err := windows.CreatePseudoConsole(size, inputRead, outputWrite, 0, &hpc); err != nil {
		_ = windows.CloseHandle(inputRead)
		_ = windows.CloseHandle(inputWrite)
		_ = windows.CloseHandle(outputRead)
		_ = windows.CloseHandle(outputWrite)
		return nil, fmt.Errorf("pty: CreatePseudoConsole: %w", err)
	}
	// ConPTY owns these ends; close duplicates in the parent.
	_ = windows.CloseHandle(inputRead)
	_ = windows.CloseHandle(outputWrite)

	// Attach ConPTY to a new process via PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE.
	attrList, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		windows.ClosePseudoConsole(hpc)
		_ = windows.CloseHandle(outputRead)
		_ = windows.CloseHandle(inputWrite)
		return nil, fmt.Errorf("pty: NewProcThreadAttributeList: %w", err)
	}
	defer attrList.Delete()

	// O uso de unsafe.Pointer(hpc) é seguro aqui porque hpc é um handle opaco do Windows (não um ponteiro Go).
	// O linter pode alertar sobre "possible misuse of unsafe.Pointer", mas este padrão é amplamente aceito para interoperabilidade com APIs nativas.
	if err := attrList.Update(
		windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE,
		unsafe.Pointer(hpc),
		unsafe.Sizeof(hpc),
	); err != nil {
		windows.ClosePseudoConsole(hpc)
		_ = windows.CloseHandle(outputRead)
		_ = windows.CloseHandle(inputWrite)
		return nil, fmt.Errorf("pty: UpdateProcThreadAttribute: %w", err)
	}

	var siEx windows.StartupInfoEx
	siEx.Cb = uint32(unsafe.Sizeof(siEx))
	siEx.Flags = windows.STARTF_USESTDHANDLES
	siEx.ProcThreadAttributeList = attrList.List()

	// Resolve the executable path the same way exec.Cmd would.
	argv0, err := exec.LookPath(cmd.Path)
	if err != nil {
		argv0 = cmd.Path
	}
	argv0p, err := windows.UTF16PtrFromString(argv0)
	if err != nil {
		return nil, fmt.Errorf("pty: argv0: %w", err)
	}

	cmdLineStr := windows.ComposeCommandLine(append([]string{cmd.Path}, cmd.Args[1:]...))
	cmdLinep, err := windows.UTF16PtrFromString(cmdLineStr)
	if err != nil {
		return nil, fmt.Errorf("pty: commandLine: %w", err)
	}

	var pi windows.ProcessInformation
	// EXTENDED_STARTUPINFO_PRESENT tells CreateProcess to read
	// StartupInfoEx.ProcThreadAttributeList.
	const creationFlags = windows.CREATE_UNICODE_ENVIRONMENT | windows.EXTENDED_STARTUPINFO_PRESENT

	if err := windows.CreateProcess(
		argv0p,
		cmdLinep,
		nil, // process security attributes
		nil, // thread security attributes
		false,
		creationFlags,
		nil, // inherit environment
		nil, // inherit working directory
		// CreateProcess accepts *StartupInfo; StartupInfoEx embeds it as the first field.
		(*windows.StartupInfo)(unsafe.Pointer(&siEx)),
		&pi,
	); err != nil {
		windows.ClosePseudoConsole(hpc)
		_ = windows.CloseHandle(outputRead)
		_ = windows.CloseHandle(inputWrite)
		return nil, fmt.Errorf("pty: CreateProcess: %w", err)
	}

	// Close the thread handle immediately; we only need the process handle.
	_ = windows.CloseHandle(pi.Thread)

	// Wrap the process in an os.Process so cmd.Wait() can track it.
	// os.FindProcess on Windows attaches to the kernel process object.
	proc, err := os.FindProcess(int(pi.ProcessId))
	if err != nil {
		_ = windows.CloseHandle(pi.Process)
		windows.ClosePseudoConsole(hpc)
		_ = windows.CloseHandle(outputRead)
		_ = windows.CloseHandle(inputWrite)
		return nil, fmt.Errorf("pty: FindProcess: %w", err)
	}

	// Hang the process on cmd so callers can use cmd.Wait().
	cmd.Process = proc

	pty := &PTY{
		// Controller: read process output from outputRead (via the write pipe).
		Controller: os.NewFile(uintptr(outputRead), "pty-controller"),
		// Worker: write input to the process through inputWrite.
		Worker: os.NewFile(uintptr(inputWrite), "pty-worker"),
	}
	pty.close = func() error {
		windows.ClosePseudoConsole(hpc)
		// Close the *os.File so Go's runtime tracks the closure properly,
		// unblocking any pending Reads and preventing double-closes.
		return pty.Controller.Close()
	}
	return pty, nil
}
