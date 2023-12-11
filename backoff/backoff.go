package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	Timeout  string `short:"t" long:"timeout" description:"Timeout to limit the command to"`
	UseShell bool   `short:"s" long:"shell" description:"Use shell to run command instead of execve format"`
	Verbose  bool   `short:"v" long:"verbose" description:"make output more verbose"`
}

func System(command []string) *exec.Cmd {
	cmd := strings.Join(command, " ")
	c := exec.Command("sh", "-c", cmd)
	return c
}

func Run(timeout string, cmd []string) error {
	if len(cmd) == 0 {
		return errors.New("No command passed")
	}
	b := backoff.NewExponentialBackOff()
	if len(timeout) != 0 {
		d, err := time.ParseDuration(timeout)
		if err != nil {
			return err
		}
		if opts.Verbose {
			log.Printf("Timeout: %v\n", d)
		}
		b.MaxElapsedTime = d
	}

	f := func() error {
		var c *exec.Cmd
		if opts.UseShell {
			c = System(cmd)
		} else {
			c = exec.Command(cmd[0], cmd[1:]...)
		}

		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		err := c.Run()
		if opts.Verbose {
			log.Printf("Command: %v :: error: %v\n", cmd, err)
		}
		return err
	}
	return backoff.Retry(f, b)
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if err := Run(opts.Timeout, args); err != nil {
		log.Fatal(err)
	}
}
