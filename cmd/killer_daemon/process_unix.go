// HUGE thank you to mitchellh:
// https://github.com/mitchellh/go-ps
// I tried to reinvent the whell here for no reason
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Process interface {
	// proc ID
	Pid() int
	// Parent proc ID
	PPid() int
	// exe name
	Executable() string
	// return process state
	State() rune
}

type UnixProcess struct {
	pid    int
	ppid   int
	state  rune
	pgrp   int
	sid    int
	binary string
}

func (p *UnixProcess) Pid() int {
	return p.pid
}

func (p *UnixProcess) Sid() int {
	return p.sid
}

func (p *UnixProcess) Pgrp() int {
	return p.pgrp
}

func (p *UnixProcess) PPid() int {
	return p.ppid
}

func (p *UnixProcess) Executable() string {
	return p.binary
}

func (p *UnixProcess) State() rune {
	return p.state
}

func findProcess(pid int) (Process, error) {
	dir := fmt.Sprintf("/proc/%d", pid)
	_, err := os.Stat(dir)
	if err != nil {
		// check if exists - if not return the error
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return newUnixProcess(pid)
}

func processes() ([]Process, error) {
	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	results := make([]Process, 0, 50)
	for {
		names, err := d.Readdirnames(10)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, name := range names {
			if name[0] < '0' || name[0] > '9' {
				continue
			}
			pid, err := strconv.ParseInt(name, 10, 0)
			if err != nil {
				continue
			}

			p, err := newUnixProcess(int(pid))
			if err != nil {
				continue
			}

			results = append(results, p)
		}
	}
	return results, nil
}

func newUnixProcess(pid int) (*UnixProcess, error) {
	p := &UnixProcess{pid: pid}
	return p, p.Refresh()
}
