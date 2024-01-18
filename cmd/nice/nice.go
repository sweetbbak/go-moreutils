package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Nice    int  `short:"n" long:"nice" description:"set nice level (-20 to 19)"`
	GetNice int  `short:"p" long:"pid" description:"get nice level of a process"`
	Setsid  bool `short:"s" long:"setsid" description:"create a new session and detach from controlling terminal"`
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

const (
	IOPRIO_CLASS_NONE int = 0
	IOPRIO_CLASS_RT   int = 1
	IOPRIO_CLASS_BE   int = 2
	IOPRIO_CLASS_IDLE int = 3

	IOPRIO_WHO_PROCESS = 1
	IOPRIO_WHO_PGRP    = 2
	IOPRIO_WHO_USER    = 3
	IOPRIO_CLASS_SHIFT = 13
)

// p is the PID of the process
func getNice(p int) (int, error) {
	return syscall.Getpriority(syscall.PRIO_PROCESS, p)
}

func setNice(pid int, priority int) error {
	// r1, _, e1 := syscall.Syscall(syscall.SYS_IOPRIO_SET, ioprioWhoProcess, uintptr(pid), uintptr(ioprioPrioValue(ioprioClassBe, priority)))
	// if int(r1) == -1 {
	// 	return e1
	// }
	// return nil

	return syscall.Setpriority(syscall.PRIO_PROCESS, pid, priority)
}

func validateNice(n int) bool {
	return n >= -20 && n <= 19
}

func Nice(args []string) error {
	if opts.GetNice != -1 {
		p, err := getNice(opts.GetNice)
		if err != nil {
			return err
		}
		fmt.Println(p)
		return nil
	}

	if !validateNice(opts.Nice) {
		return fmt.Errorf("Nice level must be betweein -20 and 19")
	}

	setIoPriority(opts.Nice)
	proc, err := Start(opts.Setsid, args...)
	if err != nil {
		return err
	}

	err = setNice(proc.Pid, opts.Nice)
	if err != nil {
		fmt.Println(err)
	}

	err = setNice(os.Getpid(), opts.Nice)
	if err != nil {
		fmt.Println(err)
	}

	n, err := getNice(proc.Pid)
	fmt.Println("proc nice level: ", n)
	n, err = getNice(os.Getpid())
	fmt.Println("proc nice level: ", n)

	return nil
}

func Start(setsid bool, args ...string) (p *os.Process, err error) {
	if args[0], err = exec.LookPath(args[0]); err == nil {

		var procAttr os.ProcAttr
		sys := syscall.SysProcAttr{
			Setsid: setsid,
			// AmbientCaps: []uintptr{6},
		}

		cwd, _ := os.Getwd()

		procAttr = os.ProcAttr{
			Dir:   cwd,
			Env:   os.Environ(),
			Sys:   &sys,
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		}

		p, err := os.StartProcess(args[0], args, &procAttr)
		if err == nil {
			return p, nil
		}
	}

	return nil, err
}

func init() {
	opts.GetNice = -1
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Nice(args); err != nil {
		log.Fatal(err)
	}
}
