package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Recursive bool `short:"R" long:"recursive" description:"copy the file mode from an existing reference file"`
	Verbose   bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Chgrp(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Not enough arguments, needs: chgrp [options] <group> <file>")
	}

	gid, err := lookupgid(args[0])
	if err != nil {
		return err
	}

	Debug("Provided Group: %s - Found GID%v\n", args[0], gid)

	if opts.Recursive {
		return recursiveChgrp(args[1:], gid)
	}

	for _, path := range args[1:] {
		path = os.ExpandEnv(path)
		if err := changeGroup(path, gid); err != nil {
			return err
		}
		Debug("Changed group of file: %s - %v\n", path, gid)
	}
	return nil
}

func recursiveChgrp(paths []string, gid int) error {
	for _, path := range paths {
		path = os.ExpandEnv(path)
		err := filepath.Walk(path, func(ppath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if err := changeGroup(ppath, gid); err != nil {
				return err
			}
			Debug("Changed group of file: %s - %v\n", ppath, gid)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func changeGroup(path string, gid int) error {
	var statt syscall.Stat_t
	if err := syscall.Stat(path, &statt); err != nil {
		return err
	}
	return os.Chown(path, int(statt.Uid), gid)
}

func lookupgid(str string) (int, error) {
	group, err := user.LookupGroupId(str)
	if err != nil {
		group, err = user.LookupGroup(str)
		if err != nil {
			return 0, fmt.Errorf("Unable to lookup provided group ID: %s: %v", str, err)
		}
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return 0, fmt.Errorf("Unable to decode GID: %v", err)
	}

	return gid, nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Chgrp(args); err != nil {
		log.Fatal(err)
	}
}
