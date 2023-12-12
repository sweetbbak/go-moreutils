package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Force  bool `short:"f" long:"force" description:"forcefully overwrite target files if they exist"`
	Sym    bool `short:"s" long:"symbolic" description:"create a symbolic link"`
	Backup bool `short:"b" long:"backup" description:"create a backup file if the target file already exists"`
}

func Backup(path string) error {
	return os.Rename(path, path+".bak")
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FileNotExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true
	}
	return false
}

func Symlink(args []string) error {
	src := os.ExpandEnv(args[0])
	dest := os.ExpandEnv(args[1])

	if FileNotExists(src) {
		return fmt.Errorf("Target file does not exist")
	}

	if FileExists(dest) {
		if opts.Backup {
			err := Backup(dest)
			if err != nil {
				return err
			}
		} else if !opts.Force {
			return fmt.Errorf("Destination file exists, use [-f|--force] [-b|--backup] or remove the destination file")
		} else {
			if err := os.Remove(dest); err != nil {
				return err
			}
		}
	}

	if opts.Sym {
		os.Symlink(src, dest)
	} else {
		os.Link(src, dest)
	}

	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err == flags.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		fmt.Printf("error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if len(args) < 2 {
		fmt.Printf("usage: ln [options] source destination\n")
		os.Exit(1)
	}

	if err := Symlink(args); err != nil {
		log.Fatal(err)
	}
}
