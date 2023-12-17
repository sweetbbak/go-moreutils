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
	Verbose   bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Readlink(args []string) error {
	for _, file := range args {
		err := readlink(file)
		if err != nil {
			Debug("%v\n", err)
			fmt.Println(file)
		}
	}
	return nil
}

func readlink(file string) error {
	Debug("FILE: %v\n", file)
	fi, err := os.Lstat(file)
	if err != nil {
		Debug("%v\n", err)
		return err
	}

	var path string

	if fi.Mode()&os.ModeSymlink != 0 {
		path, err = os.Readlink(file)
		if err != nil {
			Debug("error reading symlink: %v\n", err)
			path = fi.Name()
		}
	} else {
		path = fi.Name()
	}

	if opts.Follow {
		path, err = filepath.EvalSymlinks(file)
		// path, err = filepath.Abs(file)
		Debug("%v\n", err)
	}

	delim := "\n"
	if opts.NoNewLine {
		delim = ""
	}

	fmt.Printf("%s%s", path, delim)
	return err
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
		log.Fatal(err)
	}
}
