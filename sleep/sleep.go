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
	Exec    string `short:"e" long:"exec" description:"command to run after timer equivalent to using sleep 1 && cmd"`
	Print   bool   `short:"p" long:"print" description:"print time duration to stdout"`
	Loop    int    `short:"l" long:"loop" default:"0" description:"loop the sleep and exec command X number of times"`
	Verbose bool   `short:"v" long:"verbose" description:"Print exactly what the program is doing"`
}

func IsLetter(s string) bool {
	for _, r := range s {
		switch r {
		case 'h', 's', 'm', 'n', 'u':
		}
	}
	return true
}

// parse sleep duration, sleep for that amount of time
func Sleep(args []string) (time.Duration, error) {
	var ti time.Duration

	// if no letter is provided we assume the unit 's' for seconds
	for _, arg := range args {
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

// execute a command using a shell and get the integer exit code
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

// timer for printing the timer to the screen
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

// main sleep function
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
		if opts.Verbose {
			fmt.Printf("Program exited with code [%v]", exitc)
		}

		if opts.Loop == 0 {
			os.Exit(exitc)
		} else {
			return nil
		}
	}

	return nil
}

func looper(args []string) {
	if opts.Loop == -1 {
		x := 0
		for {
			x += 1
			if opts.Verbose {
				fmt.Printf("Loop count: [%v]", x)
			}

			if err := Eepy(args); err != nil {
				log.Println(err)
			}
		}
	}

	// loop until we are done and early exit
	if opts.Loop > 0 {
		for i := 0; i < opts.Loop; i++ {
			if opts.Verbose {
				fmt.Println("Looping: ", opts.Loop, i+1)
			}

			if err := Eepy(args); err != nil {
				log.Println(err)
			}
		}
		os.Exit(0)
	}
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		// error in this context is bad cmdline options, so we print help and tack on units and examples
		if flags.WroteHelp(err) {
			fmt.Println("[\x1b[31mUNITS\x1b[0m] us, ns, ms, s, m, h (default is 's')")
			fmt.Println("[\x1b[31mEXAMPLES\x1b[0m] 1s, 1m, 99us, 500ms, 2h30m, 1h22m333ms")
		}
		os.Exit(0)
	}

	if opts.Verbose {
		fmt.Println(args)
	}

	// loop until we are done and early exit
	looper(args)

	// else we run once and log our error
	if err := Eepy(args); err != nil {
		log.Fatal(err)
	}
}
