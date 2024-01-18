package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"golang.org/x/sys/unix"
)

type Option struct {
	All      bool
	Machine  bool
	Nodename bool
	Release  bool
	Sysname  bool
	Version  bool
}

var opts struct {
	All      bool `short:"a" long:"all" description:"show all uname information"`
	Sysname  bool `short:"s" long:"sysname" description:"print the sysname information"`
	Nodename bool `short:"n" long:"node" description:"print the nodename information"`
	Release  bool `short:"r" long:"release" description:"print the release information"`
	Version  bool `short:"v" long:"version" description:"print the version information"`
	Machine  bool `short:"m" long:"machine" description:"print the machine information"`
}

func uname() error {
	var (
		err   error
		out   []string
		uname unix.Utsname
	)

	if err = unix.Uname(&uname); err != nil {
		return err
	}
	if !opts.Nodename && !opts.Release && !opts.Version && !opts.Machine {
		opts.Sysname = true
	}

	if opts.All {
		opts.Nodename = true
		opts.Release = true
		opts.Version = true
		opts.Machine = true
	}

	if opts.Sysname {
		out = append(out, strings.Trim(string(uname.Sysname[:]), "\x00"))
	}
	if opts.Nodename {
		out = append(out, strings.Trim(string(uname.Nodename[:]), "\x00"))
	}
	if opts.Release {
		out = append(out, strings.Trim(string(uname.Release[:]), "\x00"))
	}
	if opts.Version {
		out = append(out, strings.Trim(string(uname.Version[:]), "\x00"))
	}
	if opts.Machine {
		out = append(out, strings.Trim(string(uname.Machine[:]), "\x00"))
	}

	if _, err = fmt.Printf("%s\n", strings.Join(out, " ")); err != nil {
		return err
	}

	return nil
}

func run() {
	if err := uname(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	run()
}
