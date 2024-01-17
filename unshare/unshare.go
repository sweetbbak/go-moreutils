package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	ipc   bool `short:"i" long:"ipc" description:"Unshare the IPC namespace"`
	mount bool `short:"m" long:"mount" description:"Unshare the mount namespace"`
	pid   bool `short:"p" long:"pid" description:"Unshare the PID namespace"`
	net   bool `short:"n" long:"net" description:"Unshare the net namespace"`
	uts   bool `short:"U" long:"uts" description:"Unshare the uts namespace"`
	user  bool `short:"u" long:"user" description:"Unshare the user namespace"`
}

func getShell() string {
	sh := os.ExpandEnv("SHELL")
	if sh == "" {
		sh = "/bin/bash"
	}
	return sh
}

func unshare(args []string) error {
	c := exec.Command(args[0], args[1:]...)
	c.SysProcAttr = &syscall.SysProcAttr{}
	if opts.mount {
		c.SysProcAttr.Cloneflags |= syscall.CLONE_NEWNS
	}
	if opts.uts {
		c.SysProcAttr.Cloneflags |= syscall.CLONE_NEWUTS
	}
	if opts.ipc {
		c.SysProcAttr.Cloneflags |= syscall.CLONE_NEWIPC
	}
	if opts.net {
		c.SysProcAttr.Cloneflags |= syscall.CLONE_NEWNET
	}
	if opts.pid {
		c.SysProcAttr.Cloneflags |= syscall.CLONE_NEWPID
	}
	if opts.user {
		c.SysProcAttr.Cloneflags |= syscall.CLONE_NEWUSER
	}

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}
	if err := unshare(args); err != nil {
		log.Fatal(err)
	}
}
