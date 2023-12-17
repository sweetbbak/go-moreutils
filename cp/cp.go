package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Recursive        bool   `short:"r" long:"recursive" description:"recursively copy a file tree"`
	Ask              bool   `short:"i" long:"interactive" description:"ask before overwriting files"`
	Force            bool   `short:"f" long:"force" description:"forcefully overwrite files [use caution]"`
	noFollowSymlinks bool   `short:"P" long:"no-dereference" default:"true" description:"dont follow symlinks"`
	Target           string `short:"t" long:"target-dir" description:"directory to copy files into"`
	Verbose          bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func confirmOverwrite(dest string, input *bufio.Reader) (bool, error) {
	fmt.Printf("cp: overwrite %v? ", dest)
	answer, err := input.ReadString('\n')
	if err != nil {
		return false, err
	}

	if strings.ToLower(answer)[0] != 'y' {
		return false, nil
	}

	return true, nil
}

func Copy(args []string) error {
	toDir := false

	var from []string
	var to string

	if opts.Target != "" {
		from, to = args, opts.Target
	} else {
		from, to = args[:len(args)-1], args[len(args)-1]
	}

	toStat, err := os.Stat(to)
	if err == nil {
		toDir = toStat.IsDir()
	}

	if len(args) > 2 && !toDir {
		return fmt.Errorf("No target directory for multiple files")
	}

	var lastErr error
	for _, file := range from {
		dest := to
		if toDir {
			dest = filepath.Join(dest, filepath.Base(file))
		}

		if opts.Recursive {
			// recursive copy func
		} else {
			// normal copy func
		}
	}

	return lastErr
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Copy(args); err != nil {
		log.Fatal(err)
	}
}
