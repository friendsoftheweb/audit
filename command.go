package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"charm.land/huh/v2/spinner"
)

// interruptCtx is canceled when the user presses ctrl+c. It's shared by every
// command() call so that an interrupt kills the running child process and
// aborts the whole audit run, rather than just the current step.
var interruptCtx, _ = signal.NotifyContext(context.Background(), os.Interrupt)

func command(title string, args []string) ([]byte, error) {
	ctx, cancel := context.WithCancel(interruptCtx)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	var group sync.WaitGroup
	group.Add(1)

	go func() {
		defer group.Done()
		spinner.New().Type(spinner.Dots).Title(title).Context(ctx).Run()
	}()

	out, err := cmd.Output()
	cancel()
	group.Wait()

	if interruptCtx.Err() != nil {
		fmt.Println("\nInterrupted. Exiting...")
		os.Exit(130)
	}

	return out, err
}
