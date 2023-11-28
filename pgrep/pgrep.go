package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	All  bool `short:"a" long:"all" description:"show all info"`
	List bool `short:"l" long:"list" description:"print process name and PID"`
}

type syncPids struct {
	L    sync.Mutex
	Pids []Process
}

func Pgrep(expr string) ([]Process, error) {
	r, err := regexp.CompilePOSIX(expr)
	if err != nil {
		return nil, err
	}

	procs, err := processes()
	if err != nil {
		return nil, err
	}

	pids := new(syncPids)
	wg := new(sync.WaitGroup)
	wg.Add(len(procs))
	for _, proc := range procs {
		proc := proc
		go func() {
			defer wg.Done()
			if r.MatchString(proc.Executable()) {
				pids.L.Lock()
				pids.Pids = append(pids.Pids, proc)
				pids.L.Unlock()
			}
		}()
	}
	wg.Wait()
	return pids.Pids, nil
}

func decodeState(state rune) (string, error) {
	s := string(state)
	switch s {
	case "R":
		return "Running", nil
	case "S":
		return "Sleeping uninterruptible", nil
	case "D":
		return "Waiting uninterruptible", nil
	case "Z":
		return "Zombie", nil
	case "T":
		return "Stopped on a signal", nil
	case "t":
		return "Tracing stopped", nil
	case "X":
		return "Dead", nil
	case "x":
		return "Dead", nil
	case "K":
		return "Wakekill", nil
	case "W":
		return "Waking", nil
	case "P":
		return "Parked", nil
	case "I":
		return "Idle", nil
	}
	return "", fmt.Errorf("Unknown State")
}

func printProc(proc Process) error {
	if opts.All {
		state, err := decodeState(proc.State())
		if err != nil {
			return err
		}
		render := fmt.Sprintf("%v %v %v", proc.Pid(), proc.Executable(), state)
		fmt.Println(render)
		return nil
	}

	if opts.List {
		render := fmt.Sprintf("%v %v", proc.Pid(), proc.Executable())
		fmt.Println(render)
		return nil
	}

	if !opts.All || !opts.List {
		render := fmt.Sprintf("%v", proc.Pid())
		fmt.Println(render)
		return nil
	}
	return nil
}

func getProcs(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No pattern")
	}

	matches, err := Pgrep(args[0])
	if err != nil {
		return err
	}

	if len(matches) != 0 {
		for _, item := range matches {
			printProc(item)
		}
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if err := getProcs(args); err != nil {
		log.Fatal(err)
	}
}
