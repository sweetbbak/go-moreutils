package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/jessevdk/go-flags"
	"golang.org/x/sys/unix"
)

var opts struct {
	Reboot  bool `short:"r" long:"reboot" description:"reboot the system"`
	Halt    bool `short:"H" long:"halt" description:"halt the machine regarldess of which command is invoked [poweroff|halt|reboot|suspend]"`
	Suspend bool `short:"s" long:"suspend" description:"suspend the machine"`
	NoSync  bool `short:"n" long:"no-sync" description:"do not sync filesysems and just reboot (data may be lost)"`
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func reboot() error {
	syncFS()
	if err := unix.Reboot(unix.LINUX_REBOOT_CMD_RESTART); err != nil {
		return err
	}
	return nil
}

func halt() error {
	syncFS()
	if err := unix.Reboot(unix.LINUX_REBOOT_CMD_POWER_OFF); err != nil {
		return err
	}
	return nil
}

func suspend() error {
	syncFS()
	if err := unix.Reboot(unix.LINUX_REBOOT_CMD_SW_SUSPEND); err != nil {
		return err
	}
	return nil
}

func stopPID1() error {
	process, err := os.FindProcess(1)
	if err != nil {
		return err
	}

	if err := process.Signal(syscall.SIGUSR1); err != nil {
		return fmt.Errorf("could not send signal SIGUSR1 to process PID 1: %v", err)
	}
	return nil
}

func syncFS() {
	if !opts.NoSync {
		syscall.Sync()
	}
}

func sanityRun() error {
	s, err := exec.LookPath("systemctl")
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	if opts.Reboot {
		cmd = exec.Command(s, "reboot")
	}

	if opts.Halt {
		cmd = exec.Command(s, "shutdown")
	}

	return cmd.Run()
}

func Halt(args []string) error {
	if err := sanityRun(); err == nil {
		return err
	}

	if opts.Halt {
		return halt()
	}
	if opts.Reboot {
		return reboot()
	}
	if opts.Suspend {
		suspend()
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

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Halt(args); err != nil {
		log.Fatal(err)
	}
}
