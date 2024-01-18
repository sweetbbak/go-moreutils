package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

var (
	help       bool
	data       bool
	filesystem bool
	// data       = flag.Bool("data", false, "sync file data, not metadata")
	// filesystem = flag.Bool("filesystem", false, "commit filesystem caches to disk")
)

var usage = "Usage: %s [OPTIONS] [FILE]...\n\t-f, --filesystem\tsync file data, not metadata.\n\t-d, --data\t\tcommit filesystem caches to disk\n\t-h, --help\t\tshow this help message\n"

func init() {
	flag.BoolVar(&data, "data", false, "sync file data, not metadata")
	flag.BoolVar(&data, "d", false, "sync file data, not metadata")

	flag.BoolVar(&filesystem, "filesystem", false, "commit filesystem caches to disk")
	flag.BoolVar(&filesystem, "f", false, "commit filesystem caches to disk")

	flag.BoolVar(&help, "help", false, "show this help message")
	flag.BoolVar(&help, "h", false, "show this help message")

	flag.Parse()
	helpMessage()
}

func helpMessage() {
	if help == true {
		fmt.Printf(usage, os.Args[0])
		os.Exit(0)
	}
}

func doSyscall(syscallNum uintptr, args []string) error {
	for _, filename := range args {
		f, err := os.OpenFile(filename, syscall.O_RDONLY|syscall.O_NOCTTY|syscall.O_CLOEXEC, 0o644)
		if err != nil {
			return err
		}
		if _, _, err = syscall.Syscall(syscallNum, uintptr(f.Fd()), 0, 0); err.(syscall.Errno) != 0 {
			return err
		}
		f.Close()
	}
	return nil
}

func sync(data, filesys bool, args []string) error {
	switch {
	case data:
		return doSyscall(unix.SYS_FDATASYNC, args)
	case filesys:
		return doSyscall(unix.SYS_SYNCFS, args)
	default:
		syscall.Sync()
		return nil
	}
}

func main() {
	if err := sync(data, filesystem, flag.Args()); err != nil {
		log.Fatal(err)
	}
}
