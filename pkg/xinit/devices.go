package internal

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/pilebones/go-udev/crawler"
	"github.com/pilebones/go-udev/netlink"
	"golang.org/x/sys/unix"
)

// Mknod creates a filesystem node (file, device special file or named pipe) named path
// with attributes specified by mode and dev.
func Mknod(path string, mode uint32, dev int) error {
	return syscall.Mknod(path, mode, dev)
}

// Mkdev is used to build the value of linux devices (in /dev/) which specifies major
// and minor number of the newly created device special file.
// Linux device nodes are a bit weird due to backwards compat with 16 bit device nodes.
// They are, from low to high: the lower 8 bits of the minor, then 12 bits of the major,
// then the top 12 bits of the minor.
func Mkdev(major, minor int) int {
	return ((minor & 0xfff00) << 12) | ((major & 0xfff) << 8) | (minor & 0xff)
}

func CreateDevices() error {
	if err := Mknod("/dev/null", unix.S_IFCHR|0666, Mkdev(1, 3)); err != nil {
		return err
	}

	if err := Mknod("/dev/zero", unix.S_IFCHR|0666, Mkdev(1, 5)); err != nil {
		return err
	}

	if err := Mknod("/dev/random", unix.S_IFCHR|0666, Mkdev(1, 9)); err != nil {
		return err
	}

	if err := Mknod("/dev/urandom", unix.S_IFCHR|0666, Mkdev(1, 9)); err != nil {
		return err
	}

	//
	// Create symlinks for /dev/stdin /dev/stdout and /dev/stderr
	//

	if err := os.Symlink("/proc/self/fd/0", "/dev/stdin"); err != nil {
		return err
	}

	if err := os.Symlink("/proc/self/fd/1", "/dev/stdout"); err != nil {
		return err
	}

	if err := os.Symlink("/proc/self/fd/2", "/dev/stderr"); err != nil {
		return err
	}

	return nil
}

func addDevice(device crawler.Device) error {
	devname := device.Env["DEVNAME"]
	if devname == "" {
		return nil
	}

	var (
		mode  uint32
		major uint32
		minor uint32
	)

	devtype := device.Env["DEVTYPE"]
	subsystem := device.Env["SUBSYSTEM"]

	if m, err := strconv.ParseUint(device.Env["MAJOR"], 10, 32); err == nil {
		major = uint32(m)
	} else {
		return fmt.Errorf("error parsing major: %w", err)
	}
	if n, err := strconv.ParseUint(device.Env["MINOR"], 10, 32); err == nil {
		minor = uint32(n)
	} else {
		return fmt.Errorf("error parsing minor: %w", err)
	}

	if n, err := strconv.ParseUint(device.Env["DEVMODE"], 8, 32); err == nil {
		mode = uint32(n)
	} else {
		mode = uint32(0o666)
	}

	if devtype == "disk" || devtype == "partition" {
		mode |= unix.S_IFBLK
	} else if subsystem == "tty" {
		mode |= unix.S_IFCHR
	} else {
		return fmt.Errorf("unhandled device type %q", devtype)
	}

	path := filepath.Join("/dev", devname)

	if strings.Count(path, "/") > 2 {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, fs.FileMode(0755)); err != nil {
			return fmt.Errorf("error creating device path %q: %w", dir, err)
		}
	}

	dev := unix.Mkdev(major, minor)
	if err := syscall.Mknod(path, mode, int(dev)); err != nil {
		return fmt.Errorf("error creating %q device %q: %w", devtype, devname, err)
	}

	return nil
}

func FileNotExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true
	}
	return false
}

func removeDevice(device crawler.Device) error {
	devname := device.Env["DEVNAME"]
	if devname == "" {
		return nil
	}

	path := filepath.Join("/dev", devname)
	if FileNotExists(path) {
		return nil
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("error removing device %q: %w", devname, err)
	}

	return nil
}

// ScanDevices scans /sys for devices and populate /dev
func ScanDevices(ctx context.Context) error {
	log.Println("Scanning for devices...")

	queue := make(chan crawler.Device)
	errors := make(chan error)
	quit := crawler.ExistingDevices(queue, errors, nil)

	// Handling message from queue
	for {
		select {
		case device, more := <-queue:
			if !more {
				log.Println("Finished scanning devices")
				return nil
			}

			if err := addDevice(device); err != nil {
				log.Printf("error adding device %q: %s\n", device, err)
			}
		case err := <-errors:
			log.Printf("error scanning devices: %s\n", err)
		case <-ctx.Done():
			quit <- struct{}{}
			return nil
		}
	}
}

// WatchDevices watches /sys for devices and populates /dev
func WatchDevices(ctx context.Context) error {
	log.Println("Watching for devices...")

	conn := new(netlink.UEventConn)
	if err := conn.Connect(netlink.UdevEvent); err != nil {
		return fmt.Errorf("error connecting to netlink socket: %s", err)
	}
	defer conn.Close()

	queue := make(chan netlink.UEvent)
	errors := make(chan error)
	quit := conn.Monitor(queue, errors, nil)

	// Handling message from queue
	for {
		select {
		case uevent, more := <-queue:
			if !more {
				return nil
			}
			var err error
			device := crawler.Device{KObj: uevent.KObj, Env: uevent.Env}
			switch uevent.Action {
			case "add":
				err = addDevice(device)
			case "remove":
				err = removeDevice(device)
			default:
				err = fmt.Errorf("unhandlded action %q", uevent.Action)
			}
			if err != nil {
				log.Printf("error handling uevent %v: %s\n", uevent, err)
			}
		case err := <-errors:
			log.Printf("error watching devices: %s", err)
		case <-ctx.Done():
			quit <- struct{}{}
			return nil
		}
	}
}
