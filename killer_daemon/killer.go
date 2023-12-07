package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	File  string
	Regex bool
	Cmd   bool
)

var usage = `Usage:
		
Examples:
	killer -c "start-as=fullscreen"
	`

func init() {
	flag.StringVar(&File, "f", "", "path to a file with a list of process names to block")
	flag.StringVar(&File, "file", "", "path to a file with a list of process names to block")

	flag.BoolVar(&Regex, "r", false, "use a substring to match process names")
	flag.BoolVar(&Regex, "rough", false, "use a substring to match process names")

	flag.BoolVar(&Cmd, "c", false, "match args used for a process")
	flag.BoolVar(&Cmd, "cmd", false, "match args used for a process")
}

// get list of all processes
func getProcs() ([]Process, error) {
	p, err := processes()
	if err != nil {
		return nil, err
	}
	return p, nil
}

// iterate process list for matching block list items
func ProcessPid(arg string, p []Process) ([]Process, error) {
	var matches []Process
	for _, proc := range p {
		if Regex {
			if strings.Contains(proc.Executable(), arg) {
				fmt.Printf("Found match: %v of [PID] %v\n", proc.Executable(), proc.Pid())
				matches = append(matches, proc)
			}
		}

		if Cmd {
			cmdargs, err := Cmdline(proc.Pid())
			if err != nil {
				continue
			}

			parent := os.Getppid()
			if strings.Contains(cmdargs, arg) && proc.Pid() != parent && proc.PPid() != parent {
				// fmt.Println(cmdargs, arg, self, parent)
				fmt.Printf("Found substring  match: %v of [PID] %v\n", proc.Executable(), proc.Pid())
				matches = append(matches, proc)
			}
		}

		if !Cmd && !Regex {
			if proc.Executable() == arg {
				fmt.Printf("Found match: %v of [PID] %v\n", proc.Executable(), proc.Pid())
				matches = append(matches, proc)
			}
		}
	}
	return matches, nil
}

// iterate over list of matching procs and send to kill
func KillMatches(procs []Process) {
	for _, proc := range procs {
		err := kill(os.Kill, proc.Pid())
		if err != nil {
			log.Printf("unable to kill process: %v", err)
		} else {
			log.Printf("Killed process: %v of [PID] %v", proc.Executable(), proc.Pid())
		}
	}
}

// kill PID with Signal
func kill(sig os.Signal, pid int) error {
	s := sig.(syscall.Signal)
	if err := syscall.Kill(pid, s); err != nil {
		return err
	}
	return nil
}

// scan processes - send to be processed - send to be killed
func WatchDog(args []string) error {
	p, _ := getProcs()
	for _, arg := range args {
		matches, err := ProcessPid(arg, p)
		if err != nil {
			log.Println(err)
		}
		go KillMatches(matches)
	}
	return nil
}

// capture signals that kill the main program
func handleKillSignal() {
	signalChan := make(chan os.Signal, 1)
	// sigkill apparently cannot be caught
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			<-signalChan
			println("bye bye!")
			os.Exit(0)
		}
	}()
}

func main() {
	flag.Parse()

	// handle signals and exit
	handleKillSignal()

	// loop in the background while looking for procs to kill
	for {
		time.Sleep(1 * time.Second)
		if err := WatchDog(flag.Args()); err != nil {
			log.Println(err)
		}
	}
}
