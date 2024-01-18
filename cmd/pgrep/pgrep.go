package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	All       bool `short:"a" long:"list-full" description:"show all info"`
	ListAll   bool `short:"A" long:"show" description:"show all processes"`
	Memory    bool `short:"m" description:"print memory usage of process"`
	List      bool `short:"l" long:"list-name" description:"print process name and PID"`
	Pid       bool `short:"p" long:"pid" description:"print process name using the PID instead of a regex or name"`
	ParentPid int  `short:"P" long:"parent" description:"print all processes that have a PID as parent process - pgrep -P <pid>"`
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
		return "Sleeping", nil
	case "D":
		return "Waiting", nil
	case "Z":
		return "Zombie", nil
	case "T":
		return "Stopped", nil
	case "t":
		return "Tracing", nil
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
	return "Unknown", fmt.Errorf("Unknown State")
}

func printProc(proc Process) error {
	if opts.All {
		state, err := decodeState(proc.State())
		if err != nil {
			return err
		}
		pid := proc.Pid()
		cmd, _ := Cmdline(pid)
		mem, err := GetMemory(pid)
		render := fmt.Sprintf("%v\t%v\t%v\t%v\t%v", pid, proc.Executable(), cmd, state, mem)
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

func listProcs(procs []Process) {
	for _, p := range procs {
		printProc(p)
	}
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

func getProcsPid(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No pattern")
	}

	p, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println(err)
	}

	match, err := findProcess(p)
	if err != nil {
		return err
	}
	if match != nil {
		pid := match.Pid()
		exe := match.Executable()
		t, _ := StartTime(pid)
		cmd, _ := Cmdline(pid)
		fmt.Printf("pid: %v exe: %v cmd: %v Clock: %v\n", pid, exe, cmd, t)
	}

	return nil
}

func parents() ([]Process, error) {
	ppid := opts.ParentPid
	var matches []Process
	procs, err := processes()
	if err != nil {
		return nil, err
	}
	for _, item := range procs {
		if item.PPid() == ppid {
			matches = append(matches, item)
		}
	}
	return matches, nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Pid {
		if err := getProcsPid(args); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if opts.ListAll {
		if procs, err := processes(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("PID\tNAME\tCMD\tSTATE\tMEM")
			listProcs(procs)
		}
		os.Exit(0)
	}

	if opts.ParentPid != 0 {
		// take pid and find all procs that have that as ppid
		matches, err := parents()
		if err != nil {
			log.Fatal(err)
		}
		for _, proc := range matches {
			fmt.Println(proc.Pid())
		}
		os.Exit(0)
	}

	if err := getProcs(args); err != nil {
		log.Fatal(err)
	}
}
