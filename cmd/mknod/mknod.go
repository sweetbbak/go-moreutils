package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	flags "github.com/jessevdk/go-flags"
	"golang.org/x/sys/unix"
)

const defaultPerm = 0o660

var opt struct {
	Mkdevnull bool `short:"d" long:"dev-null" description:"make dev null"`
}

type Device struct {
	Permission int
	DevType    string
	Major      int
	Minor      int
}

func newDevnull() *Device {
	// var Devnull Device
	var Devnull = Device{
		Permission: defaultPerm,
		DevType:    "c",
		Major:      1,
		Minor:      3,
	}
	return &Devnull
}

func newWatchDog() *Device {
	var Devnull = Device{
		Permission: defaultPerm,
		DevType:    "c",
		Major:      10,
		Minor:      130,
	}
	return &Devnull
}

func parseDevices(args []string, devtype string) (int, error) {
	if len(args) != 4 {
		return 0, fmt.Errorf("device type %v requires major and minor number", devtype)
	}
	major, err := strconv.ParseUint(args[2], 10, 12)
	if err != nil {
		return 0, err
	}
	minor, err := strconv.ParseUint(args[3], 10, 20)
	if err != nil {
		return 0, err
	}
	return int(unix.Mkdev(uint32(major), uint32(minor))), nil
}

func mknod(args []string) error {
	if len(args) != 2 && len(args) != 4 {
		return errors.New("usage: mknod path type [major minor]")
	}
	path := args[0]
	devtype := args[1]

	var err error
	var mode uint32
	mode = defaultPerm
	var dev int

	switch devtype {
	case "b":
		// block device - major/minor is needed
		mode |= unix.S_IFBLK
		dev, err = parseDevices(args, devtype)
		if err != nil {
			return err
		}
	case "c", "u":
		mode |= unix.S_IFCHR
		dev, err = parseDevices(args, devtype)
		if err != nil {
			return err
		}
	case "p":
		mode |= unix.S_IFIFO
		if len(args) != 2 {
			return fmt.Errorf("device type %v requires no other arguments", devtype)
		}
	default:
		return fmt.Errorf("device type not recognized: %v", devtype)
	}

	if err := unix.Mknod(path, mode, dev); err != nil {
		return fmt.Errorf("%q: mode %x: %v", path, mode, err)
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(0)
	}

	if opt.Mkdevnull {
		if len(args) != 1 {
			fmt.Println("usage: mknod --dev-null [path]")
			os.Exit(1)
		}
		d := newDevnull()
		args = append(args, d.DevType, fmt.Sprint(d.Major), fmt.Sprint(d.Minor))
		if err := mknod(args); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if err := mknod(args); err != nil {
		log.Fatal(err)
	}
}
