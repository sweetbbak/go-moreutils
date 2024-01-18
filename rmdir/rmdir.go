package main

import (
	"errors"
	"log"
	"os"
	"syscall"
	"unsafe"
)

const (
	AT_FDCWD     = -0x64
	AT_REMOVEDIR = 0x200
)

func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case syscall.EAGAIN:
		// no data available try AGAIN
		return syscall.EAGAIN
	case syscall.EINVAL:
		// invalid argument
		return syscall.EINVAL
	case syscall.ENOENT:
		// invalid entri NO ENTRY
		return syscall.ENOENT
	}
	return e
}

func Unlinkat(dirfd int, path string, flags int) (err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return
	}
	_, _, e1 := syscall.Syscall(syscall.SYS_UNLINKAT, uintptr(dirfd), uintptr(unsafe.Pointer(_p0)), uintptr(flags))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

func rmdir(args []string) error {
	var errcount int
	for _, path := range args {
		err := Unlinkat(AT_FDCWD, path, AT_REMOVEDIR)
		if err != nil {
			log.Printf("error: cant remove directory: %v", err)
			errcount += 1
		}
	}
	if errcount == len(args) {
		return errors.New("Unable to remove any directories")
	}
	return nil
}

func main() {
	args := os.Args[1:]
	if err := rmdir(args); err != nil {
		log.Fatal(err)
	}
}
