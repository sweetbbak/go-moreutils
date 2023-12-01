package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	// "os"
)

var (
	File  string
	Regex bool
	Cmd   bool
)

func init() {
	flag.StringVar(&File, "f", "", "path to a file with a list of process names to block")
	flag.StringVar(&File, "file", "", "path to a file with a list of process names to block")

	flag.BoolVar(&Regex, "r", false, "use a substring to match process names")
	flag.BoolVar(&Regex, "rough", false, "use a substring to match process names")

	flag.BoolVar(&Cmd, "c", false, "match args used for a process")
	flag.BoolVar(&Cmd, "cmd", false, "match args used for a process")
}

func getProcs() ([]Process, error) {
	p, err := processes()
	if err != nil {
		return nil, err
	}
	return p, nil
}

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

			self := os.Getpid()
			parent := os.Getppid()
			if strings.Contains(cmdargs, arg) && proc.Pid() != parent && proc.PPid() != parent {
				fmt.Println(cmdargs, arg, self, parent)
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

func WatchDog(args []string) error {
	p, _ := getProcs()
	for _, arg := range args {
		ProcessPid(arg, p)
	}
	return nil
}

func main() {
	flag.Parse()
	if err := WatchDog(flag.Args()); err != nil {
		log.Fatal(err)
	}
}
