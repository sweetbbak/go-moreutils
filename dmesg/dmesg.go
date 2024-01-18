package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"golang.org/x/sys/unix"
)

var opts struct {
	clear     bool `short:"C" long:"clear" description:""`
	readClear bool `short:"c" long:"read-clear" description:""`
}

func dmesg() error {
	level := unix.SYSLOG_ACTION_READ_ALL
	if opts.clear && opts.readClear {
		return fmt.Errorf("cannot use --clear and --read-clear at the same time, they are exclusive")
	}
	if opts.readClear {
		level = unix.SYSLOG_ACTION_READ_CLEAR
	}
	if opts.clear {
		level = unix.SYSLOG_ACTION_CLEAR
	}
	b := make([]byte, 256*1024)
	o, err := unix.Klogctl(level, b)
	if err != nil {
		return fmt.Errorf("syslog failed: %v", err)
	}

	_, err = os.Stdout.Write(b[:o])
	return err
}

func main() {
	_, err := flags.Parse(&opts)
	if err == flags.ErrHelp {
		os.Exit(0)
	}

	if err := dmesg(); err != nil {
		log.Fatal(err)
	}
}
