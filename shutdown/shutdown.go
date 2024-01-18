package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"
)

var (
	errEAGAIN error = syscall.EAGAIN
	errEINVAL error = syscall.EINVAL
	errENOENT error = syscall.ENOENT
)

const (
	ENOENT                      = syscall.Errno(0x2)
	EAGAIN                      = syscall.Errno(0xb)
	EINVAL                      = syscall.Errno(0x16)
	LINUX_REBOOT_MAGIC1         = 0xfee1dead
	LINUX_REBOOT_MAGIC2         = 0x28121969
	LINUX_REBOOT_CMD_POWER_OFF  = 0x4321fedc
	LINUX_REBOOT_CMD_RESTART    = 0x1234567
	LINUX_REBOOT_CMD_SW_SUSPEND = 0xd000fce2
)

var (
	opcodes = map[string]uint{
		"halt":    LINUX_REBOOT_CMD_POWER_OFF,
		"-h":      LINUX_REBOOT_CMD_POWER_OFF,
		"reboot":  LINUX_REBOOT_CMD_RESTART,
		"-r":      LINUX_REBOOT_CMD_RESTART,
		"suspend": LINUX_REBOOT_CMD_SW_SUSPEND,
		"-s":      LINUX_REBOOT_CMD_SW_SUSPEND,
	}
)

func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case EAGAIN:
		return errEAGAIN
	case EINVAL:
		return errEINVAL
	case ENOENT:
		return errENOENT
	}
	return e
}

func Reboot(cmd int) (err error) {
	return reboot(LINUX_REBOOT_MAGIC1, LINUX_REBOOT_MAGIC2, cmd, "")
}

func reboot(magic1 uint, magic2 uint, cmd int, arg string) (err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(arg)
	if err != nil {
		return
	}
	_, _, e1 := syscall.Syscall6(syscall.SYS_REBOOT, uintptr(magic1), uintptr(magic2), uintptr(cmd), uintptr(unsafe.Pointer(_p0)), 0, 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

var usage = ""

func shutdown(dryrun bool, args ...string) (uint, error) {
	if len(args) == 0 {
		args = append(args, "halt")
	}

	op, ok := opcodes[args[0]]
	if !ok {
		return 0, fmt.Errorf(usage)
	}

	if len(args) < 2 {
		args = append(args, "now")
	}

	when := time.Now()

	switch {
	case args[1] == "now":
	case args[1][0] == '+':
		m, err := time.ParseDuration(args[1][1:] + "m")
		if err != nil {
			return 0, err
		}
		when = when.Add(m)

	default:
		t, err := time.Parse(time.RFC3339, args[1])
		if err != nil {
			return 0, err
		}
		when = t
	}

	if !dryrun {
		time.Sleep(time.Until(when))
	}
	if !dryrun {
		if err := Reboot(int(op)); err != nil {
			return 0, err
		}
	}
	return op, nil
}

func main() {
	if _, err := shutdown(false, os.Args[1:]...); err != nil {
		log.Fatal(err)
	}
}
