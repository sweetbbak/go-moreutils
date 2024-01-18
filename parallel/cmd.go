package main

import (
	"io"
	"os/exec"
	"syscall"
)

// returna pointer to an exec.cmd
func newCmd(stdout, stderr io.Writer, cmd string, args ...string) *exec.Cmd {
	c := exec.Command(cmd, args...)
	c.Stderr, c.Stdout = stderr, stdout
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return c
}
