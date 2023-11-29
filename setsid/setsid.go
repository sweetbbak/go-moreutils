package main

import (
	"log"
	"os"
	"syscall"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Wait    bool `short:"w" long:"wait" description:"wait"`
	Fork    bool `short:"f" long:"fork" description:"fork"`
	Ctty    bool `short:"c" long:"ctty" description:"controlling tty"`
	Verbose bool `short:"v" long:"verbose" description:"describe what program is doing"`
}

func Setsid(args []string) (int, error) {
	s := &syscall.SysProcAttr{
		Setsid: true,
	}
	if opts.Fork {
		s.Foreground = false
	}
	if opts.Ctty {
		s.Ctty = 0
	}
	dir, err := os.Getwd()
	if err != nil {
		dir = "/"
	}

	sy := syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdout.Fd(), os.Stdin.Fd(), os.Stderr.Fd()},
		Dir:   dir,
		Sys:   s,
	}

	pid, err := syscall.ForkExec(args[0], args[1:], &sy)
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func runSetsid(args []string) error {
	pid, err := Setsid(args)
	if err != nil {
		return err
	}
	if opts.Verbose {
		log.Printf("Started PID: [%d]", pid)
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if err := runSetsid(args); err != nil {
		log.Fatal(err)
	}
}
