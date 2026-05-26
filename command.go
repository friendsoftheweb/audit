package main

import (
	"context"
	"os/exec"
	"sync"

	"charm.land/huh/v2/spinner"
)

func command(title string, args []string) ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.Command(args[0], args[1:]...)

	var group sync.WaitGroup
	group.Add(1)

	go func() {
		defer group.Done()
		spinner.New().Type(spinner.Dots).Title(title).Context(ctx).Run()
	}()

	out, err := cmd.Output()
	cancel()
	group.Wait()

	return out, err
}
