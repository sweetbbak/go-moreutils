package main

/*
#cgo CFLAGS: -Wall
#define _GNU_SOURCE

#include <sched.h>
#include <stdlib.h>

__attribute((constructor(102))) void enter_netns(void) {
    if (unshare(CLONE_NEWNET) == -1) {
        exit(1);
    }
}
*/
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

func init() {
	if err := enableLoopback(); err != nil {
		panic(err.Error())
	}
}

func enableLoopback() error {
	return setIFFlags("lo", syscall.IFF_UP)
}

func setIFFlags(ifname string, flags uint16) error {
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_IP)
	if err != nil {
		return err
	}
	defer syscall.Close(sock)

	var ifr struct {
		Name  [syscall.IFNAMSIZ]byte
		Flags uint16
	}

	copy(ifr.Name[:], ifname)
	ifr.Flags = flags

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(sock),
		uintptr(syscall.SIOCSIFFLAGS),
		uintptr(unsafe.Pointer(&ifr)),
	)

	if errno != 0 {
		return fmt.Errorf("interface up failed: %s", errno)
	}

	return nil
}
