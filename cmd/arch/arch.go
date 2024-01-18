package main

import (
	"os"
	"runtime"
)

var version = "0.1.0"

var usage = `USAGE:
arch [OPTIONS]
	-h, --help       print this help message
	-a, --arch       print the Architecture of the machine
	-A, --all        print all the fields
	-o, --os         print the Operating System
	-v, --version    print the version information
	`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			println(usage)
		case "-a", "--arch":
			println(runtime.GOARCH)
		case "-o", "--os":
			println(runtime.GOOS)
		case "-A", "--all":
			println(runtime.GOARCH)
			println(runtime.GOOS)
		case "-v", "--version":
			println(version)
		}
	} else {
		println(runtime.GOARCH)
	}
}
