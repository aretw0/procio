package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/aretw0/procio/pty"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", "echo Wait... && timeout /t 2 && echo Hello from ConPTY!")
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", "echo Wait... && sleep 2 && echo Hello from PTY!")
	}

	fmt.Println("--> Starting process within a pseudo-terminal...")
	p, err := pty.StartPTY(cmd)
	if err != nil {
		fmt.Printf("Error starting PTY: %v\n", err)
		os.Exit(1)
	}
	defer p.Close()

	// Asynchronously copy PTY output to the real standard output
	go func() {
		// Output from a PTY might contain terminal escape sequences (coloring, cursor moves, etc.)
		_, _ = io.Copy(os.Stdout, p.Controller)
	}()

	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\n--> Process finished with error: %v\n", err)
	} else {
		fmt.Println("\n--> Process finished successfully!")
	}
}
