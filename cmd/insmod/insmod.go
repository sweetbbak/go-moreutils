package main

import (
	"log"
	"os"
	"strings"

	"mybox/pkg/kmodule"
)

const usage = `insmod [filename] [module parameters...]`

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("insmod: ERROR: missing file name")
	}

	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		println("Usage:")
		println("    " + usage)
		println("\nDescription:")
		println("   insmod is a simple program to insert a module into the Linux Kernel")
		println("   insmod takes a path to a .ko file and optional commandline parameters for the module")
		println("   for more information, see [man insmod(8)] and see dmesg for verbose errors")
		os.Exit(0)
	}

	filename := os.Args[1]
	opts := strings.Join(os.Args[2:], " ")
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := kmodule.FileInit(f, opts, 0); err != nil {
		log.Fatalf("insmod: could not load %v: %v", filename, err)
	}
}
