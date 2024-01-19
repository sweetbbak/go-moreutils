package main

import (
	"log"
	"mybox/pkg/xinit"
	"os"
	"strconv"
	"syscall"
	"time"
)

const (
	REBOOT_HALT = iota
	REBOOT_POWEROFF
	REBOOT_RESTART
)

func Exit(reboot int, services map[internal.ServiceType][]*internal.Service) {
	StopServices(services)
	KillAllProcs()
	UnmountAll()
	Reboot(reboot)
}

func GetAllProcesses() (pids []int) {
	dir, err := os.ReadDir("/proc")
	if err != nil {
		log.Panicln(err.Error())
		return
	}

	for i := len(dir) - 1; i != 0; i-- {
		if !dir[i].IsDir() {
			continue
		}

		pid, err := strconv.Atoi(dir[i].Name())
		if err != nil {
			continue
		}

		pids = append(pids, pid)
	}

	return
}

func KillAllProcs() {
	procs := GetAllProcesses()

	for i := 0; i < len(procs); i++ {
		if procs[i] == 1 {
			continue
		}

		proc, err := os.FindProcess(procs[i])
		if err != nil {
			continue
		}

		go func(pid int) {
			time.Sleep(time.Second * 5)
			if proc, err := os.FindProcess(pid); err == nil {
				proc.Kill()
				return
			} else {
				return
			}
		}(proc.Pid)
		proc.Signal(syscall.SIGTERM)
		proc.Wait()
	}
}

func StopServices(services map[internal.ServiceType][]*internal.Service) {
	for _, srvi := range services {
		for _, srv := range srvi {
			if err := srv.Stop(); err != nil {
				log.Printf("failed to stop service %s: %s", srv.Name, err)
			}
		}
	}
}

func UnmountAll() {
	mounts, err := parseMtab()
	if err != nil {
		log.Println(err)
	}

	for i := len(mounts) - 1; i != 0; i-- {
		if err := syscall.Unmount(mounts[i].target, 0); err != nil {
			log.Print("failed to unmount: %v: %v", mounts[i].target, err)
		}
	}
}

func Reboot(e int) {
	switch e {
	case REBOOT_HALT:
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_HALT)
	case REBOOT_POWEROFF:
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	case REBOOT_RESTART:
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	}
}
