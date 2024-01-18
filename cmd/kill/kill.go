package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
)

var defaultSignal = "-SIGTERM"

func kill(sig os.Signal, pids ...string) []error {
	var errs []error
	s := sig.(syscall.Signal)
	for _, p := range pids {
		pid, err := strconv.Atoi(p)
		if err != nil {
			errs = append(errs, fmt.Errorf("%v: args must be process or job ID", p))
			continue
		}
		if err := syscall.Kill(pid, s); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func list() {
	siglist()
}

func usage() {

}

func KillProcess(args []string) error {
	if len(args) < 1 {
		usage()
		return nil
	}
	op := args[0]
	pids := args[1:]
	if op[0] != '-' {
		op = defaultSignal
		pids = args[0:]
	}

	if op[0:2] == "-l" {
		if len(args) > 2 {
			usage()
			return nil
		}
		fmt.Printf("%s\n", siglist())
		return nil
	}

	if op == "-s" || op == "--signal" {
		if len(args) < 2 {
			usage()
			return nil
		}
		op = args[1]
		pids = args[2:]
	} else {
		op = op[1:]
	}

	s, ok := signums[op]
	if !ok {
		return fmt.Errorf("%v: is not a valid signature", op)
	}

	if len(pids) < 1 {
		usage()
		return nil
	}
	if err := kill(s, pids...); err != nil {
		return fmt.Errorf("some processes couldnt be killed: %v", err)
	}
	return nil
}

func main() {
	args := os.Args[1:]
	if err := KillProcess(args); err != nil {
		log.Fatal(err)
	}
}
