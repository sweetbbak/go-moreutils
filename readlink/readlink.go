package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Follow    bool `short:"f" long:"follow" description:"readlink and evaluate symlinks, follow to original file"`
	NoNewLine bool `short:"n" long:"no-newline" description:"do not output a trailing new line"`
	Zero      bool `short:"0" long:"zero" description:"separate output using NUL instead of a newline"`
	Verbose   bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Readlink(args []string) error {
	if len(args) > 1 && opts.NoNewLine {
		log.Println("Ignoring --no-newline with multiple arguments")
		opts.NoNewLine = false
	}

	for _, file := range args {
		err := readlink(file)
		if err != nil {
			Debug("%v\n", err)
			return err
		}
	}
	return nil
}

func readlink(file string) error {
	Debug("FILE: %v\n", file)
	fi, err := os.Lstat(file)
	if err != nil {
		return err
	}

	var path string

	if fi.Mode()&os.ModeSymlink != 0 {
		// buf := make([]byte, 128)
		// syscall.Readlink(file, buf)
		// fmt.Println(string(buf))

		path, err = os.Readlink(file)
		if err != nil {
			Debug("error reading symlink: %v\n", err)
			path = fi.Name()
		}
	} else {
		return fmt.Errorf("Not a symlink")
	}

	if opts.Follow {
		path, err = filepath.EvalSymlinks(path)
		Debug("%v\n", err)
	}

	delim := "\n"
	if opts.NoNewLine {
		delim = ""
	}

	if opts.Zero {
		delim = "\x00"
	}

	fmt.Printf("%s%s", path, delim)
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Readlink(args); err != nil {
		Debug("%w\n", err)
		os.Exit(1)
	}
}
