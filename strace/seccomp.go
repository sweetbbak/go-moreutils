package main

import (
	"syscall"

	sec "github.com/seccomp/libseccomp-golang"
)

func disallow(sc string) error {
	id, err := sec.GetSyscallFromName(sc)
	if err != nil {
		return err
	}

	filter, _ := sec.NewFilter(sec.ActAllow)
	filter.AddRule(id, sec.ActErrno.SetReturnCode(int16(syscall.EPERM)))
	filter.Load()
	return nil
}
