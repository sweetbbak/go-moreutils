package main

import (
	"errors"
	"flag"
	"log"
	"syscall"
	"unsafe"
)

var (
	force = flag.Bool("f", false, "Force unmount")
	lazy  = flag.Bool("l", false, "Lazy unmount")
)

// pulled mount flags from unix package
const (
	UMOUNT_NOFOLLOW = 0x8
	MNT_FORCE       = 0x1
	MNT_DETACH      = 0x2
	ENOENT          = syscall.Errno(0x2)
	EAGAIN          = syscall.Errno(0xb)
	EINVAL          = syscall.Errno(0x16)
)

var (
	errEAGAIN error = syscall.EAGAIN
	errEINVAL error = syscall.EINVAL
	errENOENT error = syscall.ENOENT
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

func Unmount(target string, flags int) (err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(target)
	if err != nil {
		return
	}
	_, _, e1 := syscall.Syscall(syscall.SYS_UMOUNT2, uintptr(unsafe.Pointer(_p0)), uintptr(flags), 0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

func umountDisk(path string, force, lazy bool) error {
	flg := UMOUNT_NOFOLLOW
	if len(path) == 0 {
		return errors.New("path cannot be empty")
	}
	if force {
		flg |= MNT_FORCE
	}
	if lazy {
		flg |= MNT_DETACH
	}
	if err := Unmount(path, flg); err != nil {
		return err
	}
	return nil
}

func umount(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: umount [-f|-l] /path/to/disk")
	}

	for _, disk := range args {
		err := umountDisk(disk, *force, *lazy)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func main() {
	flag.Parse()
	args := flag.Args()

	if err := umount(args); err != nil {
		log.Fatal(err)
	}
}
