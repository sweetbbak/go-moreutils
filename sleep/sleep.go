package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Exec string `short:"e" long:"exec" description:"command to run after timer equivalent to using sleep 1 && cmd"`
}

func Sleep(args []string) (time.Duration, error) {
	var ti time.Duration

	for _, arg := range args {
		if len(arg) == 1 {
			arg += "s"
		}

		d, err := time.ParseDuration(arg)
		if err != nil {
			return ti, err
		}
		return d, nil
	}
	return ti, nil
}

func Eepy(args []string) error {
	t, err := Sleep(args)
	if err != nil {
		return err
	}

	time.Sleep(t)
	if opts.Exec != "" {
		exitc := system(opts.Exec)
		os.Exit(exitc)
	}

	return nil
}

func system(cmd string) int {
	c := exec.Command("sh", "-c", cmd)
	c.Env = os.Environ()
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := c.Run()
	if err == nil {
		return 0
	}

	// Figure out the exit code
	if ws, ok := c.ProcessState.Sys().(syscall.WaitStatus); ok {
		if ws.Exited() {
			return ws.ExitStatus()
		}

		if ws.Signaled() {
			return -int(ws.Signal())
		}
	}
	return -1
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if err := Eepy(args); err != nil {
		log.Fatal(err)
	}
}
