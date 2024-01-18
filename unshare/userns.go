// Package userns makes the process enter a new network namespace.
package main

/*
#cgo CFLAGS: -Wall
#define _GNU_SOURCE
#include <stdlib.h>
#include <unistd.h>
#include <sched.h>

 int originalUid = 0;

__attribute((constructor(101))) void enter_userns(void) {
	originalUid = getuid();
    if (unshare(CLONE_NEWUSER) == -1) {
        exit(1);
    }
}

*/
import "C"

import (
	"fmt"
	"os"
	"strings"

	"slices"
)

func init() {
	err := WriteUsermap(map[uint32]uint32{
		OriginalUID(): 0,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create user mapping: %v", err).Error())
	}
}

func OriginalUID() uint32 {
	return uint32(C.originalUid)
}

// WriteUsermap builds a map of Host UID -> Namespace UID.
// Example:
//
//	WriteUsermap(map[uint32]uint32{userns.OriginalUID: 0, 1234: 1234})
func WriteUsermap(mapping map[uint32]uint32) error {
	lines := []string{}
	for h, c := range mapping {
		lines = append(lines, fmt.Sprintf("%d %d 1", c, h))
	}
	slices.Sort(lines)
	return os.WriteFile("/proc/self/uid_map", []byte(strings.Join(lines, "\n")), 0o644)
}
