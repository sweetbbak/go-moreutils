package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Exec  string `short:"e" long:"exec" description:"command to run after timer equivalent to using sleep 1 && cmd"`
	Print bool   `short:"p" long:"print" description:"print time duration to stdout"`
}

func IsLetter(s string) bool {
	for _, r := range s {
		switch r {
		case 'h', 's', 'm', 'n', 'u':
		}
	}
	return true
}

func Sleep(args []string) (time.Duration, error) {
	var ti time.Duration

	for _, arg := range args {
		// if len(arg) == 1 || !is_alpha(arg) {
		if len(arg) == 1 || !IsLetter(arg) {
			arg += "s"
		}

		d, err := time.ParseDuration(arg)
		if err != nil {
			fmt.Println(err)
			return ti, err
		}
		return d, nil
	}
	return ti, nil
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

func tick(d time.Duration) {
	ticker := time.NewTicker(d)
	done := make(chan bool)

	go func() {
		b := time.Now()
		fmt.Printf("\x1b[2J\x1b[H")
		for {
			select {
			case <-done:
				fmt.Printf("\x1b[2J\x1b[H")
				return
			// case t := <-ticker.C:
			// 	fmt.Println("Tick at", t)
			default:
				fmt.Printf("\x1b[H\x1b[2K%v", time.Since(b).Truncate(0.0))
			}
		}
	}()

	time.Sleep(d)
	ticker.Stop()
	done <- true
}

func Eepy(args []string) error {
	t, err := Sleep(args)
	if err != nil {
		return err
	}

	if opts.Print {
		tick(t)
	} else {
		time.Sleep(t)
	}

	if opts.Exec != "" {
		exitc := system(opts.Exec)
		os.Exit(exitc)
	}

	return nil
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
