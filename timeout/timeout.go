package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"time"
)

type cmd struct {
	args         []string
	timeout      time.Duration
	in, out, err *os.File
}

var (
	timeout = flag.Duration("t", 30*time.Second, "Timeout for command")
)

func (c *cmd) run() (int, error) {
	if len(c.args) == 0 {
		return 1, errors.New("No command to run")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	process := exec.CommandContext(ctx, c.args[0], c.args[1:]...)
	process.Stdin, process.Stdout, process.Stderr = c.in, c.out, c.err
	if err := process.Run(); err != nil {
		exitCode := 1
		var e *exec.ExitError
		if errors.As(err, &e) {
			exitCode = e.ExitCode()
		}
		return exitCode, err
	}
	return 0, nil
}

func main() {
	flag.Parse()
	c := &cmd{args: flag.Args(), in: os.Stdin, out: os.Stdout, err: os.Stderr, timeout: *timeout}
	if exitCode, err := c.run(); err != nil || exitCode != 0 {
		log.Printf("timeout [%v]: %v", *timeout, err)
		os.Exit(exitCode)
	}
}
