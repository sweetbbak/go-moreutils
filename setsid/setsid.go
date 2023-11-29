// thanks to smallnest for giving me a good idea of how this is done in Golang
// https://gist.github.com/smallnest/de6134e36b83fd6d215edf8db787235f
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

	pid, err := syscall.ForkExec(args[0], args[0:], &sy)
	if err != nil {
		return 0, err
	}
	return pid, nil
}

// command to run with setsid
func runSetsid(args []string) error {
	pid, err := Setsid(args)
	if err != nil {
		return err
	}

	if opts.Verbose {
		log.Printf("Started PID: [%d]", pid)
	}

	if opts.Wait {

		var wstat syscall.WaitStatus

		wpid, err := syscall.Wait4(pid, &wstat, 0, nil)
		if err != nil {
			return err
		}

		if opts.Verbose {
			log.Printf("Waiting for [%d]", wpid)
		}
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
