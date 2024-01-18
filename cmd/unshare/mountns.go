// Package mountns makes the process enter a new mount namespace.
package main

/*
#cgo CFLAGS: -Wall
#define _GNU_SOURCE

#include <sched.h>
#include <stdlib.h>

__attribute((constructor(103))) void enter_netns(void) {
    if (unshare(CLONE_NEWNS) == -1) {
        exit(1);
    }
}
*/
import "C"

import (
	"syscall"
)

func BindMount(src, dst string) error {
	// Mount source to target using syscall.Mount
	return syscall.Mount(src, dst, "", syscall.MS_BIND, "")
}
