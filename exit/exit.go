package main

import (
	"os"
)

func main() {

	if len(os.Args) == 2 {
		if len(os.Args[1]) < 2 {
			os.Exit(1)
		}
		if os.Args[1][0:2] == "-h" {
			println("hard exit from a program or shell by killing the process")
			println("generally it is better to use the shell built-in exit command")
		}
		os.Exit(0)
	}

	proc := os.Getppid()
	p, err := os.FindProcess(proc)
	if err != nil {
		println("unable to find process with parent pid: ", proc)
		os.Exit(1)
	}

	err = p.Kill()
	if err != nil {
		println("unable to kill process: ", p.Pid)
	}
}
