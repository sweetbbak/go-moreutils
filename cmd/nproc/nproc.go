package main

import (
	"os"
	"runtime"
)

func init() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			println("Usage: nproc")
			println("Print the number of processing units available to the current process. AKA print number of CPU cores")
			os.Exit(0)
		}
	}
}

func main() {
	println(runtime.NumCPU())
}
