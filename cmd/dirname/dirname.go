package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Zero    bool `short:"z" long:"zero" description:"command to run after timer equivalent to using sleep 1 && cmd"`
	Verbose bool `short:"v" long:"verbose" description:"Print exactly what the program is doing"`
}

var (
	sep string = "\n"
)

func getDirname(file string) string {
	return filepath.Dir(filepath.Clean(file))
}

func dirnames(args []string) error {
	for i, fi := range args {
		dirn := getDirname(fi)
		if opts.Verbose {
			fmt.Fprintf(os.Stdout, "dirname %s - filename: %s%v", dirn, args[i], sep)
		} else {
			fmt.Fprintf(os.Stdout, "%s%v", dirn, sep)
		}
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil || len(args) == 0 {
		// error in this context is bad cmdline options, so we print help and tack on units and examples
		if flags.WroteHelp(err) {
		}
		os.Exit(0)
	}

	if opts.Verbose {
		fmt.Println(args)
	}

	if opts.Zero {
		sep = "\x00"
	}

	// else we run once and log our error
	if err := dirnames(args); err != nil {
		log.Fatal(err)
	}
}
