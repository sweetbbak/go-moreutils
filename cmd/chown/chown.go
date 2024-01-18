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

func Chown(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Not enough arguments, needs: chown [options] <user> <file>")
	}

	uid, err := lookupuid(args[0])
	if err != nil {
		return err
	}

	Debug("Provided user: %s - Found UID%v\n", args[0], uid)

	if opts.Recursive {
		return recursiveChown(args[1:], uid)
	}

	for _, path := range args[1:] {
		path = os.ExpandEnv(path)
		if err := changeOwn(path, uid); err != nil {
			return err
		}
		Debug("Changed owner of file: %s - %v\n", path, uid)
	}
	return nil
}

func recursiveChown(paths []string, uid int) error {
	for _, path := range paths {
		path = os.ExpandEnv(path)
		err := filepath.Walk(path, func(ppath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if err := changeOwn(ppath, uid); err != nil {
				return err
			}
			Debug("Changed owner of file: %s - %v\n", ppath, uid)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func changeOwn(path string, uid int) error {
	var statt syscall.Stat_t
	if err := syscall.Stat(path, &statt); err != nil {
		return err
	}
	return os.Chown(path, uid, int(statt.Gid))
}

func lookupuid(str string) (int, error) {
	u, err := user.LookupId(str)
	if err != nil {
		u, err = user.Lookup(str)
		if err != nil {
			return 0, fmt.Errorf("Unable to lookup provided ID: %s: %v", str, err)
		}
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, fmt.Errorf("Unable to decode UID: %v", err)
	}

	return uid, nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Chown(args); err != nil {
		log.Fatal(err)
	}
}
