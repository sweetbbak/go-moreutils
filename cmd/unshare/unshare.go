package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	ipc    bool     `short:"i" long:"ipc" description:"Unshare the IPC namespace"`
	mount  bool     `short:"m" long:"mount" description:"Unshare the mount namespace"`
	pid    bool     `short:"p" long:"pid" description:"Unshare the PID namespace"`
	net    bool     `short:"n" long:"net" description:"Unshare the net namespace"`
	uts    bool     `short:"U" long:"uts" description:"Unshare the uts namespace"`
	user   bool     `short:"u" long:"user" description:"Unshare the user namespace"`
	env    bool     `short:"E" long:"preserve-env" description:"preserve environment variables"`
	all    bool     `short:"a" long:"all" description:"Unshare all options provided"`
	SetEnv []string `short:"e" long:"env" description:"set environment variables for command ex: (--env USER=suwu)"`
}

func getShell() string {
	sh := os.ExpandEnv("SHELL")
	if sh == "" {
		sh = "/bin/bash"
	}
	return sh
}

func setEnv(args []string) {
	for _, v := range args {
		if strings.IndexByte(v, '=') > 0 {
			kv := strings.SplitN(v, "=", 2) // split key/value pair in 2
			if err := os.Setenv(kv[0], kv[1]); err != nil {
				log.Println(err)
			}
		}
	}
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

	if args[0] == "chroot" {
		syscall.Setuid(0)
		c.SysProcAttr.Chroot = args[1]
	}

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if opts.env {
		c.Env = os.Environ()
	}

	if len(opts.SetEnv) > 0 {
		c.Env = append(c.Env, opts.SetEnv...)
	}

	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if !opts.ipc && !opts.mount && !opts.net && !opts.pid && !opts.user && !opts.uts {
		opts.pid = true
		opts.user = true
	}

	if err := unshare(args); err != nil {
		log.Fatal(err)
	}
}
