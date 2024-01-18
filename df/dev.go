package main

import (
	"golang.org/x/sys/unix"
)

func deviceNumber(path string) (uint64, error) {
	st := &unix.Stat_t{}
	err := unix.Stat(path, st)
	if err != nil {
		return 0, err
	}
	return st.Dev, nil
}
